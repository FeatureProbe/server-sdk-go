package featureprobe

type FPUser struct {
	Key   string
	attrs map[string]string
}

func NewUser(key string) FPUser {
	return FPUser{
		Key:   key,
		attrs: map[string]string{},
	}
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
