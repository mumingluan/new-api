package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

var (
	ErrKeyBatchAllUsersForbidden = errors.New("only administrators can operate on all users' keys")
	ErrKeyBatchInvalidRequest    = errors.New("invalid key batch request")
)

func keyBatchInvalid(message string) error {
	return fmt.Errorf("%w: %s", ErrKeyBatchInvalidRequest, message)
}

const (
	MaxKeyBatchDurationSeconds = int64(100 * 366 * 24 * 60 * 60)
	MaxKeyBatchStatisticsRange = int64(10 * 366 * 24 * 60 * 60)
	MaxKeyBatchStatisticsTop   = 500
	MaxKeyBatchMinTokens       = 1_000_000_000
)

type KeyBatchFilter struct {
	AllUsers          bool `json:"all_users"`
	UsedOnly          bool `json:"used_only"`
	MinRemainingQuota *int `json:"min_remaining_quota"`
}

type KeyBatchOperationRequest struct {
	Action          model.TokenBatchAction `json:"action"`
	DurationSeconds int64                  `json:"duration_seconds"`
	Quota           int                    `json:"quota"`
	Filter          KeyBatchFilter         `json:"filter"`
}

type KeyBatchStatisticsRequest struct {
	AllUsers      bool
	StartTime     int64
	EndTime       int64
	GroupBy       string
	SortBy        string
	SortOrder     string
	UserFilterID  int
	ExcludeUserID int
	Model         string
	MinTokens     int
	Top           int
}

func resolveKeyBatchFilter(userID int, role int, filter KeyBatchFilter) (model.TokenBatchFilter, error) {
	if userID <= 0 {
		return model.TokenBatchFilter{}, keyBatchInvalid("invalid user")
	}
	if filter.AllUsers && role < common.RoleAdminUser {
		return model.TokenBatchFilter{}, ErrKeyBatchAllUsersForbidden
	}
	if filter.MinRemainingQuota != nil && (*filter.MinRemainingQuota < 0 || *filter.MinRemainingQuota > common.MaxQuota) {
		return model.TokenBatchFilter{}, keyBatchInvalid(fmt.Sprintf("minimum remaining quota must be between 0 and %d", common.MaxQuota))
	}
	return model.TokenBatchFilter{
		UserID:            userID,
		AllUsers:          filter.AllUsers,
		UsedOnly:          filter.UsedOnly,
		MinRemainingQuota: filter.MinRemainingQuota,
	}, nil
}

func validateKeyBatchOperation(request KeyBatchOperationRequest) error {
	switch request.Action {
	case model.TokenBatchExtendExpiry, model.TokenBatchDeductExpiry:
		if request.DurationSeconds <= 0 || request.DurationSeconds > MaxKeyBatchDurationSeconds {
			return keyBatchInvalid(fmt.Sprintf("duration must be between 1 and %d seconds", MaxKeyBatchDurationSeconds))
		}
	case model.TokenBatchAddQuota, model.TokenBatchDeductQuota:
		if request.Quota <= 0 || request.Quota > common.MaxQuota {
			return keyBatchInvalid(fmt.Sprintf("quota must be between 1 and %d", common.MaxQuota))
		}
	default:
		return keyBatchInvalid("invalid batch action")
	}
	return nil
}

func PreviewKeyBatch(userID int, role int, request KeyBatchOperationRequest, now int64) (*model.TokenBatchPreview, error) {
	if err := validateKeyBatchOperation(request); err != nil {
		return nil, err
	}
	filter, err := resolveKeyBatchFilter(userID, role, request.Filter)
	if err != nil {
		return nil, err
	}
	return model.PreviewTokenBatch(filter, request.Action, now)
}

func ExecuteKeyBatch(userID int, role int, request KeyBatchOperationRequest, now int64) (int64, error) {
	if err := validateKeyBatchOperation(request); err != nil {
		return 0, err
	}
	filter, err := resolveKeyBatchFilter(userID, role, request.Filter)
	if err != nil {
		return 0, err
	}
	return model.ExecuteTokenBatch(filter, request.Action, request.DurationSeconds, request.Quota, now)
}

func BuildKeyBatchStatisticsQuery(userID int, role int, request KeyBatchStatisticsRequest) (model.TokenBatchStatisticsQuery, error) {
	filter, err := resolveKeyBatchFilter(userID, role, KeyBatchFilter{AllUsers: request.AllUsers})
	if err != nil {
		return model.TokenBatchStatisticsQuery{}, err
	}
	if request.StartTime <= 0 || request.EndTime <= request.StartTime {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("invalid statistics time range")
	}
	if request.EndTime-request.StartTime > MaxKeyBatchStatisticsRange {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("statistics time range is too large")
	}
	groups := map[string]bool{"token_name": true, "model_name": true, "username": true, "channel_name": true, "user_id": true}
	if !groups[request.GroupBy] {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("invalid statistics group")
	}
	sorts := map[string]bool{"request_count": true, "prompt_tokens": true, "completion_tokens": true, "quota": true}
	if !sorts[request.SortBy] {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("invalid statistics sort")
	}
	request.SortOrder = strings.ToLower(request.SortOrder)
	if request.SortOrder != "asc" && request.SortOrder != "desc" {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("invalid statistics sort order")
	}
	if request.Top <= 0 || request.Top > MaxKeyBatchStatisticsTop {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid(fmt.Sprintf("top must be between 1 and %d", MaxKeyBatchStatisticsTop))
	}
	if request.MinTokens < 0 || request.MinTokens > MaxKeyBatchMinTokens {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid(fmt.Sprintf("minimum tokens must be between 0 and %d", MaxKeyBatchMinTokens))
	}
	if request.UserFilterID < 0 || request.ExcludeUserID < 0 {
		return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("user filters must be positive")
	}
	if !request.AllUsers {
		if request.UserFilterID > 0 && request.UserFilterID != userID {
			return model.TokenBatchStatisticsQuery{}, ErrKeyBatchAllUsersForbidden
		}
		if request.ExcludeUserID == userID {
			return model.TokenBatchStatisticsQuery{}, keyBatchInvalid("cannot exclude the current user in current-user scope")
		}
	}
	return model.TokenBatchStatisticsQuery{
		UserID:        filter.UserID,
		AllUsers:      filter.AllUsers,
		StartTime:     request.StartTime,
		EndTime:       request.EndTime,
		GroupBy:       request.GroupBy,
		SortBy:        request.SortBy,
		SortOrder:     request.SortOrder,
		UserFilterID:  request.UserFilterID,
		ExcludeUserID: request.ExcludeUserID,
		Model:         strings.TrimSpace(request.Model),
		MinTokens:     request.MinTokens,
		Top:           request.Top,
	}, nil
}

func QueryKeyBatchStatistics(userID int, role int, request KeyBatchStatisticsRequest) ([]model.TokenBatchStatisticsRow, model.TokenBatchStatisticsTotals, error) {
	query, err := BuildKeyBatchStatisticsQuery(userID, role, request)
	if err != nil {
		return nil, model.TokenBatchStatisticsTotals{}, err
	}
	return model.QueryTokenBatchStatistics(query)
}
