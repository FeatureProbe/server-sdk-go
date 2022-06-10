package featureprobe

import "fmt"

type FeatureProbe struct {
	Config FPConfig
	Repo   *Repository
}

type FPConfig struct {
	RemoteUrl       string
	TogglesUrl      *string
	EventsUrl       *string
	ServerSdkKey    string
	RefreshInterval int
	WaitFirstResp   bool
}

type FPBoolDetail struct {
	Value     bool
	RuleIndex *int
	Version   *uint64
	Reason    string
}

type FPNumberDetail struct {
	Value     float64
	RuleIndex *int
	Version   *uint64
	Reason    string
}

type FPStrDetail struct {
	Value     string
	RuleIndex *int
	Version   *uint64
	Reason    string
}

type FPJsonDetail struct {
	Value     interface{}
	RuleIndex *int
	Version   *uint64
	Reason    string
}

func NewFeatureProbe(config FPConfig) (FeatureProbe, error) {
	return FeatureProbe{
		Config: config,
		Repo:   nil,
	}, nil
}

func (fp *FeatureProbe) genericValue(toggle string, user FPUser, defaultValue interface{}) interface{} {
	if fp.Repo == nil {
		return defaultValue
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return defaultValue
	}
	val, err := t.Eval(user, fp.Repo.Segments)
	if err != nil {
		return defaultValue
	}
	return val
}

func (fp *FeatureProbe) BoolValue(toggle string, user FPUser, defaultValue bool) bool {
	val := fp.genericValue(toggle, user, defaultValue)
	r, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) StrValue(toggle string, user FPUser, defaultValue string) string {
	val := fp.genericValue(toggle, user, defaultValue)
	r, ok := val.(string)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) NumberValue(toggle string, user FPUser, defaultValue float64) float64 {
	val := fp.genericValue(toggle, user, defaultValue)
	r, ok := val.(float64)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) JsonValue(toggle string, user FPUser, defaultValue interface{}) interface{} {
	val := fp.genericValue(toggle, user, defaultValue)
	return val
}

func (fp *FeatureProbe) genericDetail(toggle string, user FPUser, defaultValue interface{}) (interface{}, *int, *uint64, string) {
	value := defaultValue
	reason := fmt.Sprintf("Toggle:[%s] not exist", toggle)
	var ruleIndex *int = nil
	var version *uint64 = nil

	if fp.Repo == nil {
		return value, ruleIndex, version, reason
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return value, ruleIndex, version, reason
	}
	detail, err := t.EvalDetail(user, fp.Repo.Segments)

	ruleIndex = detail.RuleIndex
	version = detail.Version
	reason = detail.Reason

	if err != nil {
		return value, ruleIndex, version, reason
	}

	return detail.Value, ruleIndex, version, reason
}

func (fp *FeatureProbe) BoolDetail(toggle string, user FPUser, defaultValue bool) FPBoolDetail {
	value, ruleIndex, version, reason := fp.genericDetail(toggle, user, defaultValue)
	detail := FPBoolDetail{Value: defaultValue, RuleIndex: ruleIndex, Version: version, Reason: reason}

	val, ok := value.(bool)
	if !ok {
		detail.Reason = "Value type mismatch"
		return detail
	}
	detail.Value = val
	return detail
}

func (fp *FeatureProbe) StrDetail(toggle string, user FPUser, defaultValue string) FPStrDetail {
	value, ruleIndex, version, reason := fp.genericDetail(toggle, user, defaultValue)
	detail := FPStrDetail{Value: defaultValue, RuleIndex: ruleIndex, Version: version, Reason: reason}

	val, ok := value.(string)
	if !ok {
		detail.Reason = "Value type mismatch"
		return detail
	}
	detail.Value = val
	return detail
}

func (fp *FeatureProbe) NumberDetail(toggle string, user FPUser, defaultValue float64) FPNumberDetail {
	value, ruleIndex, version, reason := fp.genericDetail(toggle, user, defaultValue)
	detail := FPNumberDetail{Value: defaultValue, RuleIndex: ruleIndex, Version: version, Reason: reason}

	val, ok := value.(float64)
	if !ok {
		detail.Reason = "Value type mismatch"
		return detail
	}
	detail.Value = val
	return detail
}

func (fp *FeatureProbe) JsonDetail(toggle string, user FPUser, defaultValue interface{}) FPJsonDetail {
	value, ruleIndex, version, reason := fp.genericDetail(toggle, user, defaultValue)
	detail := FPJsonDetail{Value: value, RuleIndex: ruleIndex, Version: version, Reason: reason}
	return detail
}

func (fp *FeatureProbe) setRepoForTest(repo Repository) {
	fp.Repo = &repo
}