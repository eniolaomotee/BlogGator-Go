package service

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type DaemonConfig struct {
	PIDFile    string
	LogFile    string
	WorkingDir string
}

type Daemon struct {
	config  DaemonConfig
	manager *Manager
}

func NewDaemon(config DaemonConfig, manager *Manager) *Daemon {
	return &Daemon{
		config:  config,
		manager: manager,
	}
}

func (d *Daemon) Start() error {
	// Check if already running
	if d.isRunning() {
		return fmt.Errorf("daemon is already running")
	}

	// Write PID file
	if err := d.writePIDFile(); err != nil {
		return fmt.Errorf("couldn't write PID file: %w", err)
	}

	// Setup logging
	if d.config.LogFile != "" {
		logFile, err := os.OpenFile(d.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("couldn't open log file: %w", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	// Change working directory
	if d.config.WorkingDir != "" {
		if err := os.Chdir(d.config.WorkingDir); err != nil {
			return fmt.Errorf("couldn't change directory: %w", err)
		}
	}

	log.Printf("Daemon started (PID: %d)", os.Getpid())

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Start all services
	if err := d.manager.StartAll(); err != nil {
		return fmt.Errorf("couldn't start services: %w", err)
	}

	// Main loop

	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP:
			// Reload configuration
			log.Println("Received SIGHUP, reloading...")
			if err := d.manager.StopAll(); err != nil {
				log.Printf("Error stopping services: %v", err)
			}
			time.Sleep(1 * time.Second)
			if err := d.manager.StartAll(); err != nil {
				log.Printf("Error starting services: %v", err)
			}

		case os.Interrupt, syscall.SIGTERM:
			log.Println("Received shutdown signal")
			if err := d.Stop(); err != nil {
				log.Printf("Error stopping daemon %v", err)
			}
		}

	}
	return nil
}

func (d *Daemon) Stop() error {
	log.Println("Stopping daemon...")

	// Stop all services
	if err := d.manager.StopAll(); err != nil {
		log.Printf("Error stopping services: %v", err)
	}

	// Remove PID file
	if err := os.Remove(d.config.PIDFile); err != nil {
		log.Printf("Error removing PID file: %v", err)
	}

	log.Println("Daemon stopped")
	return nil
}

func (d *Daemon) isRunning() bool {
	if _, err := os.Stat(d.config.PIDFile); os.IsNotExist(err) {
		return false
	}

	// Read PID from file
	data, err := os.ReadFile(d.config.PIDFile)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false
	}

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (d *Daemon) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(d.config.PIDFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func (d *Daemon) GetPID() (int, error) {
	data, err := os.ReadFile(d.config.PIDFile)
	if err != nil {
		return 0, err
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, err
	}
	return pid, nil
}
