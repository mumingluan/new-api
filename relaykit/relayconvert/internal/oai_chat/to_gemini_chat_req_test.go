package oaichat

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIChatRequestToGeminiFlattensClaudeToolHistory(t *testing.T) {
	request := dto.GeneralOpenAIRequest{
		Model: "[An]claude-opus-4-6",
		Messages: []dto.Message{
			{
				Role: "assistant",
				ToolCalls: []byte(`[
					{"id":"call_0","type":"function","function":{"name":"lookup","arguments":"{\"q\":\"x\"}"}}
				]`),
			},
			{Role: "tool", ToolCallId: "call_0", Content: `{"answer":"ok"}`},
		},
	}

	got, err := OpenAIChatRequestToGeminiGenerateContent(nil, request, nil)
	require.NoError(t, err)
	require.Len(t, got.Contents, 2)
	require.Len(t, got.Contents[0].Parts, 1)
	require.Len(t, got.Contents[1].Parts, 1)

	assert.Nil(t, got.Contents[0].Parts[0].FunctionCall)
	assert.Equal(t, `Tool call call_0 (lookup): {"q":"x"}`, got.Contents[0].Parts[0].Text)
	assert.Nil(t, got.Contents[1].Parts[0].FunctionResponse)
	assert.Equal(t, `Tool result for call_0: {"answer":"ok"}`, got.Contents[1].Parts[0].Text)
}
