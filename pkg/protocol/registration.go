package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/models"
)

// RegistrationClient handles node registration with coordinator
type RegistrationClient struct {
	coordinatorURL string
	httpClient     *http.Client
	logger         *logrus.Logger
	nodeID         string
	endpoint       string
}

// NewRegistrationClient creates a new registration client
func NewRegistrationClient(coordinatorURL, nodeID, endpoint string, logger *logrus.Logger) *RegistrationClient {
	return &RegistrationClient{
		coordinatorURL: coordinatorURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:   logger,
		nodeID:   nodeID,
		endpoint: endpoint,
	}
}

// RegisterNode registers the worker node with the coordinator
func (rc *RegistrationClient) RegisterNode(ctx context.Context, capabilities models.NodeCapabilities) (*models.NodeRegistrationResponse, error) {
	// Create registration request
	request := models.NodeRegistrationRequest{
		Action:          "register",
		NodeID:          rc.nodeID,
		Capabilities:    capabilities,
		Endpoint:        rc.endpoint,
		ProtocolVersion: models.CurrentProtocolVersion,
	}

	// Marshal request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/nodes/register", rc.coordinatorURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", fmt.Sprintf("victorialogs-benchmark-worker/%s", rc.nodeID))

	// Send request
	rc.logger.WithFields(logrus.Fields{
		"node_id":         rc.nodeID,
		"coordinator_url": url,
		"capabilities":    capabilities,
	}).Info("Registering node with coordinator")

	resp, err := rc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	// Parse response
	var response models.NodeRegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("registration failed: %s", response.Message)
	}

	rc.logger.WithFields(logrus.Fields{
		"node_id": rc.nodeID,
		"message": response.Message,
	}).Info("Node registration successful")

	return &response, nil
}

// SendHeartbeat sends a heartbeat to the coordinator
func (rc *RegistrationClient) SendHeartbeat(ctx context.Context, status string, currentLoad float64, metrics map[string]float64) (*models.HeartbeatResponse, error) {
	// Create heartbeat request
	request := models.HeartbeatRequest{
		NodeID:             rc.nodeID,
		Status:             status,
		CurrentLoad:        currentLoad,
		PerformanceMetrics: metrics,
		Timestamp:          time.Now(),
	}

	// Marshal request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal heartbeat request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/nodes/%s/heartbeat", rc.coordinatorURL, rc.nodeID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", fmt.Sprintf("victorialogs-benchmark-worker/%s", rc.nodeID))

	// Send request
	resp, err := rc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send heartbeat request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("heartbeat failed with status %d", resp.StatusCode)
	}

	// Parse response
	var response models.HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode heartbeat response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("heartbeat failed: %s", response.Message)
	}

	return &response, nil
}

// UnregisterNode unregisters the worker node from the coordinator
func (rc *RegistrationClient) UnregisterNode(ctx context.Context) error {
	// Create unregistration request
	request := map[string]interface{}{
		"action":  "unregister",
		"node_id": rc.nodeID,
	}

	// Marshal request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal unregistration request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/nodes/%s/unregister", rc.coordinatorURL, rc.nodeID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", fmt.Sprintf("victorialogs-benchmark-worker/%s", rc.nodeID))

	// Send request
	rc.logger.WithField("node_id", rc.nodeID).Info("Unregistering node from coordinator")

	resp, err := rc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send unregistration request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unregistration failed with status %d", resp.StatusCode)
	}

	rc.logger.WithField("node_id", rc.nodeID).Info("Node unregistration successful")
	return nil
}

// GetNodeCapabilities collects and returns the current node's capabilities
func GetNodeCapabilities(maxConcurrent int) models.NodeCapabilities {
	return models.NodeCapabilities{
		MaxConcurrentQueries: maxConcurrent,
		SupportedQueryTypes:  getSupportedQueryTypes(),
		Resources: models.NodeResources{
			CPUCores: runtime.NumCPU(),
			MemoryMB: getAvailableMemoryMB(),
		},
	}
}

// getSupportedQueryTypes returns the list of supported query types
func getSupportedQueryTypes() []string {
	// Define supported LogsQL query types based on VictoriaLogs capabilities
	return []string{
		"basic",     // Basic field filtering: field:value
		"stats",     // Statistical queries with stats pipe
		"regex",     // Regular expression queries: field:~"regex"
		"json",      // JSON field queries: _msg:json.field
		"complex",   // Complex queries with multiple conditions
		"pipe",      // Pipe operations: | stats, | sort, etc.
		"time",      // Time-based filtering and grouping
		"wildcard",  // Wildcard queries: field:*pattern*
	}
}

// getAvailableMemoryMB returns available memory in MB
func getAvailableMemoryMB() int {
	// Try to get actual memory information
	// This is a simplified implementation that could be enhanced
	// with proper system memory detection
	
	// For containerized environments, check for memory limits
	if memLimit := getContainerMemoryLimit(); memLimit > 0 {
		return memLimit
	}
	
	// Default fallback based on typical container sizes
	cpuCount := runtime.NumCPU()
	switch {
	case cpuCount >= 8:
		return 16384 // 16GB for high-CPU nodes
	case cpuCount >= 4:
		return 8192  // 8GB for medium nodes
	case cpuCount >= 2:
		return 4096  // 4GB for small nodes
	default:
		return 2048  // 2GB for minimal nodes
	}
}

// getContainerMemoryLimit attempts to read container memory limit
func getContainerMemoryLimit() int {
	// This is a placeholder for container memory limit detection
	// In a real implementation, you would read from:
	// - /sys/fs/cgroup/memory/memory.limit_in_bytes (cgroup v1)
	// - /sys/fs/cgroup/memory.max (cgroup v2)
	// - Kubernetes downward API environment variables
	
	// For now, return 0 to indicate no limit detected
	return 0
}

// RegistrationServer handles node registration requests on coordinator side
type RegistrationServer struct {
	logger      *logrus.Logger
	nodeManager NodeManager
}

// NodeManager interface for managing registered nodes
type NodeManager interface {
	RegisterNode(nodeStatus *models.NodeStatus) error
	UpdateNodeHeartbeat(nodeID string, heartbeat *models.HeartbeatRequest) error
	UnregisterNode(nodeID string) error
	GetNodeStatus(nodeID string) (*models.NodeStatus, error)
}

// NewRegistrationServer creates a new registration server
func NewRegistrationServer(logger *logrus.Logger, nodeManager NodeManager) *RegistrationServer {
	return &RegistrationServer{
		logger:      logger,
		nodeManager: nodeManager,
	}
}

// HandleRegistration handles node registration requests
func (rs *RegistrationServer) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var request models.NodeRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		rs.logger.WithError(err).Error("Failed to decode registration request")
		rs.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateRegistrationRequest(&request); err != nil {
		rs.logger.WithError(err).WithField("node_id", request.NodeID).Warn("Invalid registration request")
		rs.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle protocol version compatibility
	compatibleVersion := models.GetCompatibleProtocolVersion(request.ProtocolVersion)
	if compatibleVersion == "" {
		rs.logger.WithFields(logrus.Fields{
			"node_id":           request.NodeID,
			"requested_version": request.ProtocolVersion,
			"supported_versions": models.SupportedProtocolVersions,
		}).Warn("Incompatible protocol version")
		
		response := models.NodeRegistrationResponse{
			Success:                 false,
			Message:                 fmt.Sprintf("Unsupported protocol version: %s", request.ProtocolVersion),
			NodeID:                  request.NodeID,
			AcceptedProtocolVersion: "",
			SupportedVersions:       models.SupportedProtocolVersions,
		}
		rs.sendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	// Create node status
	nodeStatus := &models.NodeStatus{
		NodeID:        request.NodeID,
		Status:        "active",
		LastHeartbeat: time.Now(),
		CurrentLoad:   0.0,
		Capabilities: map[string]interface{}{
			"max_concurrent_queries": request.Capabilities.MaxConcurrentQueries,
			"supported_query_types":  request.Capabilities.SupportedQueryTypes,
			"resources":              request.Capabilities.Resources,
		},
		AssignedTasks:      []string{},
		PerformanceMetrics: make(map[string]float64),
	}

	// Register node
	if err := rs.nodeManager.RegisterNode(nodeStatus); err != nil {
		rs.logger.WithError(err).WithField("node_id", request.NodeID).Error("Failed to register node")
		rs.sendErrorResponse(w, "Failed to register node", http.StatusInternalServerError)
		return
	}

	// Send success response
	response := models.NodeRegistrationResponse{
		Success:                 true,
		Message:                 "Node registered successfully",
		NodeID:                  request.NodeID,
		AcceptedProtocolVersion: compatibleVersion,
		SupportedVersions:       models.SupportedProtocolVersions,
	}

	rs.sendJSONResponse(w, response, http.StatusOK)
	rs.logger.WithField("node_id", request.NodeID).Info("Node registered successfully")
}

// HandleHeartbeat handles node heartbeat requests
func (rs *RegistrationServer) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract node ID from URL path
	// This is a simplified implementation - in production you might use a router
	nodeID := extractNodeIDFromPath(r.URL.Path)
	if nodeID == "" {
		rs.sendErrorResponse(w, "Node ID is required", http.StatusBadRequest)
		return
	}

	// Parse request
	var request models.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		rs.logger.WithError(err).Error("Failed to decode heartbeat request")
		rs.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate node ID matches
	if request.NodeID != nodeID {
		rs.sendErrorResponse(w, "Node ID mismatch", http.StatusBadRequest)
		return
	}

	// Update heartbeat
	if err := rs.nodeManager.UpdateNodeHeartbeat(nodeID, &request); err != nil {
		rs.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to update heartbeat")
		rs.sendErrorResponse(w, "Failed to update heartbeat", http.StatusInternalServerError)
		return
	}

	// Send success response
	response := models.HeartbeatResponse{
		Success:   true,
		Message:   "Heartbeat received",
		Timestamp: time.Now(),
	}

	rs.sendJSONResponse(w, response, http.StatusOK)
}

// HandleUnregistration handles node unregistration requests
func (rs *RegistrationServer) HandleUnregistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract node ID from URL path
	nodeID := extractNodeIDFromPath(r.URL.Path)
	if nodeID == "" {
		rs.sendErrorResponse(w, "Node ID is required", http.StatusBadRequest)
		return
	}

	// Unregister node
	if err := rs.nodeManager.UnregisterNode(nodeID); err != nil {
		rs.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to unregister node")
		rs.sendErrorResponse(w, "Failed to unregister node", http.StatusInternalServerError)
		return
	}

	// Send success response
	response := map[string]interface{}{
		"success": true,
		"message": "Node unregistered successfully",
	}

	rs.sendJSONResponse(w, response, http.StatusOK)
	rs.logger.WithField("node_id", nodeID).Info("Node unregistered successfully")
}

// sendJSONResponse sends a JSON response
func (rs *RegistrationServer) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse sends an error response
func (rs *RegistrationServer) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	rs.sendJSONResponse(w, response, statusCode)
}

// validateRegistrationRequest validates a node registration request
func validateRegistrationRequest(request *models.NodeRegistrationRequest) error {
	if request.NodeID == "" {
		return fmt.Errorf("node ID is required")
	}
	
	if request.Endpoint == "" {
		return fmt.Errorf("node endpoint is required")
	}
	
	if request.Action != "register" {
		return fmt.Errorf("invalid action: expected 'register', got '%s'", request.Action)
	}
	
	// Validate capabilities
	if request.Capabilities.MaxConcurrentQueries <= 0 {
		return fmt.Errorf("max concurrent queries must be positive, got %d", request.Capabilities.MaxConcurrentQueries)
	}
	
	if len(request.Capabilities.SupportedQueryTypes) == 0 {
		return fmt.Errorf("supported query types cannot be empty")
	}
	
	// Validate resources
	if request.Capabilities.Resources.CPUCores <= 0 {
		return fmt.Errorf("CPU cores must be positive, got %d", request.Capabilities.Resources.CPUCores)
	}
	
	if request.Capabilities.Resources.MemoryMB <= 0 {
		return fmt.Errorf("memory MB must be positive, got %d", request.Capabilities.Resources.MemoryMB)
	}
	
	// Validate endpoint format (basic validation)
	if !strings.HasPrefix(request.Endpoint, "http://") && !strings.HasPrefix(request.Endpoint, "https://") {
		return fmt.Errorf("endpoint must be a valid HTTP/HTTPS URL")
	}
	
	return nil
}

// extractNodeIDFromPath extracts node ID from URL path
// This is a simplified implementation
func extractNodeIDFromPath(path string) string {
	// Expected path format: /api/v1/nodes/{nodeID}/heartbeat or /api/v1/nodes/{nodeID}/unregister
	parts := strings.Split(path, "/")
	if len(parts) >= 5 && parts[1] == "api" && parts[2] == "v1" && parts[3] == "nodes" {
		return parts[4]
	}
	return ""
}