package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/eniolaomotee/BlogGator-Go/service"
)

func ServiceManagerHandler(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		printServiceHelp()
		return nil
	}
	actions := cmd.Args[0]

	// get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("couldn't execute path: %w", err)
	}

	// Create logs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("couldn't get home directory: %w", err)
	}
	logDir := filepath.Join(homeDir, ".gator", "logs")
	pidFile := filepath.Join(homeDir, ".gator", "gator.pid")

	// create service manager
	manager := service.NewManager(logDir)

	// Parse gator interval
	aggInterval := "1m"
	if len(cmd.Args) > 1 && actions == "start" {
		aggInterval = cmd.Args[1]
	}

	// Add aggregator service
	err = manager.AddService(service.ServiceConfig{
		Name:           "gator-gg",
		Command:        executable,
		Args:           []string{"agg", aggInterval}, // run aggregator every 1 minute
		MaxRestarts:    10,
		RestartDelay:   5 * time.Second,
		CrashThreshold: 30 * time.Second,
		AutoRestart:    true,
	})
	if err != nil {
		return fmt.Errorf("couldn't add service: %w", err)
	}

	switch actions {
	case "start":
		return handleServiceStart(manager, pidFile, logDir)
	case "stop":
		return handleServiceStop(pidFile)
	case "restart":
		if err := handleServiceStop(pidFile); err != nil {
			fmt.Printf("warning: %v", err)
		}
		time.Sleep(2 * time.Second)
		return handleServiceStart(manager, pidFile, logDir)
	case "status":
		return handleServiceStatus(pidFile, logDir)
	case "logs":
		lines := 50
		if len(cmd.Args) > 1 {
			fmt.Sscanf(cmd.Args[1], "%d", &lines)
		}
		return handleServiceLogs(logDir, lines)
	case "tail":
		return handleServiceTail(logDir)
	default:
		return fmt.Errorf("unknown action %s", actions)
	}
}

func printServiceHelp() {
	fmt.Println("Gator Service Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gator service start [interval]  - Start the service manager")
	fmt.Println("  gator service stop              - Stop the service manager")
	fmt.Println("  gator service restart           - Restart the service manager")
	fmt.Println("  gator service status            - Show service status")
	fmt.Println("  gator service logs [lines]      - Show service logs")
	fmt.Println("  gator service tail              - Tail service logs in real-time")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gator service start             - Start with 1m interval")
	fmt.Println("  gator service start 30s         - Start with 30s interval")
	fmt.Println("  gator service logs 100          - Show last 100 log lines")
}

func handleServiceStart(manager *service.Manager, pidFile, logDir string) error {
	// Check if running
	if isServiceRunning(pidFile) {
		fmt.Println("service is already running")
	}
	fmt.Println("Starting gator service manager...")
	fmt.Printf("PID file: %s", pidFile)
	fmt.Printf(" Logs: %s/gator-agg.log\n", logDir)
	fmt.Println()

	// Create deamon
	deamon := service.NewDaemon(service.DaemonConfig{
		PIDFile:    pidFile,
		LogFile:    filepath.Join(logDir, "gator-daemon.log"),
		WorkingDir: "",
	}, manager)

	// start daemon
	if err := deamon.Start(); err != nil {
		return fmt.Errorf("couldn't start daemon %s", err)
	}
	return nil
}

func handleServiceStop(pidFile string) error {
	if !isServiceRunning(pidFile) {
		fmt.Println("Service is not running")
		return nil
	}

	fmt.Println(" Stopping Gator service manager...")

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("couldn't read PID file: %w", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return fmt.Errorf("couldn't parse PID: %w", err)
	}

	// Find process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("couldn't find process: %w", err)
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("couldn't stop process: %w", err)
	}

	// Wait for process to stop
	for i := 0; i < 10; i++ {
		if !isServiceRunning(pidFile) {
			fmt.Println("Service stopped successfully")
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Force kill if still running
	fmt.Println("Service didn't stop gracefully, forcing...")
	if err := process.Kill(); err != nil {
		return fmt.Errorf("couldn't kill process: %w", err)
	}

	os.Remove(pidFile)
	fmt.Println("Service stopped")
	return nil
}

func handleServiceStatus(pidFile, logDir string) error {
	fmt.Println("=== Gator Service Status ===")

	if isServiceRunning(pidFile) {
		data, _ := os.ReadFile(pidFile)
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)

		fmt.Printf("Status: Running\n")
		fmt.Printf("PID: %d\n", pid)
		fmt.Printf("Log file: %s/gator-agg.log\n", logDir)

		// Show uptime if possible
		if info, err := os.Stat(pidFile); err == nil {
			uptime := time.Since(info.ModTime())
			fmt.Printf("Uptime: %v\n", uptime.Round(time.Second))
		}
	} else {
		fmt.Println("Status: Stopped")
	}

	return nil
}
func handleServiceLogs(logDir string, lines int) error {
	logFile := filepath.Join(logDir, "gator-agg.log")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("No logs found. Service may not have been started yet.")
		return nil
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("couldn't read log file: %w", err)
	}

	logLines := strings.Split(string(content), "\n")
	start := len(logLines) - lines
	if start < 0 {
		start = 0
	}

	fmt.Printf("=== Last %d lines of gator-agg.log ===\n\n", len(logLines)-start)
	for _, line := range logLines[start:] {
		if line != "" {
			fmt.Println(line)
		}
	}

	return nil
}

func handleServiceTail(logDir string) error {
	logFile := filepath.Join(logDir, "gator-agg.log")

	fmt.Printf("Tailing %s (Ctrl+C to stop)...\n\n", logFile)

	// Use tail command if available
	cmd := exec.Command("tail", "-f", logFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func isServiceRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
