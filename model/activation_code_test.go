package model

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupActivationCodeFixture(t *testing.T) *User {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&ActivationCode{}, &ActivationLog{}, &Token{}, &User{}))
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ActivationLog{}).Error)
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ActivationCode{}).Error)
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Token{}).Error)
	require.NoError(t, DB.Exec("DELETE FROM users").Error)
	t.Cleanup(func() {
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ActivationLog{}).Error)
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ActivationCode{}).Error)
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Token{}).Error)
		require.NoError(t, DB.Exec("DELETE FROM users").Error)
	})

	user := &User{
		Username: "activation-owner",
		Password: "password",
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(user).Error)
	return user
}

func TestCreateActivationCodesEnforcesOwnerPrefix(t *testing.T) {
	user := setupActivationCodeFixture(t)
	expiresAt := common.GetTimestamp() + 86400

	codes, err := CreateUserActivationCodes(
		user.Id,
		2,
		30,
		"test",
		"activation-group",
		expiresAt,
		[]string{"TEST0001", fmt.Sprintf("%d_TEST0002", user.Id)},
	)
	require.NoError(t, err)
	require.Len(t, codes, 2)
	assert.Equal(t, fmt.Sprintf("%d_TEST0001", user.Id), codes[0].Code)
	assert.Equal(t, fmt.Sprintf("%d_TEST0002", user.Id), codes[1].Code)
	assert.Equal(t, "activation-group", codes[0].Group)

	_, err = CreateUserActivationCodes(
		user.Id,
		1,
		30,
		"test",
		"activation-group",
		expiresAt,
		[]string{"999_TEST0003"},
	)
	require.ErrorContains(t, err, "前缀必须")
}

func TestCreateActivationCodesGeneratesSixteenCharacterSuffix(t *testing.T) {
	user := setupActivationCodeFixture(t)

	codes, err := CreateUserActivationCodes(
		user.Id,
		1,
		30,
		"test",
		"activation-group",
		common.GetTimestamp()+86400,
		nil,
	)
	require.NoError(t, err)
	require.Len(t, codes, 1)

	prefix := fmt.Sprintf("%d_", user.Id)
	require.True(t, strings.HasPrefix(codes[0].Code, prefix))
	assert.Len(t, strings.TrimPrefix(codes[0].Code, prefix), 16)
}

func TestActivationLookupUsesPrefixedOwnerAndFallsBackForLegacyCode(t *testing.T) {
	user := setupActivationCodeFixture(t)
	now := common.GetTimestamp()
	codes := []ActivationCode{
		{
			UserId:      user.Id,
			Code:        fmt.Sprintf("%d_STANDARD", user.Id),
			Days:        30,
			Channel:     "standard",
			Group:       "default",
			ExpiredTime: now + 86400,
			Status:      ActivationCodeStatusActive,
			CreatedTime: now,
		},
		{
			UserId:      user.Id,
			Code:        "LEGACY-CODE",
			Days:        7,
			Channel:     "legacy",
			Group:       "default",
			ExpiredTime: now + 86400,
			Status:      ActivationCodeStatusActive,
			CreatedTime: now,
		},
	}
	require.NoError(t, DB.Create(&codes).Error)

	standard, err := GetActivationPrecheck(codes[0].Code, "10001", "", now)
	require.NoError(t, err)
	assert.Equal(t, user.Id, standard.UserId)
	assert.Equal(t, "standard", standard.Channel)

	legacy, err := GetActivationPrecheck("LEGACY-CODE", "10002", "", now)
	require.NoError(t, err)
	assert.Equal(t, user.Id, legacy.UserId)
	assert.Equal(t, "legacy", legacy.Channel)
}

func TestRedeemActivationCodeCreatesOwnerTokenExactlyOnce(t *testing.T) {
	user := setupActivationCodeFixture(t)
	now := common.GetTimestamp()
	code := ActivationCode{
		UserId:      user.Id,
		Code:        fmt.Sprintf("%d_CONCURRENT", user.Id),
		Days:        30,
		Channel:     "general",
		Group:       "activation-group",
		ExpiredTime: now + 86400,
		Status:      ActivationCodeStatusActive,
		CreatedTime: now,
	}
	require.NoError(t, DB.Create(&code).Error)

	const workers = 5
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			if _, err := RedeemActivationCode(code.Code, "10000", "127.0.0.1", now); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, successes)

	var tokens []Token
	require.NoError(t, DB.Where("user_id = ?", user.Id).Find(&tokens).Error)
	require.Len(t, tokens, 1)
	assert.Equal(t, "10000 general", tokens[0].Name)
	assert.Equal(t, "activation-group", tokens[0].Group)

	var storedCode ActivationCode
	require.NoError(t, DB.First(&storedCode, code.Id).Error)
	assert.Equal(t, ActivationCodeStatusUsed, storedCode.Status)

	var logs []ActivationLog
	require.NoError(t, DB.Where("activation_code_id = ?", code.Id).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, user.Id, logs[0].UserId)
	assert.Equal(t, tokens[0].Id, logs[0].TokenId)
}

func TestRenewActivationCodeRejectsDifferentTokenGroupWithoutConsumingCode(t *testing.T) {
	user := setupActivationCodeFixture(t)
	now := common.GetTimestamp()
	code := ActivationCode{
		UserId:      user.Id,
		Code:        fmt.Sprintf("%d_RENEW_MISMATCH", user.Id),
		Days:        30,
		Channel:     "general",
		Group:       "vip",
		ExpiredTime: now + 86400,
		Status:      ActivationCodeStatusActive,
		CreatedTime: now,
	}
	token := Token{
		UserId:      user.Id,
		Name:        "existing",
		Key:         "renew-mismatch-key",
		Status:      common.TokenStatusEnabled,
		CreatedTime: now,
		ExpiredTime: now + 3600,
		Group:       "default",
	}
	require.NoError(t, DB.Create(&code).Error)
	require.NoError(t, DB.Create(&token).Error)

	_, err := GetActivationPrecheck(code.Code, "", "sk-"+token.Key, now)
	require.ErrorIs(t, err, ErrActivationGroupMismatch)
	_, err = RenewActivationCode(code.Code, "sk-"+token.Key, "127.0.0.1", now)
	require.ErrorIs(t, err, ErrActivationGroupMismatch)

	var storedCode ActivationCode
	require.NoError(t, DB.First(&storedCode, code.Id).Error)
	assert.Equal(t, ActivationCodeStatusActive, storedCode.Status)
	var storedToken Token
	require.NoError(t, DB.First(&storedToken, token.Id).Error)
	assert.Equal(t, token.ExpiredTime, storedToken.ExpiredTime)
}

func TestRenewActivationCodeAcceptsMatchingTokenGroup(t *testing.T) {
	user := setupActivationCodeFixture(t)
	now := common.GetTimestamp()
	code := ActivationCode{
		UserId:      user.Id,
		Code:        fmt.Sprintf("%d_RENEW_MATCH", user.Id),
		Days:        30,
		Channel:     "general",
		Group:       "vip",
		ExpiredTime: now + 86400,
		Status:      ActivationCodeStatusActive,
		CreatedTime: now,
	}
	token := Token{
		UserId:      user.Id,
		Name:        "existing",
		Key:         "renew-match-key",
		Status:      common.TokenStatusEnabled,
		CreatedTime: now,
		ExpiredTime: now + 3600,
		Group:       "vip",
	}
	require.NoError(t, DB.Create(&code).Error)
	require.NoError(t, DB.Create(&token).Error)

	result, err := RenewActivationCode(code.Code, "sk-"+token.Key, "127.0.0.1", now)
	require.NoError(t, err)
	assert.Greater(t, result.NewExpiredTime, token.ExpiredTime)
}

func TestMigrateActivationCodeGroupsBackfillsLegacyRows(t *testing.T) {
	user := setupActivationCodeFixture(t)
	now := common.GetTimestamp()
	code := ActivationCode{
		UserId:      user.Id,
		Code:        fmt.Sprintf("%d_LEGACY_GROUP", user.Id),
		Days:        30,
		Channel:     "general",
		ExpiredTime: now + 86400,
		Status:      ActivationCodeStatusActive,
		CreatedTime: now,
	}
	require.NoError(t, DB.Create(&code).Error)
	require.NoError(t, migrateActivationCodeGroups())

	var storedCode ActivationCode
	require.NoError(t, DB.First(&storedCode, code.Id).Error)
	assert.Equal(t, legacyActivationCodeGroup, storedCode.Group)
}
