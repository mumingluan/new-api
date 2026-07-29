package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupKeyBatchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB, previousLogDB := DB, LOG_DB
	previousRedis := common.RedisEnabled
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Token{}, &Log{}, &Channel{}))
	DB, LOG_DB = db, db
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.RedisEnabled = previousRedis
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func TestTokenBatchRespectsUserScopeAndTokenKinds(t *testing.T) {
	db := setupKeyBatchTestDB(t)
	now := int64(1_800_000_000)
	tokens := []Token{
		{UserId: 1, Key: "user1-normal", Name: "normal", ExpiredTime: now - 100, RemainQuota: 100, UsedQuota: 50},
		{UserId: 1, Key: "user1-permanent", Name: "permanent", ExpiredTime: -1, RemainQuota: 200},
		{UserId: 1, Key: "user1-unlimited", Name: "unlimited", ExpiredTime: now + 100, RemainQuota: 999, UsedQuota: 100, UnlimitedQuota: true},
		{UserId: 2, Key: "user2-normal", Name: "other", ExpiredTime: now + 100, RemainQuota: 300, UsedQuota: 10},
	}
	require.NoError(t, db.Create(&tokens).Error)

	filter := TokenBatchFilter{UserID: 1}
	preview, err := PreviewTokenBatch(filter, TokenBatchExtendExpiry, now)
	require.NoError(t, err)
	assert.Equal(t, int64(3), preview.MatchedTokens)
	assert.Equal(t, int64(2), preview.ActionableTokens)
	assert.Equal(t, int64(1), preview.AffectedUsers)
	assert.Equal(t, int64(2), preview.UsedTokens)
	assert.Equal(t, int64(1), preview.PermanentTokens)
	assert.Equal(t, int64(1), preview.UnlimitedTokens)

	affected, err := ExecuteTokenBatch(filter, TokenBatchAddQuota, 0, 50, now)
	require.NoError(t, err)
	assert.Equal(t, int64(2), affected)

	var got []Token
	require.NoError(t, db.Order("id").Find(&got).Error)
	assert.Equal(t, 150, got[0].RemainQuota)
	assert.Equal(t, 250, got[1].RemainQuota)
	assert.Equal(t, 999, got[2].RemainQuota)
	assert.Equal(t, 300, got[3].RemainQuota)
}

func TestTokenBatchFiltersAndDeductionFloors(t *testing.T) {
	db := setupKeyBatchTestDB(t)
	now := int64(1_800_000_000)
	minQuota := 100
	tokens := []Token{
		{UserId: 1, Key: "used-high", Name: "used-high", ExpiredTime: now + 30, RemainQuota: 150, UsedQuota: 10},
		{UserId: 1, Key: "unused-high", Name: "unused-high", ExpiredTime: now + 30, RemainQuota: 150},
		{UserId: 1, Key: "used-low", Name: "used-low", ExpiredTime: now + 30, RemainQuota: 50, UsedQuota: 10},
	}
	require.NoError(t, db.Create(&tokens).Error)
	filter := TokenBatchFilter{UserID: 1, UsedOnly: true, MinRemainingQuota: &minQuota}

	affected, err := ExecuteTokenBatch(filter, TokenBatchDeductQuota, 0, 200, now)
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	var got []Token
	require.NoError(t, db.Order("id").Find(&got).Error)
	assert.Zero(t, got[0].RemainQuota)
	assert.Equal(t, 150, got[1].RemainQuota)
	assert.Equal(t, 50, got[2].RemainQuota)

	filter.MinRemainingQuota = nil
	affected, err = ExecuteTokenBatch(filter, TokenBatchDeductExpiry, 60, 0, now)
	require.NoError(t, err)
	assert.Equal(t, int64(2), affected)
	require.NoError(t, db.Order("id").Find(&got).Error)
	assert.Equal(t, now, got[0].ExpiredTime)
	assert.Equal(t, now+30, got[1].ExpiredTime)
	assert.Equal(t, now, got[2].ExpiredTime)
}

func TestTokenBatchStatisticsIncludesZeroSidedUsageAndKeepsFullTotals(t *testing.T) {
	db := setupKeyBatchTestDB(t)
	channels := []Channel{{Name: "Primary"}, {Name: "Backup"}}
	require.NoError(t, db.Create(&channels).Error)
	logs := []Log{
		{UserId: 1, Username: "alice", Type: LogTypeConsume, CreatedAt: 100, TokenName: "alpha", ModelName: "gpt-a", PromptTokens: 10, CompletionTokens: 0, Quota: 100, ChannelId: channels[0].Id},
		{UserId: 1, Username: "alice", Type: LogTypeConsume, CreatedAt: 110, TokenName: "alpha", ModelName: "gpt-b", PromptTokens: 0, CompletionTokens: 20, Quota: 200, ChannelId: channels[1].Id},
		{UserId: 2, Username: "bob", Type: LogTypeConsume, CreatedAt: 120, TokenName: "beta", ModelName: "gpt-b", PromptTokens: 30, CompletionTokens: 40, Quota: 300, ChannelId: channels[1].Id},
		{UserId: 1, Username: "alice", Type: LogTypeManage, CreatedAt: 130, TokenName: "ignored", ModelName: "gpt-a", Quota: 999},
	}
	require.NoError(t, db.Create(&logs).Error)

	rows, totals, err := QueryTokenBatchStatistics(TokenBatchStatisticsQuery{
		UserID: 1, StartTime: 90, EndTime: 200, GroupBy: "model_name", SortBy: "quota", SortOrder: "desc", Top: 1,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "gpt-b", rows[0].Name)
	assert.Equal(t, int64(2), totals.RequestCount)
	assert.Equal(t, int64(10), totals.PromptTokens)
	assert.Equal(t, int64(20), totals.CompletionTokens)
	assert.Equal(t, int64(300), totals.Quota)
	assert.Equal(t, int64(1), totals.UniqueUsers)

	rows, totals, err = QueryTokenBatchStatistics(TokenBatchStatisticsQuery{
		AllUsers: true, StartTime: 90, EndTime: 200, GroupBy: "channel_name", SortBy: "request_count", SortOrder: "desc", Top: 10,
	})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "Backup", rows[0].Name)
	assert.Equal(t, int64(3), totals.RequestCount)
	assert.Equal(t, int64(2), totals.UniqueUsers)
}
