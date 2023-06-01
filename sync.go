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
	auth               string
	togglesUrl         string
	RefreshInterval    time.Duration
	repository         *Repository
	httpClient         http.Client
	mu                 sync.Mutex
	startOnce          sync.Once
	stopOnce           sync.Once
	setInitializedOnce sync.Once
	isInitialized      bool
	stopChan           chan struct{}
	ticker             *time.Ticker
	enablePolling      bool
}

func NewSynchronizer(url string, RefreshInterval time.Duration, auth string, repo *Repository) Synchronizer {
	return Synchronizer{
		auth:            auth,
		togglesUrl:      url,
		RefreshInterval: RefreshInterval,
		httpClient:      newHttpClient(RefreshInterval),
		repository:      repo,
		stopChan:        make(chan struct{}),
		enablePolling:   true,
	}
}

func NewCustomRepoSynchronizer(repo *Repository) Synchronizer {
	return Synchronizer{
		repository:    repo,
		stopChan:      make(chan struct{}),
		enablePolling: false,
	}
}

func (s *Synchronizer) Start(ready chan<- struct{}) {
	var readyOnce sync.Once
	notifyReady := func() {
		readyOnce.Do(func() {
			close(ready)
		})
	}
	if !s.enablePolling {
		s.isInitialized = true
		notifyReady()
		return
	}
	s.startOnce.Do(func() {
		s.ticker = time.NewTicker(s.RefreshInterval)
		go func() {
			for {
				select {
				case <-s.stopChan:
					return
				case <-s.ticker.C:
					err := s.FetchRemoteRepo()
					if err == nil {
						s.setInitializedOnce.Do(func() {
							// first sync success
							s.isInitialized = true
							notifyReady()
						})
					}
				}
			}
		}()
	})
}

// Initialized return false means not successfully fetch remote resource
func (s *Synchronizer) Initialized() bool {
	return s.isInitialized
}

func (s *Synchronizer) Stop() {
	if s.stopChan != nil {
		s.stopOnce.Do(func() {
			close(s.stopChan)
			s.isInitialized = false
		})
	}
}

// FetchRemoteRepo fetch remote repo and update local repo
func (s *Synchronizer) FetchRemoteRepo() error {
	req, err := http.NewRequest(http.MethodGet, s.togglesUrl, nil)

	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	req.Header.Add("Authorization", s.auth)
	req.Header.Add("User-Agent", USER_AGENT)
	s.mu.Lock()
	resp, err := s.httpClient.Do(req)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	s.mu.Lock()
	repoData := RepositoryData{}
	err = json.Unmarshal(bodyBytes, repoData)
	s.repository.flush(repoData)
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return err
}
