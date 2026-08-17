package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenQuotaCanExceedBillingClamp(t *testing.T) {
	highQuota := int64(common.MaxQuota) + 1

	t.Run("create", func(t *testing.T) {
		user := setupTokenAutoGroupsControllerTest(t)
		request := map[string]any{
			"name":            "high-quota-create",
			"expired_time":    -1,
			"remain_quota":    highQuota,
			"unlimited_quota": false,
			"group":           "default",
		}

		ctx, recorder := newAuthenticatedContext(t, http.MethodPost, "/api/token/", request, user.Id)
		AddToken(ctx)

		response := decodeAPIResponse(t, recorder)
		require.True(t, response.Success, response.Message)
		var token model.Token
		require.NoError(t, model.DB.Where("name = ?", "high-quota-create").First(&token).Error)
		assert.Equal(t, highQuota, int64(token.RemainQuota))
	})

	t.Run("update", func(t *testing.T) {
		user := setupTokenAutoGroupsControllerTest(t)
		token := seedToken(t, model.DB, user.Id, "high-quota-update", "highquota12345678")
		request := map[string]any{
			"id":              token.Id,
			"name":            token.Name,
			"expired_time":    -1,
			"remain_quota":    highQuota,
			"unlimited_quota": false,
			"group":           "default",
		}

		ctx, recorder := newAuthenticatedContext(t, http.MethodPut, "/api/token/", request, user.Id)
		UpdateToken(ctx)

		response := decodeAPIResponse(t, recorder)
		require.True(t, response.Success, response.Message)
		require.NoError(t, model.DB.First(token, token.Id).Error)
		assert.Equal(t, highQuota, int64(token.RemainQuota))
	})
}
