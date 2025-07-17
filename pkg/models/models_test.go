package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNodeStatusSerialization(t *testing.T) {
	// Create a sample NodeStatus
	nodeStatus := NodeStatus{
		NodeID:        "worker-001",
		Status:        "active",
		LastHeartbeat: time.Now(),
		CurrentLoad:   0.75,
		Capabilities: map[string]interface{}{
			"max_concurrent_queries": 100,
			"supported_query_types":  []string{"basic", "stats", "regex"},
		},
		AssignedTasks: []string{"task-001", "task-002"},
		PerformanceMetrics: map[string]float64{
			"avg_response_time": 125.5,
			"qps":               49.8,
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(nodeStatus)
	if err != nil {
		t.Fatalf("Failed to marshal NodeStatus: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled NodeStatus
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal NodeStatus: %v", err)
	}

	// Verify key fields
	if unmarshaled.NodeID != nodeStatus.NodeID {
		t.Errorf("Expected NodeID %s, got %s", nodeStatus.NodeID, unmarshaled.NodeID)
	}
	if unmarshaled.Status != nodeStatus.Status {
		t.Errorf("Expected Status %s, got %s", nodeStatus.Status, unmarshaled.Status)
	}
	if unmarshaled.CurrentLoad != nodeStatus.CurrentLoad {
		t.Errorf("Expected CurrentLoad %f, got %f", nodeStatus.CurrentLoad, unmarshaled.CurrentLoad)
	}
}

func TestTaskExecutionLifecycle(t *testing.T) {
	// Create a new task execution
	task := TaskExecution{
		TaskID:  "benchmark-001",
		NodeID:  "worker-001",
		Status:  "pending",
		StartTime: time.Now(),
		Config: map[string]interface{}{
			"concurrent_users": 25,
			"duration":         300,
		},
	}

	// Verify initial state
	if task.Status != "pending" {
		t.Errorf("Expected initial status 'pending', got %s", task.Status)
	}
	if task.EndTime != nil {
		t.Error("Expected EndTime to be nil for pending task")
	}

	// Simulate task completion
	endTime := time.Now()
	task.Status = "completed"
	task.EndTime = &endTime
	task.Results = map[string]interface{}{
		"total_queries": 1500,
		"avg_response_time": 125.5,
	}

	// Verify completed state
	if task.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", task.Status)
	}
	if task.EndTime == nil {
		t.Error("Expected EndTime to be set for completed task")
	}
	if task.Results == nil {
		t.Error("Expected Results to be set for completed task")
	}
}

func TestNodeRegistrationRequest(t *testing.T) {
	// Create a registration request
	request := NodeRegistrationRequest{
		Action: "register",
		NodeID: "worker-001",
		Capabilities: NodeCapabilities{
			MaxConcurrentQueries: 100,
			SupportedQueryTypes:  []string{"basic", "stats", "regex"},
			Resources: NodeResources{
				CPUCores: 4,
				MemoryMB: 8192,
			},
		},
		Endpoint: "http://worker-001:8080",
	}

	// Test JSON serialization
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal NodeRegistrationRequest: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled NodeRegistrationRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal NodeRegistrationRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.Action != request.Action {
		t.Errorf("Expected Action %s, got %s", request.Action, unmarshaled.Action)
	}
	if unmarshaled.NodeID != request.NodeID {
		t.Errorf("Expected NodeID %s, got %s", request.NodeID, unmarshaled.NodeID)
	}
	if unmarshaled.Capabilities.MaxConcurrentQueries != request.Capabilities.MaxConcurrentQueries {
		t.Errorf("Expected MaxConcurrentQueries %d, got %d", 
			request.Capabilities.MaxConcurrentQueries, 
			unmarshaled.Capabilities.MaxConcurrentQueries)
	}
}

func TestBenchmarkMetrics(t *testing.T) {
	// Create sample metrics
	metrics := BenchmarkMetrics{
		TotalQueries:      1500,
		SuccessfulQueries: 1485,
		FailedQueries:     15,
		AvgResponseTime:   125.5,
		P95ResponseTime:   280.2,
		P99ResponseTime:   450.1,
		QPS:               49.8,
		ErrorRate:         0.01,
	}

	// Test JSON serialization
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal BenchmarkMetrics: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled BenchmarkMetrics
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal BenchmarkMetrics: %v", err)
	}

	// Verify calculations
	expectedErrorRate := float64(metrics.FailedQueries) / float64(metrics.TotalQueries)
	if unmarshaled.ErrorRate != expectedErrorRate {
		t.Errorf("Expected ErrorRate %f, got %f", expectedErrorRate, unmarshaled.ErrorRate)
	}

	// Verify all queries accounted for
	if unmarshaled.SuccessfulQueries+unmarshaled.FailedQueries != unmarshaled.TotalQueries {
		t.Error("Successful + Failed queries should equal Total queries")
	}
}

func TestTaskAssignmentStructure(t *testing.T) {
	// Create a task assignment
	assignment := TaskAssignment{
		TaskID: "benchmark-001",
		NodeAssignments: map[string]NodeTask{
			"worker-001": {
				ConcurrentUsers: 25,
				Duration:        300,
				QueryTemplates:  []string{"basic_filter", "json_field"},
				TargetQPS:       50,
			},
			"worker-002": {
				ConcurrentUsers: 25,
				Duration:        300,
				QueryTemplates:  []string{"error_requests", "stats_query"},
				TargetQPS:       50,
			},
		},
		GlobalConfig: map[string]interface{}{
			"target_url": "http://vlselect:8481",
			"auth":       map[string]string{"type": "bearer", "token": "xxx"},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(assignment)
	if err != nil {
		t.Fatalf("Failed to marshal TaskAssignment: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled TaskAssignment
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal TaskAssignment: %v", err)
	}

	// Verify structure
	if len(unmarshaled.NodeAssignments) != 2 {
		t.Errorf("Expected 2 node assignments, got %d", len(unmarshaled.NodeAssignments))
	}

	// Verify node task details
	if task, exists := unmarshaled.NodeAssignments["worker-001"]; exists {
		if task.ConcurrentUsers != 25 {
			t.Errorf("Expected ConcurrentUsers 25, got %d", task.ConcurrentUsers)
		}
		if len(task.QueryTemplates) != 2 {
			t.Errorf("Expected 2 query templates, got %d", len(task.QueryTemplates))
		}
	} else {
		t.Error("Expected worker-001 assignment to exist")
	}
}

func TestProtocolVersionCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		requestedVersion string
		expectedResult   string
		description      string
	}{
		{
			name:            "current version",
			requestedVersion: CurrentProtocolVersion,
			expectedResult:   CurrentProtocolVersion,
			description:      "Current version should be accepted as-is",
		},
		{
			name:            "beta version",
			requestedVersion: ProtocolVersionBeta,
			expectedResult:   ProtocolVersionBeta,
			description:      "Beta version should be accepted if supported",
		},
		{
			name:            "empty version",
			requestedVersion: "",
			expectedResult:   CurrentProtocolVersion,
			description:      "Empty version should default to current version",
		},
		{
			name:            "backward compatible version",
			requestedVersion: "v0.9.0",
			expectedResult:   CurrentProtocolVersion,
			description:      "Backward compatible version should return current version",
		},
		{
			name:            "unsupported version",
			requestedVersion: "v2.0.0",
			expectedResult:   "",
			description:      "Unsupported version should return empty string",
		},
		{
			name:            "invalid version format",
			requestedVersion: "invalid",
			expectedResult:   "",
			description:      "Invalid version format should return empty string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetCompatibleProtocolVersion(test.requestedVersion)
			if result != test.expectedResult {
				t.Errorf("Expected '%s', got '%s' - %s", test.expectedResult, result, test.description)
			}
		})
	}
}

func TestIsProtocolVersionSupported(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{CurrentProtocolVersion, true},
		{ProtocolVersionBeta, true},
		{"v2.0.0", false},
		{"", false},
		{"invalid", false},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := IsProtocolVersionSupported(test.version)
			if result != test.expected {
				t.Errorf("For version '%s', expected %v, got %v", test.version, test.expected, result)
			}
		})
	}
}