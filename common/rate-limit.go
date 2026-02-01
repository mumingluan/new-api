package common

import (
	"sync"
	"time"
)

type InMemoryRateLimiter struct {
	store              map[string]*[]int64
	mutex              sync.Mutex
	expirationDuration time.Duration
}

func (l *InMemoryRateLimiter) Init(expirationDuration time.Duration) {
	if l.store == nil {
		l.mutex.Lock()
		if l.store == nil {
			l.store = make(map[string]*[]int64)
			l.expirationDuration = expirationDuration
			if expirationDuration > 0 {
				go l.clearExpiredItems()
			}
		}
		l.mutex.Unlock()
	}
}

func (l *InMemoryRateLimiter) clearExpiredItems() {
	for {
		time.Sleep(l.expirationDuration)
		l.mutex.Lock()
		now := time.Now().Unix()
		for key := range l.store {
			queue := l.store[key]
			size := len(*queue)
			if size == 0 || now-(*queue)[size-1] > int64(l.expirationDuration.Seconds()) {
				delete(l.store, key)
			}
		}
		l.mutex.Unlock()
	}
}

// Request parameter duration's unit is seconds
func (l *InMemoryRateLimiter) Request(key string, maxRequestNum int, duration int64) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// [old <-- new]
	queue, ok := l.store[key]
	now := time.Now().Unix()
	if ok {
		if len(*queue) < maxRequestNum {
			*queue = append(*queue, now)
			return true
		} else {
			if now-(*queue)[0] >= duration {
				*queue = (*queue)[1:]
				*queue = append(*queue, now)
				return true
			} else {
				return false
			}
		}
	} else {
		s := make([]int64, 0, maxRequestNum)
		l.store[key] = &s
		*(l.store[key]) = append(*(l.store[key]), now)
	}
	return true
}

// Check 只检查是否达到限制，不记录请求（用于成功请求数限制的预检查）
func (l *InMemoryRateLimiter) Check(key string, maxRequestNum int, duration int64) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	queue, ok := l.store[key]
	if !ok {
		return true // 没有记录，允许
	}
	now := time.Now().Unix()
	if len(*queue) < maxRequestNum {
		return true // 未达到限制
	}
	// 检查最老的记录是否已过期
	if now-(*queue)[0] >= duration {
		return true // 时间窗口已过，允许
	}
	return false // 在时间窗口内已达到限制
}
