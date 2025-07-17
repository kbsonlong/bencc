package coordinator

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/config"
)

// Service represents the coordinator service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	
	// Core components (to be implemented)
	nodeManager      *NodeManager
	taskScheduler    *TaskScheduler
	resultAggregator *ResultAggregator
	apiServer        *APIServer
}

// NewService creates a new coordinator service
func NewService(cfg *config.Config, logger *logrus.Logger) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: logger,
	}
	
	// Initialize components (placeholder for now)
	service.nodeManager = NewNodeManager(cfg, logger)
	service.taskScheduler = NewTaskScheduler(cfg, logger)
	service.resultAggregator = NewResultAggregator(cfg, logger)
	service.apiServer = NewAPIServer(cfg, logger)
	
	return service, nil
}

// Start starts the coordinator service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting coordinator service...")
	
	// Start all components
	if err := s.nodeManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start node manager: %w", err)
	}
	
	if err := s.taskScheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task scheduler: %w", err)
	}
	
	if err := s.resultAggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start result aggregator: %w", err)
	}
	
	if err := s.apiServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	
	s.logger.Info("Coordinator service started successfully")
	
	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Stop stops the coordinator service
func (s *Service) Stop() {
	s.logger.Info("Stopping coordinator service...")
	
	// Stop all components
	if s.apiServer != nil {
		s.apiServer.Stop()
	}
	if s.resultAggregator != nil {
		s.resultAggregator.Stop()
	}
	if s.taskScheduler != nil {
		s.taskScheduler.Stop()
	}
	if s.nodeManager != nil {
		s.nodeManager.Stop()
	}
	
	s.logger.Info("Coordinator service stopped")
}