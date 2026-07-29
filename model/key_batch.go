package model

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type TokenBatchAction string

const (
	TokenBatchExtendExpiry TokenBatchAction = "extend_expiry"
	TokenBatchAddQuota     TokenBatchAction = "add_quota"
	TokenBatchDeductExpiry TokenBatchAction = "deduct_expiry"
	TokenBatchDeductQuota  TokenBatchAction = "deduct_quota"
)

type TokenBatchFilter struct {
	UserID            int
	AllUsers          bool
	UsedOnly          bool
	MinRemainingQuota *int
}

type TokenBatchPreview struct {
	MatchedTokens       int64 `json:"matched_tokens"`
	ActionableTokens    int64 `json:"actionable_tokens"`
	AffectedUsers       int64 `json:"affected_users"`
	UsedTokens          int64 `json:"used_tokens"`
	PermanentTokens     int64 `json:"permanent_tokens"`
	UnlimitedTokens     int64 `json:"unlimited_tokens"`
	TotalRemainingQuota int64 `json:"total_remaining_quota"`
	TotalUsedQuota      int64 `json:"total_used_quota"`
}

type TokenBatchStatisticsQuery struct {
	UserID        int
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

type TokenBatchStatisticsRow struct {
	Name             string `json:"name" gorm:"column:name"`
	RequestCount     int64  `json:"request_count" gorm:"column:request_count"`
	PromptTokens     int64  `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens" gorm:"column:completion_tokens"`
	Quota            int64  `json:"quota" gorm:"column:quota"`
	UniqueUsers      int64  `json:"unique_users" gorm:"column:unique_users"`
}

type TokenBatchStatisticsTotals struct {
	RequestCount     int64 `json:"request_count" gorm:"column:request_count"`
	PromptTokens     int64 `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens" gorm:"column:completion_tokens"`
	Quota            int64 `json:"quota" gorm:"column:quota"`
	UniqueUsers      int64 `json:"unique_users" gorm:"column:unique_users"`
}

type tokenBatchPreviewAggregate struct {
	MatchedTokens       int64 `gorm:"column:matched_tokens"`
	AffectedUsers       int64 `gorm:"column:affected_users"`
	UsedTokens          int64 `gorm:"column:used_tokens"`
	PermanentTokens     int64 `gorm:"column:permanent_tokens"`
	UnlimitedTokens     int64 `gorm:"column:unlimited_tokens"`
	TotalRemainingQuota int64 `gorm:"column:total_remaining_quota"`
	TotalUsedQuota      int64 `gorm:"column:total_used_quota"`
}

func applyTokenBatchFilter(tx *gorm.DB, filter TokenBatchFilter) *gorm.DB {
	if !filter.AllUsers {
		tx = tx.Where("user_id = ?", filter.UserID)
	}
	if filter.UsedOnly {
		tx = tx.Where("used_quota > 0")
	}
	if filter.MinRemainingQuota != nil {
		tx = tx.Where("remain_quota > ?", *filter.MinRemainingQuota)
	}
	return tx
}

func applyTokenBatchActionFilter(tx *gorm.DB, action TokenBatchAction, now int64) *gorm.DB {
	switch action {
	case TokenBatchExtendExpiry:
		return tx.Where("expired_time != ?", -1)
	case TokenBatchAddQuota:
		return tx.Where("unlimited_quota = ? AND remain_quota < ?", false, common.MaxQuota)
	case TokenBatchDeductExpiry:
		return tx.Where("expired_time != ? AND expired_time != ?", -1, now)
	case TokenBatchDeductQuota:
		return tx.Where("unlimited_quota = ? AND remain_quota > 0", false)
	default:
		return tx.Where("1 = 0")
	}
}

func PreviewTokenBatch(filter TokenBatchFilter, action TokenBatchAction, now int64) (*TokenBatchPreview, error) {
	base := applyTokenBatchFilter(DB.Model(&Token{}), filter)
	var aggregate tokenBatchPreviewAggregate
	err := base.Select(
		"COUNT(*) AS matched_tokens, "+
			"COUNT(DISTINCT user_id) AS affected_users, "+
			"COALESCE(SUM(CASE WHEN used_quota > 0 THEN 1 ELSE 0 END), 0) AS used_tokens, "+
			"COALESCE(SUM(CASE WHEN expired_time = ? THEN 1 ELSE 0 END), 0) AS permanent_tokens, "+
			"COALESCE(SUM(CASE WHEN unlimited_quota = ? THEN 1 ELSE 0 END), 0) AS unlimited_tokens, "+
			"COALESCE(SUM(remain_quota), 0) AS total_remaining_quota, "+
			"COALESCE(SUM(used_quota), 0) AS total_used_quota",
		-1,
		true,
	).Scan(&aggregate).Error
	if err != nil {
		return nil, err
	}

	var actionable int64
	err = applyTokenBatchActionFilter(applyTokenBatchFilter(DB.Model(&Token{}), filter), action, now).Count(&actionable).Error
	if err != nil {
		return nil, err
	}

	return &TokenBatchPreview{
		MatchedTokens:       aggregate.MatchedTokens,
		ActionableTokens:    actionable,
		AffectedUsers:       aggregate.AffectedUsers,
		UsedTokens:          aggregate.UsedTokens,
		PermanentTokens:     aggregate.PermanentTokens,
		UnlimitedTokens:     aggregate.UnlimitedTokens,
		TotalRemainingQuota: aggregate.TotalRemainingQuota,
		TotalUsedQuota:      aggregate.TotalUsedQuota,
	}, nil
}

func ExecuteTokenBatch(filter TokenBatchFilter, action TokenBatchAction, durationSeconds int64, quota int, now int64) (int64, error) {
	var affectedTokens []Token
	var rowsAffected int64

	err := DB.Transaction(func(tx *gorm.DB) error {
		selection := applyTokenBatchActionFilter(applyTokenBatchFilter(tx.Model(&Token{}), filter), action, now)
		if err := lockForUpdate(selection.Select("id", "key")).Find(&affectedTokens).Error; err != nil {
			return err
		}
		if len(affectedTokens) == 0 {
			return nil
		}

		updates := map[string]interface{}{}
		switch action {
		case TokenBatchExtendExpiry:
			const maxUnix = int64(^uint64(0) >> 1)
			updates["expired_time"] = gorm.Expr(
				"CASE WHEN expired_time < ? THEN ? WHEN expired_time > ? THEN ? ELSE expired_time + ? END",
				now, now+durationSeconds, maxUnix-durationSeconds, maxUnix, durationSeconds,
			)
		case TokenBatchAddQuota:
			updates["remain_quota"] = gorm.Expr(
				"CASE WHEN remain_quota > ? THEN ? ELSE remain_quota + ? END",
				common.MaxQuota-quota, common.MaxQuota, quota,
			)
		case TokenBatchDeductExpiry:
			updates["expired_time"] = gorm.Expr(
				"CASE WHEN expired_time <= ? THEN ? ELSE expired_time - ? END",
				now+durationSeconds, now, durationSeconds,
			)
		case TokenBatchDeductQuota:
			updates["remain_quota"] = gorm.Expr(
				"CASE WHEN remain_quota <= ? THEN 0 ELSE remain_quota - ? END",
				quota, quota,
			)
		default:
			return fmt.Errorf("unsupported token batch action %q", action)
		}

		result := applyTokenBatchActionFilter(applyTokenBatchFilter(tx.Model(&Token{}), filter), action, now).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		rowsAffected = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}

	if common.RedisEnabled && len(affectedTokens) > 0 {
		if err := invalidateTokensCache(affectedTokens); err != nil {
			common.SysError("failed to invalidate token caches after batch operation: " + err.Error())
		}
	}
	return rowsAffected, nil
}

func tokenBatchStatsBase(query TokenBatchStatisticsQuery) (*gorm.DB, error) {
	tx := LOG_DB.Model(&Log{}).
		Where("logs.type = ?", LogTypeConsume).
		Where("logs.created_at >= ? AND logs.created_at < ?", query.StartTime, query.EndTime)
	if !query.AllUsers {
		tx = tx.Where("logs.user_id = ?", query.UserID)
	}
	if query.UserFilterID > 0 {
		tx = tx.Where("logs.user_id = ?", query.UserFilterID)
	}
	if query.ExcludeUserID > 0 {
		tx = tx.Where("logs.user_id != ?", query.ExcludeUserID)
	}
	if query.Model != "" {
		model := strings.TrimSpace(query.Model)
		if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
			model = strings.ReplaceAll(model, `\`, `\\`)
			model = strings.ReplaceAll(model, `%`, `\%`)
			model = strings.ReplaceAll(model, `_`, `\_`)
			tx = tx.Where("logs.model_name LIKE ?", "%"+model+"%")
		} else {
			model = strings.ReplaceAll(model, "!", "!!")
			model = strings.ReplaceAll(model, "%", "!%")
			model = strings.ReplaceAll(model, "_", "!_")
			tx = tx.Where("logs.model_name LIKE ? ESCAPE '!'", "%"+model+"%")
		}
	}
	if query.MinTokens > 0 {
		tx = tx.Where("logs.prompt_tokens + logs.completion_tokens >= ?", query.MinTokens)
	}
	return tx, nil
}

func QueryTokenBatchStatistics(query TokenBatchStatisticsQuery) ([]TokenBatchStatisticsRow, TokenBatchStatisticsTotals, error) {
	groupFields := map[string]string{
		"token_name":   "logs.token_name",
		"model_name":   "logs.model_name",
		"username":     "logs.username",
		"channel_name": "logs.channel_id",
		"user_id":      "logs.user_id",
	}
	sortFields := map[string]string{
		"request_count":     "request_count",
		"prompt_tokens":     "prompt_tokens",
		"completion_tokens": "completion_tokens",
		"quota":             "quota",
	}
	groupField, ok := groupFields[query.GroupBy]
	if !ok {
		return nil, TokenBatchStatisticsTotals{}, fmt.Errorf("unsupported statistics group %q", query.GroupBy)
	}
	sortField, ok := sortFields[query.SortBy]
	if !ok {
		return nil, TokenBatchStatisticsTotals{}, fmt.Errorf("unsupported statistics sort %q", query.SortBy)
	}
	order := "DESC"
	if strings.EqualFold(query.SortOrder, "asc") {
		order = "ASC"
	}

	base, err := tokenBatchStatsBase(query)
	if err != nil {
		return nil, TokenBatchStatisticsTotals{}, err
	}
	if query.GroupBy == "channel_name" || query.GroupBy == "user_id" {
		base = base.Where(groupField + " > 0")
	} else {
		base = base.Where(groupField + " IS NOT NULL AND " + groupField + " != ''")
	}

	var rows []TokenBatchStatisticsRow
	err = base.Select(
		groupField + " AS name, " +
			"COUNT(*) AS request_count, " +
			"COALESCE(SUM(logs.prompt_tokens), 0) AS prompt_tokens, " +
			"COALESCE(SUM(logs.completion_tokens), 0) AS completion_tokens, " +
			"COALESCE(SUM(logs.quota), 0) AS quota, " +
			"COUNT(DISTINCT logs.user_id) AS unique_users",
	).Group(groupField).Order(sortField + " " + order).Limit(query.Top).Scan(&rows).Error
	if err != nil {
		return nil, TokenBatchStatisticsTotals{}, err
	}

	if query.GroupBy == "channel_name" && len(rows) > 0 {
		channelIDs := make([]int, 0, len(rows))
		for _, row := range rows {
			var channelID int
			if _, err := fmt.Sscan(row.Name, &channelID); err == nil && channelID > 0 {
				channelIDs = append(channelIDs, channelID)
			}
		}
		if len(channelIDs) > 0 {
			var channels []struct {
				ID   int    `gorm:"column:id"`
				Name string `gorm:"column:name"`
			}
			if err := DB.Table("channels").Select("id, name").Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
				return nil, TokenBatchStatisticsTotals{}, err
			}
			channelNames := make(map[string]string, len(channels))
			for _, channel := range channels {
				channelNames[fmt.Sprint(channel.ID)] = channel.Name
			}
			for i := range rows {
				if name := channelNames[rows[i].Name]; name != "" {
					rows[i].Name = name
				}
			}
		}
	}

	totalsBase, err := tokenBatchStatsBase(query)
	if err != nil {
		return nil, TokenBatchStatisticsTotals{}, err
	}
	var totals TokenBatchStatisticsTotals
	err = totalsBase.Select(
		"COUNT(*) AS request_count, " +
			"COALESCE(SUM(logs.prompt_tokens), 0) AS prompt_tokens, " +
			"COALESCE(SUM(logs.completion_tokens), 0) AS completion_tokens, " +
			"COALESCE(SUM(logs.quota), 0) AS quota, " +
			"COUNT(DISTINCT logs.user_id) AS unique_users",
	).Scan(&totals).Error
	return rows, totals, err
}
