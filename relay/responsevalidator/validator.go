package responsevalidator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
)

type Result struct {
	Output   bool
	Filtered bool
	Terminal bool
}

func (r Result) Valid() bool {
	return r.Output || r.Filtered
}

func rawPresent(raw json.RawMessage) bool {
	value := strings.TrimSpace(string(raw))
	return value != "" && value != "null" && value != "{}" && value != "[]"
}

func validArguments(arguments string) bool {
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		return true
	}
	var value any
	return common.Unmarshal([]byte(arguments), &value) == nil
}

func ValidateToolCalls(calls []dto.ToolCallResponse) error {
	for i, call := range calls {
		if strings.TrimSpace(call.Function.Name) == "" {
			return fmt.Errorf("tool call %d has no function name", i)
		}
		if !validArguments(call.Function.Arguments) {
			return fmt.Errorf("tool call %d has invalid JSON arguments", i)
		}
	}
	return nil
}

func parseToolCalls(raw json.RawMessage) ([]dto.ToolCallResponse, error) {
	if !rawPresent(raw) {
		return nil, nil
	}
	var calls []dto.ToolCallResponse
	if err := common.Unmarshal(raw, &calls); err != nil {
		return nil, err
	}
	return calls, ValidateToolCalls(calls)
}

func validLegacyFunctionCall(raw json.RawMessage) (bool, error) {
	if !rawPresent(raw) {
		return false, nil
	}
	var call struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	if err := common.Unmarshal(raw, &call); err != nil {
		return false, err
	}
	if strings.TrimSpace(call.Name) == "" {
		return false, fmt.Errorf("legacy function call has no function name")
	}
	if !validArguments(call.Arguments) {
		return false, fmt.Errorf("legacy function call has invalid JSON arguments")
	}
	return true, nil
}

func OpenAIChatResponse(response *dto.OpenAITextResponse) (Result, error) {
	var result Result
	if response == nil {
		return result, fmt.Errorf("response is nil")
	}
	for _, choice := range response.Choices {
		if choice.Text != "" || choice.Message.StringContent() != "" ||
			choice.Message.GetReasoningContent() != "" ||
			(choice.Message.Refusal != nil && *choice.Message.Refusal != "") ||
			rawPresent(choice.Message.Audio) {
			result.Output = true
		}
		if strings.EqualFold(choice.FinishReason, "content_filter") {
			result.Filtered = true
		}
		if choice.FinishReason != "" {
			result.Terminal = true
		}
		calls, err := parseToolCalls(choice.Message.ToolCalls)
		if err != nil {
			return result, err
		}
		if len(calls) > 0 {
			result.Output = true
		}
		legacy, err := validLegacyFunctionCall(choice.Message.FunctionCall)
		if err != nil {
			return result, err
		}
		result.Output = result.Output || legacy
	}
	return result, nil
}

type streamedToolCall struct {
	name      string
	arguments strings.Builder
}

type OpenAIStreamState struct {
	Result
	tools map[string]*streamedToolCall
}

func NewOpenAIStreamState() *OpenAIStreamState {
	return &OpenAIStreamState{tools: make(map[string]*streamedToolCall)}
}

func (s *OpenAIStreamState) Observe(response *dto.ChatCompletionsStreamResponse) error {
	if s == nil || response == nil {
		return fmt.Errorf("response is nil")
	}
	for _, choice := range response.Choices {
		delta := choice.Delta
		if delta.GetContentString() != "" || delta.GetReasoningContent() != "" ||
			(delta.Refusal != nil && *delta.Refusal != "") || rawPresent(delta.Audio) {
			s.Output = true
		}
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			s.Terminal = true
			if strings.EqualFold(*choice.FinishReason, "content_filter") {
				s.Filtered = true
			}
		}
		if legacy, err := validLegacyFunctionCall(delta.FunctionCall); err != nil {
			return err
		} else if legacy {
			s.Output = true
		}
		for position, call := range delta.ToolCalls {
			index := position
			if call.Index != nil {
				index = *call.Index
			}
			key := fmt.Sprintf("%d:%d", choice.Index, index)
			tool := s.tools[key]
			if tool == nil {
				tool = &streamedToolCall{}
				s.tools[key] = tool
			}
			if call.Function.Name != "" {
				tool.name = call.Function.Name
			}
			tool.arguments.WriteString(call.Function.Arguments)
		}
	}
	return nil
}

func (s *OpenAIStreamState) ObserveCompletion(response *dto.CompletionsStreamResponse) error {
	if s == nil || response == nil {
		return fmt.Errorf("response is nil")
	}
	for _, choice := range response.Choices {
		if choice.Text != "" {
			s.Output = true
		}
		if choice.FinishReason != "" {
			s.Terminal = true
			if strings.EqualFold(choice.FinishReason, "content_filter") {
				s.Filtered = true
			}
		}
	}
	return nil
}

func (s *OpenAIStreamState) Validate(done bool) error {
	if err := s.ValidateOutput(); err != nil {
		return err
	}
	if !s.Terminal && !done {
		return fmt.Errorf("stream ended before a terminal event")
	}
	return nil
}

// ValidateOutput verifies that a stream contains billable semantic output.
// Terminal framing is intentionally checked by Validate so relay handlers can
// accept and bill a non-empty stream that ended abnormally while still
// recording the abnormal StreamStatus.
func (s *OpenAIStreamState) ValidateOutput() error {
	if s == nil {
		return fmt.Errorf("stream state is nil")
	}
	for key, tool := range s.tools {
		if strings.TrimSpace(tool.name) == "" {
			return fmt.Errorf("tool call %s has no function name", key)
		}
		if !validArguments(tool.arguments.String()) {
			return fmt.Errorf("tool call %s has incomplete JSON arguments", key)
		}
		s.Output = true
	}
	if !s.Valid() {
		return fmt.Errorf("stream contained no semantic output")
	}
	return nil
}

func (s *OpenAIStreamState) ToolCount() int {
	if s == nil {
		return 0
	}
	return len(s.tools)
}

func ClaudeResponse(response *dto.ClaudeResponse) (Result, error) {
	var result Result
	if response == nil {
		return result, fmt.Errorf("response is nil")
	}
	if response.Completion != "" {
		result.Output = true
	}
	if response.StopReason != "" {
		result.Terminal = true
		if strings.Contains(strings.ToLower(response.StopReason), "refusal") {
			result.Filtered = true
		}
	}
	for i, block := range response.Content {
		switch block.Type {
		case "text":
			result.Output = result.Output || block.GetText() != ""
		case "thinking", "redacted_thinking":
			result.Output = true
		case "tool_use":
			if strings.TrimSpace(block.Name) == "" {
				return result, fmt.Errorf("tool_use block %d has no name", i)
			}
			result.Output = true
		case "server_tool_use", "web_search_tool_result":
			result.Output = true
		default:
			if block.Type != "" {
				result.Output = true
			}
		}
	}
	return result, nil
}

type ClaudeStreamState struct {
	Result
	tools map[int]*streamedToolCall
}

func NewClaudeStreamState() *ClaudeStreamState {
	return &ClaudeStreamState{tools: make(map[int]*streamedToolCall)}
}

func (s *ClaudeStreamState) Observe(response *dto.ClaudeResponse) error {
	if s == nil || response == nil {
		return fmt.Errorf("response is nil")
	}
	switch response.Type {
	case "content_block_start":
		if response.ContentBlock == nil {
			return fmt.Errorf("content_block_start has no content block")
		}
		block := response.ContentBlock
		switch block.Type {
		case "tool_use", "server_tool_use":
			if strings.TrimSpace(block.Name) == "" {
				return fmt.Errorf("%s block has no name", block.Type)
			}
			index := response.GetIndex()
			s.tools[index] = &streamedToolCall{name: block.Name}
			s.Output = true
		case "text":
			s.Output = s.Output || block.GetText() != ""
		case "thinking", "redacted_thinking", "web_search_tool_result":
			s.Output = true
		}
	case "content_block_delta":
		if response.Delta == nil {
			return fmt.Errorf("content_block_delta has no delta")
		}
		if response.Delta.Text != nil && *response.Delta.Text != "" {
			s.Output = true
		}
		if response.Delta.Thinking != nil && *response.Delta.Thinking != "" {
			s.Output = true
		}
		if response.Delta.Type == "input_json_delta" {
			index := response.GetIndex()
			tool := s.tools[index]
			if tool == nil {
				return fmt.Errorf("tool arguments received before tool_use block %d", index)
			}
			if response.Delta.PartialJson != nil {
				tool.arguments.WriteString(*response.Delta.PartialJson)
			}
		}
	case "message_delta":
		if response.Delta != nil && response.Delta.StopReason != nil {
			s.Terminal = true
			if strings.Contains(strings.ToLower(*response.Delta.StopReason), "refusal") {
				s.Filtered = true
			}
		}
	case "message_stop":
		s.Terminal = true
	}
	return nil
}

func (s *ClaudeStreamState) Validate() error {
	if err := s.ValidateOutput(); err != nil {
		return err
	}
	if !s.Terminal {
		return fmt.Errorf("stream ended before message_stop")
	}
	return nil
}

func (s *ClaudeStreamState) ValidateOutput() error {
	if s == nil {
		return fmt.Errorf("stream state is nil")
	}
	for index, tool := range s.tools {
		if strings.TrimSpace(tool.name) == "" {
			return fmt.Errorf("tool call %d has no name", index)
		}
		if !validArguments(tool.arguments.String()) {
			return fmt.Errorf("tool call %d has incomplete JSON arguments", index)
		}
		s.Output = true
	}
	if !s.Valid() {
		return fmt.Errorf("stream contained no semantic output")
	}
	return nil
}

func GeminiResponse(response *dto.GeminiChatResponse) (Result, error) {
	var result Result
	if response == nil {
		return result, fmt.Errorf("response is nil")
	}
	if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != nil {
		result.Filtered = true
		result.Terminal = true
	}
	for candidateIndex, candidate := range response.Candidates {
		if candidate.FinishReason != nil && *candidate.FinishReason != "" {
			result.Terminal = true
			switch strings.ToUpper(*candidate.FinishReason) {
			case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
				result.Filtered = true
			}
		}
		for partIndex, part := range candidate.Content.Parts {
			switch {
			case part.FunctionCall != nil:
				if strings.TrimSpace(part.FunctionCall.FunctionName) == "" {
					return result, fmt.Errorf("candidate %d part %d function call has no name", candidateIndex, partIndex)
				}
				result.Output = true
			case part.Text != "":
				result.Output = true
			case part.InlineData != nil && part.InlineData.Data != "":
				result.Output = true
			case part.FileData != nil && part.FileData.FileUri != "":
				result.Output = true
			case part.ExecutableCode != nil && part.ExecutableCode.Code != "":
				result.Output = true
			case part.CodeExecutionResult != nil && (part.CodeExecutionResult.Output != "" || part.CodeExecutionResult.Outcome != ""):
				result.Output = true
			}
		}
	}
	return result, nil
}

func ResponsesResponse(response *dto.OpenAIResponsesResponse) (Result, error) {
	var result Result
	if response == nil {
		return result, fmt.Errorf("response is nil")
	}
	var status string
	if len(response.Status) > 0 {
		_ = common.Unmarshal(response.Status, &status)
	}
	switch status {
	case "completed":
		result.Terminal = true
	case "failed", "cancelled":
		return result, fmt.Errorf("responses API returned status %s", status)
	case "incomplete":
		result.Terminal = true
		if response.IncompleteDetails != nil && response.IncompleteDetails.Reason == "content_filter" {
			result.Filtered = true
		}
	}
	for i, output := range response.Output {
		switch output.Type {
		case "function_call", "custom_tool_call":
			if strings.TrimSpace(output.Name) == "" {
				return result, fmt.Errorf("output item %d tool call has no name", i)
			}
			if !validArguments(output.ArgumentsString()) {
				return result, fmt.Errorf("output item %d tool call has invalid JSON arguments", i)
			}
			result.Output = true
		case "image_generation_call", "computer_call", "web_search_call", "file_search_call", "code_interpreter_call":
			result.Output = true
		default:
			for _, content := range output.Content {
				if content.Text != "" || content.Refusal != "" {
					result.Output = true
				}
			}
		}
	}
	return result, nil
}

type ResponsesStreamState struct {
	Result
	tools map[string]*streamedToolCall
}

func NewResponsesStreamState() *ResponsesStreamState {
	return &ResponsesStreamState{tools: make(map[string]*streamedToolCall)}
}

func responsesToolKey(event *dto.ResponsesStreamResponse) string {
	if event.ItemID != "" {
		return event.ItemID
	}
	if event.Item != nil && event.Item.ID != "" {
		return event.Item.ID
	}
	if event.OutputIndex != nil {
		return fmt.Sprintf("output:%d", *event.OutputIndex)
	}
	return "tool:0"
}

func responsesToolKeys(event *dto.ResponsesStreamResponse) []string {
	keys := make([]string, 0, 3)
	if event.ItemID != "" {
		keys = append(keys, event.ItemID)
	}
	if event.Item != nil && event.Item.ID != "" {
		keys = append(keys, event.Item.ID)
	}
	if event.OutputIndex != nil {
		keys = append(keys, fmt.Sprintf("output:%d", *event.OutputIndex))
	}
	if len(keys) == 0 {
		keys = append(keys, responsesToolKey(event))
	}
	return keys
}

func (s *ResponsesStreamState) toolFor(event *dto.ResponsesStreamResponse) *streamedToolCall {
	keys := responsesToolKeys(event)
	var tool *streamedToolCall
	for _, key := range keys {
		if s.tools[key] != nil {
			tool = s.tools[key]
			break
		}
	}
	if tool == nil {
		tool = &streamedToolCall{}
	}
	for _, key := range keys {
		s.tools[key] = tool
	}
	return tool
}

func (s *ResponsesStreamState) Observe(event *dto.ResponsesStreamResponse) error {
	if s == nil || event == nil {
		return fmt.Errorf("response event is nil")
	}
	switch event.Type {
	case "response.failed", "response.error":
		return fmt.Errorf("responses stream returned %s", event.Type)
	case "response.completed", "response.incomplete":
		s.Terminal = true
		if event.Response != nil {
			result, err := ResponsesResponse(event.Response)
			if err != nil {
				return err
			}
			s.Output = s.Output || result.Output
			s.Filtered = s.Filtered || result.Filtered
		}
	case "response.output_text.delta", "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		if event.Delta != "" {
			s.Output = true
		}
	case "response.image_generation_call.partial_image":
		// Partial image bytes are semantic output even though only completed
		// image items are eligible for billing.
		s.Output = true
	case "response.output_item.added", "response.output_item.done":
		if event.Item == nil {
			return nil
		}
		switch event.Item.Type {
		case "function_call", "custom_tool_call":
			if strings.TrimSpace(event.Item.Name) == "" {
				return fmt.Errorf("stream tool call has no name")
			}
			tool := s.toolFor(event)
			tool.name = event.Item.Name
			arguments := event.Item.ArgumentsString()
			if arguments != "" && tool.arguments.Len() == 0 {
				tool.arguments.WriteString(arguments)
			}
		case "image_generation_call", "computer_call", "web_search_call", "file_search_call", "code_interpreter_call":
			s.Output = true
		default:
			for _, content := range event.Item.Content {
				if content.Text != "" || content.Refusal != "" {
					s.Output = true
				}
			}
		}
	case "response.function_call_arguments.delta", "response.custom_tool_call_input.delta":
		tool := s.toolFor(event)
		tool.arguments.WriteString(event.Delta)
	case "response.function_call_arguments.done", "response.custom_tool_call_input.done":
		tool := s.toolFor(event)
		if tool.arguments.Len() == 0 {
			tool.arguments.WriteString(event.Delta)
		}
	}
	return nil
}

func (s *ResponsesStreamState) Validate() error {
	if err := s.ValidateOutput(); err != nil {
		return err
	}
	if !s.Terminal {
		return fmt.Errorf("stream ended before response.completed or response.incomplete")
	}
	return nil
}

func (s *ResponsesStreamState) ValidateOutput() error {
	if s == nil {
		return fmt.Errorf("stream state is nil")
	}
	seen := make(map[*streamedToolCall]bool)
	for key, tool := range s.tools {
		if seen[tool] {
			continue
		}
		seen[tool] = true
		if strings.TrimSpace(tool.name) == "" {
			return fmt.Errorf("tool call %s has no name", key)
		}
		if !validArguments(tool.arguments.String()) {
			return fmt.Errorf("tool call %s has incomplete JSON arguments", key)
		}
		s.Output = true
	}
	if !s.Valid() {
		return fmt.Errorf("stream contained no semantic output")
	}
	return nil
}
