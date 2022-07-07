package featureprobe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Synchronizer struct {
	auth            string
	togglesUrl      string
	RefreshInterval time.Duration
	repository      *Repository
	httpClient      http.Client
	mu              sync.Mutex
	startOnce       sync.Once
	stopOnce        sync.Once
	stopChan        chan struct{}
	ticker          *time.Ticker
}

func NewSynchronizer(url string, RefreshInterval time.Duration, auth string, repo *Repository) Synchronizer {
	return Synchronizer{
		auth:            auth,
		togglesUrl:      url,
		RefreshInterval: RefreshInterval,
		httpClient:      newHttpClient(RefreshInterval),
		repository:      repo,
		stopChan:        make(chan struct{}),
	}
}

//TODO: create error message channel?
func (s *Synchronizer) Start() {
	s.startOnce.Do(func() {
		s.ticker = time.NewTicker(s.RefreshInterval * time.Millisecond)
		go func() {
			for {
				select {
				case <-s.stopChan:
					return
				case <-s.ticker.C:
					s.fetchRemoteRepo()
				}
			}
		}()
	})
}

func (s *Synchronizer) Stop() {
	if s.stopChan != nil {
		s.stopOnce.Do(func() {
			close(s.stopChan)
		})
	}
}

func (s *Synchronizer) fetchRemoteRepo() {
	req, err := http.NewRequest(http.MethodGet, s.togglesUrl, nil)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	req.Header.Add("Authorization", s.auth)
	req.Header.Add("User-Agent", USER_AGENT)
	s.mu.Lock()
	resp, err := s.httpClient.Do(req)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	s.mu.Lock()
	err = json.Unmarshal(bodyBytes, s.repository)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}
