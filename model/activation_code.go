package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	ActivationCodeStatusActive   = 1
	ActivationCodeStatusUsed     = 2
	ActivationCodeStatusDisabled = 3
)

var (
	ErrActivationCodeInvalid       = errors.New("激活码无效或已被使用")
	ErrActivationCodeExpired       = errors.New("激活码已过期")
	ErrActivationCodeOwnerMismatch = errors.New("激活码与令牌不属于同一用户")
	ErrActivationTokenExists       = errors.New("当前 QQ 号已存在密钥")
	ErrActivationTokenNotFound     = errors.New("密钥未找到")
	ErrActivationOwnerUnavailable  = errors.New("激活码所属用户不可用")
	ErrActivationRequestInvalid    = errors.New("激活请求参数无效")
)

type ActivationCode struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id" gorm:"index;not null"`
	Code        string `json:"code" gorm:"type:varchar(255);uniqueIndex;not null"`
	Days        int    `json:"days" gorm:"not null"`
	Channel     string `json:"channel" gorm:"type:varchar(100);index;not null"`
	ExpiredTime int64  `json:"expired_time" gorm:"bigint;index;not null"`
	Status      int    `json:"status" gorm:"index;not null"`
	CreatedTime int64  `json:"created_time" gorm:"bigint;index;not null"`
	UsedTime    int64  `json:"used_time" gorm:"bigint;index;not null"`
}

type ActivationLog struct {
	Id               int    `json:"id"`
	LegacySourceId   *int64 `json:"-" gorm:"uniqueIndex"`
	ActivationCodeId int    `json:"activation_code_id" gorm:"index;not null"`
	UserId           int    `json:"user_id" gorm:"index;not null"`
	ActivationCode   string `json:"activation_code" gorm:"type:varchar(255);index;not null"`
	Action           string `json:"action" gorm:"type:varchar(16);index;not null"`
	Days             int    `json:"days" gorm:"not null"`
	Identifier       string `json:"identifier" gorm:"type:varchar(255);index;not null"`
	TokenId          int    `json:"token_id" gorm:"index;not null"`
	TokenKey         string `json:"api_key" gorm:"type:varchar(128);not null"`
	ClientIp         string `json:"client_ip" gorm:"type:varchar(100);index;not null"`
	UsedTime         int64  `json:"used_time" gorm:"bigint;index;not null"`
}

type ActivationCodeFilters struct {
	Search      string
	Channel     string
	Status      int
	Days        int
	CreatedFrom int64
	CreatedTo   int64
	ExpiresFrom int64
	ExpiresTo   int64
}

type ActivationPrecheck struct {
	Code           string `json:"-"`
	UserId         int    `json:"-"`
	Channel        string `json:"channel"`
	Days           int    `json:"days"`
	OldExpiredTime int64  `json:"old_expired_time,omitempty"`
	NewExpiredTime int64  `json:"new_expired_time"`
}

type ActivationResult struct {
	ApiKey         string `json:"api_key,omitempty"`
	Channel        string `json:"channel,omitempty"`
	ExpiredTime    int64  `json:"expired_time,omitempty"`
	NewExpiredTime int64  `json:"new_expired_time,omitempty"`
}

func activationCodeOwnerPrefix(code string) (int, bool) {
	prefix, _, ok := strings.Cut(strings.TrimSpace(code), "_")
	if !ok || prefix == "" {
		return 0, false
	}
	userId, err := strconv.Atoi(prefix)
	return userId, err == nil && userId > 0
}

func findActivationCode(db *gorm.DB, code string, forUpdate bool) (*ActivationCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, gorm.ErrRecordNotFound
	}
	query := db
	if forUpdate {
		query = lockForUpdate(query)
	}
	var activationCode ActivationCode
	if userId, ok := activationCodeOwnerPrefix(code); ok {
		err := query.Where("user_id = ? AND code = ?", userId, code).First(&activationCode).Error
		if err == nil {
			return &activationCode, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	err := query.Where("code = ?", code).First(&activationCode).Error
	return &activationCode, err
}

func validateUsableActivationCode(code *ActivationCode, now int64) error {
	if code.Status != ActivationCodeStatusActive {
		return ErrActivationCodeInvalid
	}
	if code.ExpiredTime > 0 && code.ExpiredTime < now {
		return ErrActivationCodeExpired
	}
	return nil
}

func GetActivationPrecheck(code, qq, apiKey string, now int64) (*ActivationPrecheck, error) {
	activationCode, err := findActivationCode(DB, code, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivationCodeInvalid
		}
		return nil, err
	}
	if err := validateUsableActivationCode(activationCode, now); err != nil {
		return nil, err
	}

	result := &ActivationPrecheck{
		Code:    activationCode.Code,
		UserId:  activationCode.UserId,
		Channel: activationCode.Channel,
		Days:    activationCode.Days,
	}
	if strings.TrimSpace(qq) != "" {
		name := fmt.Sprintf("%s %s", strings.TrimSpace(qq), activationCode.Channel)
		var count int64
		if err := DB.Model(&Token{}).Where("user_id = ? AND name = ?", activationCode.UserId, name).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrActivationTokenExists
		}
		result.NewExpiredTime = time.Unix(now, 0).AddDate(0, 0, activationCode.Days).Add(2 * time.Hour).Unix()
		return result, nil
	}

	key := strings.TrimPrefix(strings.TrimSpace(apiKey), "sk-")
	if key == "" {
		return nil, fmt.Errorf("%w: 缺少 QQ 号或 API 密钥", ErrActivationRequestInvalid)
	}
	var token Token
	if err := DB.Where(commonKeyCol+" = ?", key).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivationTokenNotFound
		}
		return nil, err
	}
	if token.UserId != activationCode.UserId {
		return nil, ErrActivationCodeOwnerMismatch
	}
	result.OldExpiredTime = token.ExpiredTime
	base := now
	if token.ExpiredTime > now {
		base = token.ExpiredTime
	}
	result.NewExpiredTime = time.Unix(base, 0).AddDate(0, 0, activationCode.Days).Unix()
	return result, nil
}

func RedeemActivationCode(code, qq, clientIp string, now int64) (*ActivationResult, error) {
	key, err := common.GenerateKey()
	if err != nil {
		return nil, err
	}
	var result ActivationResult
	err = DB.Transaction(func(tx *gorm.DB) error {
		activationCode, err := findActivationCode(tx, code, true)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrActivationCodeInvalid
			}
			return err
		}
		if err := validateUsableActivationCode(activationCode, now); err != nil {
			return err
		}
		var owner User
		if err := tx.Select("id", "status").Where("id = ?", activationCode.UserId).First(&owner).Error; err != nil {
			return err
		}
		if owner.Status != common.UserStatusEnabled {
			return ErrActivationOwnerUnavailable
		}
		qq = strings.TrimSpace(qq)
		if qq == "" {
			return fmt.Errorf("%w: QQ 号不能为空", ErrActivationRequestInvalid)
		}
		name := fmt.Sprintf("%s %s", qq, activationCode.Channel)
		var count int64
		if err := tx.Model(&Token{}).Where("user_id = ? AND name = ?", owner.Id, name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrActivationTokenExists
		}
		claimed := tx.Model(&ActivationCode{}).
			Where("id = ? AND status = ?", activationCode.Id, ActivationCodeStatusActive).
			Updates(map[string]any{"status": ActivationCodeStatusUsed, "used_time": now})
		if claimed.Error != nil {
			return claimed.Error
		}
		if claimed.RowsAffected != 1 {
			return ErrActivationCodeInvalid
		}

		expiredTime := time.Unix(now, 0).AddDate(0, 0, activationCode.Days).Add(2 * time.Hour).Unix()
		token := Token{
			UserId:         owner.Id,
			Name:           name,
			Key:            key,
			Status:         common.TokenStatusEnabled,
			CreatedTime:    now,
			AccessedTime:   now,
			ExpiredTime:    expiredTime,
			UnlimitedQuota: true,
			Group:          "普通用户",
		}
		if err := tx.Create(&token).Error; err != nil {
			return err
		}
		logEntry := ActivationLog{
			ActivationCodeId: activationCode.Id,
			UserId:           owner.Id,
			ActivationCode:   activationCode.Code,
			Action:           "create",
			Days:             activationCode.Days,
			Identifier:       name,
			TokenId:          token.Id,
			TokenKey:         "sk-" + key,
			ClientIp:         clientIp,
			UsedTime:         now,
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			return err
		}
		result = ActivationResult{ApiKey: logEntry.TokenKey, Channel: activationCode.Channel, ExpiredTime: expiredTime}
		return nil
	})
	return &result, err
}

func RenewActivationCode(code, apiKey, clientIp string, now int64) (*ActivationResult, error) {
	key := strings.TrimPrefix(strings.TrimSpace(apiKey), "sk-")
	if key == "" {
		return nil, fmt.Errorf("%w: API 密钥不能为空", ErrActivationRequestInvalid)
	}
	var result ActivationResult
	err := DB.Transaction(func(tx *gorm.DB) error {
		activationCode, err := findActivationCode(tx, code, true)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrActivationCodeInvalid
			}
			return err
		}
		if err := validateUsableActivationCode(activationCode, now); err != nil {
			return err
		}
		var token Token
		if err := lockForUpdate(tx).Where(commonKeyCol+" = ?", key).First(&token).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrActivationTokenNotFound
			}
			return err
		}
		if token.UserId != activationCode.UserId {
			return ErrActivationCodeOwnerMismatch
		}
		claimed := tx.Model(&ActivationCode{}).
			Where("id = ? AND status = ?", activationCode.Id, ActivationCodeStatusActive).
			Updates(map[string]any{"status": ActivationCodeStatusUsed, "used_time": now})
		if claimed.Error != nil {
			return claimed.Error
		}
		if claimed.RowsAffected != 1 {
			return ErrActivationCodeInvalid
		}
		base := now
		if token.ExpiredTime > now {
			base = token.ExpiredTime
		}
		newExpiredTime := time.Unix(base, 0).AddDate(0, 0, activationCode.Days).Unix()
		if err := tx.Model(&token).Updates(map[string]any{
			"expired_time": newExpiredTime,
			"status":       common.TokenStatusEnabled,
		}).Error; err != nil {
			return err
		}
		logEntry := ActivationLog{
			ActivationCodeId: activationCode.Id,
			UserId:           activationCode.UserId,
			ActivationCode:   activationCode.Code,
			Action:           "renew",
			Days:             activationCode.Days,
			Identifier:       "sk-" + key,
			TokenId:          token.Id,
			TokenKey:         "sk-" + key,
			ClientIp:         clientIp,
			UsedTime:         now,
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			return err
		}
		result = ActivationResult{NewExpiredTime: newExpiredTime}
		return nil
	})
	if err == nil && common.RedisEnabled {
		_ = cacheDeleteToken(key)
	}
	return &result, err
}

func GetActivationLogByCode(code string) (*ActivationLog, error) {
	var logEntry ActivationLog
	err := DB.Where("activation_code = ?", strings.TrimSpace(code)).Order("used_time desc").First(&logEntry).Error
	return &logEntry, err
}

func activationCodeListQuery(userId int, filters ActivationCodeFilters) *gorm.DB {
	query := DB.Model(&ActivationCode{}).Where("user_id = ?", userId)
	if search := strings.TrimSpace(filters.Search); search != "" {
		search = strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(search)
		query = query.Where("code LIKE ? ESCAPE '!'", "%"+search+"%")
	}
	if filters.Channel != "" {
		query = query.Where("channel = ?", filters.Channel)
	}
	if filters.Status > 0 {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Days > 0 {
		query = query.Where("days = ?", filters.Days)
	}
	if filters.CreatedFrom > 0 {
		query = query.Where("created_time >= ?", filters.CreatedFrom)
	}
	if filters.CreatedTo > 0 {
		query = query.Where("created_time <= ?", filters.CreatedTo)
	}
	if filters.ExpiresFrom > 0 {
		query = query.Where("expired_time >= ?", filters.ExpiresFrom)
	}
	if filters.ExpiresTo > 0 {
		query = query.Where("expired_time <= ?", filters.ExpiresTo)
	}
	return query
}

func ListUserActivationCodes(userId, offset, limit int, filters ActivationCodeFilters) ([]ActivationCode, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	query := activationCodeListQuery(userId, filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var codes []ActivationCode
	err := query.Order("created_time desc").Offset(offset).Limit(limit).Find(&codes).Error
	return codes, total, err
}

func ListAllUserActivationCodes(userId int, filters ActivationCodeFilters) ([]ActivationCode, error) {
	var codes []ActivationCode
	err := activationCodeListQuery(userId, filters).Order("created_time desc").Find(&codes).Error
	return codes, err
}

func ListUserActivationLogs(userId, offset, limit int, search, action string) ([]ActivationLog, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := DB.Model(&ActivationLog{}).Where("user_id = ?", userId)
	if search = strings.TrimSpace(search); search != "" {
		search = strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(search)
		pattern := "%" + search + "%"
		query = query.Where("(activation_code LIKE ? ESCAPE '!' OR identifier LIKE ? ESCAPE '!')", pattern, pattern)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []ActivationLog
	err := query.Order("used_time desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func CreateUserActivationCodes(userId, count, days int, channel string, expiredTime int64, suffixes []string) ([]ActivationCode, error) {
	if userId <= 0 || count <= 0 || count > 1000 || days <= 0 || expiredTime <= common.GetTimestamp() {
		return nil, errors.New("激活码参数无效")
	}
	channel = strings.TrimSpace(channel)
	if channel == "" || len(channel) > 100 {
		return nil, errors.New("渠道名称无效")
	}
	if len(suffixes) > 0 && len(suffixes) != count {
		return nil, errors.New("自定义码数量与创建数量不一致")
	}
	now := common.GetTimestamp()
	codes := make([]ActivationCode, 0, count)
	for i := 0; i < count; i++ {
		suffix := ""
		if len(suffixes) > 0 {
			suffix = strings.TrimSpace(suffixes[i])
		} else {
			random, err := common.GenerateRandomCharsKey(16)
			if err != nil {
				return nil, err
			}
			suffix = strings.ToUpper(random)
		}
		if suffix == "" || len(suffix) > 200 || strings.ContainsAny(suffix, " \t\r\n,") {
			return nil, fmt.Errorf("激活码内容 %q 无效", suffix)
		}
		expectedPrefix := strconv.Itoa(userId) + "_"
		fullCode := expectedPrefix + suffix
		if strings.HasPrefix(suffix, expectedPrefix) {
			fullCode = suffix
		} else if _, hasPrefix := activationCodeOwnerPrefix(suffix); hasPrefix {
			return nil, fmt.Errorf("激活码前缀必须是 %d", userId)
		}
		codes = append(codes, ActivationCode{
			UserId:      userId,
			Code:        fullCode,
			Days:        days,
			Channel:     channel,
			ExpiredTime: expiredTime,
			Status:      ActivationCodeStatusActive,
			CreatedTime: now,
		})
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&codes).Error
	})
	return codes, err
}

func UpdateUserActivationCodes(userId int, ids []int, codes []string, updates map[string]any) (int64, error) {
	if len(ids) == 0 && len(codes) == 0 {
		return 0, errors.New("未选择激活码")
	}
	query := DB.Model(&ActivationCode{}).Where("user_id = ?", userId)
	query = query.Where("status <> ?", ActivationCodeStatusUsed)
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	} else {
		query = query.Where("code IN ?", codes)
	}
	result := query.Updates(updates)
	return result.RowsAffected, result.Error
}

func DeleteUserActivationCodes(userId int, ids []int, codes []string) (int64, error) {
	if len(ids) == 0 && len(codes) == 0 {
		return 0, errors.New("未选择激活码")
	}
	query := DB.Where("user_id = ? AND status <> ?", userId, ActivationCodeStatusUsed)
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	} else {
		query = query.Where("code IN ?", codes)
	}
	result := query.Delete(&ActivationCode{})
	return result.RowsAffected, result.Error
}
