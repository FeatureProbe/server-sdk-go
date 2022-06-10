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
	var h = saltHash("key", "salt", 10000)
	assert.Equal(t, h, 2647)
}

func TestMatchSegmentCondition(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key11").With("city", "4")
	toggle := repo.Toggles["json_toggle"]
	detail, _ := toggle.EvalDetail(user, repo.Segments)
	v, ok := detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_1"], "v2")
}

func TestNotMatchSegmentCondition(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key11").With("city", "100")
	toggle := repo.Toggles["json_toggle"]
	toggle.Eval(user, repo.Segments)
	detail, _ := toggle.EvalDetail(user, repo.Segments)
	assert.Equal(t, detail.Reason, "default")
}

func TestNoSegments(t *testing.T) {
	c := Condition{
		Type:      "segment",
		Subject:   "subject",
		Predicate: "name",
		Objects:   nil,
	}

	user := NewUser("key11").With("city", "100")
	r := c.matchSegmentCondition(user, nil)
	assert.False(t, r)

}

func TestMultiConditions(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key").With("city", "1").With("os", "linux")
	toggle := repo.Toggles["multi_condition_toggle"]
	r, _ := toggle.Eval(user, repo.Segments)
	v, ok := r.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_0"], "")

	user = NewUser("key").With("city", "1").With("os", "linux")
	toggle = repo.Toggles["multi_condition_toggle"]
	detail, _ := toggle.EvalDetail(user, repo.Segments)
	v, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, v["variation_0"], "")

	user = NewUser("key").With("os", "linux")
	detail, _ = toggle.EvalDetail(user, repo.Segments)
	_, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, detail.Reason, "default")

	user = NewUser("key").With("city", "1")
	detail, _ = toggle.EvalDetail(user, repo.Segments)
	_, ok = detail.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, detail.Reason, "default")
}

func TestDisabledToggle(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key").With("city", "100")
	toggle := repo.Toggles["disabled_toggle"]
	detail, _ := toggle.EvalDetail(user, repo.Segments)
	assert.Equal(t, detail.Reason, "disabled")

	_, err = toggle.Eval(user, repo.Segments)
	assert.Empty(t, err)
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

	user := NewUser("key").With("name", "key")

	params := evalParams{
		Key:        "not care",
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
		Salt:         "salt",
	}

	user := NewUser("key").With("name", "key")

	params := evalParams{
		Key:        "not care",
		User:       user,
		Variations: nil,
		Segments:   nil,
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

	user := NewUser("key").With("name", "key")

	params := evalParams{
		Key:        "not care",
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

	user := NewUser("key")

	params := evalParams{
		Key:  "not care",
		User: user,
		Variations: []interface{}{
			"a", "b",
		},
		Segments: nil,
	}

	v, err := serve.selectVariation(params)
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

	user := NewUser("not care").With("name", "world")

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

	user := NewUser("not care").With("name", "not_in")

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

	user := NewUser("not care")

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

	user := NewUser("not care").With("name", "not in")

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

	user := NewUser("not care").With("name", "bob world")

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

	user := NewUser("not care").With("name", "bob")

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

	user := NewUser("not care").With("name", "bob")

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

	user := NewUser("not care").With("name", "bob world")

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

	user := NewUser("not care").With("name", "world bob")

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

	user := NewUser("not care").With("name", "bob")

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

	user := NewUser("not care").With("name", "bob")

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

	user := NewUser("not care").With("name", "world bob")

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

	user := NewUser("not care").With("name", "alice world bob")

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

	user := NewUser("not care").With("name", "alice bob")

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

	user := NewUser("not care").With("name", "alice world bob")

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

	user := NewUser("not care").With("name", "alice world bob")

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

	user := NewUser("not care").With("name", "alice orld bob hello3")

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

	user := NewUser("not care").With("name", "alice orld bob hello")

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

	user := NewUser("not care").With("name", "\\\\\\")

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

	user := NewUser("not care").With("name", "123")

	r := condition.matchStringCondition(user, condition.Predicate)
	assert.False(t, r)
}

func TestUnknownConditionType(t *testing.T) {
	c := Condition{
		Type:      "unkown",
		Subject:   "subject",
		Predicate: "name",
		Objects:   nil,
	}
	u := NewUser("key")
	b := c.meet(u, nil)
	assert.False(t, b)
}

func TestMatchEqualString(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key").With("city", "1")
	toggle := repo.Toggles["json_toggle"]
	r, _ := toggle.Eval(user, repo.Segments)
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
	user := NewUser("key")
	_, err = toggle.Eval(user, nil)
	assert.Error(t, err)

	_, err = toggle.EvalDetail(user, nil)
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
	user := NewUser("key").With("city", "1")
	_, err = toggle.Eval(user, nil)
	assert.Error(t, err)

	_, err = toggle.EvalDetail(user, nil)
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
	user := NewUser("key").With("city", "1")
	_, err = toggle.Eval(user, nil)
	assert.Error(t, err)

	_, err = toggle.EvalDetail(user, nil)
	assert.Error(t, err)
}
