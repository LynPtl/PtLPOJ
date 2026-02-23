package models

// Problem represents an in-memory problem definition loaded from problems.json
type Problem struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Difficulty    string   `json:"difficulty"`
	Tags          []string `json:"tags"`
	TimeLimitMs   int      `json:"time_limit_ms"`
	MemoryLimitKb int      `json:"memory_limit_kb"`
	CaseCount     int      `json:"case_count"`
}
