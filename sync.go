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
	once            sync.Once
}

func NewSynchronizer(url string, RefreshInterval time.Duration, auth string, repo *Repository) Synchronizer {
	return Synchronizer{
		auth:            auth,
		togglesUrl:      url,
		RefreshInterval: RefreshInterval,
		httpClient:      newHttpClient(RefreshInterval),
		repository:      repo,
	}
}

//TODO: create error message channel?
func (s *Synchronizer) Start() {
	s.once.Do(func() {
		go s.doSynchronize()
	})
}

func (s *Synchronizer) doSynchronize() {
	for {
		req, err := http.NewRequest(http.MethodGet, s.togglesUrl, nil)
		if err != nil {
			fmt.Printf("%s\n", err)
			break
		}
		req.Header.Add("Authorization", s.auth)
		s.mu.Lock()
		resp, err := s.httpClient.Do(req)
		s.mu.Unlock()
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		if resp == nil || resp.Body == nil {
			continue
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		s.mu.Lock()
		err = json.Unmarshal(bodyBytes, s.repository)
		s.mu.Unlock()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		time.Sleep(s.RefreshInterval * time.Millisecond)
	}
}
