package signals

import (
	"os"
	"os/signal"

	"owl-engine/pkg/xlogs"
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, shutdownSignals...)
	go func() {
		ch := <-c
		close(stop)

		xlogs.Infof("received %v, exiting gracefully...", ch)
	}()

	return stop
}
