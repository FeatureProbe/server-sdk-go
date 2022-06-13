package featureprobe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Synchronizer struct {
	auth       string
	togglesUrl string
	refreshMs  time.Duration
	repository *Repository
	httpClient http.Client
	mu         sync.Mutex
}

func NewSynchronizer(url string, refreshMs time.Duration, auth string, repo *Repository) Synchronizer {
	return Synchronizer{
		auth:       auth,
		togglesUrl: url,
		refreshMs:  refreshMs,
		httpClient: http.Client{},
		repository: repo,
	}
}

// create error channel ?
func (s *Synchronizer) StartSynchronize() {
	go s.doSynchronize()
}

func (s *Synchronizer) doSynchronize() {
	for {
		req, err := http.NewRequest(http.MethodGet, s.togglesUrl, nil)
		if err != nil {
			fmt.Errorf("%s", err)
			break
		}
		req.Header.Add("Authorization", s.auth)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			fmt.Errorf("%s", err)
		}

		if resp == nil || resp.Body == nil {
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		s.mu.Lock()
		err = json.Unmarshal(bodyBytes, s.repository)
		s.mu.Unlock()
		if err != nil {
			fmt.Errorf("%s", err)
		}
		//fmt.Println(string(bodyBytes))
		time.Sleep(s.refreshMs * time.Millisecond)
	}
}
