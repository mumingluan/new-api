package geminichat

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamResponseGeminiChatToOpenAIPreservesFunctionCallID(t *testing.T) {
	response, _ := StreamResponseGeminiChat2OpenAI(&dto.GeminiChatResponse{
		Candidates: []dto.GeminiChatCandidate{{
			Index: 0,
			Content: dto.GeminiChatContent{Parts: []dto.GeminiPart{{
				FunctionCall: &dto.FunctionCall{
					ID:           "call_stable",
					FunctionName: "lookup",
					Arguments:    map[string]interface{}{"q": "x"},
				},
			}}},
		}},
	})

	require.Len(t, response.Choices, 1)
	require.Len(t, response.Choices[0].Delta.ToolCalls, 1)
	call := response.Choices[0].Delta.ToolCalls[0]
	assert.Equal(t, "call_stable", call.ID)
	assert.Equal(t, "lookup", call.Function.Name)
	assert.JSONEq(t, `{"q":"x"}`, call.Function.Arguments)
}
