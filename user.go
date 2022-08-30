package featureprobe

import (
	"strconv"
	"time"
)

type FPUser struct {
	key   string
	attrs map[string]string
}

func NewUser() FPUser {
	return FPUser{
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
	currnet := time.Now().UnixNano()
	return strconv.FormatInt(currnet, 10)
}

func (u FPUser) With(key string, value string) FPUser {
	u.attrs[key] = value
	return u
}

func (u FPUser) GetAll() map[string]string {
	return u.attrs
}

func (u FPUser) Get(key string) string {
	return u.attrs[key]
}
