package featureprobe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSyncWaitFirstResp(t *testing.T) {
	repo, jsonStr := setup(t)
	var repo2 Repository
	synchronizer := NewSynchronizer("https://featureprobe.com/api/toggles", 1000, "sdk_key", &repo2)

	httpmock.ActivateNonDefault(&synchronizer.httpClient)
	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, jsonStr))

	synchronizer.Start(true)
	defer synchronizer.Stop()
	time.Sleep(4 * time.Second)
	count := httpmock.GetTotalCallCount()

	assert.True(t, count > 2)
	synchronizer.mu.Lock() // for go test -race
	assert.Equal(t, repo, repo2)
	httpmock.DeactivateAndReset()
	synchronizer.mu.Unlock()
}

func TestSyncNotWaitFirstResp(t *testing.T) {
	repo, jsonStr := setup(t)
	var repo2 Repository
	synchronizer := NewSynchronizer("https://featureprobe.com/api/toggles", 1000, "sdk_key", &repo2)

	httpmock.ActivateNonDefault(&synchronizer.httpClient)
	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, jsonStr))

	synchronizer.Start()
	defer synchronizer.Stop()
	time.Sleep(4 * time.Second)
	count := httpmock.GetTotalCallCount()
	fmt.Println(count)

	assert.True(t, count > 2)
	synchronizer.mu.Lock() // for go test -race
	assert.Equal(t, repo, repo2)
	httpmock.DeactivateAndReset()
	synchronizer.mu.Unlock()
}

func TestSyncInvalidJson(t *testing.T) {
	var repo2 Repository
	synchronizer := NewSynchronizer("https://featureprobe.com/api/toggles", 100, "sdk_key", &repo2)

	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, `{ `))
	httpmock.ActivateNonDefault(&synchronizer.httpClient)

	synchronizer.Start()
	defer synchronizer.Stop()
	time.Sleep(1 * time.Second)
	count := httpmock.GetTotalCallCount()

	assert.True(t, count >= 1)
	synchronizer.mu.Lock()
	httpmock.DeactivateAndReset()
	synchronizer.mu.Unlock()
}

func TestSyncInvalidUrl(t *testing.T) {
	var repo2 Repository
	synchronizer := NewSynchronizer(string([]byte{1, 2, 3}), 100, "sdk_key", &repo2)
	_, jsonStr := setup(t)

	httpmock.ActivateNonDefault(&synchronizer.httpClient)
	synchronizer.Start()
	defer synchronizer.Stop()
	httpmock.RegisterResponder("GET", "https://featureprobe.com/api/toggles",
		httpmock.NewStringResponder(200, jsonStr))

	time.Sleep(1 * time.Second)
	synchronizer.mu.Lock()
	httpmock.DeactivateAndReset()
	synchronizer.mu.Unlock()
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
