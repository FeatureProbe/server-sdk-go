package featureprobe

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
