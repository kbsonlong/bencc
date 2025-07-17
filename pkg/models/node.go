package models

import (
	"time"
)

// Protocol version constants
const (
	// CurrentProtocolVersion is the current protocol version
	CurrentProtocolVersion = "v1.0.0"
	
	// MinSupportedProtocolVersion is the minimum supported protocol version
	MinSupportedProtocolVersion = "v1.0.0"
	
	// ProtocolVersionBeta is the beta version for testing new features
	ProtocolVersionBeta = "v1.1.0-beta"
)

// SupportedProtocolVersions lists all supported protocol versions
var SupportedProtocolVersions = []string{
	"v1.0.0",
	"v1.1.0-beta", // For future compatibility testing
}

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
	Action          string           `json:"action"`
	NodeID          string           `json:"node_id"`
	Capabilities    NodeCapabilities `json:"capabilities"`
	Endpoint        string           `json:"endpoint"`
	ProtocolVersion string           `json:"protocol_version"`
}

// NodeRegistrationResponse represents the response to a node registration
type NodeRegistrationResponse struct {
	Success                bool     `json:"success"`
	Message                string   `json:"message"`
	NodeID                 string   `json:"node_id"`
	AcceptedProtocolVersion string   `json:"accepted_protocol_version"`
	SupportedVersions      []string `json:"supported_versions"`
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

// IsProtocolVersionSupported checks if a protocol version is supported
func IsProtocolVersionSupported(version string) bool {
	for _, supportedVersion := range SupportedProtocolVersions {
		if supportedVersion == version {
			return true
		}
	}
	return false
}

// GetCompatibleProtocolVersion returns the best compatible protocol version
// If the requested version is supported, it returns that version
// Otherwise, it returns the current version if compatible, or empty string if incompatible
func GetCompatibleProtocolVersion(requestedVersion string) string {
	if requestedVersion == "" {
		return CurrentProtocolVersion
	}
	
	if IsProtocolVersionSupported(requestedVersion) {
		return requestedVersion
	}
	
	// Implement basic semantic version compatibility
	// For now, we support backward compatibility within the same major version
	if isBackwardCompatible(requestedVersion) {
		return CurrentProtocolVersion
	}
	
	return ""
}

// isBackwardCompatible checks if the requested version is backward compatible
func isBackwardCompatible(requestedVersion string) bool {
	// Simple backward compatibility check
	// In production, you might want to use a proper semver library
	switch requestedVersion {
	case "v0.9.0", "v0.9.1": // Legacy versions that are compatible
		return true
	default:
		return false
	}
}