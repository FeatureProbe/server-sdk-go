package featureprobe

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTogglesUnmarshal(t *testing.T) {
	var toggles Toggles
	bytes, _ := ioutil.ReadFile("./resources/fixtures/toggles.json")
	err := json.Unmarshal(bytes, &toggles)
	assert.Equal(t, nil, err)
}

func TestSaltHash(t *testing.T) {
	var h = salt_hash("key", "salt", 10000)
	assert.Equal(t, h, 2647)
}
