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

func (fp *FeatureProbe) BoolValue(toggle string, user FPUser, defaultValue bool) bool {
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
	r, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) StrValue(toggle string, user FPUser, defaultValue string) string {
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
	r, ok := val.(string)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) NumberValue(toggle string, user FPUser, defaultValue float64) float64 {
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
	r, ok := val.(float64)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) JsonValue(toggle string, user FPUser, defaultValue interface{}) interface{} {
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

func (fp *FeatureProbe) BoolDetail(toggle string, user FPUser, defaultValue bool) FPBoolDetail {
	notExist := FPBoolDetail{
		Value:     defaultValue,
		RuleIndex: nil,
		Version:   nil,
		Reason:    fmt.Sprintf("Toggle:[%s] not exist", toggle),
	}
	if fp.Repo == nil {
		return notExist
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return notExist
	}
	detail, err := t.EvalDetail(user, fp.Repo.Segments)
	r := FPBoolDetail{
		Value:     defaultValue,
		RuleIndex: detail.RuleIndex,
		Version:   detail.Version,
		Reason:    detail.Reason,
	}
	if err != nil {
		return r
	}
	val, ok := detail.Value.(bool)
	if !ok {
		r.Reason = "Value type mismatch"
		return r
	}
	r.Value = val
	return r
}

func (fp *FeatureProbe) StrDetail(toggle string, user FPUser, defaultValue string) FPStrDetail {
	notExist := FPStrDetail{
		Value:     defaultValue,
		RuleIndex: nil,
		Version:   nil,
		Reason:    fmt.Sprintf("Toggle:[%s] not exist", toggle),
	}
	if fp.Repo == nil {
		return notExist
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return notExist
	}
	detail, err := t.EvalDetail(user, fp.Repo.Segments)
	r := FPStrDetail{
		Value:     defaultValue,
		RuleIndex: detail.RuleIndex,
		Version:   detail.Version,
		Reason:    detail.Reason,
	}
	if err != nil {
		return r
	}
	val, ok := detail.Value.(string)
	if !ok {
		r.Reason = "Value type mismatch"
		return r
	}
	r.Value = val
	return r
}

func (fp *FeatureProbe) NumberDetail(toggle string, user FPUser, defaultValue float64) FPNumberDetail {
	notExist := FPNumberDetail{
		Value:     defaultValue,
		RuleIndex: nil,
		Version:   nil,
		Reason:    fmt.Sprintf("Toggle:[%s] not exist", toggle),
	}
	if fp.Repo == nil {
		return notExist
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return notExist
	}
	detail, err := t.EvalDetail(user, fp.Repo.Segments)
	r := FPNumberDetail{
		Value:     defaultValue,
		RuleIndex: detail.RuleIndex,
		Version:   detail.Version,
		Reason:    detail.Reason,
	}
	if err != nil {
		return r
	}
	val, ok := detail.Value.(float64)
	if !ok {
		r.Reason = "Value type mismatch"
		return r
	}
	r.Value = val
	return r
}

func (fp *FeatureProbe) JsonDetail(toggle string, user FPUser, defaultValue interface{}) FPJsonDetail {
	notExist := FPJsonDetail{
		Value:     defaultValue,
		RuleIndex: nil,
		Version:   nil,
		Reason:    fmt.Sprintf("Toggle:[%s] not exist", toggle),
	}
	if fp.Repo == nil {
		return notExist
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return notExist
	}
	detail, err := t.EvalDetail(user, fp.Repo.Segments)
	r := FPJsonDetail{
		Value:     defaultValue,
		RuleIndex: detail.RuleIndex,
		Version:   detail.Version,
		Reason:    detail.Reason,
	}
	if err != nil {
		return r
	}
	r.Value = detail.Value
	return r
}

func (fp *FeatureProbe) setRepoForTest(repo Repository) {
	fp.Repo = &repo
}
