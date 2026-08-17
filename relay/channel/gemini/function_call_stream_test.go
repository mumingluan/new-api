package gemini

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeminiFunctionCallStreamStateJoinsSplitNameAndArguments(t *testing.T) {
	state := geminiFunctionCallStreamState{}
	nameChunk := &dto.GeminiChatResponse{
		Candidates: []dto.GeminiChatCandidate{{
			Index: 0,
			Content: dto.GeminiChatContent{Parts: []dto.GeminiPart{{
				FunctionCall: &dto.FunctionCall{
					FunctionName: "lookup",
					Arguments:    map[string]interface{}{},
				},
			}}},
		}},
	}
	assert.True(t, state.normalize(nameChunk))

	firstCall := nameChunk.Candidates[0].Content.Parts[0].FunctionCall
	require.NotEmpty(t, firstCall.ID)
	assert.Equal(t, "lookup", firstCall.FunctionName)

	argumentsChunk := &dto.GeminiChatResponse{
		Candidates: []dto.GeminiChatCandidate{{
			Index: 0,
			Content: dto.GeminiChatContent{Parts: []dto.GeminiPart{{
				FunctionCall: &dto.FunctionCall{
					Arguments: map[string]interface{}{"q": "x"},
				},
			}}},
		}},
	}
	assert.False(t, state.normalize(argumentsChunk))

	secondCall := argumentsChunk.Candidates[0].Content.Parts[0].FunctionCall
	assert.Equal(t, firstCall.ID, secondCall.ID)
	assert.Equal(t, "lookup", secondCall.FunctionName)
	assert.Equal(t, map[string]interface{}{"q": "x"}, secondCall.Arguments)
}
