package featureprobe

type Toggles struct {
	Toggles  map[string]Toggle  `json:"toggles"`
	Segments map[string]Segment `json:"segments,omitempty"`
}

type Toggle struct {
	Key          string                   `json:"key"`
	Enabled      bool                     `json:"enabled"`
	Version      uint64                   `json:"version"`
	ForClient    bool                     `json:"forClient"`
	DefaultServe Serve                    `json:"defaultServe"`
	Rules        []Rule                   `json:"rules"`
	Variations   []map[string]interface{} `json:"variations"`
}

type Segment struct {
	Key     string `json:"key"`
	UniqId  string `json:"uniqueId"`
	Version uint64 `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Serve struct {
	Select uint16 `json:"select,omitempty"`
	Split  Split  `json:"split,omitempty"`
}

type Rule struct {
	Serve      Serve       `json:"serve"`
	Conditions []Condition `json:"conditions"`
}

type Split struct {
	Distribution [][][]uint32 `json:"distribution"`
	BucketBy     string       `json:"bucketBy,omitempty"`
	Salt         string       `json:"salt,omitempty"`
}

type Condition struct {
	Type    string   `json:"type"`
	Subject string   `json:"subject"`
	Predict string   `json:"predict"`
	Objects []string `json:"objects"`
}
