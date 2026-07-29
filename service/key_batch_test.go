package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveKeyBatchFilterDefaultsAdminsToCurrentUser(t *testing.T) {
	filter, err := resolveKeyBatchFilter(42, common.RoleAdminUser, KeyBatchFilter{})
	require.NoError(t, err)
	assert.Equal(t, 42, filter.UserID)
	assert.False(t, filter.AllUsers)
}

func TestResolveKeyBatchFilterRejectsAllUsersForCommonUser(t *testing.T) {
	_, err := resolveKeyBatchFilter(42, common.RoleCommonUser, KeyBatchFilter{AllUsers: true})
	assert.ErrorIs(t, err, ErrKeyBatchAllUsersForbidden)
}

func TestValidateKeyBatchOperationBounds(t *testing.T) {
	assert.NoError(t, validateKeyBatchOperation(KeyBatchOperationRequest{Action: model.TokenBatchAddQuota, Quota: 1}))
	assert.Error(t, validateKeyBatchOperation(KeyBatchOperationRequest{Action: model.TokenBatchAddQuota, Quota: 0}))
	assert.NoError(t, validateKeyBatchOperation(KeyBatchOperationRequest{Action: model.TokenBatchExtendExpiry, DurationSeconds: 60}))
	assert.Error(t, validateKeyBatchOperation(KeyBatchOperationRequest{Action: model.TokenBatchExtendExpiry, DurationSeconds: MaxKeyBatchDurationSeconds + 1}))
}

func TestBuildKeyBatchStatisticsQueryProtectsUserScope(t *testing.T) {
	base := KeyBatchStatisticsRequest{
		StartTime: 1, EndTime: 2, GroupBy: "token_name", SortBy: "request_count", SortOrder: "desc", Top: 50,
	}
	base.UserFilterID = 99
	_, err := BuildKeyBatchStatisticsQuery(42, common.RoleAdminUser, base)
	assert.True(t, errors.Is(err, ErrKeyBatchAllUsersForbidden))

	base.AllUsers = true
	query, err := BuildKeyBatchStatisticsQuery(42, common.RoleAdminUser, base)
	require.NoError(t, err)
	assert.True(t, query.AllUsers)
	assert.Equal(t, 99, query.UserFilterID)
}
