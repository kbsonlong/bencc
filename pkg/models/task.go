package models

import (
	"time"
)

// TaskExecution represents a benchmark task execution
type TaskExecution struct {
	TaskID       string                 `json:"task_id"`
	NodeID       string                 `json:"node_id"`
	Status       string                 `json:"status"` // "pending", "running", "completed", "failed"
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Config       map[string]interface{} `json:"config"`
	Results      map[string]interface{} `json:"results,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// TaskAssignment represents task assignment to worker nodes
type TaskAssignment struct {
	TaskID          string                 `json:"task_id"`
	NodeAssignments map[string]NodeTask    `json:"node_assignments"`
	GlobalConfig    map[string]interface{} `json:"global_config"`
}

// NodeTask represents a task assigned to a specific node
type NodeTask struct {
	ConcurrentUsers int      `json:"concurrent_users"`
	Duration        int      `json:"duration"` // seconds
	QueryTemplates  []string `json:"query_templates"`
	TargetQPS       float64  `json:"target_qps"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	NodeID    string                 `json:"node_id"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
	Status    string                 `json:"status"`
}

// BenchmarkMetrics represents performance metrics collected during benchmark
type BenchmarkMetrics struct {
	TotalQueries      int64   `json:"total_queries"`
	SuccessfulQueries int64   `json:"successful_queries"`
	FailedQueries     int64   `json:"failed_queries"`
	AvgResponseTime   float64 `json:"avg_response_time"`
	P95ResponseTime   float64 `json:"p95_response_time"`
	P99ResponseTime   float64 `json:"p99_response_time"`
	QPS               float64 `json:"qps"`
	ErrorRate         float64 `json:"error_rate"`
}

// QueryExecution represents a single query execution result
type QueryExecution struct {
	QueryID      string        `json:"query_id"`
	QueryType    string        `json:"query_type"`
	Query        string        `json:"query"`
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ResponseSize int64         `json:"response_size"`
}