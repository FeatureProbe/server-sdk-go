package featureprobe

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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
	Key           string        `json:"key"`
	Enabled       bool          `json:"enabled"`
	Version       uint64        `json:"version"`
	ForClient     bool          `json:"forClient"`
	DisabledServe Serve         `json:"disabledServe"`
	DefaultServe  Serve         `json:"defaultServe"`
	Rules         []Rule        `json:"rules"`
	Variations    []interface{} `json:"variations"`
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
	Value     interface{}
	RuleIndex *int
	Version   *uint64
	Reason    string
}

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

func (t *Toggle) Eval(user FPUser, segments map[string]Segment) (interface{}, error) {
	params := evalParams{
		User:       user,
		Segments:   segments,
		Variations: t.Variations,
	}

	if !t.Enabled {
		return t.DisabledServe.SelectVariation(params)
	}

	for _, rule := range t.Rules {
		serve, err := rule.ServeVariation(params)
		if err != nil {
			return nil, err
		}
		if serve != nil {
			return serve, nil
		}
	}
	return t.DefaultServe.SelectVariation(params)
}

func (t *Toggle) EvalDetail(user FPUser, segments map[string]Segment) (EvalDetail, error) {
	params := evalParams{
		User:       user,
		Segments:   segments,
		Variations: t.Variations,
	}

	if !t.Enabled {
		serve, _ := t.DisabledServe.SelectVariation(params)
		return EvalDetail{
			Value:     serve,
			Version:   &t.Version,
			RuleIndex: nil,
			Reason:    "disabled",
		}, nil
	}

	for index, rule := range t.Rules {
		serve, err := rule.ServeVariation(params)
		if err != nil {
			return EvalDetail{
				Value:     nil,
				Version:   &t.Version,
				RuleIndex: &index,
				Reason:    err.Error(),
			}, err
		}
		if serve != nil {
			return EvalDetail{
				Value:     serve,
				RuleIndex: &index,
				Version:   &t.Version,
				Reason:    fmt.Sprintf("rule %d ", index),
			}, nil
		}
	}

	serve, err := t.DefaultServe.SelectVariation(params)
	if err != nil {
		return EvalDetail{
			Value:     nil,
			RuleIndex: nil,
			Version:   &t.Version,
			Reason:    err.Error(),
		}, err
	}
	return EvalDetail{
		Value:     serve,
		RuleIndex: nil,
		Version:   &t.Version,
		Reason:    "default",
	}, nil
}

func (s *Serve) SelectVariation(params evalParams) (interface{}, error) {
	var index int
	if s.Select != nil {
		index = *s.Select
	} else {
		i, err := s.Split.FindIndex(params)
		if err != nil {
			return nil, err
		}
		index = i
	}

	len := len(params.Variations)
	if index >= len {
		return nil, fmt.Errorf("index %d overflow, variations count is %d", index, len)
	}
	return params.Variations[index], nil
}

func (s *Split) FindIndex(params evalParams) (int, error) {
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
	fmt.Println(bucketIndex)

	variation := -1
	for v, d := range s.Distribution {
		for _, r := range d {
			if r.Lower <= bucketIndex && r.Upper > bucketIndex {
				variation = v
			}
		}
	}

	return variation, nil
}

func (s *Split) hashKey(params evalParams) (string, error) {
	var hashKey string
	user := params.User
	if len(s.BucketBy) == 0 {
		hashKey = user.Key
	} else {
		bucketBy := s.BucketBy
		key := user.Get(bucketBy)
		if len(key) != 0 {
			hashKey = key
		} else {
			return "", fmt.Errorf("user with id: %s does not have attribute named: [%s]", user.Key, key)
		}
	}
	return hashKey, nil
}

func (r *Rule) ServeVariation(params evalParams) (interface{}, error) {
	for _, c := range r.Conditions {
		if !c.Meet(params.User, params.Segments) {
			return nil, nil
		}
	}
	return r.Serve.SelectVariation(params)
}

func (c *Condition) Meet(user FPUser, segments map[string]Segment) bool {
	if c.Type == "string" {
		if c.MatchStringCondition(user, c.Predicate) {
			return true
		}
	}
	// TODO: type segments
	return false
}

func (c *Condition) MatchStringCondition(user FPUser, predict string) bool {
	customValue := user.Get(c.Subject)
	if len(customValue) == 0 {
		return false
	}

	switch {
	case predict == "is one of":
		return c.MatchObjects(func(o string) bool { return customValue == o })
	case predict == "starts with":
		return c.MatchObjects(func(o string) bool { return strings.HasPrefix(customValue, o) })
	case predict == "ends with":
		return c.MatchObjects(func(o string) bool { return strings.HasSuffix(customValue, o) })
	case predict == "contains":
		return c.MatchObjects(func(o string) bool { return strings.Contains(customValue, o) })
	case predict == "matches regex":
		return c.MatchObjects(func(o string) bool {
			matched, err := regexp.Match(o, []byte(customValue))
			if err != nil {
				return false
			}
			return matched
		})
	case predict == "is not any of":
		return !c.MatchStringCondition(user, "is one of")
	case predict == "does not start with":
		return !c.MatchStringCondition(user, "starts with")
	case predict == "does not end with":
		return !c.MatchStringCondition(user, "ends with")
	case predict == "does not contain":
		return !c.MatchStringCondition(user, "contains")
	case predict == "does not match regex":
		return !c.MatchStringCondition(user, "matches regex")
	}

	return false
}

func (c *Condition) MatchObjects(f func(string) bool) bool {
	for _, o := range c.Objects {
		if f(o) {
			return true
		}
	}
	return false
}
