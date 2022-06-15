package featureprobe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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
	once       sync.Once
}

func NewSynchronizer(url string, refreshMs time.Duration, auth string, repo *Repository) Synchronizer {
	return Synchronizer{
		auth:       auth,
		togglesUrl: url,
		refreshMs:  refreshMs,
		httpClient: newHttpClient(refreshMs),
		repository: repo,
	}
}

//TODO: create error message channel?
func (s *Synchronizer) StartSynchronize() {
	s.once.Do(func() {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("%s\n", err)
				}
			}()
			for {
				s.doSynchronize()
				time.Sleep(s.refreshMs * time.Millisecond)
			}
		}()
	})
}

func (s *Synchronizer) doSynchronize() {
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
	}
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
