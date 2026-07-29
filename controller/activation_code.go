package controller

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type activationRequest struct {
	ActivationCode string `json:"activation_code"`
	QQ             string `json:"qq"`
	ApiKey         string `json:"api_key"`
}

type activationCreateRequest struct {
	Count       int      `json:"count"`
	Days        int      `json:"days"`
	Channel     string   `json:"channel"`
	ExpiredTime int64    `json:"expired_time"`
	Codes       []string `json:"codes"`
}

type activationBatchRequest struct {
	Ids         []int    `json:"ids"`
	Codes       []string `json:"codes"`
	Days        *int     `json:"days"`
	Channel     *string  `json:"channel"`
	ExpiredTime *int64   `json:"expired_time"`
	Status      *int     `json:"status"`
}

func activationClientIp(c *gin.Context) string {
	if realIp := strings.TrimSpace(c.GetHeader("X-Real-IP")); realIp != "" {
		return realIp
	}
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}
	return c.ClientIP()
}

func activationPublicError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if !errors.Is(err, model.ErrActivationCodeInvalid) &&
		!errors.Is(err, model.ErrActivationCodeExpired) &&
		!errors.Is(err, model.ErrActivationCodeOwnerMismatch) &&
		!errors.Is(err, model.ErrActivationTokenExists) &&
		!errors.Is(err, model.ErrActivationTokenNotFound) &&
		!errors.Is(err, model.ErrActivationOwnerUnavailable) &&
		!errors.Is(err, model.ErrActivationRequestInvalid) &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusInternalServerError
		common.SysError("activation code request failed: " + err.Error())
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func PrecheckActivationCode(c *gin.Context) {
	var request activationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		activationPublicError(c, fmt.Errorf("%w: 请求格式无效", model.ErrActivationRequestInvalid))
		return
	}
	result, err := model.GetActivationPrecheck(request.ActivationCode, request.QQ, request.ApiKey, common.GetTimestamp())
	if err != nil {
		activationPublicError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":            true,
		"channel":          result.Channel,
		"days":             result.Days,
		"client_ip":        activationClientIp(c),
		"new_expired_time": result.NewExpiredTime,
		"old_expired_time": result.OldExpiredTime,
	})
}

func RedeemActivationCode(c *gin.Context) {
	var request activationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		activationPublicError(c, fmt.Errorf("%w: 请求格式无效", model.ErrActivationRequestInvalid))
		return
	}
	result, err := model.RedeemActivationCode(request.ActivationCode, request.QQ, activationClientIp(c), common.GetTimestamp())
	if err != nil {
		activationPublicError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func RenewActivationCode(c *gin.Context) {
	var request activationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		activationPublicError(c, fmt.Errorf("%w: 请求格式无效", model.ErrActivationRequestInvalid))
		return
	}
	result, err := model.RenewActivationCode(request.ActivationCode, request.ApiKey, activationClientIp(c), common.GetTimestamp())
	if err != nil {
		activationPublicError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"new_expired_time": result.NewExpiredTime, "message": "续费成功"})
}

func QueryActivationCode(c *gin.Context) {
	var request activationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		activationPublicError(c, fmt.Errorf("%w: 请求格式无效", model.ErrActivationRequestInvalid))
		return
	}
	logEntry, err := model.GetActivationLogByCode(request.ActivationCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			activationPublicError(c, fmt.Errorf("%w: 该激活码未被使用或不存在", model.ErrActivationRequestInvalid))
			return
		}
		activationPublicError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"activation_code": logEntry.ActivationCode,
		"used_time":       time.Unix(logEntry.UsedTime, 0).Format("2006-01-02 15:04"),
		"used_ip":         logEntry.ClientIp,
		"action":          logEntry.Action,
		"api_key":         logEntry.TokenKey,
	})
}

func activationFilters(c *gin.Context) model.ActivationCodeFilters {
	status, _ := strconv.Atoi(c.Query("status"))
	days, _ := strconv.Atoi(c.Query("days"))
	createdFrom, _ := strconv.ParseInt(c.Query("created_from"), 10, 64)
	createdTo, _ := strconv.ParseInt(c.Query("created_to"), 10, 64)
	expiresFrom, _ := strconv.ParseInt(c.Query("expires_from"), 10, 64)
	expiresTo, _ := strconv.ParseInt(c.Query("expires_to"), 10, 64)
	return model.ActivationCodeFilters{
		Search:      c.Query("search"),
		Channel:     c.Query("channel"),
		Status:      status,
		Days:        days,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		ExpiresFrom: expiresFrom,
		ExpiresTo:   expiresTo,
	}
}

func ListActivationCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	codes, total, err := model.ListUserActivationCodes(
		c.GetInt("id"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
		activationFilters(c),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func ListActivationLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	logs, total, err := model.ListUserActivationLogs(
		c.GetInt("id"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
		c.Query("search"),
		c.Query("action"),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}

func CreateActivationCodes(c *gin.Context) {
	var request activationCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	if len(request.Codes) > 0 {
		request.Count = len(request.Codes)
	}
	codes, err := model.CreateUserActivationCodes(
		c.GetInt("id"),
		request.Count,
		request.Days,
		request.Channel,
		request.ExpiredTime,
		request.Codes,
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, codes)
}

func UpdateActivationCodes(c *gin.Context) {
	var request activationBatchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	updates := map[string]any{}
	if request.Days != nil {
		if *request.Days <= 0 {
			common.ApiErrorMsg(c, "有效天数必须大于 0")
			return
		}
		updates["days"] = *request.Days
	}
	if request.Channel != nil {
		channel := strings.TrimSpace(*request.Channel)
		if channel == "" || len(channel) > 100 {
			common.ApiErrorMsg(c, "渠道名称无效")
			return
		}
		updates["channel"] = channel
	}
	if request.ExpiredTime != nil {
		if *request.ExpiredTime <= common.GetTimestamp() {
			common.ApiErrorMsg(c, "过期时间必须晚于当前时间")
			return
		}
		updates["expired_time"] = *request.ExpiredTime
	}
	if request.Status != nil {
		if *request.Status != model.ActivationCodeStatusActive && *request.Status != model.ActivationCodeStatusDisabled {
			common.ApiErrorMsg(c, "状态无效")
			return
		}
		updates["status"] = *request.Status
	}
	if len(updates) == 0 {
		common.ApiErrorMsg(c, "没有可更新的字段")
		return
	}
	count, err := model.UpdateUserActivationCodes(c.GetInt("id"), request.Ids, request.Codes, updates)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"updated": count})
}

func DeleteActivationCodes(c *gin.Context) {
	var request activationBatchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	count, err := model.DeleteUserActivationCodes(c.GetInt("id"), request.Ids, request.Codes)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"deleted": count})
}

func ExportActivationCodes(c *gin.Context) {
	codes, err := model.ListAllUserActivationCodes(c.GetInt("id"), activationFilters(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=activation-codes.csv")
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{"code", "days", "channel", "status", "expires_at", "created_at", "used_at"})
	for _, code := range codes {
		_ = writer.Write([]string{
			code.Code,
			strconv.Itoa(code.Days),
			code.Channel,
			strconv.Itoa(code.Status),
			fmt.Sprintf("%d", code.ExpiredTime),
			fmt.Sprintf("%d", code.CreatedTime),
			fmt.Sprintf("%d", code.UsedTime),
		})
	}
	writer.Flush()
}
