package wtoken

import (
	"sync"
	"time"
)

var Store TokenStore

func init() {
	Store = NewTokenStore()
	StartTokenCleanup()
}

type TokenInfo struct {
	Token      string
	Expiration time.Time
}

type tokenStore struct {
	store sync.Map
}

type TokenStore interface {
	SaveToken(userName, token string, duration time.Duration)
	RetrieveToken(userName string) (string, bool)
	cleanExpiredTokens()
}

func NewTokenStore() TokenStore {
	return &tokenStore{}
}

func StartTokenCleanup() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				Store.cleanExpiredTokens()
			}
		}
	}()
}

func (ts *tokenStore) SaveToken(userName, token string, duration time.Duration) {
	ts.store.Store(userName, &TokenInfo{
		Token:      token,
		Expiration: time.Now().Add(duration),
	})
}

func (ts *tokenStore) RetrieveToken(userName string) (string, bool) {
	val, ok := ts.store.Load(userName)
	if !ok {
		return "", false
	}
	info := val.(*TokenInfo)
	if time.Now().After(info.Expiration) {
		ts.store.Delete(userName)
		return "", false
	}
	return info.Token, true
}

func (ts *tokenStore) cleanExpiredTokens() {
	now := time.Now()
	ts.store.Range(func(key, value interface{}) bool {
		info := value.(*TokenInfo)
		if now.After(info.Expiration) {
			ts.store.Delete(key)
		}
		return true
	})
}
