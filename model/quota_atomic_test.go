package model

import (
	"errors"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecreaseUserQuotaAtomicallyRejectsConcurrentOverdraft(t *testing.T) {
	truncateTables(t)

	user := &User{
		Id:       901,
		Username: "atomic-user-quota",
		Status:   common.UserStatusEnabled,
		Quota:    100,
	}
	require.NoError(t, DB.Create(user).Error)

	oldBatchUpdateEnabled := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = true
	t.Cleanup(func() {
		common.BatchUpdateEnabled = oldBatchUpdateEnabled
	})

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- DecreaseUserQuota(user.Id, 75, false)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	var successCount int
	var insufficientCount int
	for err := range errs {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, ErrUserQuotaInsufficient):
			insufficientCount++
		default:
			require.NoError(t, err)
		}
	}

	var stored User
	require.NoError(t, DB.Select("quota").First(&stored, user.Id).Error)
	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, insufficientCount)
	assert.Equal(t, 25, stored.Quota)
}

func TestDecreaseTokenQuotaAtomicallyRejectsConcurrentOverdraft(t *testing.T) {
	truncateTables(t)

	token := &Token{
		Id:          902,
		UserId:      901,
		Key:         "atomic-limited-token",
		Name:        "atomic limited token",
		Status:      common.TokenStatusEnabled,
		RemainQuota: 100,
	}
	require.NoError(t, DB.Create(token).Error)

	oldBatchUpdateEnabled := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = true
	t.Cleanup(func() {
		common.BatchUpdateEnabled = oldBatchUpdateEnabled
	})

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- DecreaseTokenQuota(token.Id, token.Key, 75)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	var successCount int
	var insufficientCount int
	for err := range errs {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, ErrTokenQuotaInsufficient):
			insufficientCount++
		default:
			require.NoError(t, err)
		}
	}

	var stored Token
	require.NoError(t, DB.Select("remain_quota", "used_quota").First(&stored, token.Id).Error)
	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, insufficientCount)
	assert.Equal(t, 25, stored.RemainQuota)
	assert.Equal(t, 75, stored.UsedQuota)
}

func TestDecreaseUnlimitedTokenQuotaTracksUsageWithoutNegativeBalance(t *testing.T) {
	truncateTables(t)

	token := &Token{
		Id:             903,
		UserId:         901,
		Key:            "atomic-unlimited-token",
		Name:           "atomic unlimited token",
		Status:         common.TokenStatusEnabled,
		RemainQuota:    0,
		UnlimitedQuota: true,
	}
	require.NoError(t, DB.Create(token).Error)

	require.NoError(t, DecreaseTokenQuota(token.Id, token.Key, 75))

	var stored Token
	require.NoError(t, DB.Select("remain_quota", "used_quota").First(&stored, token.Id).Error)
	assert.Zero(t, stored.RemainQuota)
	assert.Equal(t, 75, stored.UsedQuota)

	require.NoError(t, IncreaseTokenQuota(token.Id, token.Key, 75))
	require.NoError(t, DB.Select("remain_quota", "used_quota").First(&stored, token.Id).Error)
	assert.Zero(t, stored.RemainQuota)
	assert.Zero(t, stored.UsedQuota)
}
