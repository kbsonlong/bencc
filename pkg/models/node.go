package models

import (
	"time"
)

// NodeStatus represents the status of a worker node
type NodeStatus struct {
	NodeID             string                 `json:"node_id"`
	Status             string                 `json:"status"` // "active", "busy", "failed", "offline"
	LastHeartbeat      time.Time              `json:"last_heartbeat"`
	CurrentLoad        float64                `json:"current_load"` // 0.0 - 1.0
	Capabilities       map[string]interface{} `json:"capabilities"`
	AssignedTasks      []string               `json:"assigned_tasks"`
	PerformanceMetrics map[string]float64     `json:"performance_metrics"`
}

// NodeCapabilities represents the capabilities of a worker node
type NodeCapabilities struct {
	MaxConcurrentQueries int      `json:"max_concurrent_queries"`
	SupportedQueryTypes  []string `json:"supported_query_types"`
	Resources            NodeResources `json:"resources"`
}

// NodeResources represents the resource information of a node
type NodeResources struct {
	CPUCores  int `json:"cpu_cores"`
	MemoryMB  int `json:"memory_mb"`
}

// NodeRegistrationRequest represents a node registration request
type NodeRegistrationRequest struct {
	Action       string           `json:"action"`
	NodeID       string           `json:"node_id"`
	Capabilities NodeCapabilities `json:"capabilities"`
	Endpoint     string           `json:"endpoint"`
}

// NodeRegistrationResponse represents the response to a node registration
type NodeRegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NodeID  string `json:"node_id"`
}

// HeartbeatRequest represents a heartbeat request from a worker node
type HeartbeatRequest struct {
	NodeID             string             `json:"node_id"`
	Status             string             `json:"status"`
	CurrentLoad        float64            `json:"current_load"`
	PerformanceMetrics map[string]float64 `json:"performance_metrics"`
	Timestamp          time.Time          `json:"timestamp"`
}

// HeartbeatResponse represents the response to a heartbeat
type HeartbeatResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}