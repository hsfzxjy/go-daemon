package daemon

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// ErrStop should be returned signal handler function
// for termination of handling signals.
var ErrStop = errors.New("stop serve signals")

// SignalHandlerFunc is the interface for signal handler functions.
type SignalHandlerFunc func(sig os.Signal) (err error)

// SetSigHandler sets handler for the given signals.
// SIGTERM has the default handler, he returns ErrStop.
func SetSigHandler(handler SignalHandlerFunc, signals ...os.Signal) {
	for _, sig := range signals {
		handlers[sig] = handler
	}
}

// ServeSignals calls handlers for system signals.
func ServeSignals(ctx context.Context) (err error) {
	signals := make([]os.Signal, 0, len(handlers))
	for sig := range handlers {
		signals = append(signals, sig)
	}

	ch := make(chan os.Signal, 8)
	signal.Notify(ch, signals...)

LOOP:
	for {
		select {
		case sig, ok := <-ch:
			if !ok {
				break LOOP
			}
			err = handlers[sig](sig)
			if err != nil {
				break LOOP
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	signal.Stop(ch)

	if err == ErrStop {
		err = nil
	}

	return
}

var handlers = make(map[os.Signal]SignalHandlerFunc)

func init() {
	handlers[syscall.SIGTERM] = sigtermDefaultHandler
}

func sigtermDefaultHandler(sig os.Signal) error {
	return ErrStop
}
