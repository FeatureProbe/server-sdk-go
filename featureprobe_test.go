package featureprobe

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFeatureProbe(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	config := FPConfig{
		RemoteUrl:       "",
		TogglesUrl:      nil,
		EventsUrl:       nil,
		ServerSdkKey:    "",
		RefreshInterval: 1,
		WaitFirstResp:   true,
	}

	_, err = NewFeatureProbe(config)
	assert.Empty(t, err)
}

func TestEvalNilRepo(t *testing.T) {
	fp := setupFeatureProbe(t)
	user := NewUser("key11").With("city", "4")

	val := fp.BoolValue("bool_toggle", user, true)
	assert.Equal(t, true, val)
	detail := fp.BoolDetail("bool_toggle", user, true)
	assert.Equal(t, true, detail.Value)

	val1 := fp.StrValue("string_toggle", user, "1")
	assert.Equal(t, "1", val1)
	detail1 := fp.StrDetail("string_toggle", user, "1")
	assert.Equal(t, "1", detail1.Value)

	val2 := fp.NumberValue("number_toggle", user, 1.0)
	assert.Equal(t, 1.0, val2)
	detail2 := fp.NumberDetail("number_toggle", user, 1.0)
	assert.Equal(t, 1.0, detail2.Value)

	val3 := fp.JsonValue("json_toggle", user, nil)
	assert.Equal(t, nil, val3)
	detail3 := fp.JsonDetail("json_toggle", user, nil)
	assert.Equal(t, nil, detail3.Value)
}

func TestEval(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key11").With("city", "4")

	fp := setupFeatureProbe(t)
	fp.setRepoForTest(repo)

	val := fp.BoolValue("bool_toggle", user, true)
	assert.Equal(t, false, val)
	detail := fp.BoolDetail("bool_toggle", user, true)
	assert.Equal(t, false, detail.Value)

	val1 := fp.StrValue("string_toggle", user, "1")
	assert.Equal(t, "2", val1)
	detail1 := fp.StrDetail("string_toggle", user, "1")
	assert.Equal(t, "2", detail1.Value)

	val2 := fp.NumberValue("number_toggle", user, 1.0)
	assert.Equal(t, 2.0, val2)
	detail2 := fp.NumberDetail("number_toggle", user, 1.0)
	assert.Equal(t, 2.0, detail2.Value)

	val3 := fp.JsonValue("json_toggle", user, nil)
	assert.NotEmpty(t, val3)
	detail3 := fp.JsonDetail("json_toggle", user, nil)
	assert.NotEmpty(t, detail3.Value)
}

func TestEvalTypeMismatch(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key11").With("city", "4")
	fp := setupFeatureProbe(t)
	fp.setRepoForTest(repo)

	val := fp.BoolValue("number_toggle", user, true)
	assert.Equal(t, true, val)
	detail := fp.BoolDetail("number_toggle", user, true)
	assert.Equal(t, true, detail.Value)

	val1 := fp.StrValue("number_toggle", user, "1")
	assert.Equal(t, "1", val1)
	detail1 := fp.StrDetail("number_toggle", user, "1")
	assert.Equal(t, "1", detail1.Value)

	val2 := fp.NumberValue("bool_toggle", user, 1.0)
	assert.Equal(t, 1.0, val2)
	detail2 := fp.NumberDetail("bool_toggle", user, 1.0)
	assert.Equal(t, 1.0, detail2.Value)
}

func TestEvalNotExist(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	user := NewUser("key11").With("city", "4")
	fp := setupFeatureProbe(t)
	fp.setRepoForTest(repo)

	val := fp.BoolValue("not_exist_toggle", user, true)
	assert.Equal(t, true, val)
	detail := fp.BoolDetail("not_exist_toggle", user, true)
	assert.Equal(t, true, detail.Value)

	val1 := fp.StrValue("not_exist_toggle", user, "1")
	assert.Equal(t, "1", val1)
	detail1 := fp.StrDetail("not_exist_toggle", user, "1")
	assert.Equal(t, "1", detail1.Value)

	val2 := fp.NumberValue("not_exist_toggle", user, 1.0)
	assert.Equal(t, 1.0, val2)
	detail2 := fp.NumberDetail("not_exist_toggle", user, 1.0)
	assert.Equal(t, 1.0, detail2.Value)

	val3 := fp.JsonValue("not_exist_toggle", user, nil)
	assert.Equal(t, nil, val3)
	detail3 := fp.JsonDetail("not_exist_toggle", user, nil)
	assert.Equal(t, nil, detail3.Value)
}

func TestOutOfIndexToggle(t *testing.T) {
	jsonStr := `
{
	"segments": {},
	"toggles": {
		"disabled_toggle": {
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
			"variations": [{},
				{
					"disabled_key": "disabled_value"
				}
			]
		}
	}
}`
	var repo Repository
	err := json.Unmarshal([]byte(jsonStr), &repo)
	assert.Equal(t, nil, err)

	fp := setupFeatureProbe(t)
	fp.setRepoForTest(repo)

	user := NewUser("key11").With("city", "4")

	v := fp.BoolValue("disabled_toggle", user, false)
	detail := fp.BoolDetail("disabled_toggle", user, false)
	assert.Equal(t, v, false)
	assert.Equal(t, detail.Value, false)
	assert.True(t, strings.Contains(detail.Reason, "overflow"))

	v2 := fp.StrValue("disabled_toggle", user, "1")
	detail2 := fp.StrDetail("disabled_toggle", user, "1")
	assert.Equal(t, v2, "1")
	assert.Equal(t, detail2.Value, "1")
	assert.True(t, strings.Contains(detail2.Reason, "overflow"))

	v3 := fp.NumberValue("disabled_toggle", user, 1.0)
	detail3 := fp.NumberDetail("disabled_toggle", user, 1.0)
	assert.Equal(t, v3, 1.0)
	assert.Equal(t, detail3.Value, 1.0)
	assert.True(t, strings.Contains(detail3.Reason, "overflow"))

	v4 := fp.JsonValue("disabled_toggle", user, nil)
	detail4 := fp.JsonDetail("disabled_toggle", user, nil)
	assert.Equal(t, v4, nil)
	assert.Equal(t, detail4.Value, nil)
	assert.True(t, strings.Contains(detail4.Reason, "overflow"))
}

func setupFeatureProbe(t *testing.T) FeatureProbe {
	config := FPConfig{
		RemoteUrl:       "",
		TogglesUrl:      nil,
		EventsUrl:       nil,
		ServerSdkKey:    "",
		RefreshInterval: 1,
		WaitFirstResp:   true,
	}

	fp, err := NewFeatureProbe(config)
	assert.Empty(t, err)
	return fp
}
