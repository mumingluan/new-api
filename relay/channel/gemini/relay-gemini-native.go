package gemini

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relay/responsevalidator"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func GeminiTextGenerationHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	logger.LogDebug(c, "Gemini native response body: %s", responseBody)

	// 解析为 Gemini 原生响应格式
	var geminiResponse dto.GeminiChatResponse
	err = common.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
		return nil, types.NewOpenAIError(errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason), types.ErrorCodePromptBlocked, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
	}

	validity, validationErr := responsevalidator.GeminiResponse(&geminiResponse)
	if validationErr != nil {
		return nil, types.NewOpenAIError(validationErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if !validity.Valid() {
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
	}

	// 计算使用量（优先上游 UsageMetadata，缺失时本地估算并保留 Gemini 计费语义）
	usage := buildUsageFromGeminiResponse(c, info, &geminiResponse)

	service.IOCopyBytesGracefully(c, resp, responseBody)

	return &usage, nil
}

func NativeGeminiEmbeddingHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	logger.LogDebug(c, "Gemini native embedding response body: %s", responseBody)

	usage := service.ResponseText2Usage(c, "", info.UpstreamModelName, info.GetEstimatePromptTokens())

	if info.IsGeminiBatchEmbedding {
		var geminiResponse dto.GeminiBatchEmbeddingResponse
		err = common.Unmarshal(responseBody, &geminiResponse)
		if err != nil {
			return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
	} else {
		var geminiResponse dto.GeminiEmbeddingResponse
		err = common.Unmarshal(responseBody, &geminiResponse)
		if err != nil {
			return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
	}

	service.IOCopyBytesGracefully(c, resp, responseBody)

	return usage, nil
}

func GeminiTextGenerationStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	helper.SetEventStreamHeaders(c)

	var validationErr error
	var validationAPIError *types.NewAPIError
	sawTerminal := false
	usage, err := geminiStreamHandler(c, info, resp, func(data string, geminiResponse *dto.GeminiChatResponse) bool {
		if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
			validationAPIError = types.NewOpenAIError(
				errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason),
				types.ErrorCodePromptBlocked,
				http.StatusBadRequest,
				types.ErrOptionWithSkipRetry(),
			)
			return false
		}
		validity, currentErr := responsevalidator.GeminiResponse(geminiResponse)
		if currentErr != nil {
			validationErr = currentErr
			return false
		}
		sawTerminal = sawTerminal || validity.Terminal
		if !validity.Valid() {
			return true
		}
		err := helper.StringData(c, data)
		if err != nil {
			logger.LogError(c, "failed to write stream data: "+err.Error())
			return false
		}
		info.SendResponseCount++
		return true
	})

	if err != nil {
		return usage, err
	}
	if validationAPIError != nil {
		return nil, validationAPIError
	}
	if validationErr != nil {
		return nil, types.NewOpenAIError(validationErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if info.SendResponseCount == 0 {
		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
	}
	if !sawTerminal {
		return nil, types.NewOpenAIError(errors.New("Gemini stream ended before a terminal finish reason"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	return usage, nil
}
