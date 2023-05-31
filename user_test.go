package featureprobe

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestFPUser(t *testing.T) {
	var user = NewUser().StableRollout("uniqueUserKey")
	user.With("city", "1").With("os", "linux")
	assert.Equal(t, "1", user.Get("city"))
	assert.Equal(t, 2, len(user.GetAll()))
	assert.Equal(t, "uniqueUserKey", user.Key())
}

func TestAutoGenerateUserKey(t *testing.T) {
	var user = NewUser()
	assert.Equal(t, 19, len(user.Key()))
}

func TestCurrentWriteUserAttr(t *testing.T) {
	var user = NewUser()
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			user.With("foo", "bar")
		}()
	}

	wg.Wait()
}
