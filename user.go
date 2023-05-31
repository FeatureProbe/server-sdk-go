package featureprobe

import (
	"strconv"
	"sync"
	"time"
)

type FPUser struct {
	mu    *sync.RWMutex
	key   string
	attrs map[string]string
}

func NewUser() FPUser {
	return FPUser{
		mu:    &sync.RWMutex{},
		attrs: map[string]string{},
	}
}

func (u FPUser) StableRollout(key string) FPUser {
	u.key = key
	return u
}

func (u FPUser) Key() string {
	if len(u.key) == 0 {
		u.key = u.generateKey()
	}
	return u.key
}

func (u FPUser) generateKey() string {
	current := time.Now().UnixNano()
	return strconv.FormatInt(current, 10)
}

func (u FPUser) With(key string, value string) FPUser {
	u.mu.Lock()
	u.attrs[key] = value
	u.mu.Unlock()

	return u
}

func (u FPUser) GetAll() map[string]string {
	u.mu.RLock()
	snapshot := make(map[string]string, len(u.attrs))
	for k, v := range u.attrs {
		snapshot[k] = v
	}
	u.mu.RUnlock()

	return snapshot
}

func (u FPUser) Get(key string) string {
	u.mu.RLock()
	v := u.attrs[key]
	u.mu.RUnlock()

	return v
}

func (u FPUser) ContainAttr(key string) bool {
	u.mu.RLock()
	_, ok := u.attrs[key]
	u.mu.RUnlock()

	return ok
}

func (u FPUser) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"key":   u.Key(),
		"attrs": u.GetAll(),
	}
}
