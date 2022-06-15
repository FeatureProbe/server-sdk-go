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
	capacity       uint64
	incomingEvents []AccessEvent
	packedData     []packedData
	httpClient     http.Client
	mu             sync.Mutex
	once           sync.Once
}

type AccessEvent struct {
	time    uint64
	key     string
	value   interface{}
	index   uint64
	version uint64
	reason  string
}

type packedData struct {
	events []AccessEvent
	access access
}

type access struct {
	startTime uint64
	endTime   uint64
	counters  map[string][]toggleCounter
}

type toggleCounter struct {
	value   interface{}
	version uint64
	index   uint64
	count   uint64
}

type variation struct {
	key     string
	index   uint64
	version uint64
}

type countValue struct {
	count uint64
	value interface{}
}

func NewEventRecorder(eventsUrl string, auth string, flushMs time.Duration, capacity uint64) EventRecorder {
	return EventRecorder{
		auth:           auth,
		eventsUrl:      eventsUrl,
		flushMs:        flushMs,
		capacity:       capacity,
		incomingEvents: []AccessEvent{},
		packedData:     []packedData{},
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
		events := []AccessEvent{}
		e.mu.Lock()
		events, e.incomingEvents = e.incomingEvents, events
		e.mu.Unlock()

		packedData := e.buildPackedData(events)
		body, err := json.Marshal(packedData)
		if err != nil {
			fmt.Println("flush error: json invalid")
		}
		fmt.Println(body)

		req, err := http.NewRequest(http.MethodPost, e.eventsUrl, bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("%s\n", err)
			break
		}
		req.Header.Add("Authorization", e.auth)
		e.mu.Lock()
		_, err = e.httpClient.Do(req)
		e.mu.Unlock()
		fmt.Printf("%s\n", err)

		time.Sleep(e.flushMs * time.Millisecond)
	}
}

func (e *EventRecorder) buildPackedData(events []AccessEvent) []packedData {
	access := e.buildAccess(events)
	return []packedData{packedData{access: access, events: events}}
}

func (e *EventRecorder) buildAccess(events []AccessEvent) access {
	var startTime *uint64 = nil
	var endTime *uint64 = nil

	counters := e.buildCounters(startTime, endTime, events)
	access := access{
		startTime: *startTime,
		endTime:   *endTime,
		counters:  map[string][]toggleCounter{},
	}

	for k, v := range counters {
		counter := toggleCounter{
			index:   k.index,
			version: k.version,
			count:   v.count,
			value:   v.value,
		}
		c, ok := access.counters[k.key]
		if !ok {
			access.counters[k.key] = []toggleCounter{counter}
		} else {
			access.counters[k.key] = append(c, counter)
		}
	}
	return access
}

func (e *EventRecorder) buildCounters(startTime *uint64, endTime *uint64, events []AccessEvent) map[variation]countValue {
	counters := map[variation]countValue{}

	for _, e := range events {
		if startTime == nil || *startTime < e.time {
			startTime = &e.time
		}
		if endTime == nil || *endTime > e.time {
			endTime = &e.time
		}

		v := variation{key: e.key, version: e.version, index: e.index}
		c, ok := counters[v]
		if !ok {
			counters[v] = countValue{count: 0, value: e.value}
		} else {
			c.count += 1
		}
	}
	return counters
}

func (e *EventRecorder) RecordAccess(event AccessEvent) {
	e.mu.Lock()
	e.incomingEvents = append(e.incomingEvents, event)
}
