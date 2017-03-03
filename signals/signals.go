package signals

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

// NewCancelOnSigTermSigIntContext returns a context that is cancelled when the SIGTERM or SIGINT signal is received.
func NewCancelOnSigTermSigIntContext() context.Context {
	return NewCancelOnSignalsContext(syscall.SIGTERM, syscall.SIGINT)
}

// NewCancelOnSignalsContext returns a context that is cancelled when any of the provided signals are received.
func NewCancelOnSignalsContext(sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// Use a buffer of 1 in case we are not ready when the signal arrives
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sig...)

	go func() {
		<-signals
		cancel()
	}()

	return ctx
}

// RegisterStackTraceWriterOnSigQuit starts a goroutine that listens for the SIGQUIT (kill -3) signal and writes a
// pprof-formatted snapshot of all running goroutines when the signal is received.
func RegisterStackTraceWriterOnSigQuit(out io.Writer) {
	RegisterStackTraceWriter(out, syscall.SIGQUIT)
}

// RegisterStackTraceWriter starts a goroutine that listens for the specified signals and writes a pprof-formatted
// snapshot of all running goroutines to out when any of the provided signals are received.
func RegisterStackTraceWriter(out io.Writer, sig ...os.Signal) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sig...)

	go func() {
		for {
			<-signals
			err := pprof.Lookup("goroutine").WriteTo(out, 2)
			if err != nil {
				fmt.Fprintln(out, "Failed to dump goroutines")
			}
		}
	}()
}
