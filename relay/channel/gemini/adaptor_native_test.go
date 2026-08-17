package gemini

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestConvertGeminiRequestCleansNativeAndBatchToolSchemas(t *testing.T) {
	tools := `[
		{
			"functionDeclarations": [
				{
					"name": "lookup",
					"parameters": {
						"type": "object",
						"properties": {
							"query": {"type": "string", "const": "fixed"},
							"count": {"type": "integer", "exclusiveMinimum": 0}
						},
						"patternProperties": {"^x-": {"type": "string"}}
					}
				}
			]
		}
	]`
	request := &dto.GeminiChatRequest{
		Tools: []byte(tools),
		Requests: []dto.GeminiChatRequest{
			{Tools: []byte(tools)},
		},
	}

	converted, err := (&Adaptor{}).ConvertGeminiRequest(nil, &relaycommon.RelayInfo{}, request)
	require.NoError(t, err)
	require.Same(t, request, converted)

	for _, rawTools := range [][]byte{request.Tools, request.Requests[0].Tools} {
		assert.False(t, gjson.GetBytes(rawTools, "0.functionDeclarations.0.parameters.patternProperties").Exists())
		assert.False(t, gjson.GetBytes(rawTools, "0.functionDeclarations.0.parameters.properties.query.const").Exists())
		assert.Equal(t, "fixed", gjson.GetBytes(rawTools, "0.functionDeclarations.0.parameters.properties.query.enum.0").String())
		assert.False(t, gjson.GetBytes(rawTools, "0.functionDeclarations.0.parameters.properties.count.exclusiveMinimum").Exists())
		assert.False(t, gjson.GetBytes(rawTools, "0.functionDeclarations.0.parameters.properties.count.minimum").Exists())
	}
}
