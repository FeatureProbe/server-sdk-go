package featureprobe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFPUser(t *testing.T) {
	var user = NewUser("uniqueUserKey")
	user.With("city", "1").With("os", "linux")
	assert.Equal(t, "1", user.Get("city"))
	assert.Equal(t, 2, len(user.GetAll()))
}
