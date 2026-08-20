package llmcontract

import (
	"maps"
	"testing"

	"github.com/voocel/agentcore/schema"
)

func TestStrictObjectsInjectsRecursively(t *testing.T) {
	nested := schema.Object(
		schema.Property("id", schema.String("ID")).Required(),
	)
	original := schema.Object(
		schema.Property("name", schema.String("名称")).Required(),
		schema.Property("items", schema.Array("列表", nested)).Required(),
		schema.Property("feedback", Nullable(schema.Object(
			schema.Property("note", schema.String("说明")).Required(),
		))).Required(),
	)

	out := StrictObjects(original)

	if out["additionalProperties"] != false {
		t.Fatal("顶层 object 未注入 additionalProperties:false")
	}
	props := out["properties"].(map[string]any)
	itemSchema := props["items"].(map[string]any)["items"].(map[string]any)
	if itemSchema["additionalProperties"] != false {
		t.Fatal("数组元素 object 未注入 additionalProperties:false")
	}
	feedback := props["feedback"].(map[string]any)
	if feedback["additionalProperties"] != false {
		t.Fatal("可空联合 object（type 含 object 与 null）未注入 additionalProperties:false")
	}
	if _, ok := original["additionalProperties"]; ok {
		t.Fatal("StrictObjects 不应修改传入 schema")
	}
	if err := ValidateStrictWire(out); err != nil {
		t.Fatalf("归一化结果应通过 wire 校验: %v", err)
	}
	if err := ValidateStrictWire(original); err == nil {
		t.Fatal("未归一化的 schema 不应通过 wire 校验")
	}
}

func TestStrictObjectsRewritesNullableEnumAsAnyOf(t *testing.T) {
	original := schema.Object(
		schema.Property("strand", Nullable(schema.Enum("主导叙事线", "quest", "fire"))).Required(),
	)

	out := StrictObjects(original)

	strand := out["properties"].(map[string]any)["strand"].(map[string]any)
	if _, hasType := strand["type"]; hasType {
		t.Fatal("改写后不应保留联合 type")
	}
	if _, hasEnum := strand["enum"]; hasEnum {
		t.Fatal("改写后顶层不应保留 enum")
	}
	branches, ok := strand["anyOf"].([]any)
	if !ok || len(branches) != 2 {
		t.Fatalf("应改写为两分支 anyOf，实际 %v", strand["anyOf"])
	}
	first := branches[0].(map[string]any)
	if first["type"] != "string" {
		t.Fatalf("枚举分支 type 应为 string，实际 %v", first["type"])
	}
	if enum, ok := first["enum"].([]any); !ok || len(enum) != 2 || enum[0] != "quest" || enum[1] != "fire" {
		t.Fatalf("枚举分支应只含非 null 值，实际 %v", first["enum"])
	}
	if second := branches[1].(map[string]any); second["type"] != "null" {
		t.Fatalf("第二分支应为 null 类型，实际 %v", second["type"])
	}
	if strand["description"] != "主导叙事线" {
		t.Fatal("description 应保留在节点顶层")
	}
	if err := ValidateStrictWire(out); err != nil {
		t.Fatalf("归一化结果应通过 wire 校验: %v", err)
	}
	if err := ValidateStrictWire(StrictObjects(original)); err != nil {
		t.Fatalf("二次归一化应幂等且通过 wire 校验: %v", err)
	}
	withAdditional := maps.Clone(original)
	withAdditional["additionalProperties"] = false
	if err := ValidateStrictWire(withAdditional); err == nil {
		t.Fatal("未改写的联合 type + enum 不应通过 wire 校验")
	}
	if _, hasType := original["properties"].(map[string]any)["strand"].(map[string]any)["type"]; !hasType {
		t.Fatal("StrictObjects 不应修改传入 schema")
	}
}

func TestStrictObjectsKeepsExplicitAdditionalProperties(t *testing.T) {
	s := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{"extra": map[string]any{"type": "string"}},
		"required":             []string{"extra"},
		"additionalProperties": map[string]any{"type": "string"},
	}
	out := StrictObjects(s)
	if _, isMap := out["additionalProperties"].(map[string]any); !isMap {
		t.Fatal("已显式声明的 additionalProperties 应保持原值")
	}
}
