package featureprobe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func loadRepoFromFile() (repo Repository, err error) {
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	repoData := RepositoryData{}
	err = json.Unmarshal(bytes, &repoData)
	repo.flush(repoData)
	return
}

func TestTogglesUnmarshal(t *testing.T) {
	var toggles Toggles
	bytes, _ := ioutil.ReadFile("./resources/fixtures/toggles.json")
	err := json.Unmarshal(bytes, &toggles)
	assert.Equal(t, nil, err)
}

func TestSaltHash(t *testing.T) {
	var h = saltHash("key", "salt", 10000)
	assert.Equal(t, h, 2647)
}

func TestMatchInSegmentCondition(t *testing.T) {
	repo, err := loadRepoFromFile()
	assert.Equal(t, nil, err)

	user := NewUser().With("city", "4")
	toggle, _ := repo.getToggle("json_toggle")
	detail, _ := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	v, ok := detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_1"], "v2")
}

func TestMatchNotInSegmentCondition(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "100")
	toggle, _ := repo.getToggle("not_in_segment")
	detail, _ := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	v, ok := detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["not_in"], true)
}

func TestNotMatchInSegmentCondition(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "100")
	toggle, _ := repo.getToggle("json_toggle")
	_, _ = toggle.eval(user, repo.getToggles(), repo.getSegments(), nil, 10)
	detail, _ := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Equal(t, detail.Reason, "default")
}

func TestNoSegments(t *testing.T) {
	c := Condition{
		Type:      "segment",
		Subject:   "subject",
		Predicate: "name",
		Objects:   nil,
	}

	user := NewUser().With("city", "100")
	r := c.matchSegmentCondition(user, "is in", nil)
	assert.False(t, r)

	r = c.matchSegmentCondition(user, "is not in", nil)
	assert.False(t, r)
}

func TestSegmentsUnknownPredicate(t *testing.T) {
	c := Condition{
		Type:      "segment",
		Subject:   "subject",
		Predicate: "name",
		Objects:   nil,
	}

	segments := map[string]Segment{}

	user := NewUser().StableRollout("key11").With("city", "100")
	r := c.matchSegmentCondition(user, "unknown", segments)
	assert.False(t, r)
}

func TestMultiConditions(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().StableRollout("key11").With("city", "1").With("os", "linux")
	toggle, _ := repo.getToggle("multi_condition_toggle")
	r, _ := toggle.eval(user, repo.getToggles(), repo.getSegments(), nil, 10)
	v, ok := r.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_0"], "")

	user = NewUser().StableRollout("key").With("city", "1").With("os", "linux")
	toggle, _ = repo.getToggle("multi_condition_toggle")
	detail, _ := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	v, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_0"], "")

	user = NewUser().StableRollout("key").With("os", "linux")
	detail, _ = toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	_, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, detail.Reason, "default")

	user = NewUser().StableRollout("key").With("city", "1")
	detail, _ = toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	_, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, detail.Reason, "default")
}

func TestDisabledToggle(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "100")
	toggle, _ := repo.getToggle("disabled_toggle")
	detail, _ := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Equal(t, detail.Reason, "disabled")

	_, err := toggle.eval(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Empty(t, err)
}

func TestPrerequisiteToggleMatched(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "1")
	toggle, _ := repo.getToggle("prerequisite_toggle")

	detail, err := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Empty(t, err)
	assert.Equal(t, detail.Value, "2")
}

func TestPrerequisiteToggleNotMatchedShouldBeReturnDefaultValue(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "6")
	toggle, _ := repo.getToggle("not_match_prerequisite_toggle")

	detail, err := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Empty(t, err)
	assert.Equal(t, detail.Reason, "default")
	assert.Equal(t, detail.Value, "1")
}

func TestPrerequisiteToggleNotExistShouldBeReturnDefaultValue(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "6")
	toggle, _ := repo.getToggle("prerequisite_not_exist_toggle")

	detail, err := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 10)
	assert.Empty(t, err)
	assert.Equal(t, detail.Reason, "prerequisite toggle not exist")
	assert.Equal(t, detail.Value, "1")
}

func TestPrerequisiteToggleDeepOverlowShouldBeReturnDefaultValue(t *testing.T) {
	repo, _ := loadRepoFromFile()

	user := NewUser().With("city", "6")
	toggle, _ := repo.getToggle("prerequisite_deep_overflow")

	detail, err := toggle.evalDetail(user, repo.getToggles(), repo.getSegments(), nil, 5)
	assert.Empty(t, err)
	assert.Equal(t, detail.Reason, "prerequisite deep overflow")
	assert.Equal(t, detail.Value, "1")
}

func TestDistributionNoSalt(t *testing.T) {
	distribution := [][]Range{
		{Range{Lower: 0, Upper: 2647}},
		{Range{Lower: 2647, Upper: 2648}},
		{Range{Lower: 2648, Upper: 10000}},
	}

	split := Split{
		Distribution: distribution,
		BucketBy:     "name",
		Salt:         "",
	}

	user := NewUser().StableRollout("key").With("name", "key")

	params := EvalParam{
		User:       user,
		Variations: nil,
		Segments:   nil,
	}

	index, _ := split.findIndex(params)
	assert.Equal(t, index, 2)
}

func TestDistributionInExactBucket(t *testing.T) {
	distribution := [][]Range{
		{Range{Lower: 0, Upper: 2647}},
		{Range{Lower: 2647, Upper: 2648}},
		{Range{Lower: 2648, Upper: 10000}},
	}

	split := Split{
		Distribution: distribution,
		BucketBy:     "name",
		Salt:         "",
	}

	user := NewUser().StableRollout("key").With("name", "key")

	params := EvalParam{
		User:       user,
		Variations: nil,
		Segments:   nil,
		Key:        "salt",
	}

	index, _ := split.findIndex(params)
	assert.Equal(t, index, 1)
}

func TestDistributionInNoneBucket(t *testing.T) {
	distribution := [][]Range{
		{Range{Lower: 0, Upper: 2647}},
		{Range{Lower: 2648, Upper: 10000}},
	}

	split := Split{
		Distribution: distribution,
		BucketBy:     "name",
		Salt:         "salt",
	}

	user := NewUser().StableRollout("key").With("name", "key")

	params := EvalParam{
		User:       user,
		Variations: nil,
		Segments:   nil,
	}

	_, err := split.findIndex(params)
	assert.Error(t, err)
}

func TestSelectVariationFail(t *testing.T) {
	distribution := [][]Range{
		{Range{Lower: 0, Upper: 5000}},
		{Range{Lower: 5000, Upper: 10000}},
	}

	split := Split{
		Distribution: distribution,
		BucketBy:     "name",
		Salt:         "salt",
	}
	serve := Serve{
		Split:  &split,
		Select: nil,
	}

	user := NewUser().StableRollout("key")

	params := EvalParam{
		User: user,
		Variations: []interface{}{
			"a", "b",
		},
		Segments: nil,
	}

	v, _, err := serve.selectVariation(params)
	assert.Equal(t, v, nil)
	assert.Error(t, err)
}

func TestMatchIsOneOf(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "is one of",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "world")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchIsOneOf(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "is one of",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "not_in")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestUserMissKeyIsNotOneOf(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "is not any of",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser()

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchIsNotAnyOf(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "is not any of",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "not in")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestMatchEndsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "ends with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob world")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchEndsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "ends with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchNotEndsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not end with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchNotEndsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not end with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob world")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchStartsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "starts with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "world bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchStartsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "starts with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchNotStartsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not start with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchNotStartsWith(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not start with",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "world bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchCondition(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "contains",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "alice world bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchContainsCondition(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "contains",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "alice bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchNotContainsCondition(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not contain",
		Objects: []string{
			"hello", "world",
		},
	}

	user := NewUser().With("name", "alice world bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestMatchRegex(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "matches regex",
		Objects: []string{
			"hello", "world.*",
		},
	}

	user := NewUser().With("name", "alice world bob")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestMatchRegexFirstObject(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "matches regex",
		Objects: []string{
			"hello\\d", "world.*",
		},
	}

	user := NewUser().With("name", "alice orld bob hello3")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestNotMatchRegex(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "does not match regex",
		Objects: []string{
			"hello\\d", "world.*",
		},
	}

	user := NewUser().With("name", "alice orld bob hello")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.True(t, r)
}

func TestInvalidRegex(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "matches regex",
		Objects: []string{
			"\\\\\\",
		},
	}

	user := NewUser().With("name", "\\\\\\")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestUnknownPredicate(t *testing.T) {
	condition := Condition{
		Type:      "string",
		Subject:   "name",
		Predicate: "unknown",
		Objects: []string{
			"123",
		},
	}

	user := NewUser().With("name", "123")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestDatetimeBefore(t *testing.T) {
	now := time.Now().Unix()
	condition := Condition{
		Type:      "datetime",
		Subject:   "datetime",
		Predicate: "before",
		Objects: []string{
			fmt.Sprintf("%d", now+1),
		},
	}

	user := NewUser()
	r := condition.meet(user, nil)
	assert.True(t, r)

	user.With("datetime", fmt.Sprintf("%d", now))
	r = condition.meet(user, nil)
	assert.True(t, r)

	user.With("datetime", fmt.Sprintf("%d", now+1))
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestDatetimeAfter(t *testing.T) {
	now := time.Now().Unix()
	condition := Condition{
		Type:      "datetime",
		Subject:   "datetime",
		Predicate: "after",
		Objects: []string{
			fmt.Sprintf("%d", now),
		},
	}

	user := NewUser()
	r := condition.meet(user, nil)
	assert.True(t, r)

	user.With("datetime", fmt.Sprintf("%d", now))
	r = condition.meet(user, nil)
	assert.True(t, r)

	user.With("datetime", fmt.Sprintf("%d", now+1))
	r = condition.meet(user, nil)
	assert.True(t, r)

	user.With("datetime", fmt.Sprintf("%d", now-1))
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestDatetimeInvalidCustomValue(t *testing.T) {
	condition := Condition{
		Type:      "datetime",
		Subject:   "datetime",
		Predicate: "after",
		Objects:   []string{},
	}

	user := NewUser().With("datetime", "a")
	r := condition.meet(user, nil)
	assert.False(t, r)
}

func TestDatetimeInvalid(t *testing.T) {
	condition := Condition{
		Type:      "datetime",
		Subject:   "datetime",
		Predicate: "after",
		Objects: []string{
			"a",
		},
	}

	user := NewUser()
	r := condition.meet(user, nil)
	assert.False(t, r)
}

func TestDatetimeUnknownPredicate(t *testing.T) {
	condition := Condition{
		Type:      "datetime",
		Subject:   "datetime",
		Predicate: "a",
		Objects: []string{
			"a",
		},
	}

	user := NewUser()
	r := condition.meet(user, nil)
	assert.False(t, r)
}

func TestNumberEqual(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "=",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "1")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestNumberNotEqual(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "!=",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "1")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestNumberGreaterThan(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: ">",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "1")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestNumberGreaterThanOrEqualTo(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: ">=",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "0")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "1")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestNumberLessThan(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "<",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "1")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestNumberLessThanOrEqualTo(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "<=",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "1")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "2")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "3")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("price", "4")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestNumberInvalid(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "?",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "a")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser()
	r = condition.meet(user, nil)
	assert.False(t, r)

	condition = Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: ">",
		Objects: []string{
			"a",
		},
	}

	user = NewUser().With("price", "1")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestNumberUnknownPredicate(t *testing.T) {
	condition := Condition{
		Type:      "number",
		Subject:   "price",
		Predicate: "?",
		Objects: []string{
			"1", "2", "3",
		},
	}

	user := NewUser().With("price", "0")
	r := condition.meet(user, nil)
	assert.False(t, r)
}

func TestSemVerEqual(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: "=",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "1.0.0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "2.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestSemVerNotEqual(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: "!=",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "1.0.0")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "2.0.0")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestSemVerGreaterThan(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: ">",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "1.0.0")
	r := condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "2.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestSemVerGreaterThanOrEqualTo(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: ">=",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "1.0.0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "2.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)
}

func TestSemVerLessThan(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: "<",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "0.1.0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "3.0.0")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestSemVerLessThanOrEqualTo(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: "<=",
		Objects: []string{
			"1.0.0", "2.0.0", "3.0.0",
		},
	}

	user := NewUser().With("version", "0.1.0")
	r := condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "1.1.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "2.0.0")
	r = condition.meet(user, nil)
	assert.True(t, r)

	user = NewUser().With("version", "4.1.0")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestSemVerInvalid(t *testing.T) {
	condition := Condition{
		Type:      "semver",
		Subject:   "version",
		Predicate: ">",
		Objects: []string{
			"invalid",
		},
	}

	user := NewUser().With("version", "0.1.0")
	r := condition.meet(user, nil)
	assert.False(t, r)

	condition.Predicate = "?"
	user = NewUser().With("version", "0.1.0")
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser()
	r = condition.meet(user, nil)
	assert.False(t, r)

	user = NewUser().With("version", "invalid_version")
	r = condition.meet(user, nil)
	assert.False(t, r)
}

func TestUnknownConditionType(t *testing.T) {
	c := Condition{
		Type:      "unknown",
		Subject:   "subject",
		Predicate: "name",
		Objects:   nil,
	}
	u := NewUser()
	b := c.meet(u, nil)
	assert.False(t, b)
}

func TestMatchEqualString(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	repoData := RepositoryData{}
	err := json.Unmarshal(bytes, &repoData)
	repo.flush(repoData)
	assert.Equal(t, nil, err)

	user := NewUser().With("city", "1")
	toggle, _ := repo.getToggle("json_toggle")
	r, _ := toggle.eval(user, repo.getToggles(), repo.getSegments(), nil, 10)
	v, ok := r.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_0"], "c2")
}

func TestInvalidJsonRange(t *testing.T) {
	var r Range
	jsonStr := `{"a": 123}`
	err := json.Unmarshal([]byte(jsonStr), &r)
	assert.Error(t, err)

	jsonStr = `[100]`
	err = json.Unmarshal([]byte(jsonStr), &r)
	assert.Error(t, err)

}

func TestDisabledOutOfRangeToggle(t *testing.T) {
	var toggle Toggle
	jsonStr := `
		{
            "key": "disabled_toggle",
            "enabled": false,
            "version": 1,
            "disabledServe": {
                "select": 2
            },
            "defaultServe": {
                "select": 2
            },
            "rules": [],
            "variations": [
                {},
                {
                    "disabled_key": "disabled_value"
                }
            ]
        }`
	err := json.Unmarshal([]byte(jsonStr), &toggle)
	assert.Empty(t, err)
	user := NewUser()
	_, err = toggle.eval(user, nil, nil, nil, 10)
	assert.Error(t, err)

	_, err = toggle.evalDetail(user, nil, nil, nil, 10)
	assert.Error(t, err)
}

func TestEnabledOutOfRangeToggle(t *testing.T) {
	var toggle Toggle
	jsonStr := `
		{
            "key": "disabled_toggle",
            "enabled": true,
            "version": 1,
            "disabledServe": {
                "select": 2
            },
            "defaultServe": {
                "select": 2
            },
            "rules": [{
			  	"serve": {
					"select": 2
			  	},
			  	"conditions": [
					{
					  	"type": "string",
					  	"subject": "city",
					  	"predicate": "is one of",
					  	"objects": [
						 	"1",
						 	"2",
						 	"3"
					  ]
					}
				]
			}],
            "variations": [
                {},
                {
                    "disabled_key": "disabled_value"
                }
            ]
        }`
	err := json.Unmarshal([]byte(jsonStr), &toggle)
	assert.Empty(t, err)
	user := NewUser().With("city", "1")
	_, err = toggle.eval(user, nil, nil, nil, 10)
	assert.Error(t, err)

	_, err = toggle.evalDetail(user, nil, nil, nil, 10)
	assert.Error(t, err)
}

func TestDefaultServeOutOfRangeToggle(t *testing.T) {
	var toggle Toggle
	jsonStr := `
		{
            "key": "disabled_toggle",
            "enabled": true,
            "version": 1,
            "disabledServe": {
                "select": 2
            },
            "defaultServe": {
                "select": 2
            },
            "rules": [],
            "variations": [
                {},
                {
                    "disabled_key": "disabled_value"
                }
            ]
        }`
	err := json.Unmarshal([]byte(jsonStr), &toggle)
	assert.Empty(t, err)
	user := NewUser().With("city", "1")
	_, err = toggle.eval(user, nil, nil, nil, 10)
	assert.Error(t, err)

	_, err = toggle.evalDetail(user, nil, nil, nil, 10)
	assert.Error(t, err)
}

func TestClearRepo(t *testing.T) {
	repo, _ := loadRepoFromFile()
	assert.True(t, len(repo.getSegments()) > 0)
	assert.True(t, len(repo.getToggles()) > 0)

	repo.Clear()

	assert.Equal(t, 0, len(repo.getSegments()))
	assert.Equal(t, 0, len(repo.getToggles()))
}
