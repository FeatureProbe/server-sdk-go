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
	flushMs        time.Duration
	capacity       int64
	incomingEvents []AccessEvent
	packedData     []PackedData
	httpClient     http.Client
	mu             sync.Mutex
	once           sync.Once
}

type AccessEvent struct {
	Time    int64       `json:"time"`
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Index   *int64      `json:"index,omitempty"`
	Version *int64      `json:"version,omitempty"`
	Reason  string      `json:"reason"`
}

type PackedData struct {
	Events []AccessEvent `json:"events"`
	Access Access        `json:"access"`
}

type Access struct {
	StartTime int64                      `json:"startTime"`
	EndTime   int64                      `json:"endTime"`
	Counters  map[string][]ToggleCounter `json:"counters"`
}

type ToggleCounter struct {
	Value   interface{} `json:"Value"`
	Version *int64      `json:"version,omitempty"`
	Index   *int64      `json:"index,omitempty"`
	Count   int64       `json:"Count"`
}

type Variation struct {
	Key     string `json:"key"`
	Index   *int64 `json:"index"`
	Version *int64 `json:"version"`
}

type CountValue struct {
	Count int64       `json:"count"`
	Value interface{} `json:"value"`
}

func NewEventRecorder(eventsUrl string, flushMs time.Duration, auth string) EventRecorder {
	return EventRecorder{
		auth:           auth,
		eventsUrl:      eventsUrl,
		flushMs:        flushMs,
		incomingEvents: []AccessEvent{},
		packedData:     []PackedData{},
		httpClient:     newHttpClient(flushMs),
	}
}

func (e *EventRecorder) Start() {
	e.once.Do(func() {
		go e.doFlush()
	})
}

func (e *EventRecorder) doFlush() {
	for {
		events := make([]AccessEvent, 0)
		e.mu.Lock()
		events, e.incomingEvents = e.incomingEvents, events
		e.mu.Unlock()

		if len(events) != 0 {
			packedData := e.buildPackedData(events)
			body, _ := json.Marshal(packedData)

			req, err := http.NewRequest(http.MethodPost, e.eventsUrl, bytes.NewBuffer(body))
			if err != nil {
				fmt.Printf("%s\n", err)
				break
			}
			req.Header.Add("Authorization", e.auth)
			req.Header.Set("Content-Type", "application/json")
			e.mu.Lock()
			_, _ = e.httpClient.Do(req)
			e.mu.Unlock()
		}

		time.Sleep(e.flushMs * time.Millisecond)
	}
}

func (e *EventRecorder) buildPackedData(events []AccessEvent) []PackedData {
	access := e.buildAccess(events)
	p := PackedData{Access: access, Events: events}
	return []PackedData{p}
}

func (e *EventRecorder) buildAccess(events []AccessEvent) Access {
	counters, startTime, endTime := e.buildCounters(events)
	access := Access{
		StartTime: startTime,
		EndTime:   endTime,
		Counters:  map[string][]ToggleCounter{},
	}

	for k, v := range counters {
		counter := ToggleCounter{
			Index:   k.Index,
			Version: k.Version,
			Count:   v.Count,
			Value:   v.Value,
		}
		c, ok := access.Counters[k.Key]
		if !ok {
			access.Counters[k.Key] = []ToggleCounter{counter}
		} else {
			access.Counters[k.Key] = append(c, counter)
		}
	}
	return access
}

func (e *EventRecorder) buildCounters(events []AccessEvent) (map[Variation]CountValue, int64, int64) {
	var startTime *int64 = nil
	var endTime *int64 = nil
	counters := map[Variation]CountValue{}

	for _, event := range events {
		if startTime == nil || *startTime < event.Time {
			startTime = &event.Time
		}
		if endTime == nil || *endTime > event.Time {
			endTime = &event.Time
		}

		v := Variation{Key: event.Key, Version: event.Version, Index: event.Index}
		c, ok := counters[v]
		if !ok {
			counters[v] = CountValue{Count: 1, Value: event.Value}
		} else {
			c.Count += 1
		}
	}
	return counters, *startTime, *endTime
}

func (e *EventRecorder) RecordAccess(event AccessEvent) {
	e.mu.Lock()
	e.incomingEvents = append(e.incomingEvents, event)
	e.mu.Unlock()
}
