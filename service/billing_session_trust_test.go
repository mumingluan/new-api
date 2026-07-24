package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBillingSessionDoesNotTrustChargeAboveTrustThreshold(t *testing.T) {
	trustQuota := common.GetTrustQuota()
	info := &relaycommon.RelayInfo{
		UserId:         1,
		UserQuota:      trustQuota * 10,
		TokenUnlimited: true,
	}
	session := &BillingSession{
		relayInfo: info,
		funding:   &WalletFunding{userId: info.UserId},
	}

	ctx, _ := gin.CreateTestContext(nil)

	assert.True(t, session.shouldTrust(ctx, trustQuota))
	assert.False(t, session.shouldTrust(ctx, trustQuota+1))
}
