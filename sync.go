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
	closeOne        sync.Once
	stopChan        chan struct{}
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
	ticker := time.NewTimer(s.RefreshInterval * time.Millisecond)
	s.startOnce.Do(func() {
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-s.stopChan:
					return
				case <-ticker.C:
					s.fetchRemoteRepo()
				}
			}
		}()
	})
}

func (s *Synchronizer) Stop() {
	if s.stopChan == nil {
		return
	}
	s.closeOne.Do(func() {
		close(s.stopChan)
	})
}

func (s *Synchronizer) fetchRemoteRepo() {

	req, err := http.NewRequest(http.MethodGet, s.togglesUrl, nil)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	req.Header.Add("Authorization", s.auth)
	s.mu.Lock()
	resp, err := s.httpClient.Do(req)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	if resp == nil || resp.Body == nil {
		return
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	s.mu.Lock()
	err = json.Unmarshal(bodyBytes, s.repository)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}
