package featureprobe

import (
	"context"
	"fmt"
	socketio "github.com/socket-iox/socket-io-client-go"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var VERSION string = "1.1.0"
var USER_AGENT string = "Go/" + VERSION

type FeatureProbe struct {
	Config   FPConfig
	Repo     *Repository
	Syncer   *Synchronizer
	Socket   *socketio.Client
	Recorder *EventRecorder
}

type FPConfig struct {
	RemoteUrl            string
	TogglesUrl           string
	EventsUrl            string
	RealtimeUrl          string
	ServerSdkKey         string
	RefreshInterval      time.Duration
	StartWait            time.Duration
	Repo                 *Repository
	MaxPrerequisitesDeep int
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

func NewFeatureProbe(config FPConfig) FeatureProbe {
	ready := make(chan struct{}, 1)
	setServerUrls(&config)
	timeout := config.RefreshInterval
	eventRecorder := NewEventRecorder(config.EventsUrl, timeout, config.ServerSdkKey)
	eventRecorder.Start()

	//setup realtime connection
	u, err := url.Parse(config.RealtimeUrl)
	var socket *socketio.Client
	if err == nil {
		s := socketio.Client{NameSpace: &u.Path}
		socket = &s
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), config.StartWait)
	defer cancelFunc()
	toggleSyncer := Synchronizer{}
	repo := Repository{}
	if config.Repo == nil {
		toggleSyncer = NewSynchronizer(config.TogglesUrl, config.RefreshInterval, config.ServerSdkKey, &repo)
	} else {
		repo = *config.Repo
		toggleSyncer = NewCustomRepoSynchronizer(config.Repo)
	}
	toggleSyncer.Start(ready)
	if config.MaxPrerequisitesDeep == 0 {
		config.MaxPrerequisitesDeep = 20
	}
	client := FeatureProbe{
		Config:   config,
		Repo:     &repo,
		Syncer:   &toggleSyncer,
		Recorder: &eventRecorder,
		Socket:   socket,
	}

	go client.connectSocket()

	if config.StartWait > 0 {
		for {
			select {
			case <-ready:
				return client
			case <-ctx.Done():
				go func() { <-ready }()
				// log. timeout encountered waiting for FeatureProbe client initialization
				return client
			}
		}
	}
	go func() { <-ready }()
	return client
}

func setServerUrls(config *FPConfig) {
	if !strings.HasSuffix(config.RemoteUrl, "/") {
		config.RemoteUrl += "/"
	}
	if len(config.EventsUrl) == 0 {
		config.EventsUrl = config.RemoteUrl + "api/events"
	}
	if len(config.RealtimeUrl) == 0 {
		config.RealtimeUrl = config.RemoteUrl + "realtime"
	}
	if len(config.TogglesUrl) == 0 {
		config.TogglesUrl = config.RemoteUrl + "api/server-sdk/toggles"
	}
}

func newToggleForTest(key string, value interface{}) Toggle {
	s := 0
	return Toggle{
		Key:           key,
		Enabled:       true,
		DefaultServe:  Serve{Select: &s},
		DisabledServe: Serve{Select: &s},
		Version:       0,
		ForClient:     false,
		Variations:    []interface{}{value},
		Rules:         []Rule{},
	}
}

func NewFeatureProbeForTest(toggles map[string]interface{}) FeatureProbe {
	repo := Repository{}
	repo.Toggles = map[string]Toggle{}
	for key, value := range toggles {
		repo.Toggles[key] = newToggleForTest(key, value)
	}
	return FeatureProbe{
		Repo: &repo,
	}
}

func (fp *FeatureProbe) BoolValue(toggle string, user FPUser, defaultValue bool) bool {
	val, _, _, _ := fp.genericDetail(toggle, user, defaultValue)
	r, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) StrValue(toggle string, user FPUser, defaultValue string) string {
	val, _, _, _ := fp.genericDetail(toggle, user, defaultValue)
	r, ok := val.(string)
	if !ok {
		return defaultValue
	}
	return r
}

func (fp *FeatureProbe) NumberValue(toggle string, user FPUser, defaultValue float64) float64 {
	val, _, _, _ := fp.genericDetail(toggle, user, defaultValue)
	i, ok := val.(int)
	if ok {
		return float64(i)
	}
	f, ok := val.(float64)
	if !ok {
		return defaultValue
	}
	return f
}

func (fp *FeatureProbe) JsonValue(toggle string, user FPUser, defaultValue interface{}) interface{} {
	val, _, _, _ := fp.genericDetail(toggle, user, defaultValue)
	return val
}

func (fp *FeatureProbe) Track(eventName string, user FPUser, value *float64) {
	if fp.Recorder != nil {
		fp.Recorder.RecordCustom(CustomEvent{
			Kind:  "custom",
			Time:  time.Now().UnixNano() / 1e6,
			User:  user.Key(),
			Name:  eventName,
			Value: value,
		})
	}
}

func (fp *FeatureProbe) genericDetail(toggle string, user FPUser, defaultValue interface{}) (interface{}, *int, *uint64, string) {
	reason := fmt.Sprintf("Toggle:[%s] not exist", toggle)
	var ruleIndex *int = nil
	var version *uint64 = nil
	var variationIndex *int = nil

	if fp.Repo == nil {
		return defaultValue, ruleIndex, version, reason
	}
	t, ok := fp.Repo.Toggles[toggle]
	if !ok {
		return defaultValue, ruleIndex, version, reason
	}
	detail, _ := t.evalDetail(user, fp.Repo.Toggles, fp.Repo.Segments, defaultValue, fp.Config.MaxPrerequisitesDeep)

	variationIndex = detail.VariationIndex
	ruleIndex = detail.RuleIndex
	version = detail.Version
	reason = detail.Reason

	if fp.Recorder != nil && variationIndex != nil {
		fp.trackEvent(t, user, detail)
	}
	return detail.Value, ruleIndex, version, reason
}

func (fp *FeatureProbe) trackEvent(toggle Toggle, user FPUser, evalDetail EvalDetail) {
	nowTime := time.Now().UnixNano() / 1e6
	fp.Recorder.RecordAccess(AccessEvent{
		Kind:           "access",
		Time:           nowTime,
		User:           user.Key(),
		Key:            toggle.Key,
		Value:          evalDetail.Value,
		VariationIndex: evalDetail.VariationIndex,
		Version:        evalDetail.Version,
	}, toggle.TrackAccessEvents)

	if fp.Repo.DebugUntilTime > 0 && fp.Repo.DebugUntilTime >= uint64(nowTime) {
		fp.Recorder.RecordDebugAccess(DebugEvent{
			Kind:           "debug",
			Time:           nowTime,
			User:           user.Key(),
			Key:            toggle.Key,
			UserDetail:     user.ToMap(),
			Value:          evalDetail.Value,
			VariationIndex: evalDetail.VariationIndex,
			RuleIndex:      evalDetail.RuleIndex,
			Version:        evalDetail.Version,
			Reason:         evalDetail.Reason,
		})
	}
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

func newHttpClient(timeout time.Duration) http.Client {
	return http.Client{
		Timeout: timeout * time.Millisecond,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// Initialized return false means not successfully fetch remote resource
func (fp *FeatureProbe) Initialized() bool {
	return fp.Syncer.Initialized()
}

func (fp *FeatureProbe) Close() {
	if fp.Syncer != nil {
		fp.Syncer.Stop()
	}
	if fp.Repo != nil {
		fp.Repo.Clear()
	}
	if fp.Recorder != nil {
		fp.Recorder.Stop()
	}
}

func (fp *FeatureProbe) connectSocket() {
	url := fp.Config.RealtimeUrl
	client := fp.Socket
	client.On("connect", func(client *socketio.Client, data []string) {
		client.Emit("register", map[string]string{"key": fp.Config.ServerSdkKey})
	})

	client.On("update", func(client *socketio.Client, data []string) {
		fp.Syncer.FetchRemoteRepo()
	})

	if err := client.Connect(url, "websocket"); err != nil {
		fmt.Printf("realtime socket connect err: %s\n", err)
	}
}
