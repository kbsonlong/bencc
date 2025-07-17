package worker

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/config"
)

// NodeRegistrar handles registration with coordinator
type NodeRegistrar struct {
	config *config.Config
	logger *logrus.Logger
}

// NewNodeRegistrar creates a new node registrar
func NewNodeRegistrar(cfg *config.Config, logger *logrus.Logger) *NodeRegistrar {
	return &NodeRegistrar{
		config: cfg,
		logger: logger,
	}
}

// Start starts the node registrar
func (nr *NodeRegistrar) Start(ctx context.Context) error {
	nr.logger.WithField("coordinator_url", nr.config.Worker.CoordinatorURL).Info("Node registrar started")
	return nil
}

// Stop stops the node registrar
func (nr *NodeRegistrar) Stop() {
	nr.logger.Info("Node registrar stopped")
}

// TaskExecutor executes benchmark tasks
type TaskExecutor struct {
	config *config.Config
	logger *logrus.Logger
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(cfg *config.Config, logger *logrus.Logger) *TaskExecutor {
	return &TaskExecutor{
		config: cfg,
		logger: logger,
	}
}

// Start starts the task executor
func (te *TaskExecutor) Start(ctx context.Context) error {
	te.logger.Info("Task executor started")
	return nil
}

// Stop stops the task executor
func (te *TaskExecutor) Stop() {
	te.logger.Info("Task executor stopped")
}

// QueryEngine executes LogsQL queries
type QueryEngine struct {
	config *config.Config
	logger *logrus.Logger
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(cfg *config.Config, logger *logrus.Logger) *QueryEngine {
	return &QueryEngine{
		config: cfg,
		logger: logger,
	}
}

// Start starts the query engine
func (qe *QueryEngine) Start(ctx context.Context) error {
	qe.logger.WithField("target_url", qe.config.Target.URL).Info("Query engine started")
	return nil
}

// Stop stops the query engine
func (qe *QueryEngine) Stop() {
	qe.logger.Info("Query engine stopped")
}

// MetricsCollector collects performance metrics
type MetricsCollector struct {
	config *config.Config
	logger *logrus.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(cfg *config.Config, logger *logrus.Logger) *MetricsCollector {
	return &MetricsCollector{
		config: cfg,
		logger: logger,
	}
}

// Start starts the metrics collector
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.logger.Info("Metrics collector started")
	return nil
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() {
	mc.logger.Info("Metrics collector stopped")
}