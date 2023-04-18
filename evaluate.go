package featureprobe

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/masterminds/semver"
)

type Repository struct {
	Toggles  map[string]Toggle  `json:"toggles"`
	Segments map[string]Segment `json:"segments"`
}

type Toggles struct {
	Toggles  map[string]Toggle  `json:"toggles"`
	Segments map[string]Segment `json:"segments,omitempty"`
}

type Toggle struct {
	Key               string         `json:"key"`
	Enabled           bool           `json:"enabled"`
	TrackAccessEvents bool           `json:"trackAccessEvents"`
	LastModified      uint64         `json:"lastModified"`
	Version           uint64         `json:"version"`
	ForClient         bool           `json:"forClient"`
	DisabledServe     Serve          `json:"disabledServe"`
	DefaultServe      Serve          `json:"defaultServe"`
	Rules             []Rule         `json:"rules"`
	Variations        []interface{}  `json:"variations"`
	Prerequisites     []Prerequisite `json:"prerequisites"`
}

type Segment struct {
	Key     string `json:"key"`
	UniqId  string `json:"uniqueId"`
	Version uint64 `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Serve struct {
	Select *int   `json:"select,omitempty"`
	Split  *Split `json:"split,omitempty"`
}

type Rule struct {
	Serve      Serve       `json:"serve"`
	Conditions []Condition `json:"conditions"`
}

type Split struct {
	Distribution [][]Range `json:"distribution"`
	BucketBy     string    `json:"bucketBy,omitempty"`
	Salt         string    `json:"salt,omitempty"`
}

type Range struct {
	Lower int `json:"-"`
	Upper int `json:"-"`
}

type Condition struct {
	Type      string   `json:"type"`
	Subject   string   `json:"subject"`
	Predicate string   `json:"predicate"`
	Objects   []string `json:"objects"`
}

type evalParams struct {
	Key        string
	IsDetail   bool
	User       FPUser
	Variations []interface{}
	Segments   map[string]Segment
}

type EvalDetail struct {
	Value          interface{}
	RuleIndex      *int
	VariationIndex *int
	Version        *uint64
	Reason         string
}

type Prerequisite struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

var (
	ErrPrerequisiteNotExist     = errors.New("prerequisite toggle not exist")
	ErrPrerequisiteDeepOverflow = errors.New("prerequisite deep overflow")
)

func saltHash(key string, salt string, bucketSize uint32) int {
	h := sha1.New()
	h.Write([]byte(key + salt))
	bytes := h.Sum(nil)
	size := len(bytes)
	value := binary.BigEndian.Uint32(bytes[size-4 : size])
	// avoid negative number mod
	mod := int64(value) % int64(bucketSize)
	return int(mod)
}

func (r *Range) UnmarshalJSON(data []byte) error {
	var raw []int
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("invalid distribution range")
	}
	*r = Range{
		Lower: raw[0],
		Upper: raw[1],
	}

	return nil
}

func (t *Toggle) Eval(user FPUser, toggles map[string]Toggle, segments map[string]Segment, defaultValue interface{}, deep int) (interface{}, error) {
	detail, err := t.evalDetail(user, toggles, segments, defaultValue, deep)
	return detail.Value, err
}

func (t *Toggle) evalDetail(user FPUser, toggles map[string]Toggle, segments map[string]Segment, defaultValue interface{}, deep int) (EvalDetail, error) {
	detail, err := t.doEvalDetail(user, toggles, segments, defaultValue, deep)
	if err == nil {
		return detail, nil
	}
	if err == ErrPrerequisiteDeepOverflow || err == ErrPrerequisiteNotExist {
		detail, err = t.createDefaultEvalDetail(evalParams{
			User:       user,
			Segments:   segments,
			Variations: t.Variations,
			Key:        t.Key,
		}, defaultValue)
		if err != nil {
			detail.Reason = err.Error()
		}
		return detail, nil
	}
	return detail, err
}

func (t *Toggle) prerequisite(user FPUser, toggles map[string]Toggle, segments map[string]Segment, defaultValue interface{}, deep int) (bool, error) {
	if t.Prerequisites == nil && len(t.Prerequisites) == 0 {
		return true, nil
	}
	for _, prerequisite := range t.Prerequisites {
		toggle, exists := toggles[prerequisite.Key]
		if !exists {
			return false, ErrPrerequisiteNotExist
		}
		result, err := toggle.doEvalDetail(user, toggles, segments, defaultValue, deep-1)
		if err != nil {
			return false, err
		}
		if result.Value == nil || fmt.Sprintf("%v", result.Value) != fmt.Sprintf("%v", prerequisite.Value) {
			return false, nil
		}
	}
	return true, nil
}

func (t *Toggle) doEvalDetail(user FPUser, toggles map[string]Toggle, segments map[string]Segment, defaultValue interface{}, deep int) (EvalDetail, error) {
	if deep <= 0 {
		return t.buildEvalDetail(defaultValue, nil, nil, ""), ErrPrerequisiteDeepOverflow
	}
	params := evalParams{
		User:       user,
		Segments:   segments,
		Variations: t.Variations,
		Key:        t.Key,
	}
	if !t.Enabled {
		serve, index, err := t.DisabledServe.selectVariation(params)
		if err != nil {
			return t.buildEvalDetail(defaultValue, nil, nil, err.Error()), err
		}
		return t.buildEvalDetail(serve, nil, index, "disabled"), nil
	}
	match, err := t.prerequisite(user, toggles, segments, defaultValue, deep)
	if err != nil {
		return t.buildEvalDetail(defaultValue, nil, nil, ""), err
	}
	if !match {
		return t.createDefaultEvalDetail(params, defaultValue)
	}
	for ruleIndex, rule := range t.Rules {
		serve, vi, err := rule.serveVariation(params)
		if err != nil {
			return t.buildEvalDetail(defaultValue, &ruleIndex, nil, err.Error()), err
		}
		if serve != nil {
			return t.buildEvalDetail(serve, &ruleIndex, vi, fmt.Sprintf("rule %d ", ruleIndex)), nil
		}
	}
	return t.createDefaultEvalDetail(params, defaultValue)
}

func (t *Toggle) createDefaultEvalDetail(params evalParams, defaultValue interface{}) (EvalDetail, error) {
	serve, vi, err := t.DefaultServe.selectVariation(params)
	if err != nil {
		return t.buildEvalDetail(defaultValue, nil, nil, err.Error()), err
	}
	return t.buildEvalDetail(serve, nil, vi, "default"), nil
}

func (t *Toggle) buildEvalDetail(value interface{}, ruleIndex *int, variationIndex *int, reason string) EvalDetail {
	return EvalDetail{
		Value:          value,
		VariationIndex: variationIndex,
		RuleIndex:      ruleIndex,
		Version:        &t.Version,
		Reason:         reason,
	}

}

func (s *Serve) selectVariation(params evalParams) (interface{}, *int, error) {
	var index *int = nil
	if s.Select != nil {
		index = s.Select
	} else {
		i, err := s.Split.findIndex(params)
		if err != nil {
			return nil, nil, err
		}
		index = &i
	}

	length := len(params.Variations)
	if *index >= length {
		return nil, nil, fmt.Errorf("index %d overflow, variations count is %d", index, length)
	}
	return params.Variations[*index], index, nil
}

func (s *Split) findIndex(params evalParams) (int, error) {
	hashKey, err := s.hashKey(params)
	if err != nil {
		return -1, err
	}

	var salt string
	if len(s.Salt) == 0 {
		salt = params.Key
	} else {
		salt = s.Salt
	}

	bucketIndex := saltHash(hashKey, salt, 10000)

	variation := s.getVariation(bucketIndex)

	if variation == -1 {
		return variation, fmt.Errorf("not find hash_bucket in distribution")
	}

	return variation, nil
}

func (s *Split) getVariation(bucketIndex int) int {
	for v, d := range s.Distribution {
		for _, r := range d {
			if r.Lower <= bucketIndex && bucketIndex < r.Upper {
				return v
			}
		}
	}
	return -1
}

func (s *Split) hashKey(params evalParams) (string, error) {
	var hashKey string
	user := params.User
	if len(s.BucketBy) == 0 {
		hashKey = user.Key()
	} else {
		bucketBy := s.BucketBy
		key := user.Get(bucketBy)
		if len(key) != 0 {
			hashKey = key
		} else {
			return "", fmt.Errorf("user with id: %s does not have attribute named: [%s]", user.Key(), key)
		}
	}
	return hashKey, nil
}

func (r *Rule) serveVariation(params evalParams) (interface{}, *int, error) {
	for _, c := range r.Conditions {
		if !c.meet(params.User, params.Segments) {
			return nil, nil, nil
		}
	}
	return r.Serve.selectVariation(params)
}

func (c *Condition) meet(user FPUser, segments map[string]Segment) bool {
	switch c.Type {
	case "string":
		return c.matchStringCondition(user, c.Predicate)
	case "segment":
		return c.matchSegmentCondition(user, c.Predicate, segments)
	case "datetime":
		return c.matchDatetimeCondition(user, c.Predicate)
	case "semver":
		return c.matchSemverCondition(user, c.Predicate)
	case "number":
		return c.matchNumberCondition(user, c.Predicate)
	}

	return false
}

func (c *Condition) matchStringCondition(user FPUser, predicate string) bool {
	customValue := user.Get(c.Subject)
	if len(customValue) == 0 {
		return false
	}

	switch predicate {
	case "is one of":
		return c.matchObjects(func(o string) bool { return customValue == o })
	case "starts with":
		return c.matchObjects(func(o string) bool { return strings.HasPrefix(customValue, o) })
	case "ends with":
		return c.matchObjects(func(o string) bool { return strings.HasSuffix(customValue, o) })
	case "contains":
		return c.matchObjects(func(o string) bool { return strings.Contains(customValue, o) })
	case "matches regex":
		return c.matchObjects(func(o string) bool {
			matched, err := regexp.Match(o, []byte(customValue))
			if err != nil {
				return false
			}
			return matched
		})
	case "is not any of":
		return !c.matchStringCondition(user, "is one of")
	case "does not start with":
		return !c.matchStringCondition(user, "starts with")
	case "does not end with":
		return !c.matchStringCondition(user, "ends with")
	case "does not contain":
		return !c.matchStringCondition(user, "contains")
	case "does not match regex":
		return !c.matchStringCondition(user, "matches regex")
	}

	return false
}

func (c *Condition) matchSegmentCondition(user FPUser, predicate string, segments map[string]Segment) bool {
	if segments == nil {
		return false
	}
	switch predicate {
	case "is in":
		return c.userInSegments(user, segments)
	case "is not in":
		return !c.userInSegments(user, segments)
	}
	return false
}

func (c *Condition) userDatetime(user FPUser) (int64, error) {
	customValue := user.Get(c.Subject)
	if len(customValue) == 0 {
		return time.Now().Unix(), nil
	}
	return strconv.ParseInt(customValue, 10, 64)
}

func (c *Condition) matchDatetimeCondition(user FPUser, predicate string) bool {
	cv, err := c.userDatetime(user)
	if err != nil {
		return false
	}
	switch predicate {
	case "after":
		return c.matchDatetimeObjects(func(o int64) bool { return cv >= o })
	case "before":
		return c.matchDatetimeObjects(func(o int64) bool { return cv < o })
	}
	return false
}

func (c *Condition) matchSemverCondition(user FPUser, predicate string) bool {
	customValue := user.Get(c.Subject)
	if len(customValue) == 0 {
		return false
	}
	cv, err := semver.NewVersion(customValue)
	if err != nil {
		return false
	}

	switch predicate {
	case "=":
		return c.matchSemVerObjects(func(o *semver.Version) bool { return cv.Equal(o) })
	case "!=":
		return !c.matchSemverCondition(user, "=")
	case ">":
		return c.matchSemVerObjects(func(o *semver.Version) bool { return cv.GreaterThan(o) })
	case ">=":
		return c.matchSemVerObjects(func(o *semver.Version) bool { return cv.GreaterThan(o) || cv.Equal(o) })
	case "<":
		return c.matchSemVerObjects(func(o *semver.Version) bool { return cv.LessThan(o) })
	case "<=":
		return c.matchSemVerObjects(func(o *semver.Version) bool { return cv.LessThan(o) || cv.Equal(o) })
	}

	return false

}

func (c *Condition) matchNumberCondition(user FPUser, predicate string) bool {
	customValue := user.Get(c.Subject)
	if len(customValue) == 0 {
		return false
	}
	cv, err := strconv.ParseFloat(customValue, 32)
	if err != nil {
		return false
	}

	switch predicate {
	case "=":
		return c.matchNumberObjects(func(o float64) bool { return cv == o })
	case "!=":
		return !c.matchNumberCondition(user, "=")
	case ">":
		return c.matchNumberObjects(func(o float64) bool { return cv > o })
	case ">=":
		return c.matchNumberObjects(func(o float64) bool { return cv >= o })
	case "<":
		return c.matchNumberObjects(func(o float64) bool { return cv < o })
	case "<=":
		return c.matchNumberObjects(func(o float64) bool { return cv <= o })
	}

	return false
}

func (c *Condition) userInSegments(user FPUser, segments map[string]Segment) bool {
	for _, segmentKey := range c.Objects {
		segment, ok := segments[segmentKey]
		if ok {
			if segment.contains(user) {
				return true
			}
		}
	}
	return false
}

func (c *Condition) matchObjects(f func(string) bool) bool {
	for _, o := range c.Objects {
		if f(o) {
			return true
		}
	}
	return false
}

func (c *Condition) matchDatetimeObjects(f func(int64) bool) bool {
	for _, o := range c.Objects {
		co, err := strconv.ParseInt(o, 10, 64)
		if err != nil {
			return false
		}
		if f(co) {
			return true
		}
	}
	return false
}

func (c *Condition) matchNumberObjects(f func(float64) bool) bool {
	for _, o := range c.Objects {
		co, err := strconv.ParseFloat(o, 32)
		if err != nil {
			return false
		}
		if f(co) {
			return true
		}
	}
	return false
}

func (c *Condition) matchSemVerObjects(f func(*semver.Version) bool) bool {
	for _, o := range c.Objects {
		co, err := semver.NewVersion(o)
		if err != nil {
			return false
		}
		if f(co) {
			return true
		}
	}
	return false
}

func (s *Segment) contains(user FPUser) bool {
	for _, rule := range s.Rules {
		if rule.allow(user) {
			return true
		}
	}
	return false
}

func (r *Rule) allow(user FPUser) bool {
	for _, condition := range r.Conditions {
		if condition.meet(user, nil) {
			return true
		}
	}
	return false
}

func (repo *Repository) Clear() {
	repo.Toggles = make(map[string]Toggle)
	repo.Segments = make(map[string]Segment)
}
