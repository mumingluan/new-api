package claude

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHandleClaudeResponseData_RejectsToolUseStopReasonWithEmptyContent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	info := &relaycommon.RelayInfo{RelayFormat: types.RelayFormatOpenAI}
	claudeInfo := &ClaudeResponseInfo{Usage: &dto.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}}

	payload := dto.ClaudeResponse{
		Id:         "msg_1",
		Type:       "message",
		Role:       "assistant",
		StopReason: "tool_use",
		Content:    []dto.ClaudeMediaMessage{},
		Completion: "",
		Usage:      &dto.ClaudeUsage{InputTokens: 1, OutputTokens: 1},
	}
	data, err := common.Marshal(payload)
	require.NoError(t, err)

	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	apiErr := HandleClaudeResponseData(c, info, claudeInfo, resp, data)
	require.NotNil(t, apiErr)
	require.Equal(t, types.ErrorCodeEmptyResponse, apiErr.GetErrorCode())
}
