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
	variationIndex1 := 0
	variationIndex2 := 1
	ruleIndex := 0
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex1,
		RuleIndex:      &ruleIndex,
		Version:        &version1,
		Reason:         "default",
	}, true)

	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex1,
		RuleIndex:      &ruleIndex,
		Version:        &version1,
		Reason:         "default",
	}, true)

	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex1,
		RuleIndex:      &ruleIndex,
		Version:        &version1,
		Reason:         "default",
	}, false)

	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value2",
		VariationIndex: &variationIndex2,
		RuleIndex:      &ruleIndex,
		Version:        &version1,
		Reason:         "default",
	}, true)

	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle2",
		Value:          "some_value",
		VariationIndex: &variationIndex2,
		RuleIndex:      &ruleIndex,
		Version:        &version2,
		Reason:         "default",
	}, true)

	recorder.RecordCustom(CustomEvent{
		Kind:  "custom",
		Time:  time.Now().Unix(),
		User:  "some_user",
		Name:  "some_event",
		Value: nil,
	})

	assert.True(t, len(recorder.access.Counters) == 2)
	assert.True(t, recorder.access.Counters["some_toggle"][0].Count == 3)
	assert.True(t, len(recorder.access.Counters["some_toggle"]) == 2)
	assert.True(t, len(recorder.incomingEvents) == 5)

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
	version := uint64(1)
	variationIndex := 0
	ruleIndex := 0
	recorder := NewEventRecorder(string([]byte{1, 2, 3}), 1000, "sdk_key")
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex,
		RuleIndex:      &ruleIndex,
		Version:        &version,
		Reason:         "default",
	}, true)
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex,
		RuleIndex:      &ruleIndex,
		Version:        &version,
		Reason:         "default",
	}, true)

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
	version := uint64(1)
	variationIndex := 0
	ruleIndex := 0
	recorder := NewEventRecorder("https://featureprobe.com/api/events", 1000, "sdk_key")
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex,
		RuleIndex:      &ruleIndex,
		Version:        &version,
		Reason:         "default",
	}, true)
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex,
		RuleIndex:      &ruleIndex,
		Version:        &version,
		Reason:         "default",
	}, true)

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
	version := uint64(1)
	variationIndex := 0
	ruleIndex := 0
	recorder := NewEventRecorder("https://featureprobe.com/api/events", 5000, "sdk_key")
	recorder.Start()
	recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           time.Now().Unix(),
		User:           "some_user",
		Key:            "some_toggle",
		Value:          "some_value",
		VariationIndex: &variationIndex,
		RuleIndex:      &ruleIndex,
		Version:        &version,
		Reason:         "default",
	}, true)
	httpmock.ActivateNonDefault(&recorder.httpClient)
	httpmock.RegisterResponder("POST", "https://featureprobe.com/api/events",
		httpmock.NewStringResponder(200, "{"))

	recorder.Stop()

	count := httpmock.GetTotalCallCount()
	assert.Equal(t, 1, count)
	defer httpmock.DeactivateAndReset()
}
