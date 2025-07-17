package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/models"
)

// MockNodeManager implements NodeManager interface for testing
type MockNodeManager struct {
	nodes       map[string]*models.NodeStatus
	heartbeats  map[string]*models.HeartbeatRequest
	shouldError bool
}

func NewMockNodeManager() *MockNodeManager {
	return &MockNodeManager{
		nodes:      make(map[string]*models.NodeStatus),
		heartbeats: make(map[string]*models.HeartbeatRequest),
	}
}

func (m *MockNodeManager) RegisterNode(nodeStatus *models.NodeStatus) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	m.nodes[nodeStatus.NodeID] = nodeStatus
	return nil
}

func (m *MockNodeManager) UpdateNodeHeartbeat(nodeID string, heartbeat *models.HeartbeatRequest) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	m.heartbeats[nodeID] = heartbeat
	if node, exists := m.nodes[nodeID]; exists {
		node.LastHeartbeat = heartbeat.Timestamp
		node.Status = heartbeat.Status
		node.CurrentLoad = heartbeat.CurrentLoad
		node.PerformanceMetrics = heartbeat.PerformanceMetrics
	}
	return nil
}

func (m *MockNodeManager) UnregisterNode(nodeID string) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	delete(m.nodes, nodeID)
	delete(m.heartbeats, nodeID)
	return nil
}

func (m *MockNodeManager) GetNodeStatus(nodeID string) (*models.NodeStatus, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	if node, exists := m.nodes[nodeID]; exists {
		return node, nil
	}
	return nil, fmt.Errorf("node not found")
}

func TestRegistrationClient_RegisterNode(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/nodes/register" {
			t.Errorf("Expected path /api/v1/nodes/register, got %s", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse and verify request body
		var request models.NodeRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if request.Action != "register" {
			t.Errorf("Expected action 'register', got %s", request.Action)
		}
		if request.NodeID != "test-node" {
			t.Errorf("Expected NodeID 'test-node', got %s", request.NodeID)
		}
		if request.ProtocolVersion != models.CurrentProtocolVersion {
			t.Errorf("Expected ProtocolVersion '%s', got %s", models.CurrentProtocolVersion, request.ProtocolVersion)
		}

		// Send success response
		response := models.NodeRegistrationResponse{
			Success: true,
			Message: "Node registered successfully",
			NodeID:  request.NodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create registration client
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests
	client := NewRegistrationClient(server.URL, "test-node", "http://test-node:8080", logger)

	// Test registration
	capabilities := models.NodeCapabilities{
		MaxConcurrentQueries: 100,
		SupportedQueryTypes:  []string{"basic", "stats"},
		Resources: models.NodeResources{
			CPUCores: 4,
			MemoryMB: 8192,
		},
	}

	ctx := context.Background()
	response, err := client.RegisterNode(ctx, capabilities)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful registration, got success=%v", response.Success)
	}
	if response.NodeID != "test-node" {
		t.Errorf("Expected NodeID 'test-node', got %s", response.NodeID)
	}
}

func TestRegistrationClient_SendHeartbeat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		expectedPath := "/api/v1/nodes/test-node/heartbeat"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse and verify request body
		var request models.HeartbeatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if request.NodeID != "test-node" {
			t.Errorf("Expected NodeID 'test-node', got %s", request.NodeID)
		}
		if request.Status != "active" {
			t.Errorf("Expected status 'active', got %s", request.Status)
		}
		if request.CurrentLoad != 0.5 {
			t.Errorf("Expected CurrentLoad 0.5, got %f", request.CurrentLoad)
		}

		// Send success response
		response := models.HeartbeatResponse{
			Success:   true,
			Message:   "Heartbeat received",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create registration client
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client := NewRegistrationClient(server.URL, "test-node", "http://test-node:8080", logger)

	// Test heartbeat
	metrics := map[string]float64{
		"avg_response_time": 125.5,
		"qps":               49.8,
	}

	ctx := context.Background()
	response, err := client.SendHeartbeat(ctx, "active", 0.5, metrics)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful heartbeat, got success=%v", response.Success)
	}
}

func TestRegistrationClient_UnregisterNode(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		expectedPath := "/api/v1/nodes/test-node/unregister"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create registration client
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client := NewRegistrationClient(server.URL, "test-node", "http://test-node:8080", logger)

	// Test unregistration
	ctx := context.Background()
	err := client.UnregisterNode(ctx)
	if err != nil {
		t.Fatalf("Unregistration failed: %v", err)
	}
}

func TestGetNodeCapabilities(t *testing.T) {
	capabilities := GetNodeCapabilities(100)

	if capabilities.MaxConcurrentQueries != 100 {
		t.Errorf("Expected MaxConcurrentQueries 100, got %d", capabilities.MaxConcurrentQueries)
	}

	expectedTypes := []string{"basic", "stats", "regex", "json", "complex", "pipe", "time", "wildcard"}
	if len(capabilities.SupportedQueryTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported query types, got %d", len(expectedTypes), len(capabilities.SupportedQueryTypes))
	}

	// Verify all expected query types are present
	typeMap := make(map[string]bool)
	for _, queryType := range capabilities.SupportedQueryTypes {
		typeMap[queryType] = true
	}
	
	for _, expectedType := range expectedTypes {
		if !typeMap[expectedType] {
			t.Errorf("Expected query type '%s' not found in capabilities", expectedType)
		}
	}

	if capabilities.Resources.CPUCores <= 0 {
		t.Errorf("Expected positive CPU cores, got %d", capabilities.Resources.CPUCores)
	}

	if capabilities.Resources.MemoryMB <= 0 {
		t.Errorf("Expected positive memory MB, got %d", capabilities.Resources.MemoryMB)
	}
}

func TestRegistrationServer_HandleRegistration(t *testing.T) {
	// Create mock node manager
	nodeManager := NewMockNodeManager()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewRegistrationServer(logger, nodeManager)

	// Create test request
	request := models.NodeRegistrationRequest{
		Action: "register",
		NodeID: "test-node",
		Capabilities: models.NodeCapabilities{
			MaxConcurrentQueries: 100,
			SupportedQueryTypes:  []string{"basic", "stats"},
			Resources: models.NodeResources{
				CPUCores: 4,
				MemoryMB: 8192,
			},
		},
		Endpoint: "http://test-node:8080",
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", bytes.NewBuffer(requestData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Handle request
	server.HandleRegistration(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.NodeRegistrationResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful registration, got success=%v", response.Success)
	}
	if response.NodeID != "test-node" {
		t.Errorf("Expected NodeID 'test-node', got %s", response.NodeID)
	}

	// Verify node was registered in mock manager
	if _, exists := nodeManager.nodes["test-node"]; !exists {
		t.Error("Expected node to be registered in node manager")
	}
}

func TestRegistrationServer_HandleHeartbeat(t *testing.T) {
	// Create mock node manager with pre-registered node
	nodeManager := NewMockNodeManager()
	nodeManager.nodes["test-node"] = &models.NodeStatus{
		NodeID: "test-node",
		Status: "active",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewRegistrationServer(logger, nodeManager)

	// Create test request
	request := models.HeartbeatRequest{
		NodeID:      "test-node",
		Status:      "active",
		CurrentLoad: 0.5,
		PerformanceMetrics: map[string]float64{
			"avg_response_time": 125.5,
		},
		Timestamp: time.Now(),
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/test-node/heartbeat", bytes.NewBuffer(requestData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Handle request
	server.HandleHeartbeat(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.HeartbeatResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful heartbeat, got success=%v", response.Success)
	}

	// Verify heartbeat was recorded in mock manager
	if _, exists := nodeManager.heartbeats["test-node"]; !exists {
		t.Error("Expected heartbeat to be recorded in node manager")
	}
}

func TestRegistrationServer_HandleUnregistration(t *testing.T) {
	// Create mock node manager with pre-registered node
	nodeManager := NewMockNodeManager()
	nodeManager.nodes["test-node"] = &models.NodeStatus{
		NodeID: "test-node",
		Status: "active",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewRegistrationServer(logger, nodeManager)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/test-node/unregister", nil)
	w := httptest.NewRecorder()

	// Handle request
	server.HandleUnregistration(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify node was unregistered from mock manager
	if _, exists := nodeManager.nodes["test-node"]; exists {
		t.Error("Expected node to be unregistered from node manager")
	}
}

func TestExtractNodeIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/nodes/test-node/heartbeat", "test-node"},
		{"/api/v1/nodes/worker-001/unregister", "worker-001"},
		{"/api/v1/nodes/node-123/heartbeat", "node-123"},
		{"/invalid/path", ""},
		{"/api/v1/nodes", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := extractNodeIDFromPath(test.path)
		if result != test.expected {
			t.Errorf("For path %s, expected %s, got %s", test.path, test.expected, result)
		}
	}
}

func TestRegistrationServer_ErrorHandling(t *testing.T) {
	// Test with error-prone node manager
	nodeManager := NewMockNodeManager()
	nodeManager.shouldError = true

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewRegistrationServer(logger, nodeManager)

	// Test registration error with valid request that will trigger node manager error
	request := models.NodeRegistrationRequest{
		Action: "register",
		NodeID: "test-node",
		Capabilities: models.NodeCapabilities{
			MaxConcurrentQueries: 100,
			SupportedQueryTypes:  []string{"basic", "stats"},
			Resources: models.NodeResources{
				CPUCores: 4,
				MemoryMB: 8192,
			},
		},
		Endpoint:        "http://test-node:8080",
		ProtocolVersion: models.CurrentProtocolVersion,
	}

	requestData, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", bytes.NewBuffer(requestData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleRegistration(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d for registration error, got %d", http.StatusInternalServerError, w.Code)
	}

	// Test heartbeat error
	heartbeatReq := models.HeartbeatRequest{
		NodeID: "test-node",
		Status: "active",
	}

	heartbeatData, _ := json.Marshal(heartbeatReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/nodes/test-node/heartbeat", bytes.NewBuffer(heartbeatData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.HandleHeartbeat(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d for heartbeat error, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestValidateRegistrationRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     models.NodeRegistrationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{"basic", "stats"},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 8192,
					},
				},
				Endpoint: "http://test-node:8080",
			},
			expectError: false,
		},
		{
			name: "missing node ID",
			request: models.NodeRegistrationRequest{
				Action:   "register",
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "node ID is required",
		},
		{
			name: "missing endpoint",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
			},
			expectError: true,
			errorMsg:    "node endpoint is required",
		},
		{
			name: "invalid action",
			request: models.NodeRegistrationRequest{
				Action:   "invalid",
				NodeID:   "test-node",
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "invalid action: expected 'register', got 'invalid'",
		},
		{
			name: "invalid max concurrent queries",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 0,
					SupportedQueryTypes:  []string{"basic"},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 8192,
					},
				},
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "max concurrent queries must be positive, got 0",
		},
		{
			name: "empty supported query types",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 8192,
					},
				},
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "supported query types cannot be empty",
		},
		{
			name: "invalid CPU cores",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{"basic"},
					Resources: models.NodeResources{
						CPUCores: 0,
						MemoryMB: 8192,
					},
				},
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "CPU cores must be positive, got 0",
		},
		{
			name: "invalid memory",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{"basic"},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 0,
					},
				},
				Endpoint: "http://test-node:8080",
			},
			expectError: true,
			errorMsg:    "memory MB must be positive, got 0",
		},
		{
			name: "invalid endpoint format",
			request: models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{"basic"},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 8192,
					},
				},
				Endpoint: "invalid-endpoint",
			},
			expectError: true,
			errorMsg:    "endpoint must be a valid HTTP/HTTPS URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRegistrationRequest(&test.request)
			
			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != test.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", test.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestProtocolVersionCompatibility(t *testing.T) {
	// Create mock node manager
	nodeManager := NewMockNodeManager()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewRegistrationServer(logger, nodeManager)

	tests := []struct {
		name            string
		protocolVersion string
		expectSuccess   bool
		expectedVersion string
	}{
		{
			name:            "current version",
			protocolVersion: models.CurrentProtocolVersion,
			expectSuccess:   true,
			expectedVersion: models.CurrentProtocolVersion,
		},
		{
			name:            "supported beta version",
			protocolVersion: models.ProtocolVersionBeta,
			expectSuccess:   true,
			expectedVersion: models.ProtocolVersionBeta,
		},
		{
			name:            "backward compatible version",
			protocolVersion: "v0.9.0",
			expectSuccess:   true,
			expectedVersion: models.CurrentProtocolVersion,
		},
		{
			name:            "unsupported version",
			protocolVersion: "v2.0.0",
			expectSuccess:   false,
			expectedVersion: "",
		},
		{
			name:            "empty version defaults to current",
			protocolVersion: "",
			expectSuccess:   true,
			expectedVersion: models.CurrentProtocolVersion,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create test request
			request := models.NodeRegistrationRequest{
				Action: "register",
				NodeID: "test-node",
				Capabilities: models.NodeCapabilities{
					MaxConcurrentQueries: 100,
					SupportedQueryTypes:  []string{"basic", "stats"},
					Resources: models.NodeResources{
						CPUCores: 4,
						MemoryMB: 8192,
					},
				},
				Endpoint:        "http://test-node:8080",
				ProtocolVersion: test.protocolVersion,
			}

			requestData, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", bytes.NewBuffer(requestData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Handle request
			server.HandleRegistration(w, req)

			// Parse response
			var response models.NodeRegistrationResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if test.expectSuccess {
				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				}
				if !response.Success {
					t.Errorf("Expected successful registration, got success=%v, message=%s", response.Success, response.Message)
				}
				if response.AcceptedProtocolVersion != test.expectedVersion {
					t.Errorf("Expected accepted version '%s', got '%s'", test.expectedVersion, response.AcceptedProtocolVersion)
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
				}
				if response.Success {
					t.Errorf("Expected failed registration, got success=%v", response.Success)
				}
			}
		})
	}
}

func TestGetSupportedQueryTypes(t *testing.T) {
	queryTypes := getSupportedQueryTypes()
	
	expectedTypes := []string{"basic", "stats", "regex", "json", "complex", "pipe", "time", "wildcard"}
	
	if len(queryTypes) != len(expectedTypes) {
		t.Errorf("Expected %d query types, got %d", len(expectedTypes), len(queryTypes))
	}
	
	// Verify all expected types are present
	typeMap := make(map[string]bool)
	for _, queryType := range queryTypes {
		typeMap[queryType] = true
	}
	
	for _, expectedType := range expectedTypes {
		if !typeMap[expectedType] {
			t.Errorf("Expected query type '%s' not found", expectedType)
		}
	}
}