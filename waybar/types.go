package main

type Config struct {
	APIURL   string   `yaml:"api_url"`
	Backends []string `yaml:"backends"`
}

type MemoryPressure struct {
	Enabled          bool    `json:"enabled"`
	CurrentBytes     float64 `json:"current_bytes"`
	SoftBytes        float64 `json:"soft_bytes"`
	HardBytes        float64 `json:"hard_bytes"`
	CurrentFormatted string  `json:"current_formatted"`
	SoftFormatted    string  `json:"soft_formatted"`
	HardFormatted    string  `json:"hard_formatted"`
	PressureLevel    string  `json:"pressure_level"`
}

type ActiveTask struct {
	RequestID       string  `json:"request_id"`
	ElapsedSeconds  float64 `json:"elapsed_seconds,omitempty"`
	GeneratedTokens int     `json:"generated_tokens,omitempty"`
	TokensPerSecond float64 `json:"tokens_per_second,omitempty"`
	LastActivityAge float64 `json:"last_activity_age_seconds,omitempty"`
	PromptTokens    int     `json:"prompt_tokens,omitempty"`
	MaxTokens       int     `json:"max_tokens,omitempty"`

	PreProcessed int     `json:"processed,omitempty"`
	PreTotal     int     `json:"total,omitempty"`
	PreSpeed     float64 `json:"speed,omitempty"`
	PreETA       float64 `json:"eta,omitempty"`
	PreElapsed   float64 `json:"elapsed,omitempty"`
	PrePhase     string  `json:"phase,omitempty"`
}

type Machine struct {
	Label                string         `json:"label"`
	Host                 string         `json:"host"`
	Port                 int            `json:"port"`
	Model                string         `json:"model"`
	ModelMemoryUsed      float64        `json:"model_memory_used"`
	ModelMemoryMax       float64        `json:"model_memory_max"`
	MemoryPressure       MemoryPressure `json:"memory_pressure"`
	Waiting              []ActiveTask   `json:"waiting"`
	Prefilling           []ActiveTask   `json:"prefilling"`
	Generating           []ActiveTask   `json:"generating"`
	ActiveRequests       int            `json:"active_requests"`
	WaitingRequests      int            `json:"waiting_requests"`
	TotalActiveRequests  int            `json:"total_active_requests"`
	TotalWaitingRequests int            `json:"total_waiting_requests"`
	Busy                 bool           `json:"busy"`
	Maintenance          bool           `json:"maintenance"`
}

type APIResponse map[string]Machine

type MachineState struct {
	Status     string
	Icon       string
	Model      string
	ETA        string
	MemoryUsed string
	MemoryMax  string
}

type OutputState struct {
	APIResponse APIResponse    `json:"api_response"`
	Machines    []MachineState `json:"machines"`
}

type Output struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}
