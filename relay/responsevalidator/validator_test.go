package responsevalidator

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/stretchr/testify/require"
)

func decode[T any](t *testing.T, payload string) *T {
	t.Helper()
	var value T
	require.NoError(t, common.Unmarshal([]byte(payload), &value))
	return &value
}

func TestOpenAIChatResponseSemanticOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		valid   bool
		wantErr bool
	}{
		{"empty choices", `{"choices":[]}`, false, false},
		{"empty assistant shell", `{"choices":[{"message":{"role":"assistant"},"finish_reason":"stop"}]}`, false, false},
		{"text", `{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`, true, false},
		{"completion text", `{"choices":[{"text":"ok","finish_reason":"stop"}]}`, true, false},
		{"refusal", `{"choices":[{"message":{"refusal":"no"},"finish_reason":"stop"}]}`, true, false},
		{"filtered", `{"choices":[{"message":{},"finish_reason":"content_filter"}]}`, true, false},
		{"tool only", `{"choices":[{"message":{"tool_calls":[{"type":"function","function":{"name":"lookup","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`, true, false},
		{"tool missing name", `{"choices":[{"message":{"tool_calls":[{"type":"function","function":{"arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`, false, true},
		{"tool malformed arguments", `{"choices":[{"message":{"tool_calls":[{"type":"function","function":{"name":"lookup","arguments":"{"}}]},"finish_reason":"tool_calls"}]}`, false, true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := OpenAIChatResponse(decode[dto.OpenAITextResponse](t, test.payload))
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.valid, result.Valid())
		})
	}
}

func TestOpenAIStreamStateToolOnlyAndTerminalRules(t *testing.T) {
	t.Parallel()

	state := NewOpenAIStreamState()
	require.NoError(t, state.Observe(decode[dto.ChatCompletionsStreamResponse](t,
		`{"choices":[{"index":0,"delta":{"role":"assistant"}}]}`)))
	require.Error(t, state.Validate(true))

	state = NewOpenAIStreamState()
	require.NoError(t, state.Observe(decode[dto.ChatCompletionsStreamResponse](t,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"name":"lookup","arguments":"{"}}]}}]}`)))
	require.NoError(t, state.Observe(decode[dto.ChatCompletionsStreamResponse](t,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"}"}}]},"finish_reason":"tool_calls"}]}`)))
	require.NoError(t, state.Validate(false))
	require.Equal(t, 1, state.ToolCount())

	incomplete := NewOpenAIStreamState()
	require.NoError(t, incomplete.Observe(decode[dto.ChatCompletionsStreamResponse](t,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"name":"lookup","arguments":"{"}}]},"finish_reason":"tool_calls"}]}`)))
	require.Error(t, incomplete.Validate(false))

	unterminated := NewOpenAIStreamState()
	require.NoError(t, unterminated.Observe(decode[dto.ChatCompletionsStreamResponse](t,
		`{"choices":[{"index":0,"delta":{"content":"partial"}}]}`)))
	require.Error(t, unterminated.Validate(false))
	require.NoError(t, unterminated.Validate(true))
}

func TestClaudeSemanticOutput(t *testing.T) {
	t.Parallel()

	result, err := ClaudeResponse(decode[dto.ClaudeResponse](t,
		`{"type":"message","stop_reason":"tool_use","content":[]}`))
	require.NoError(t, err)
	require.False(t, result.Valid())

	result, err = ClaudeResponse(decode[dto.ClaudeResponse](t,
		`{"type":"message","stop_reason":"tool_use","content":[{"type":"tool_use","name":"lookup","input":{}}]}`))
	require.NoError(t, err)
	require.True(t, result.Valid())

	state := NewClaudeStreamState()
	require.NoError(t, state.Observe(decode[dto.ClaudeResponse](t, `{"type":"message_start"}`)))
	require.NoError(t, state.Observe(decode[dto.ClaudeResponse](t, `{"type":"message_stop"}`)))
	require.Error(t, state.Validate())

	state = NewClaudeStreamState()
	require.NoError(t, state.Observe(decode[dto.ClaudeResponse](t,
		`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","name":"lookup"}}`)))
	require.NoError(t, state.Observe(decode[dto.ClaudeResponse](t,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{}"}}`)))
	require.NoError(t, state.Observe(decode[dto.ClaudeResponse](t, `{"type":"message_stop"}`)))
	require.NoError(t, state.Validate())
}

func TestGeminiSemanticOutput(t *testing.T) {
	t.Parallel()

	result, err := GeminiResponse(decode[dto.GeminiChatResponse](t,
		`{"candidates":[{"content":{"role":"model","parts":[]},"finishReason":"STOP"}]}`))
	require.NoError(t, err)
	require.False(t, result.Valid())

	result, err = GeminiResponse(decode[dto.GeminiChatResponse](t,
		`{"candidates":[{"content":{"parts":[{"functionCall":{"name":"lookup","args":{}}}]},"finishReason":"STOP"}]}`))
	require.NoError(t, err)
	require.True(t, result.Valid())

	_, err = GeminiResponse(decode[dto.GeminiChatResponse](t,
		`{"candidates":[{"content":{"parts":[{"functionCall":{"args":{}}}]},"finishReason":"STOP"}]}`))
	require.Error(t, err)

	result, err = GeminiResponse(decode[dto.GeminiChatResponse](t,
		`{"candidates":[{"content":{"parts":[]},"finishReason":"SAFETY"}]}`))
	require.NoError(t, err)
	require.True(t, result.Valid())
}

func TestResponsesSemanticOutput(t *testing.T) {
	t.Parallel()

	result, err := ResponsesResponse(decode[dto.OpenAIResponsesResponse](t,
		`{"status":"completed","output":[]}`))
	require.NoError(t, err)
	require.False(t, result.Valid())

	result, err = ResponsesResponse(decode[dto.OpenAIResponsesResponse](t,
		`{"status":"completed","output":[{"type":"function_call","name":"lookup","arguments":"{}"}]}`))
	require.NoError(t, err)
	require.True(t, result.Valid())

	_, err = ResponsesResponse(decode[dto.OpenAIResponsesResponse](t,
		`{"status":"failed","output":[]}`))
	require.Error(t, err)

	state := NewResponsesStreamState()
	require.NoError(t, state.Observe(decode[dto.ResponsesStreamResponse](t,
		`{"type":"response.created"}`)))
	require.NoError(t, state.Observe(decode[dto.ResponsesStreamResponse](t,
		`{"type":"response.completed","response":{"status":"completed","output":[]}}`)))
	require.Error(t, state.Validate())

	state = NewResponsesStreamState()
	require.NoError(t, state.Observe(decode[dto.ResponsesStreamResponse](t,
		`{"type":"response.output_item.added","output_index":0,"item":{"id":"call_1","type":"function_call","name":"lookup"}}`)))
	require.NoError(t, state.Observe(decode[dto.ResponsesStreamResponse](t,
		`{"type":"response.function_call_arguments.delta","item_id":"call_1","delta":"{}"}`)))
	require.NoError(t, state.Observe(decode[dto.ResponsesStreamResponse](t,
		`{"type":"response.completed","response":{"status":"completed","output":[]}}`)))
	require.NoError(t, state.Validate())
}
