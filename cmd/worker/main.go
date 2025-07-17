package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/victorialogs-benchmark/distributed-benchmark/internal/worker"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/config"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/logger"
)

var (
	configFile      string
	logLevel        string
	coordinatorURL  string
	nodeID          string
)

var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "VictoriaLogs Distributed Benchmark Worker",
	Long:  "Worker node for executing distributed VictoriaLogs benchmark queries",
	RunE:  runWorker,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&coordinatorURL, "coordinator-url", "", "Coordinator service URL")
	rootCmd.PersistentFlags().StringVar(&nodeID, "node-id", "", "Unique node identifier")
}

func runWorker(cmd *cobra.Command, args []string) error {
	// Initialize logger
	log := logger.NewLogger(logLevel)
	
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Override config with command line flags
	if coordinatorURL != "" {
		cfg.Worker.CoordinatorURL = coordinatorURL
	}
	if nodeID != "" {
		cfg.Worker.NodeID = nodeID
	}
	
	// Auto-generate node ID if not provided
	if cfg.Worker.NodeID == "" {
		hostname, _ := os.Hostname()
		cfg.Worker.NodeID = fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())
	}
	
	// Create worker service
	workerService, err := worker.NewService(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create worker service: %w", err)
	}
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start worker service
	go func() {
		if err := workerService.Start(ctx); err != nil {
			log.WithError(err).Error("Worker service failed")
			cancel()
		}
	}()
	
	log.WithField("node_id", cfg.Worker.NodeID).Info("Worker service started successfully")
	
	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutdown signal received, stopping worker...")
	
	// Graceful shutdown
	cancel()
	workerService.Stop()
	
	log.Info("Worker service stopped")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}