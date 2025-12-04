package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusRestart ServiceStatus = "restart"
	StatusCrashed ServiceStatus = "crashed"
)

type ServiceConfig struct {
	Name           string
	Command        string
	Args           []string
	MaxRestarts    int           // maximum restarts before giving up
	RestartDelay   time.Duration // delay between restarts
	CrashThreshold time.Duration // if service crashes within this time , count as crash
	LogFile        string        // Path to log file
	AutoRestart    bool          //Auto-restart on crash

}

type Service struct {
	config       ServiceConfig
	cmd          *exec.Cmd
	status       ServiceStatus
	restartCount int
	startTime    time.Time
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	logFile      *os.File
}

type Manager struct {
	services map[string]*Service
	mu       sync.RWMutex
	logDir   string
}

func NewManager(logDir string) *Manager {
	// Create a log directory if one doesnt' exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning couldn't create log directory %v", err)
	}

	m := &Manager{
		services: make(map[string]*Service),
		logDir:   logDir,
	}
	return m
}

func (m *Manager) AddService(config ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.services[config.Name]
	if exists {
		return fmt.Errorf("service %s already exists", config.Name)
	}

	// Set defaults
	if config.MaxRestarts == 0 {
		config.MaxRestarts = 5
	}
	if config.RestartDelay == 0 {
		config.RestartDelay = 5 * time.Second
	}
	if config.CrashThreshold == 0 {
		config.CrashThreshold = 30 * time.Second
	}

	if config.LogFile == "" {
		config.LogFile = fmt.Sprintf("%s/%s.log", m.logDir, config.Name)
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		config:       config,
		status:       StatusStopped,
		restartCount: 0,
		ctx:          ctx,
		cancel:       cancel,
	}
	m.services[config.Name] = service

	return nil
}

func (m *Manager) Start(serviceName string) error {
	m.mu.RLock()
	service, exists := m.services[serviceName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	return service.start()

}

func (m *Manager) Stop(serviceName string) error {
	m.mu.RLock()
	service, exists := m.services[serviceName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	return service.stop()
}

func (m *Manager) Restart(serviceName string) error {
	err := m.Stop(serviceName)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	return m.Start(serviceName)
}

func (m *Manager) GetStatus(serviceName string) (ServiceStatus, error) {
	m.mu.RLock()
	service, exists := m.services[serviceName]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	service.mu.RLock()
	defer service.mu.RUnlock()
	return service.status, nil
}

func (m *Manager) GetAllStatus() map[string]ServiceStatus {
	m.mu.RLock()
	defer m.mu.Lock()

	status := make(map[string]ServiceStatus)
	for name, service := range m.services {
		service.mu.RLock()
		status[name] = service.status
		service.mu.RUnlock()
	}

	return status
}

func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RLock()

	for name := range m.services {
		if err := m.Start(name); err != nil {
			return fmt.Errorf("failed to start %s: %w", name, err)
		}
	}

	return nil
}

func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	for name := range m.services {
		if err := m.Stop(name); err != nil {
			lastErr = err
			log.Printf("Error stopping %s: %v", name, err)
		}
	}
	return lastErr
}

func (m *Manager) Run() error {
	// Start all services
	if err := m.StartAll(); err != nil {
		return err
	}

	// setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Service manager running. Press ctrl + c to stop")

	// Waiting for shutdown
	<-sigChan
	log.Println("Shutdown Signal received, Stopping all services....")

	// stop all services
	if err := m.StopAll(); err != nil {
		log.Printf("Error occured during shutdown %v", err)
	}

	log.Println("All services stopped")
	return nil
}

// Service Methods
func (s *Service) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusRunning {
		return fmt.Errorf("service %s is already running", s.config.Name)
	}

	// Open log file
	logFile, err := os.OpenFile(s.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open log file: %w", err)
	}
	s.logFile = logFile

	//Create command
	s.cmd = exec.CommandContext(s.ctx, s.config.Command, s.config.Args...)
	s.cmd.Stderr = logFile
	s.cmd.Stdout = logFile

	// start the process
	if err := s.cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("couldn't start process %w", err)
	}
	s.status = StatusRunning
	s.startTime = time.Now()

	log.Printf("[%s] Service started (PID: %d)", s.config.Name, s.cmd.Process.Pid)

	//Monitor the process via a goroutine
	go s.monitor()

	return nil

}

func (s *Service) stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusStopped {
		return nil
	}

	s.status = StatusStopped
	s.cancel() // cancel context

	if s.cmd != nil && s.cmd.Process == nil {
		// try graceful shutdown first
		log.Printf("[%s] sending SIGTERM...", s.config.Name)
		if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("[%s] Error sending SIGTERM: %v", s.config.Name, err)
		}

		// wait for graceful shutdown with timeout
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-time.After(10 * time.Second):
			//Force kill if graceful shutdown fails
			log.Printf("[%s] Graceful shutdown timeout, forcing kil...", s.config.Name)
			if err := s.cmd.Process.Kill(); err != nil {
				log.Printf("[%s] Error killing process: %v", s.config.Name, err)
			}
		case <-done:
			log.Printf("[%s] Service stopped gracefully", s.config.Name)
		}
	}

	if s.logFile != nil {
		s.logFile.Close()
	}

	return nil
}

func (s *Service) monitor() {
	err := s.cmd.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()

	// if already stopped, don't restart
	if s.status == StatusStopped {
		return
	}

	runtime := time.Since(s.startTime)

	// check if this was a crash (service died quickly)
	if runtime < s.config.CrashThreshold {
		s.restartCount++
		log.Printf("[%s] Service crashed after %v (restart count:%d/%d)", s.config.Name, runtime, s.restartCount, s.config.MaxRestarts)
	} else {
		//Service ran for a while, reset restart counter
		s.restartCount = 0
		log.Printf("[%s] Service exited after %v", s.config.Name, runtime)
	}

	if err != nil {
		log.Printf("[%s] Exit error: %v", s.config.Name, err)
	}

	// Check if we should restart
	if !s.config.AutoRestart {
		s.status = StatusStopped
		log.Printf("[%s] Auto-restart disabled, service stopped", s.config.Name)
		return
	}

	if s.restartCount >= s.config.MaxRestarts {
		s.status = StatusCrashed
		log.Printf("[%s] Max restarts reached (%d), giving up", s.config.Name, s.config.MaxRestarts)
		return
	}

	// Schedule restart
	s.status = StatusRestart
	log.Printf("[%s] Restarting in %v...", s.config.Name, s.config.RestartDelay)

	s.mu.Unlock()
	time.Sleep(s.config.RestartDelay)
	s.mu.Lock()

	// Check again if we should still restart
	if s.status != StatusRestart {
		return
	}

	// Close old log file
	if s.logFile != nil {
		s.logFile.Close()
	}

	// Reopen log file
	logFile, err := os.OpenFile(s.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("[%s] Couldn't reopen log file: %v", s.config.Name, err)
		s.status = StatusCrashed
		return
	}
	s.logFile = logFile

	// Create new command
	s.cmd = exec.CommandContext(s.ctx, s.config.Command, s.config.Args...)
	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile

	// Start the process
	if err := s.cmd.Start(); err != nil {
		log.Printf("[%s] Restart failed: %v", s.config.Name, err)
		s.status = StatusCrashed
		return
	}

	s.status = StatusRunning
	s.startTime = time.Now()
	log.Printf("[%s] Service restarted (PID: %d)", s.config.Name, s.cmd.Process.Pid)

	// Continue monitoring
	go s.monitor()

}

func (s *Service) GetInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]interface{}{
		"name":          s.config.Name,
		"status":        s.status,
		"restart_count": s.restartCount,
		"log_file":      s.config.LogFile,
	}

	if s.status == StatusRunning {
		info["uptime"] = time.Since(s.startTime).String()
		if s.cmd != nil && s.cmd.Process != nil {
			info["pid"] = s.cmd.Process.Pid
		}
	}

	return info
}

func (m *Manager) GetService(name string) (*Service, bool) {
	m.mu.Lock()
	defer m.mu.RUnlock()
	service, exists := m.services[name]
	return service, exists
}
