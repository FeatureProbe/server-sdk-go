package featureprobe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type EventRecorder struct {
	auth           string
	eventsUrl      string
	flushInterval  time.Duration
	incomingEvents []interface{}
	access         Access
	httpClient     http.Client
	mu             sync.Mutex
	wg             sync.WaitGroup
	startOnce      sync.Once
	stopOnce       sync.Once
	stopChan       chan struct{}
	ticker         *time.Ticker
}

type AccessEvent struct {
	Kind              string      `json:"kind"`
	Time              int64       `json:"time"`
	User              string      `json:"user"`
	Key               string      `json:"key"`
	Value             interface{} `json:"value"`
	VariationIndex    *int        `json:"variationIndex"`
	RuleIndex         *int        `json:"ruleIndex"`
	Version           *uint64     `json:"version"`
	Reason            string      `json:"reason"`
	TrackAccessEvents bool        `json:"-"`
}

type CustomEvent struct {
	Kind  string   `json:"kind"`
	Time  int64    `json:"time"`
	User  string   `json:"user"`
	Name  string   `json:"name"`
	Value *float64 `json:"value"`
}

type PackedData struct {
	Events []interface{} `json:"events"`
	Access Access        `json:"access"`
}

type Access struct {
	StartTime int64                      `json:"startTime"`
	EndTime   int64                      `json:"endTime"`
	Counters  map[string][]ToggleCounter `json:"counters"`
}

type ToggleCounter struct {
	Value   interface{} `json:"value"`
	Version *uint64     `json:"version"`
	Index   *int        `json:"index"`
	Count   int         `json:"count"`
}

type Variation struct {
	Key     string  `json:"key"`
	Index   *int    `json:"index"`
	Version *uint64 `json:"version"`
}

type CountValue struct {
	Count int         `json:"count"`
	Value interface{} `json:"value"`
}

func NewEventRecorder(eventsUrl string, flushInterval time.Duration, auth string) EventRecorder {
	return EventRecorder{
		auth:           auth,
		eventsUrl:      eventsUrl,
		flushInterval:  flushInterval,
		incomingEvents: []interface{}{},
		access:         newAccess(),
		httpClient:     newHttpClient(flushInterval),
		stopChan:       make(chan struct{}),
	}
}

func newAccess() Access {
	return Access{
		Counters: make(map[string][]ToggleCounter),
	}
}

func nowToggleCounter(value interface{}, version *uint64, index *int) ToggleCounter {
	return ToggleCounter{
		value,
		version,
		index,
		1,
	}
}

func (e *EventRecorder) Start() {
	e.wg.Add(1)
	e.startOnce.Do(func() {
		e.ticker = time.NewTicker(e.flushInterval)
		go func() {
			for {
				select {
				case <-e.stopChan:
					e.doFlush()
					e.wg.Done()
					return
				case <-e.ticker.C:
					e.doFlush()
				}
			}
		}()
	})
}

func (e *EventRecorder) doFlush() {
	events := make([]interface{}, 0)
	e.mu.Lock()
	events, e.incomingEvents = e.incomingEvents, events
	packedData := e.buildPackedData(events)
	e.access = newAccess()
	e.mu.Unlock()
	if len(events) == 0 && len(packedData[0].Access.Counters) == 0 {
		return
	}
	body, _ := json.Marshal(packedData)
	fmt.Println(string(body))
	req, err := http.NewRequest(http.MethodPost, e.eventsUrl, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	req.Header.Add("Authorization", e.auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", USER_AGENT)
	_, err = e.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Report event fails: %s\n", err)
	}
}

func (e *EventRecorder) buildPackedData(events []interface{}) []PackedData {
	e.access.EndTime = time.Now().UnixNano() / 1e6
	p := PackedData{Access: e.access, Events: events}
	return []PackedData{p}
}

func (e *EventRecorder) addAccess(event AccessEvent) {
	if len(e.access.Counters) == 0 {
		e.access.StartTime = time.Now().UnixNano() / 1e6
	}
	counters, ok := e.access.Counters[event.Key]
	if ok {
		for index, counter := range counters {
			if *counter.Version == *event.Version && *counter.Index == *event.VariationIndex {
				counters[index].Count = counter.Count + 1
				return
			}
		}
		e.access.Counters[event.Key] = append(counters,
			nowToggleCounter(event.Value, event.Version, event.VariationIndex))
	} else {
		e.access.Counters[event.Key] = []ToggleCounter{
			nowToggleCounter(event.Value, event.Version, event.VariationIndex)}
	}
}

func (e *EventRecorder) RecordAccess(event AccessEvent) {
	e.mu.Lock()
	if event.TrackAccessEvents {
		e.incomingEvents = append(e.incomingEvents, event)
	}
	e.addAccess(event)
	e.mu.Unlock()
}

func (e *EventRecorder) RecordCustom(event CustomEvent) {
	e.mu.Lock()
	e.incomingEvents = append(e.incomingEvents, event)
	e.mu.Unlock()
}

func (e *EventRecorder) Stop() {
	if e.stopChan != nil {
		e.stopOnce.Do(func() {
			close(e.stopChan)
		})
	}
	e.wg.Wait()
}
