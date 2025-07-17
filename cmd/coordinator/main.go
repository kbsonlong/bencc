package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/victorialogs-benchmark/distributed-benchmark/internal/coordinator"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/config"
	"github.com/victorialogs-benchmark/distributed-benchmark/pkg/logger"
)

var (
	configFile string
	logLevel   string
)

var rootCmd = &cobra.Command{
	Use:   "coordinator",
	Short: "VictoriaLogs Distributed Benchmark Coordinator",
	Long:  "Coordinator service for managing distributed VictoriaLogs benchmark testing",
	RunE:  runCoordinator,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
}

func runCoordinator(cmd *cobra.Command, args []string) error {
	// Initialize logger
	log := logger.NewLogger(logLevel)
	
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Create coordinator service
	coordinatorService, err := coordinator.NewService(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create coordinator service: %w", err)
	}
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start coordinator service
	go func() {
		if err := coordinatorService.Start(ctx); err != nil {
			log.WithError(err).Error("Coordinator service failed")
			cancel()
		}
	}()
	
	log.Info("Coordinator service started successfully")
	
	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutdown signal received, stopping coordinator...")
	
	// Graceful shutdown
	cancel()
	coordinatorService.Stop()
	
	log.Info("Coordinator service stopped")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}