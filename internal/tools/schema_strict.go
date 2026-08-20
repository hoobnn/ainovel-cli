package tools

import (
	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/llmcontract"
)

// strictObject 构建 strict 工具的顶层 schema：在 schema.Object 的基础上递归为
// 所有 object 节点显式补 additionalProperties:false。声明 StrictSchema()=true
// 的工具必须用它（而不是裸 schema.Object）构建 Schema() 返回值——工具 schema
// 由 agentcore/litellm 原样透传，Anthropic strict 校验缺该键会以 HTTP 400
// 拒绝整个请求，且不做服务端补全。
func strictObject(props ...schema.Prop) map[string]any {
	return llmcontract.StrictObjects(schema.Object(props...))
}
