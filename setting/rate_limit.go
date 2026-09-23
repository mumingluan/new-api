package setting

import (
	"fmt"
	"math"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// maxRateLimitDurationSeconds is the largest window the count cap is computed
// against (24h). Token-bucket capacity is count*duration; this keeps that
// product inside int64 when the window is at most a day.
const maxRateLimitDurationSeconds = 24 * 60 * 60

// maxModelRequestRateLimitCount is math.MaxInt64 / maxRateLimitDurationSeconds.
// It is the largest count that cannot overflow int64(count)*duration for a
// window of at most 24 hours.
const maxModelRequestRateLimitCount int64 = math.MaxInt64 / maxRateLimitDurationSeconds

var ModelRequestRateLimitEnabled = false

// Per-key minute rate limit settings.
var TokenRateLimitEnabled = false
var TokenRateLimitDurationMinutes = 1
var TokenRateLimitCount = 0
var TokenRateLimitSuccessCount = 0
var TokenRateLimitGroup = map[string][2]int{}
var TokenRateLimitMutex sync.RWMutex

// Per-key daily rate limit settings.
var TokenDailyRateLimitEnabled = false
var TokenDailyRateLimitCount = 0
var TokenDailyRateLimitSuccessCount = 0
var TokenDailyRateLimitGroup = map[string][2]int{}
var TokenDailyRateLimitMutex sync.RWMutex
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := common.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return common.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := ModelRequestRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string][2]int)
	err := common.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkModelRequestRateLimitGroup {
		if limits[0] < 0 || limits[1] < 1 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if int64(limits[0]) > maxModelRequestRateLimitCount || int64(limits[1]) > maxModelRequestRateLimitCount {
			return fmt.Errorf("group %s [%d, %d] exceeds max rate limit %d", group, limits[0], limits[1], maxModelRequestRateLimitCount)
		}
	}

	return nil
}

func parseTokenRateLimitGroup(jsonStr string) (map[string][2]int, error) {
	values := make(map[string][2]int)
	if err := common.Unmarshal([]byte(jsonStr), &values); err != nil {
		return nil, err
	}
	for group, limits := range values {
		if limits[0] < 0 || limits[1] < 0 {
			return nil, fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if int64(limits[0]) > maxModelRequestRateLimitCount || int64(limits[1]) > maxModelRequestRateLimitCount {
			return nil, fmt.Errorf("group %s [%d, %d] exceeds max rate limit %d", group, limits[0], limits[1], maxModelRequestRateLimitCount)
		}
	}
	return values, nil
}

func updateTokenRateLimitGroup(target *map[string][2]int, mutex *sync.RWMutex, jsonStr string) error {
	values, err := parseTokenRateLimitGroup(jsonStr)
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	*target = values
	return nil
}

// TokenRateLimitGroup2JSONString returns minute-level per-key group overrides.
func TokenRateLimitGroup2JSONString() string {
	TokenRateLimitMutex.RLock()
	defer TokenRateLimitMutex.RUnlock()
	jsonBytes, err := common.Marshal(TokenRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling token rate limit group: " + err.Error())
	}
	return string(jsonBytes)
}

// UpdateTokenRateLimitGroupByJSONString replaces minute-level per-key overrides.
func UpdateTokenRateLimitGroupByJSONString(jsonStr string) error {
	return updateTokenRateLimitGroup(&TokenRateLimitGroup, &TokenRateLimitMutex, jsonStr)
}

// GetTokenRateLimit returns minute-level overrides for a group.
func GetTokenRateLimit(group string) (int, int, bool) {
	TokenRateLimitMutex.RLock()
	defer TokenRateLimitMutex.RUnlock()
	value, ok := TokenRateLimitGroup[group]
	return value[0], value[1], ok
}

// CheckTokenRateLimitGroup validates minute-level per-key overrides.
func CheckTokenRateLimitGroup(jsonStr string) error {
	_, err := parseTokenRateLimitGroup(jsonStr)
	return err
}

// TokenDailyRateLimitGroup2JSONString returns daily per-key group overrides.
func TokenDailyRateLimitGroup2JSONString() string {
	TokenDailyRateLimitMutex.RLock()
	defer TokenDailyRateLimitMutex.RUnlock()
	jsonBytes, err := common.Marshal(TokenDailyRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling token daily rate limit group: " + err.Error())
	}
	return string(jsonBytes)
}

// UpdateTokenDailyRateLimitGroupByJSONString replaces daily per-key overrides.
func UpdateTokenDailyRateLimitGroupByJSONString(jsonStr string) error {
	return updateTokenRateLimitGroup(&TokenDailyRateLimitGroup, &TokenDailyRateLimitMutex, jsonStr)
}

// GetTokenDailyRateLimit returns daily overrides for a group.
func GetTokenDailyRateLimit(group string) (int, int, bool) {
	TokenDailyRateLimitMutex.RLock()
	defer TokenDailyRateLimitMutex.RUnlock()
	value, ok := TokenDailyRateLimitGroup[group]
	return value[0], value[1], ok
}

// CheckTokenDailyRateLimitGroup validates daily per-key overrides.
func CheckTokenDailyRateLimitGroup(jsonStr string) error {
	_, err := parseTokenRateLimitGroup(jsonStr)
	return err
}
