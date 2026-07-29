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
		expiresAt,
		[]string{"TEST0001", fmt.Sprintf("%d_TEST0002", user.Id)},
	)
	require.NoError(t, err)
	require.Len(t, codes, 2)
	assert.Equal(t, fmt.Sprintf("%d_TEST0001", user.Id), codes[0].Code)
	assert.Equal(t, fmt.Sprintf("%d_TEST0002", user.Id), codes[1].Code)

	_, err = CreateUserActivationCodes(
		user.Id,
		1,
		30,
		"test",
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
			ExpiredTime: now + 86400,
			Status:      ActivationCodeStatusActive,
			CreatedTime: now,
		},
		{
			UserId:      user.Id,
			Code:        "LEGACY-CODE",
			Days:        7,
			Channel:     "legacy",
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
	assert.Equal(t, "普通用户", tokens[0].Group)

	var storedCode ActivationCode
	require.NoError(t, DB.First(&storedCode, code.Id).Error)
	assert.Equal(t, ActivationCodeStatusUsed, storedCode.Status)

	var logs []ActivationLog
	require.NoError(t, DB.Where("activation_code_id = ?", code.Id).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, user.Id, logs[0].UserId)
	assert.Equal(t, tokens[0].Id, logs[0].TokenId)
}
