// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"sync"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources/window"
	"github.com/palantir/witchcraft-go-health/status"
	"github.com/rs/zerolog"
)

var (
	_ io.WriteCloser           = (*TCPWriter)(nil)
	_ status.HealthCheckSource = (*TCPWriter)(nil)
)

const (
	errWriterClosed = "writer is closed"

	// TCPWriterHealthCheckName is the name used for the external health check
	TCPWriterHealthCheckName health.CheckType = "TCP_LOGGER_CONNECTION_STATUS"
)

// envelopeSerializerFunc provides a way to change the serialization method for the provided payload.
type envelopeSerializerFunc func(payload []byte) ([]byte, error)

type TCPWriter struct {
	provider           ConnProvider
	envelopeSerializer envelopeSerializerFunc

	// closedChan is used to signal that the writer is shutting down
	closedChan chan struct{}

	mu   sync.RWMutex // guards conn below
	conn net.Conn

	health window.ErrorHealthCheckSource
}

// NewTCPWriter returns an io.WriteCloser that writes logs to a TCP socket and wraps
// them with the provided envelope metadata. TCP connections are retrieved using the
// ConnProvider and connection state is managed internally.
func NewTCPWriter(metadata LogEnvelopeMetadata, provider ConnProvider) *TCPWriter {
	return newTCPWriterInternal(provider, zerologSerializer(metadata))
}

func newTCPWriterInternal(provider ConnProvider, serializerFunc envelopeSerializerFunc) *TCPWriter {
	return &TCPWriter{
		envelopeSerializer: serializerFunc,
		provider:           provider,
		closedChan:         make(chan struct{}),
		conn:               nil,
		health: window.MustNewErrorHealthCheckSource(
			TCPWriterHealthCheckName,
			window.HealthyIfNoRecentErrors,
			window.WithFailingHealthStateValue(health.HealthState_WARNING)),
	}
}

// Write implements the io.Writer interface for use with TCP sockets.
// The provided input is wrapped in a LogEnvelopeV1 and serialized as JSON before writing to the underlying socket.
// If there is a connection error before or during writing, the connection will be closed and an error will be returned.
// If a subsequent Write is called after Close, then this will return immediately with an error.
func (d *TCPWriter) Write(p []byte) (n int, err error) {
	defer func() {
		d.health.Submit(err)
	}()

	if d.closed() {
		return 0, werror.Error(errWriterClosed)
	}

	// remove the trailing new-line delimiter if it exists before wrapping in the envelope
	envelope, err := d.envelopeSerializer(bytes.TrimSuffix(p, []byte("\n")))
	if err != nil {
		return 0, werror.Wrap(err, "failed to serialize the envelope")
	}

	conn, err := d.getConn()
	if err != nil {
		return 0, err
	}

	var total int
	for total < len(envelope) {
		n, err := conn.Write(envelope[total:])
		total += n
		if err != nil {
			if nerr, ok := err.(net.Error); !ok || nerr.Timeout() || !nerr.Temporary() {
				// permanent error or timeout so close the connection
				_ = d.closeConn()
				return total, err
			}
			return total, err
		}
	}
	return len(p), nil
}

func (d *TCPWriter) getConn() (net.Conn, error) {
	// Fast path: connection is already established and cached
	d.mu.RLock()
	conn := d.conn
	d.mu.RUnlock()
	if d.conn != nil {
		return conn, nil
	}

	// No active connection, so use the provider to get a new net.Conn, and cache it.
	d.mu.Lock()
	defer d.mu.Unlock()
	newConn, err := d.provider.GetConn()
	if err != nil {
		return nil, err
	}
	d.conn = newConn
	return newConn, nil
}

func (d *TCPWriter) closeConn() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn != nil {
		err := d.conn.Close()
		d.conn = nil
		return err
	}
	return nil
}

func (d *TCPWriter) closed() bool {
	select {
	case <-d.closedChan:
		return true
	default:
		return false
	}
}

// Close will close any existing client connections and shuts down the writer from any future writes.
func (d *TCPWriter) Close() error {
	close(d.closedChan)
	return d.closeConn()
}

func (d *TCPWriter) HealthStatus(ctx context.Context) health.HealthStatus {
	return d.health.HealthStatus(ctx)
}

func zerologSerializer(metadata LogEnvelopeMetadata) envelopeSerializerFunc {
	// create a new top-level logger that writes to ioutil.Discard since each
	// serialization will write to its own local buffer instead of a single writer
	logger := zerolog.New(ioutil.Discard).With().
		Str("type", "envelope.1").
		Str("deployment", metadata.Deployment).
		Str("environment", metadata.Environment).
		Str("environmentId", metadata.EnvironmentID).
		Str("host", metadata.Host).
		Str("nodeId", metadata.NodeID).
		Str("service", metadata.Service).
		Str("serviceId", metadata.ServiceID).
		Str("stack", metadata.Stack).
		Str("stackId", metadata.StackID).
		Str("product", metadata.Product).
		Str("productVersion", metadata.ProductVersion).
		Logger()
	return func(p []byte) ([]byte, error) {
		var buf bytes.Buffer
		l := logger.Output(&buf)
		if p != nil {
			l = l.With().RawJSON("payload", p).Logger()
		}
		l.Log().Send()
		return buf.Bytes(), nil
	}
}
