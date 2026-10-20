package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Endpoint struct {
	Label  string
	Host   string
	Port   int
	APIKey string
}

type MemoryPressure struct {
	Enabled          bool   `json:"enabled"`
	CurrentBytes     int64  `json:"current_bytes"`
	SoftBytes        int64  `json:"soft_bytes"`
	HardBytes        int64  `json:"hard_bytes"`
	CurrentFormatted string `json:"current_formatted,omitempty"`
	SoftFormatted    string `json:"soft_formatted,omitempty"`
	HardFormatted    string `json:"hard_formatted,omitempty"`
	PressureLevel    string `json:"pressure_level"`
}

type ActiveModel struct {
	ID              string            `json:"id"`
	Waiting         []json.RawMessage `json:"waiting"`
	Prefilling      []json.RawMessage `json:"prefilling"`
	Generating      []json.RawMessage `json:"generating"`
	ActiveRequests  int               `json:"active_requests"`
	WaitingRequests int               `json:"waiting_requests"`
}

type ActiveModels struct {
	ModelMemoryUsed      int64           `json:"model_memory_used"`
	ModelMemoryMax       int64           `json:"model_memory_max"`
	MemoryPressure       *MemoryPressure `json:"memory_pressure"`
	TotalActiveRequests  int             `json:"total_active_requests"`
	TotalWaitingRequests int             `json:"total_waiting_requests"`
	Models               []ActiveModel   `json:"models"`
}

type Stats struct {
	ActiveModels ActiveModels `json:"active_models"`
}

type EndpointStatus struct {
	Label                string            `json:"label"`
	Host                 string            `json:"host"`
	Port                 int               `json:"port"`
	Model                string            `json:"model"`
	ModelMemoryUsed      int64             `json:"model_memory_used"`
	ModelMemoryMax       int64             `json:"model_memory_max"`
	MemoryPressure       *MemoryPressure   `json:"memory_pressure"`
	Waiting              []json.RawMessage `json:"waiting"`
	Prefilling           []json.RawMessage `json:"prefilling"`
	Generating           []json.RawMessage `json:"generating"`
	ActiveRequests       int               `json:"active_requests"`
	WaitingRequests      int               `json:"waiting_requests"`
	TotalActiveRequests  int               `json:"total_active_requests"`
	TotalWaitingRequests int               `json:"total_waiting_requests"`
	Busy                 bool              `json:"busy"`
	Maintenance          bool              `json:"maintenance"`
	Error                string            `json:"error,omitempty"`
}

func newAuthedClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}
}

func computeBaseUrl(ep Endpoint) string {
	return fmt.Sprintf("http://%s:%d", ep.Host, ep.Port)
}

func omlxLogin(ep Endpoint, st EndpointStatus) (*http.Client, error) {
	client := newAuthedClient()
	base := computeBaseUrl(ep)

	loginBody, _ := json.Marshal(map[string]string{"api_key": ep.APIKey})
	loginURL := base + "/admin/api/login"
	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(loginBody))
	if err != nil {
		st.Error = "login request: " + err.Error()
		st.Maintenance = true
		return client, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		st.Error = "login: " + err.Error()
		st.Maintenance = true
		return client, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		st.Error = fmt.Sprintf("login: HTTP %d", resp.StatusCode)
		st.Maintenance = true
		return client, err
	}

	return client, nil
}

func fetchOmlxStats(ep Endpoint, client *http.Client, st EndpointStatus) (Stats, error) {
	var stats Stats
	base := computeBaseUrl(ep)

	statsURL := base + "/admin/api/stats"
	req2, err := http.NewRequest(http.MethodGet, statsURL, nil)
	if err != nil {
		st.Error = "stats request: " + err.Error()
		return stats, err
	}
	resp2, err := client.Do(req2)
	if err != nil {
		st.Error = "stats: " + err.Error()
		return stats, err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		st.Error = fmt.Sprintf("stats: HTTP %d", resp2.StatusCode)
		return stats, err
	}

	if err := json.NewDecoder(resp2.Body).Decode(&stats); err != nil {
		st.Error = "stats decode: " + err.Error()
		return stats, err
	}

	return stats, nil
}

func poll(ep Endpoint) EndpointStatus {
	st := EndpointStatus{Label: ep.Label, Host: ep.Host, Port: ep.Port}
	client, err := omlxLogin(ep, st)
	if err != nil {
		return st
	}
	stats, err := fetchOmlxStats(ep, client, st)
	if err != nil {
		return st
	}
	st.ModelMemoryUsed = stats.ActiveModels.ModelMemoryUsed
	st.ModelMemoryMax = stats.ActiveModels.ModelMemoryMax
	st.MemoryPressure = stats.ActiveModels.MemoryPressure
	st.TotalActiveRequests = stats.ActiveModels.TotalActiveRequests
	st.TotalWaitingRequests = stats.ActiveModels.TotalWaitingRequests
	if len(stats.ActiveModels.Models) > 0 {
		m := stats.ActiveModels.Models[0]
		st.Model = m.ID
		st.Waiting = m.Waiting
		st.Prefilling = m.Prefilling
		st.Generating = m.Generating
		st.ActiveRequests = m.ActiveRequests
		st.WaitingRequests = m.WaitingRequests
	}
	st.Busy = len(st.Prefilling) > 0 || len(st.Generating) > 0
	st.Maintenance = false
	return st
}

func endpointsFromEnv() []Endpoint {
	spec := os.Getenv("OMLX_BACKENDS")
	var endpoints []Endpoint
	for i, group := range strings.Split(spec, ";") {
		fields := strings.Split(group, ",")
		if len(fields) != 4 {
			log.Fatalf("OMLX_BACKENDS group %d has %d fields, want 4 (label,host,port,apiKey): %q", i+1, len(fields), group)
		}
		port, err := strconv.Atoi(strings.TrimSpace(fields[2]))
		if err != nil {
			log.Fatalf("OMLX_BACKENDS group %d has non-integer port %q", i+1, fields[2])
		}
		endpoints = append(endpoints, Endpoint{
			Label:  strings.TrimSpace(fields[0]),
			Host:   strings.TrimSpace(fields[1]),
			Port:   port,
			APIKey: strings.TrimSpace(fields[3]),
		})
	}
	return endpoints
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	endpoints := endpointsFromEnv()
	if len(endpoints) == 0 {
		log.Fatal("OMLX_BACKENDS must specify at least one backend (label,host,port,apiKey)")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statuses := make(map[string]*EndpointStatus, len(endpoints))

		var wg sync.WaitGroup
		var mu sync.Mutex
		for i := range endpoints {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ep := endpoints[i]
				res := poll(ep)
				mu.Lock()
				statuses[ep.Label] = &res
				mu.Unlock()
			}(i)
		}
		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})

	log.Printf("omlx-status listening on :%s (%d backends)", port, len(endpoints))
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
