package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
)

// geminiVideoResult holds either a remote URL or inline base64 video data.
type geminiVideoResult struct {
	URL      string // remote video URL (preferred if available)
	Data     []byte // decoded video bytes from bytesBase64Encoded
	MimeType string // e.g. "video/mp4"
}

func getGeminiVideoResult(channel *model.Channel, task *model.Task, apiKey string) (*geminiVideoResult, error) {
	if channel == nil || task == nil {
		return nil, fmt.Errorf("invalid channel or task")
	}

	// Try extracting URL from stored task data first
	if url := extractGeminiVideoURLFromTaskData(task); url != "" {
		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
	}

	baseURL := constant.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() != "" {
		baseURL = channel.GetBaseURL()
	}

	adaptor := relay.GetTaskAdaptor(constant.TaskPlatform(strconv.Itoa(channel.Type)))
	if adaptor == nil {
		return nil, fmt.Errorf("gemini task adaptor not found")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api key not available for task")
	}

	proxy := channel.GetSetting().Proxy
	resp, err := adaptor.FetchTask(baseURL, apiKey, map[string]any{
		"task_id": task.GetUpstreamTaskID(),
		"action":  task.Action,
	}, proxy)
	if err != nil {
		return nil, fmt.Errorf("fetch task failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read task response failed: %w", err)
	}

	// Try remote URL from parsed result
	taskInfo, parseErr := adaptor.ParseTaskResult(body)
	if parseErr == nil && taskInfo != nil && taskInfo.RemoteUrl != "" {
		return &geminiVideoResult{URL: ensureAPIKey(taskInfo.RemoteUrl, apiKey)}, nil
	}

	// Try remote URL from raw payload
	if url := extractGeminiVideoURLFromPayload(body); url != "" {
		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
	}

	// Try inline base64 video data (Veo returns bytesBase64Encoded directly)
	if result := extractGeminiVideoBase64FromPayload(body); result != nil {
		return result, nil
	}

	if parseErr != nil {
		return nil, fmt.Errorf("parse task result failed: %w", parseErr)
	}

	return nil, fmt.Errorf("gemini video data not found")
}

// extractGeminiVideoBase64FromPayload extracts inline base64 video from fetchPredictOperation response.
func extractGeminiVideoBase64FromPayload(body []byte) *geminiVideoResult {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	resp, ok := payload["response"].(map[string]any)
	if !ok {
		return nil
	}

	// Check response.videos[0].bytesBase64Encoded
	if videos, ok := resp["videos"].([]any); ok && len(videos) > 0 {
		if vm, ok := videos[0].(map[string]any); ok {
			if b64, ok := vm["bytesBase64Encoded"].(string); ok && b64 != "" {
				data, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return nil
				}
				mime := "video/mp4"
				if m, ok := vm["mimeType"].(string); ok && m != "" {
					mime = m
				}
				return &geminiVideoResult{Data: data, MimeType: mime}
			}
		}
	}

	// Check response.bytesBase64Encoded directly
	if b64, ok := resp["bytesBase64Encoded"].(string); ok && b64 != "" {
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil
		}
		return &geminiVideoResult{Data: data, MimeType: "video/mp4"}
	}

	return nil
}

func extractGeminiVideoURLFromTaskData(task *model.Task) string {
	if task == nil || len(task.Data) == 0 {
		return ""
	}
	var payload map[string]any
	if err := common.Unmarshal(task.Data, &payload); err != nil {
		return ""
	}
	return extractGeminiVideoURLFromMap(payload)
}

func extractGeminiVideoURLFromPayload(body []byte) string {
	var payload map[string]any
	if err := common.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return extractGeminiVideoURLFromMap(payload)
}

func extractGeminiVideoURLFromMap(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if uri, ok := payload["uri"].(string); ok && uri != "" {
		return uri
	}
	if resp, ok := payload["response"].(map[string]any); ok {
		if uri := extractGeminiVideoURLFromResponse(resp); uri != "" {
			return uri
		}
	}
	return ""
}

func extractGeminiVideoURLFromResponse(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	if gvr, ok := resp["generateVideoResponse"].(map[string]any); ok {
		if uri := extractGeminiVideoURLFromGeneratedSamples(gvr); uri != "" {
			return uri
		}
	}
	if videos, ok := resp["videos"].([]any); ok {
		for _, video := range videos {
			if vm, ok := video.(map[string]any); ok {
				if uri, ok := vm["uri"].(string); ok && uri != "" {
					return uri
				}
			}
		}
	}
	if uri, ok := resp["video"].(string); ok && uri != "" {
		return uri
	}
	if uri, ok := resp["uri"].(string); ok && uri != "" {
		return uri
	}
	return ""
}

func extractGeminiVideoURLFromGeneratedSamples(gvr map[string]any) string {
	if gvr == nil {
		return ""
	}
	if samples, ok := gvr["generatedSamples"].([]any); ok {
		for _, sample := range samples {
			if sm, ok := sample.(map[string]any); ok {
				if video, ok := sm["video"].(map[string]any); ok {
					if uri, ok := video["uri"].(string); ok && uri != "" {
						return uri
					}
				}
			}
		}
	}
	return ""
}

func getVertexVideoURL(channel *model.Channel, task *model.Task) (string, error) {
	if channel == nil || task == nil {
		return "", fmt.Errorf("invalid channel or task")
	}
	if url := strings.TrimSpace(task.GetResultURL()); url != "" && !isTaskProxyContentURL(url, task.TaskID) {
		return url, nil
	}
	if url := extractVertexVideoURLFromTaskData(task); url != "" {
		return url, nil
	}

	baseURL := constant.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() != "" {
		baseURL = channel.GetBaseURL()
	}

	adaptor := relay.GetTaskAdaptor(constant.TaskPlatform(strconv.Itoa(channel.Type)))
	if adaptor == nil {
		return "", fmt.Errorf("vertex task adaptor not found")
	}

	key := getVertexTaskKey(channel, task)
	if key == "" {
		return "", fmt.Errorf("vertex key not available for task")
	}

	resp, err := adaptor.FetchTask(baseURL, key, map[string]any{
		"task_id": task.GetUpstreamTaskID(),
		"action":  task.Action,
	}, channel.GetSetting().Proxy)
	if err != nil {
		return "", fmt.Errorf("fetch task failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read task response failed: %w", err)
	}

	taskInfo, parseErr := adaptor.ParseTaskResult(body)
	if parseErr == nil && taskInfo != nil && strings.TrimSpace(taskInfo.Url) != "" {
		return taskInfo.Url, nil
	}
	if url := extractVertexVideoURLFromPayload(body); url != "" {
		return url, nil
	}
	if parseErr != nil {
		return "", fmt.Errorf("parse task result failed: %w", parseErr)
	}
	return "", fmt.Errorf("vertex video url not found")
}

func isTaskProxyContentURL(url string, taskID string) bool {
	if strings.TrimSpace(url) == "" || strings.TrimSpace(taskID) == "" {
		return false
	}
	return strings.Contains(url, "/v1/videos/"+taskID+"/content")
}

func getVertexTaskKey(channel *model.Channel, task *model.Task) string {
	if task != nil {
		if key := strings.TrimSpace(task.PrivateData.Key); key != "" {
			return key
		}
	}
	if channel == nil {
		return ""
	}
	keys := channel.GetKeys()
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			return key
		}
	}
	return strings.TrimSpace(channel.Key)
}

func extractVertexVideoURLFromTaskData(task *model.Task) string {
	if task == nil || len(task.Data) == 0 {
		return ""
	}
	return extractVertexVideoURLFromPayload(task.Data)
}

func extractVertexVideoURLFromPayload(body []byte) string {
	var payload map[string]any
	if err := common.Unmarshal(body, &payload); err != nil {
		return ""
	}
	resp, ok := payload["response"].(map[string]any)
	if !ok || resp == nil {
		return ""
	}

	if videos, ok := resp["videos"].([]any); ok && len(videos) > 0 {
		if video, ok := videos[0].(map[string]any); ok && video != nil {
			if b64, _ := video["bytesBase64Encoded"].(string); strings.TrimSpace(b64) != "" {
				mime, _ := video["mimeType"].(string)
				enc, _ := video["encoding"].(string)
				return buildVideoDataURL(mime, enc, b64)
			}
		}
	}
	if b64, _ := resp["bytesBase64Encoded"].(string); strings.TrimSpace(b64) != "" {
		enc, _ := resp["encoding"].(string)
		return buildVideoDataURL("", enc, b64)
	}
	if video, _ := resp["video"].(string); strings.TrimSpace(video) != "" {
		if strings.HasPrefix(video, "data:") || strings.HasPrefix(video, "http://") || strings.HasPrefix(video, "https://") {
			return video
		}
		enc, _ := resp["encoding"].(string)
		return buildVideoDataURL("", enc, video)
	}
	return ""
}

func buildVideoDataURL(mimeType string, encoding string, base64Data string) string {
	mime := strings.TrimSpace(mimeType)
	if mime == "" {
		enc := strings.TrimSpace(encoding)
		if enc == "" {
			enc = "mp4"
		}
		if strings.Contains(enc, "/") {
			mime = enc
		} else {
			mime = "video/" + enc
		}
	}
	return "data:" + mime + ";base64," + base64Data
}

func ensureAPIKey(uri, key string) string {
	if key == "" || uri == "" {
		return uri
	}
	if strings.Contains(uri, "key=") {
		return uri
	}
	if strings.Contains(uri, "?") {
		return fmt.Sprintf("%s&key=%s", uri, key)
	}
	return fmt.Sprintf("%s?key=%s", uri, key)
}
