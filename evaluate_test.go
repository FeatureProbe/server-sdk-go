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
	//t.Log(toggles)
}

func TestSaltHash(t *testing.T) {
	var h = saltHash("key", "salt", 10000)
	assert.Equal(t, h, 2647)
}

func TestMultiConditions(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key").With("city", "1").With("os", "linux")
	toggle := repo.Toggles["multi_condition_toggle"]
	r, err := toggle.Eval(*user, repo.Segments)
	t.Log(r)
}
