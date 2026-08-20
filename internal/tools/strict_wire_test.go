package tools

import (
	"testing"

	"github.com/voocel/agentcore"
	"github.com/voocel/ainovel-cli/internal/llmcontract"
	"github.com/voocel/ainovel-cli/internal/store"
)

// TestStrictToolSchemasAreWireReady 断言所有声明 StrictSchema()=true 的工具，
// 其 Schema() 已是可直接发送的 strict 形态（每个 object 显式带
// additionalProperties:false）。工具 schema 由 agentcore/litellm 原样透传给
// Anthropic，缺该键会以 HTTP 400 拒绝整个请求（tools.N.custom: For 'object'
// type, 'additionalProperties' must be explicitly set to false）。
// 新增 strict 工具时必须加入此列表，并用 strictObject 构建 Schema()。
func TestStrictToolSchemasAreWireReady(t *testing.T) {
	st := store.NewStore(t.TempDir())
	strictTools := []agentcore.Tool{
		NewSaveBookTool(st),
		NewReviseOutlineTool(st),
		NewResolveOutlineFeedbackTool(st),
		NewAuditFoundationTool(st),
		NewDraftChapterTool(st),
		NewCommitChapterTool(st, NewStyleStatsIndex(st)),
		NewSaveReviewTool(st),
	}
	for _, tool := range strictTools {
		s, ok := tool.(agentcore.StrictSchemaTool)
		if !ok || !s.StrictSchema() {
			t.Errorf("%s 未声明 StrictSchema()=true，与本测试的清单不一致", tool.Name())
			continue
		}
		if err := llmcontract.ValidateStrictWire(tool.Schema()); err != nil {
			t.Errorf("%s schema 不满足 strict 上线形态: %v", tool.Name(), err)
		}
	}
}
