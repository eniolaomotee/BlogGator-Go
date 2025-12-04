# BlogGator Service Manager

The Gator service manager keeps your RSS aggregator running in the background with automatic restarts on crashes.

## Quick Start

```bash
# Start the service (aggregates every 1 minute)
gator service start

# Start with custom interval
gator service start 30s

# Check status
gator service status

# View logs
gator service logs

# Tail logs in real-time
gator service tail

# Stop the service
gator service stop

# Restart the service
gator service restart


Features
✅ Auto-restart on crash - Automatically restarts if the aggregator crashes
✅ Configurable restart limits - Prevents infinite restart loops
✅ Crash detection - Distinguishes between crashes and normal exits
✅ Detailed logging - All output logged to ~/.gator/logs/
✅ Graceful shutdown - Properly stops services on SIGTERM
✅ Status monitoring - Check service health and uptime



## System Service Installation
``` bash
# Install as system service
./scripts/install-service.sh

# Start service
sudo systemctl start gator

# Enable on boot
sudo systemctl enable gator

# Check status
sudo systemctl status gator

# View logs
sudo journalctl -u gator -f
```


## macOS (launchd)
``` bash
# Install as user service
./scripts/install-service.sh

# Service starts automatically
# To manually control:
launchctl start com.gator.aggregator
launchctl stop com.gator.aggregator

# View logs
tail -f ~/.gator/logs/gator-service.log
```

## Configuration
Service configuration is in service/manager.go:

MaxRestarts: Maximum restart attempts (default: 10)
RestartDelay: Delay between restarts (default: 5s)
CrashThreshold: Time threshold for crash detection (default: 30s)
AutoRestart: Enable/disable auto-restart (default: true)


## Troubleshooting
Service won't start:

``` bash
# Check if already running
gator service status

# Check logs for errors
gator service logs 100
``

## Too many restarts:
``` bash
# Check logs to identify the issue
gator service logs

# Stop the service and fix the underlying problem
gator service stop
```

## Service not stopping
``` bash
# Force stop by killing the process
kill $(cat ~/.gator/gator.pid)
rm ~/.gator/gator.pid
```

## Log Files

All logs are stored in ~/.gator/logs/:

``` bash
gator-agg.log - Aggregator output
gator-daemon.log - Service manager logs
```
