package featureprobe

import (
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	repo, jsonStr := setup(t)
	var repo2 Repository

	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, jsonStr))

	synchronizer := NewSynchronizer("https://featureprobe.com/api/toggles", 1000, "sdk_key", &repo2)
	synchronizer.StartSynchronize()

	time.Sleep(1 * time.Second)

	count := httpmock.GetTotalCallCount()

	assert.True(t, count >= 1)
	assert.Equal(t, repo, repo2)

}

func TestSyncInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var repo2 Repository

	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, `{ `))

	synchronizer := NewSynchronizer("https://featureprobe.com/api/toggles", 1000, "sdk_key", &repo2)
	synchronizer.StartSynchronize()

	time.Sleep(1 * time.Second)

	count := httpmock.GetTotalCallCount()

	assert.True(t, count >= 1)
}

func TestSyncInvalidUrl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	_, jsonStr := setup(t)
	var repo2 Repository

	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, jsonStr))

	synchronizer := NewSynchronizer(string([]byte{1, 2, 3}), 1000, "sdk_key", &repo2)
	synchronizer.StartSynchronize()

	time.Sleep(1 * time.Second)
	//TODO: check error
}

func setup(t *testing.T) (Repository, string) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	jsonStr := string(bytes)
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)
	return repo, jsonStr
}
