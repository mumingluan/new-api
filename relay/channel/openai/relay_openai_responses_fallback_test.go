package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenaiHandler_FallbackResponsesToolCallOnly(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	info := &relaycommon.RelayInfo{
		RelayMode:          relayconstant.RelayModeChatCompletions,
		RelayFormat:        types.RelayFormatOpenAI,
		OriginModelName:    "gpt-test",
		ShouldIncludeUsage: true,
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gpt-test",
			ChannelSetting:    dto.ChannelSettings{},
		},
	}
	info.SetEstimatePromptTokens(3)

	upstream := dto.OpenAIResponsesResponse{
		ID:        "resp_test",
		Object:    "response",
		CreatedAt: 123,
		Model:     "gpt-test",
		Output: []dto.ResponsesOutput{
			{
				Type:      "function_call",
				ID:        "item_1",
				Status:    "completed",
				Role:      "assistant",
				CallId:    "call_1",
				Name:      "get_weather",
				Arguments: json.RawMessage(`{"location":"SF"}`),
			},
		},
		Usage: &dto.Usage{InputTokens: 3, OutputTokens: 4, TotalTokens: 7},
	}

	body, err := common.Marshal(upstream)
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}

	usage, apiErr := OpenaiHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	require.Equal(t, 7, usage.TotalTokens)

	var out dto.OpenAITextResponse
	require.NoError(t, common.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Choices, 1)
	toolCalls := out.Choices[0].Message.ParseToolCalls()
	require.NotEmpty(t, toolCalls)
	require.Equal(t, "", out.Choices[0].Message.StringContent())
}
