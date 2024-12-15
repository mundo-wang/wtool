package wtoken

import (
	"sync"
	"time"
)

var Store *TokenStore

func init() {
	Store = NewTokenStore()
	StartTokenCleanup()
}

type TokenStore struct {
	store sync.Map // 读多写少的情况用sync.Map
}

type TokenInfo struct {
	Token      string
	Expiration time.Time
}

func NewTokenStore() *TokenStore {
	return &TokenStore{}
}

// StartTokenCleanup 启动定时清理过期token的定时任务
func StartTokenCleanup() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				Store.CleanExpiredTokens()
			}
		}
	}()
}

// StoreToken 存储token到 TokenStore
func (ts *TokenStore) StoreToken(userName, token string, duration time.Duration) {
	ts.store.Store(userName, &TokenInfo{
		Token:      token,
		Expiration: time.Now().Add(duration),
	})
}

// GetToken 检索给定userName的令牌
func (ts *TokenStore) GetToken(userName string) (string, bool) {
	val, ok := ts.store.Load(userName)
	if !ok {
		return "", false // 没有找到用户对应的token
	}
	info := val.(*TokenInfo)
	// 检查token是否过期
	if time.Now().After(info.Expiration) {
		// 如果过期则清除该token并返回false
		ts.store.Delete(userName)
		return "", false
	}
	return info.Token, true
}

// CleanExpiredTokens 从TokenStore中删除过期的令牌
func (ts *TokenStore) CleanExpiredTokens() {
	now := time.Now()
	ts.store.Range(func(key, value interface{}) bool {
		info := value.(*TokenInfo)
		if now.After(info.Expiration) {
			ts.store.Delete(key)
		}
		return true
	})
}
