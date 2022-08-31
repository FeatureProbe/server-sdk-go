package featureprobe

import (
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestEventFlush(t *testing.T) {
	recorder := NewEventRecorder("https://featureprobe.com/api/events", 1000, "sdk_key")
	version1 := uint64(1)
	version2 := uint64(2)
	recorder.RecordAccess(AccessEvent{
		Time:    time.Now().Unix(),
		Key:     "some_toggle",
		Value:   "some_value",
		Version: &version1,
		Reason:  "default",
	})
	recorder.RecordAccess(AccessEvent{
		Time:    time.Now().Unix(),
		Key:     "some_toggle",
		Value:   "some_value",
		Version: &version1,
		Reason:  "default",
	})
	recorder.RecordAccess(AccessEvent{
		Time:    time.Now().Unix(),
		Key:     "some_toggle",
		Value:   "some_value",
		Version: &version2,
		Reason:  "default",
	})

	httpmock.ActivateNonDefault(&recorder.httpClient)
	httpmock.RegisterResponder("POST", "https://featureprobe.com/api/events",
		httpmock.NewStringResponder(200, "{}"))

	recorder.Start()

	time.Sleep(2 * time.Second)
	count := httpmock.GetTotalCallCount()
	assert.True(t, count >= 1)
	defer httpmock.DeactivateAndReset()
}

func TestEventFlushInvalidUrl(t *testing.T) {
	recorder := NewEventRecorder(string([]byte{1, 2, 3}), 1000, "sdk_key")
	recorder.RecordAccess(AccessEvent{
		Time:   time.Now().Unix(),
		Key:    "some_toggle",
		Value:  "some_value",
		Reason: "default",
	})
	recorder.RecordAccess(AccessEvent{
		Time:   time.Now().Unix(),
		Key:    "some_toggle",
		Value:  "some_value",
		Reason: "default",
	})

	httpmock.ActivateNonDefault(&recorder.httpClient)
	httpmock.RegisterResponder("POST", "https://featureprobe.com/api/events",
		httpmock.NewStringResponder(200, "{}"))

	recorder.Start()

	time.Sleep(2 * time.Second)
	count := httpmock.GetTotalCallCount()
	assert.Equal(t, 0, count)
	defer httpmock.DeactivateAndReset()
}

func TestEventFlushInvalidResp(t *testing.T) {
	recorder := NewEventRecorder("https://featureprobe.com/api/events", 1000, "sdk_key")
	recorder.RecordAccess(AccessEvent{
		Time:   time.Now().Unix(),
		Key:    "some_toggle",
		Value:  "some_value",
		Reason: "default",
	})
	recorder.RecordAccess(AccessEvent{
		Time:   time.Now().Unix(),
		Key:    "some_toggle",
		Value:  "some_value",
		Reason: "default",
	})

	httpmock.ActivateNonDefault(&recorder.httpClient)
	httpmock.RegisterResponder("POST", "https://featureprobe.com/api/events",
		httpmock.NewStringResponder(200, "{"))

	recorder.Start()

	time.Sleep(2 * time.Second)
	count := httpmock.GetTotalCallCount()
	assert.True(t, count > 0)
	defer httpmock.DeactivateAndReset()
}

func TestCloseEvent(t *testing.T) {
	recorder := NewEventRecorder("https://featureprobe.com/api/events", 5000, "sdk_key")
	recorder.Start()
	recorder.RecordAccess(AccessEvent{
		Time:   time.Now().Unix(),
		Key:    "some_toggle",
		Value:  "some_value",
		Reason: "default",
	})
	httpmock.ActivateNonDefault(&recorder.httpClient)
	httpmock.RegisterResponder("POST", "https://featureprobe.com/api/events",
		httpmock.NewStringResponder(200, "{"))

	recorder.Stop()

	count := httpmock.GetTotalCallCount()
	assert.Equal(t, 1, count)
	defer httpmock.DeactivateAndReset()
}
