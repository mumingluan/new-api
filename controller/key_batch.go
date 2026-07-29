package controller

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func keyBatchError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrKeyBatchAllUsersForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	if errors.Is(err, service.ErrKeyBatchInvalidRequest) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.SysError("key batch request failed: " + err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Key batch request failed"})
}

func bindKeyBatchOperation(c *gin.Context) (service.KeyBatchOperationRequest, bool) {
	var request service.KeyBatchOperationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		keyBatchError(c, err)
		return request, false
	}
	return request, true
}

func PreviewKeyBatch(c *gin.Context) {
	request, ok := bindKeyBatchOperation(c)
	if !ok {
		return
	}
	preview, err := service.PreviewKeyBatch(c.GetInt("id"), c.GetInt("role"), request, common.GetTimestamp())
	if err != nil {
		keyBatchError(c, err)
		return
	}
	common.ApiSuccess(c, preview)
}

func ExecuteKeyBatch(c *gin.Context) {
	request, ok := bindKeyBatchOperation(c)
	if !ok {
		return
	}
	affected, err := service.ExecuteKeyBatch(c.GetInt("id"), c.GetInt("role"), request, common.GetTimestamp())
	if err != nil {
		keyBatchError(c, err)
		return
	}
	params := map[string]interface{}{
		"action":    request.Action,
		"affected":  affected,
		"all_users": request.Filter.AllUsers,
		"used_only": request.Filter.UsedOnly,
		"duration":  request.DurationSeconds,
		"quota":     request.Quota,
	}
	if request.Filter.MinRemainingQuota != nil {
		params["min_remaining_quota"] = *request.Filter.MinRemainingQuota
	}
	if request.Filter.AllUsers {
		recordManageAudit(c, "token.batch_operation", params)
	} else {
		recordUserSecurityAudit(c, c.GetInt("id"), "token.batch_operation", params)
	}
	common.ApiSuccess(c, gin.H{"affected": affected})
}

func parseKeyBatchStatisticsRequest(c *gin.Context) service.KeyBatchStatisticsRequest {
	allUsers, _ := strconv.ParseBool(c.DefaultQuery("all_users", "false"))
	startTime, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	userID, _ := strconv.Atoi(c.Query("user_id"))
	excludeUserID, _ := strconv.Atoi(c.Query("exclude_user_id"))
	minTokens, _ := strconv.Atoi(c.DefaultQuery("min_tokens", "0"))
	top, _ := strconv.Atoi(c.DefaultQuery("top", "50"))
	return service.KeyBatchStatisticsRequest{
		AllUsers:      allUsers,
		StartTime:     startTime,
		EndTime:       endTime,
		GroupBy:       c.DefaultQuery("group_by", "token_name"),
		SortBy:        c.DefaultQuery("sort_by", "request_count"),
		SortOrder:     c.DefaultQuery("sort_order", "desc"),
		UserFilterID:  userID,
		ExcludeUserID: excludeUserID,
		Model:         c.Query("model"),
		MinTokens:     minTokens,
		Top:           top,
	}
}

func GetKeyBatchStatistics(c *gin.Context) {
	rows, totals, err := service.QueryKeyBatchStatistics(c.GetInt("id"), c.GetInt("role"), parseKeyBatchStatisticsRequest(c))
	if err != nil {
		keyBatchError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": rows, "totals": totals})
}

func ExportKeyBatchStatistics(c *gin.Context) {
	rows, totals, err := service.QueryKeyBatchStatistics(c.GetInt("id"), c.GetInt("role"), parseKeyBatchStatisticsRequest(c))
	if err != nil {
		keyBatchError(c, err)
		return
	}

	filename := fmt.Sprintf("key-statistics-%s.csv", time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{"name", "request_count", "prompt_tokens", "completion_tokens", "quota", "unique_users"})
	for _, row := range rows {
		_ = writer.Write([]string{
			row.Name,
			strconv.FormatInt(row.RequestCount, 10),
			strconv.FormatInt(row.PromptTokens, 10),
			strconv.FormatInt(row.CompletionTokens, 10),
			strconv.FormatInt(row.Quota, 10),
			strconv.FormatInt(row.UniqueUsers, 10),
		})
	}
	_ = writer.Write([]string{
		"TOTAL",
		strconv.FormatInt(totals.RequestCount, 10),
		strconv.FormatInt(totals.PromptTokens, 10),
		strconv.FormatInt(totals.CompletionTokens, 10),
		strconv.FormatInt(totals.Quota, 10),
		strconv.FormatInt(totals.UniqueUsers, 10),
	})
	writer.Flush()
}
