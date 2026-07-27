package gemini

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanToolsSanitizesNativeFunctionSchemasWithoutDroppingOtherTools(t *testing.T) {
	input := []byte(`[
		{
			"functionDeclarations": [
				{
					"name": "lookup",
					"behavior": "BLOCKING",
					"parameters": {
						"type": "object",
						"patternProperties": {"^x-": {"type": "string"}},
						"properties": {
							"kind": {"type": "string", "const": "city"},
							"score": {"type": "number", "exclusiveMinimum": 0},
							"legacy": {"type": "number", "exclusiveMinimum": false},
							"nested": {
								"anyOf": [
									{"type": "object", "properties": {"value": {"type": "string", "const": "ok"}}}
								]
							}
						}
					},
					"response": {
						"type": "object",
						"properties": {"result": {"type": "string", "const": "done"}}
					}
				}
			],
			"futureToolField": {"enabled": true}
		},
		{"googleSearch": {}}
	]`)

	output, err := CleanTools(input)
	require.NoError(t, err)

	var got []map[string]interface{}
	require.NoError(t, common.Unmarshal(output, &got))
	require.Len(t, got, 2)
	assert.Equal(t, map[string]interface{}{"enabled": true}, got[0]["futureToolField"])
	assert.Equal(t, map[string]interface{}{}, got[1]["googleSearch"])

	declarations := got[0]["functionDeclarations"].([]interface{})
	declaration := declarations[0].(map[string]interface{})
	assert.Equal(t, "BLOCKING", declaration["behavior"])

	parameters := declaration["parameters"].(map[string]interface{})
	assert.NotContains(t, parameters, "patternProperties")
	assert.Equal(t, "OBJECT", parameters["type"])
	properties := parameters["properties"].(map[string]interface{})
	assert.Equal(t, "STRING", properties["kind"].(map[string]interface{})["type"])
	assert.Equal(t, []interface{}{"city"}, properties["kind"].(map[string]interface{})["enum"])
	assert.NotContains(t, properties["score"].(map[string]interface{}), "minimum")
	assert.NotContains(t, properties["legacy"].(map[string]interface{}), "minimum")
	nestedAnyOf := properties["nested"].(map[string]interface{})["anyOf"].([]interface{})
	nestedProperties := nestedAnyOf[0].(map[string]interface{})["properties"].(map[string]interface{})
	assert.Equal(t, []interface{}{"ok"}, nestedProperties["value"].(map[string]interface{})["enum"])

	response := declaration["response"].(map[string]interface{})
	responseProperties := response["properties"].(map[string]interface{})
	assert.Equal(t, []interface{}{"done"}, responseProperties["result"].(map[string]interface{})["enum"])
}

func TestCleanToolsSupportsSnakeCaseFunctionDeclarations(t *testing.T) {
	input := []byte(`{
		"function_declarations": [
			{
				"name": "lookup",
				"parameters": {
					"type": "object",
					"properties": {"query": {"type": "string", "const": "fixed"}}
				}
			}
		]
	}`)

	output, err := CleanTools(input)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, common.Unmarshal(output, &got))
	declarations := got["function_declarations"].([]interface{})
	parameters := declarations[0].(map[string]interface{})["parameters"].(map[string]interface{})
	properties := parameters["properties"].(map[string]interface{})
	assert.Equal(t, []interface{}{"fixed"}, properties["query"].(map[string]interface{})["enum"])
}

func TestCleanClaudeCompatibleFunctionParametersUsesConservativeDraftSchema(t *testing.T) {
	input := map[string]interface{}{
		"type":             "object",
		"nullable":         true,
		"propertyOrdering": []interface{}{"mode", "payload"},
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"anyOf": []interface{}{
					map[string]interface{}{"type": "null"},
					map[string]interface{}{"type": "string", "const": "run"},
					map[string]interface{}{"type": "integer"},
				},
			},
			"payload": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"value": map[string]interface{}{
						"type":             "number",
						"exclusiveMinimum": 0,
						"patternProperties": map[string]interface{}{
							".*": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
		"required": []interface{}{"mode", 7, ""},
	}

	got := CleanClaudeCompatibleFunctionParameters(input)
	assert.Equal(t, map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"run"},
			},
			"payload": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"value": map[string]interface{}{"type": "number"},
				},
			},
		},
		"required": []string{"mode"},
	}, got)
}

func TestCleanClaudeCompatibleFunctionParametersDefaultsRootToObject(t *testing.T) {
	assert.Equal(t, map[string]interface{}{"type": "object"}, CleanClaudeCompatibleFunctionParameters(nil))
	assert.Equal(t, map[string]interface{}{"type": "object"}, CleanClaudeCompatibleFunctionParameters(map[string]interface{}{}))
}
