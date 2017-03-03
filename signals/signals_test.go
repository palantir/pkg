package signals_test

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/signals"
)

func TestCancelOnSignalsContext(t *testing.T) {
	ctx, _ := signals.CancelOnSignalsContext(context.Background(), syscall.SIGHUP)
	timer := time.NewTimer(time.Second * 3)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	go func() {
		err = proc.Signal(syscall.SIGHUP)
		require.NoError(t, err)
	}()

	done := false
	select {
	case <-ctx.Done():
		done = true
	case <-timer.C:
	}

	assert.True(t, done)
}

func TestRegisterStackTraceWriterOnSignals(t *testing.T) {
	out := &bytes.Buffer{}
	signals.RegisterStackTraceWriterOnSignals(out, syscall.SIGHUP)

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGHUP)
	require.NoError(t, err)

	// add sleep because write to buffer happens on a separate channel
	time.Sleep(1 * time.Second)

	// output stack should contain current routine
	assert.Contains(t, out.String(), "signals_test.TestRegisterStackTraceWriterOnSignals")
}

func TestUnregisterStackTraceWriterOnSignals(t *testing.T) {
	out := &bytes.Buffer{}
	unregister := signals.RegisterStackTraceWriterOnSignals(out, syscall.SIGHUP)
	unregister()

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGHUP)
	require.NoError(t, err)

	// add sleep because write to buffer happens on a separate channel
	time.Sleep(1 * time.Second)

	// output stack should be empty
	assert.Empty(t, out.String())
}
