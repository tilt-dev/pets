package health

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/windmilleng/pets/internal/proc"
)

// The go stdlib returns this error if we try to signal
// a process that's still being set up.
const processStartingErr = "no data yet"

func ProcessAliveCheck(pid int) healthcheck.Check {
	return func() error {
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		// 0 is a POSIX trick to ask the OS if it knows about the process
		return process.Signal(syscall.Signal(0))
	}
}

// Perform the health check until:
// 1) the process dies, or
// 2) the tcp health check passes
func WaitForTCP(process proc.PetsProc, interval time.Duration) error {
	ctx, cancel := context.WithCancel(context.Background())

	// The healthchecks start background goroutines. Cancel them all
	// when we're done waiting.
	defer cancel()

	pid := process.Pid
	host := process.Host()
	processAliveCheck := healthcheck.AsyncWithContext(
		ctx, ProcessAliveCheck(pid), interval)
	tcpHealthCheck := healthcheck.AsyncWithContext(
		ctx, healthcheck.TCPDialCheck(host, interval), interval)
	for true {
		err := processAliveCheck()
		if err != nil {
			// if the process is still starting, that's fine. Try again later.
			if err.Error() == processStartingErr {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			return fmt.Errorf("Process died without opening a network connection")
		}

		err = tcpHealthCheck()
		if err == nil {
			// The process is alive and it passed the TCP check!
			return nil
		}

		// Otherwise, sleep and try again later.
		time.Sleep(interval)
	}
	return nil
}
