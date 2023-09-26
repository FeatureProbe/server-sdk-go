package featureprobe

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFeatureProbe(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	err := json.Unmarshal(bytes, &repo)
	assert.Equal(t, nil, err)

	config := FPConfig{
		RemoteUrl: "https://featureprobe.com/",
		RefreshInterval: 100 * time.Millisecond,
		Repo:            &repo,
	}
	client := NewFeatureProbe(config)
	assert.True(t, client.Initialized())
}

func TestEvalNilRepo(t *testing.T) {
	config := FPConfig{
		RemoteUrl: "https://featureprobe.com/",
		RefreshInterval: 100 * time.Millisecond,
	}
	fp := FeatureProbe{
		Repo:   nil,
		Config: config,
	}
	user := NewUser().StableRollout("key11").With("city", "4")

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
	repoDate := RepositoryData{}
	err := json.Unmarshal(bytes, &repoDate)
	repo.flush(repoDate)
	assert.Equal(t, nil, err)

	user := NewUser().StableRollout("key11").With("city", "4")

	fp := setupFeatureProbe(t, repo)

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

	user := NewUser().StableRollout("key11").With("city", "4")
	fp := setupFeatureProbe(t, repo)

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

	user := NewUser().With("city", "4")
	fp := setupFeatureProbe(t, repo)

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
		"overflow_bool_toggle": {
			"key": "overflow_bool_toggle",
			"enabled": true,
			"version": 1,
			"disabledServe": {
				"select": 2
			},
			"defaultServe": {
				"select": 2
			},
			"rules": [],
			"variations": [true, false]
		},
		"overflow_str_toggle": {
			"key": "overflow_str_toggle",
			"enabled": true,
			"version": 1,
			"disabledServe": {
				"select": 2
			},
			"defaultServe": {
				"select": 2
			},
			"rules": [],
			"variations": ["1", "2"]
		},
		"overflow_number_toggle": {
			"key": "overflow_number_toggle",
			"enabled": true,
			"version": 1,
			"disabledServe": {
				"select": 2
			},
			"defaultServe": {
				"select": 2
			},
			"rules": [],
			"variations": [1.0, 2.0]
		},
		"overflow_json_toggle": {
			"key": "overflow_json_toggle",
			"enabled": true,
			"version": 1,
			"disabledServe": {
				"select": 2
			},
			"defaultServe": {
				"select": 2
			},
			"rules": [],
			"variations": [{}, {}]
		}
	}
}`
	var repo Repository
	repoData := RepositoryData{}
	err := json.Unmarshal([]byte(jsonStr), &repoData)
	repo.flush(repoData)
	assert.Equal(t, nil, err)

	fp := setupFeatureProbe(t, repo)

	user := NewUser().With("city", "4")

	v := fp.BoolValue("overflow_bool_toggle", user, false)
	detail := fp.BoolDetail("overflow_bool_toggle", user, false)
	assert.Equal(t, false, v)
	assert.Equal(t, false, detail.Value)
	assert.True(t, strings.Contains(detail.Reason, "overflow"))

	v2 := fp.StrValue("overflow_str_toggle", user, "1")
	detail2 := fp.StrDetail("overflow_str_toggle", user, "1")
	assert.Equal(t, "1", v2)
	assert.Equal(t, "1", detail2.Value)
	assert.True(t, strings.Contains(detail2.Reason, "overflow"))

	v3 := fp.NumberValue("overflow_number_toggle", user, 1.0)
	detail3 := fp.NumberDetail("overflow_number_toggle", user, 1.0)
	assert.Equal(t, 1.0, v3)
	assert.Equal(t, 1.0, detail3.Value)
	assert.True(t, strings.Contains(detail3.Reason, "overflow"))

	v4 := fp.JsonValue("overflow_json_toggle", user, nil)
	detail4 := fp.JsonDetail("overflow_json_toggle", user, nil)
	assert.Equal(t, nil, v4)
	assert.Equal(t, nil, detail4.Value)
	assert.True(t, strings.Contains(detail4.Reason, "overflow"))
}

func TestUnitTestingForCaller(t *testing.T) {
	toggles := map[string]interface{}{}
	toggles["toggle0"] = 0
	toggles["toggle1"] = 1.0
	toggles["toggle2"] = true
	toggles["toggle3"] = "red"
	toggles["toggle4"] = []int{1, 2, 3}

	fp := NewFeatureProbeForTest(toggles)
	user := NewUser()

	assert.Equal(t, 2.0, fp.NumberValue("toggle0", user, 2))
	assert.Equal(t, 1.0, fp.NumberValue("toggle1", user, 2))
	assert.Equal(t, true, fp.BoolValue("toggle2", user, false))
	assert.Equal(t, "red", fp.StrValue("toggle3", user, "blue"))
	assert.Equal(t, []int{1, 2, 3}, fp.JsonValue("toggle4", user, nil))
}

func TestClientInitializedTimeout(t *testing.T) {
	config := FPConfig{
		RemoteUrl:       "http://not-found/server",
		RefreshInterval: 3 * time.Second,
		StartWait:       3 * time.Second,
	}
	fp := NewFeatureProbe(config)
	assert.Equal(t, false, fp.Initialized())
}

func TestCloseClient(t *testing.T) {
	config := FPConfig{
		RefreshInterval: 100 * time.Millisecond,
	}
	fp := NewFeatureProbe(config)

	fp.Close()
	assert.Equal(t, 0, len(fp.Repo.getToggles()))
}

func TestTrack(t *testing.T) {
	config := FPConfig{
		RemoteUrl: "http://localhost/",
		RefreshInterval: 100 * time.Millisecond,
	}
	fp := NewFeatureProbe(config)
	user := NewUser()
	value := 99.9
	fp.Track("some_event", user, &value)
	fp.Track("some_event2", user, nil)
	assert.True(t, len(fp.Recorder.incomingEvents) == 2)
	fp.Close()
}

func TestRecorderDebugEvent(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	repoData := RepositoryData{}
	json.Unmarshal(bytes, &repoData)
	repoData.DebugUntilTime = uint64((time.Now().UnixNano() / 1e6) + 5000)
	repo.flush(repoData)

	user := NewUser().With("city", "4")
	fp := setupFeatureProbe(t, repo)
	fp.BoolValue("bool_toggle", user, true)

	assert.Equal(t, 1, len(fp.Recorder.incomingEvents))
	debugEvent := (fp.Recorder.incomingEvents[0]).(DebugEvent)
	assert.Equal(t, debugEvent.Kind, "debug")
	assert.Equal(t, debugEvent.Value, false)
	assert.Equal(t, *debugEvent.RuleIndex, 1)
	assert.Equal(t, *debugEvent.VariationIndex, 1)
	assert.Equal(t, debugEvent.Reason, "rule 1 ")
}

func TestNotRecorderDebugEvent(t *testing.T) {
	var repo Repository
	bytes, _ := ioutil.ReadFile("./resources/fixtures/repo.json")
	json.Unmarshal(bytes, &repo)
	user := NewUser().With("city", "4")
	fp := setupFeatureProbe(t, repo)
	fp.BoolValue("bool_toggle", user, true)
	assert.Equal(t, 0, len(fp.Recorder.incomingEvents))
}

func TestContract(t *testing.T) {
	bytes, _ := ioutil.ReadFile("./resources/fixtures/server-sdk-specification/spec/toggle_simple_spec.json")
	var tests ContractTests
	err := json.Unmarshal(bytes, &tests)
	assert.Equal(t, nil, err)

	for _, scenario := range tests.Tests {
		t.Log("scenario: ", scenario.Scenario)
		assert.NotEmpty(t, scenario.Cases)
		repo := Repository{}
		repo.flush(scenario.Fixture)
		fp := FeatureProbe{Repo: &repo, Config: FPConfig{
			RemoteUrl: "https://localhost/",
			MaxPrerequisitesDeep: 5}}

		for _, Case := range scenario.Cases {
			t.Log("  case: ", Case.Name)
			user := NewUser().StableRollout(Case.User.Key)
			for _, kv := range Case.User.CustomValues {
				user = user.With(kv.Key, kv.Value)
			}

			switch Case.Function.Name {
			case "bool_value":
				d := Case.Function.Default.(bool)
				v := fp.BoolValue(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v)
			case "string_value":
				d := Case.Function.Default.(string)
				v := fp.StrValue(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v)
			case "number_value":
				d := Case.Function.Default.(float64)
				v := fp.NumberValue(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v)
			case "json_value":
				d := Case.Function.Default
				v := fp.JsonValue(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v)

			case "bool_detail":
				d := Case.Function.Default.(bool)
				v := fp.BoolDetail(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v.Value)
				assertBoolDetail(t, Case, v)
			case "string_detail":
				d := Case.Function.Default.(string)
				v := fp.StrDetail(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v.Value)
				assertStrDetail(t, Case, v)
			case "number_detail":
				d := Case.Function.Default.(float64)
				v := fp.NumberDetail(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v.Value)
				assertNumberDetail(t, Case, v)
			case "json_detail":
				d := Case.Function.Default
				v := fp.JsonDetail(Case.Function.Toggle, user, d)
				assert.Equal(t, Case.ExpectResult.Value, v.Value)
				assertJsonDetail(t, Case, v)
			}
		}
	}
}

func assertBoolDetail(t *testing.T, Case Case, r FPBoolDetail) {
	if Case.ExpectResult.Reason != nil {
		//assert.True(t, strings.Contains(r.Reason, *Case.ExpectResult.Reason))
		assert.Contains(t, r.Reason, *Case.ExpectResult.Reason)
	}
	if Case.ExpectResult.RuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.RuleIndex, *r.RuleIndex)
	}
	if Case.ExpectResult.NoRuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.NoRuleIndex, r.RuleIndex == nil)
	}
	if Case.ExpectResult.Version != nil {
		assert.Equal(t, *Case.ExpectResult.Version, *r.Version)
	}
}

func assertNumberDetail(t *testing.T, Case Case, r FPNumberDetail) {
	if Case.ExpectResult.Reason != nil {
		assert.True(t, strings.Contains(r.Reason, *Case.ExpectResult.Reason))

	}
	if Case.ExpectResult.RuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.RuleIndex, *r.RuleIndex)
	}
	if Case.ExpectResult.NoRuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.NoRuleIndex, r.RuleIndex == nil)
	}
	if Case.ExpectResult.Version != nil {
		assert.Equal(t, *Case.ExpectResult.Version, *r.Version)
	}
}

func assertStrDetail(t *testing.T, Case Case, r FPStrDetail) {
	// if Case.ExpectResult.Reason != nil {
	// 	assert.True(t, strings.Contains(r.Reason, *Case.ExpectResult.Reason))
	// }
	if Case.ExpectResult.RuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.RuleIndex, *r.RuleIndex)
	}
	if Case.ExpectResult.NoRuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.NoRuleIndex, r.RuleIndex == nil)
	}
	if Case.ExpectResult.Version != nil {
		assert.Equal(t, *Case.ExpectResult.Version, *r.Version)
	}
}

func assertJsonDetail(t *testing.T, Case Case, r FPJsonDetail) {
	if Case.ExpectResult.Reason != nil {
		assert.True(t, strings.Contains(r.Reason, *Case.ExpectResult.Reason))
	}
	if Case.ExpectResult.RuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.RuleIndex, *r.RuleIndex)
	}
	if Case.ExpectResult.NoRuleIndex != nil {
		assert.Equal(t, *Case.ExpectResult.NoRuleIndex, r.RuleIndex == nil)
	}
	if Case.ExpectResult.Version != nil {
		assert.Equal(t, *Case.ExpectResult.Version, *r.Version)
	}
}

func setupFeatureProbe(t *testing.T, repo Repository) *FeatureProbe {
	config := FPConfig{
		RemoteUrl: "https://featureprobe.com/",
		RefreshInterval: 1 * time.Second,
		Repo:            &repo,
	}

	fp := NewFeatureProbe(config)
	return &fp
}

type ContractTests struct {
	Tests []Scenario `json:"tests"`
}

type Scenario struct {
	Scenario string         `json:"scenario"`
	Cases    []Case         `json:"cases"`
	Fixture  RepositoryData `json:"fixture"`
}

type Case struct {
	Name         string       `json:"name"`
	User         User         `json:"user"`
	Function     Function     `json:"function"`
	ExpectResult ExpectResult `json:"expectResult"`
}

type User struct {
	Key          string     `json:"key"`
	CustomValues []KeyValue `json:"customValues"`
}

type Function struct {
	Name    string      `json:"name"`
	Toggle  string      `json:"toggle"`
	Default interface{} `json:"default"`
}

type ExpectResult struct {
	Value       interface{} `json:"value"`
	Reason      *string     `json:"reason"`
	RuleIndex   *int        `json:"ruleIndex"`
	NoRuleIndex *bool       `json:"noRuleIndex"`
	Version     *uint64     `json:"version"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
