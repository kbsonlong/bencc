package worker

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/config"
)

// Service represents the worker service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	
	// Core components (to be implemented)
	nodeRegistrar    *NodeRegistrar
	taskExecutor     *TaskExecutor
	queryEngine      *QueryEngine
	metricsCollector *MetricsCollector
}

// NewService creates a new worker service
func NewService(cfg *config.Config, logger *logrus.Logger) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: logger,
	}
	
	// Initialize components (placeholder for now)
	service.nodeRegistrar = NewNodeRegistrar(cfg, logger)
	service.taskExecutor = NewTaskExecutor(cfg, logger)
	service.queryEngine = NewQueryEngine(cfg, logger)
	service.metricsCollector = NewMetricsCollector(cfg, logger)
	
	return service, nil
}

// Start starts the worker service
func (s *Service) Start(ctx context.Context) error {
	s.logger.WithField("node_id", s.config.Worker.NodeID).Info("Starting worker service...")
	
	// Start all components
	if err := s.nodeRegistrar.Start(ctx); err != nil {
		return fmt.Errorf("failed to start node registrar: %w", err)
	}
	
	if err := s.taskExecutor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task executor: %w", err)
	}
	
	if err := s.queryEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start query engine: %w", err)
	}
	
	if err := s.metricsCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics collector: %w", err)
	}
	
	s.logger.WithField("node_id", s.config.Worker.NodeID).Info("Worker service started successfully")
	
	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Stop stops the worker service
func (s *Service) Stop() {
	s.logger.WithField("node_id", s.config.Worker.NodeID).Info("Stopping worker service...")
	
	// Stop all components
	if s.metricsCollector != nil {
		s.metricsCollector.Stop()
	}
	if s.queryEngine != nil {
		s.queryEngine.Stop()
	}
	if s.taskExecutor != nil {
		s.taskExecutor.Stop()
	}
	if s.nodeRegistrar != nil {
		s.nodeRegistrar.Stop()
	}
	
	s.logger.WithField("node_id", s.config.Worker.NodeID).Info("Worker service stopped")
}