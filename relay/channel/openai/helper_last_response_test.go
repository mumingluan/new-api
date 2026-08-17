package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
)

func TestHandleLastResponse_SendsToolCallChunkWithoutText(t *testing.T) {
	t.Parallel()

	info := &relaycommon.RelayInfo{ShouldIncludeUsage: false}

	last := dto.ChatCompletionsStreamResponse{
		Id:      "chatcmpl_test",
		Object:  "chat.completion.chunk",
		Created: 123,
		Model:   "gpt-test",
		Choices: []dto.ChatCompletionsStreamResponseChoice{
			{
				Index: 0,
				Delta: dto.ChatCompletionsStreamResponseChoiceDelta{
					ToolCalls: []dto.ToolCallResponse{
						{
							ID:   "call_1",
							Type: "function",
							Function: dto.FunctionResponse{
								Name:      "get_weather",
								Arguments: "{}",
							},
						},
					},
				},
			},
		},
		Usage: &dto.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
	}

	data, err := common.Marshal(last)
	if err != nil {
		t.Fatalf("marshal last chunk: %v", err)
	}

	var (
		responseId         string
		createAt           int64
		systemFingerprint  string
		model              string
		usage              = &dto.Usage{}
		containStreamUsage bool
		shouldSendLastResp = true
	)

	if err := handleLastResponse(string(data), &responseId, &createAt, &systemFingerprint, &model, &usage, &containStreamUsage, info, &shouldSendLastResp); err != nil {
		t.Fatalf("handleLastResponse error: %v", err)
	}

	if !containStreamUsage {
		t.Fatalf("expected containStreamUsage=true")
	}
	if !shouldSendLastResp {
		t.Fatalf("expected shouldSendLastResp=true when tool_calls exist")
	}
}
