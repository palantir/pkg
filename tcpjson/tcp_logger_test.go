// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-logging/conjure/witchcraft/api/logging"
	"github.com/palantir/witchcraft-go-logging/wlog"
	"github.com/palantir/witchcraft-go-logging/wlog/svclog/svc1log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testMetadata = LogEnvelopeMetadata{
		Type:           "envelope.1",
		Deployment:     "test-deployment",
		Environment:    "test-environment",
		EnvironmentID:  "test-environment-id",
		Host:           "test-host",
		NodeID:         "test-node-id",
		Product:        "test-product",
		ProductVersion: "test-product-version",
		Service:        "test-service",
		ServiceID:      "test-service-id",
		Stack:          "test-stack",
		StackID:        "test-stack-id",
	}
	logPayload = []byte(`{"type": "service.1","message":"test","level":"INFO"}\n`)
)

func TestWrite(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"payload-with-newline", logPayload},
		{"payload-without-newline", bytes.TrimSuffix(logPayload, []byte("\n"))},
		{"no-payload", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectedEnvelope := getEnvelopeBytes(t, tc.payload)
			provider := new(bufferedConnProvider)
			tcpWriter := NewTCPWriter(testMetadata, provider)
			n, err := tcpWriter.Write(tc.payload)
			require.NoError(t, err)
			require.Equal(t, len(tc.payload), n)
			buf := provider.buffer.Bytes()
			require.True(t, bytes.Equal(buf, expectedEnvelope))
		})
	}
}

// TestWrite_Timeout asserts that the connection is closed on a timeout error.
func TestWrite_Timeout(t *testing.T) {
	// startup an in-memory TLS server to write TCP logs to
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	// get a TCP connection using the TCPConnProvider
	uris := []string{fmt.Sprintf("%s://%s", server.Listener.Addr().Network(), server.Listener.Addr().String())}
	connProvider, err := NewTCPConnProvider(uris, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, err)
	conn, err := connProvider.GetConn()
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// create the TCPWriter with a small net conn wrapper so that we can control the state of the connection below.
	tcpWriter := NewTCPWriter(testMetadata, &netConnProvider{conn: conn})

	// initial write should succeed
	n, err := tcpWriter.Write(logPayload)
	require.NoError(t, err)
	require.Equal(t, len(logPayload), n)
	// conn is cached since it's active
	require.NotNil(t, tcpWriter.conn)

	// set a deadline which should cause the Write to timeout
	err = conn.SetDeadline(time.Now())
	require.NoError(t, err)
	_, err = tcpWriter.Write(logPayload)
	require.Error(t, err)
	require.True(t, isTimeoutError(err))
	require.False(t, isTemporaryError(err))
	// conn is nil, since it should be closed
	require.Nil(t, tcpWriter.conn)

	// subsequent writes should error since the connection should be closed
	_, err = tcpWriter.Write(logPayload)
	require.Error(t, err)
	require.False(t, isTimeoutError(err))
	require.False(t, isTemporaryError(err))
	require.True(t,
		strings.Contains(err.Error(), "use of closed network connection") ||
			strings.Contains(err.Error(), "tls: use of closed connection"),
	)
	// conn is nil, since it should be closed
	require.Nil(t, tcpWriter.conn)
}

func isTimeoutError(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}

func isTemporaryError(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}
	return false
}

// TestWriteFromSvc1log is more of an integration style test which verifies the envelopes written
// are as expected when using the TCPWriter as an io.Writer for svc1log.
func TestWriteFromSvc1log(t *testing.T) {
	provider := new(bufferedConnProvider)
	tcpWriter := NewTCPWriter(testMetadata, provider)
	logger := svc1log.NewFromCreator(tcpWriter, wlog.DebugLevel, wlog.NewJSONMarshalLoggerProvider().NewLeveledLogger)
	logger.Debug("this is a test")

	buf := provider.buffer.Bytes()
	var gotEnvelope LogEnvelopeV1
	err := json.Unmarshal(buf, &gotEnvelope)
	require.NoError(t, err)

	// Verify all envelope metadata
	assert.Equal(t, testMetadata.Type, gotEnvelope.Type)
	assert.Equal(t, testMetadata.Deployment, gotEnvelope.Deployment)
	assert.Equal(t, testMetadata.Environment, gotEnvelope.Environment)
	assert.Equal(t, testMetadata.EnvironmentID, gotEnvelope.EnvironmentID)
	assert.Equal(t, testMetadata.Host, gotEnvelope.Host)
	assert.Equal(t, testMetadata.NodeID, gotEnvelope.NodeID)
	assert.Equal(t, testMetadata.Product, gotEnvelope.Product)
	assert.Equal(t, testMetadata.ProductVersion, gotEnvelope.ProductVersion)
	assert.Equal(t, testMetadata.Service, gotEnvelope.Service)
	assert.Equal(t, testMetadata.ServiceID, gotEnvelope.ServiceID)
	assert.Equal(t, testMetadata.Stack, gotEnvelope.Stack)
	assert.Equal(t, testMetadata.StackID, gotEnvelope.StackID)

	// Verify the payload
	gotPayload := new(logging.ServiceLogV1)
	err = gotPayload.UnmarshalJSON(gotEnvelope.Payload)
	require.NoError(t, err)
	assert.Equal(t, "this is a test", gotPayload.Message)
	assert.Equal(t, logging.New_LogLevel(logging.LogLevel_DEBUG), gotPayload.Level)
}

// TestClosedWriter verifies the behavior of attempting to write when the writer is closed.
func TestClosedWriter(t *testing.T) {
	provider := new(bufferedConnProvider)
	tcpWriter := NewTCPWriter(testMetadata, provider)

	n, err := tcpWriter.Write(logPayload)
	require.NoError(t, err)
	require.Equal(t, len(logPayload), n)

	err = tcpWriter.Close()
	require.NoError(t, err)

	// Attempt a write and expect that the writer is closed
	n, err = tcpWriter.Write(logPayload)
	require.Error(t, err)
	require.EqualError(t, err, errWriterClosed)
	require.True(t, n == 0)
}

func TestHealthStatus(t *testing.T) {
	for _, tc := range []struct {
		name          string
		tcpWriterFunc func() *TCPWriter
		expected      health.HealthState_Value
	}{
		{
			name: "not started",
			tcpWriterFunc: func() *TCPWriter {
				return NewTCPWriter(LogEnvelopeMetadata{}, &bufferedConnProvider{})
			},
			expected: health.HealthState_HEALTHY,
		},
		{
			name: "shutting down",
			tcpWriterFunc: func() *TCPWriter {
				w := NewTCPWriter(LogEnvelopeMetadata{}, &bufferedConnProvider{})
				_ = w.Close()
				return w
			},
			expected: health.HealthState_HEALTHY,
		},
		{
			name: "established connection",
			tcpWriterFunc: func() *TCPWriter {
				w := NewTCPWriter(LogEnvelopeMetadata{}, &bufferedConnProvider{})
				_, err := w.Write(logPayload)
				require.NoError(t, err)
				return w
			},
			expected: health.HealthState_HEALTHY,
		},
		{
			name: "failed to get connection",
			tcpWriterFunc: func() *TCPWriter {
				w := NewTCPWriter(LogEnvelopeMetadata{}, &failingConnProvider{err: fmt.Errorf("error")})
				_, err := w.Write(logPayload)
				assert.Error(t, err)
				return w
			},
			expected: health.HealthState_WARNING,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tcpWriter := tc.tcpWriterFunc()
			gotHealthState := tcpWriter.HealthStatus(context.Background())
			assert.NotNil(t, gotHealthState)
			assert.Len(t, gotHealthState.Checks, 1)
			assert.Equal(t, tc.expected, gotHealthState.Checks[TCPWriterHealthCheckName].State.Value())
		})
	}
}

func getEnvelopeBytes(t *testing.T, payload []byte) []byte {
	envelope, err := zerologSerializer(testMetadata)(payload)
	require.NoError(t, err)
	return envelope
}

// bufferedConnProvider is a mock ConnProvider that writes to an internal
// bytes buffer instead of to the net.Conn.
type bufferedConnProvider struct {
	net.Conn
	err    error
	buffer bytes.Buffer
	// writeCount tracks the total number of writes that are called.
	// This field should only be used in testing and should only be updated/read with atomic operations.
	writeCount int32
}

func (t *bufferedConnProvider) GetConn() (net.Conn, error) {
	return t, nil
}

func (t *bufferedConnProvider) Write(d []byte) (int, error) {
	atomic.AddInt32(&t.writeCount, 1)
	return t.buffer.Write(d)
}

func (t *bufferedConnProvider) Close() error {
	return t.err
}

type failingConnProvider struct {
	err error
}

func (t *failingConnProvider) GetConn() (net.Conn, error) {
	return nil, t.err
}

type netConnProvider struct {
	conn net.Conn
}

func (n *netConnProvider) GetConn() (net.Conn, error) {
	return n.conn, nil
}

// BenchmarkEnvelopeSerializer records the total time and memory allocations for each envelope serializer.
func BenchmarkEnvelopeSerializer(b *testing.B) {
	for _, tc := range []struct {
		name           string
		serializerFunc envelopeSerializerFunc
	}{
		{"zerolog", zerologSerializer(testMetadata)},
		{"JSON-Encoder", jsonEncoderSerializer(testMetadata)},
		{"JSON-Marshaler", jsonMarshalSerializer(testMetadata)},
		{"manual", manualSerializer(testMetadata)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				_, _ = tc.serializerFunc(logPayload)
			}
		})
	}
}

// jsonEncoderSerializer returns an envelopeSerializerFunc that uses the json.Encoder to serialize the envelope.
func jsonEncoderSerializer(metadata LogEnvelopeMetadata) envelopeSerializerFunc {
	return func(p []byte) ([]byte, error) {
		var buf bytes.Buffer
		envelopeToWrite := getEnvelopeWithPayload(metadata, p)
		if err := json.NewEncoder(&buf).Encode(&envelopeToWrite); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

// jsonMarshalSerializer returns an envelopeSerializerFunc that uses the json.Marshal to serialize the envelope.
func jsonMarshalSerializer(metadata LogEnvelopeMetadata) envelopeSerializerFunc {
	return func(p []byte) ([]byte, error) {
		envelopeToWrite := getEnvelopeWithPayload(metadata, p)
		b, err := json.Marshal(&envelopeToWrite)
		if err != nil {
			return nil, err
		}
		return append(b, '\n'), nil
	}
}

// manualSerializer returns an envelopeSerializerFunc that manually injects the payload.
func manualSerializer(metadata LogEnvelopeMetadata) envelopeSerializerFunc {
	metadataJSON, _ := jsonEncoderSerializer(metadata)(nil)
	return func(p []byte) ([]byte, error) {
		// manually inject the payload into the metadataJSON
		idx := bytes.LastIndexByte(metadataJSON, '}')
		if idx == -1 {
			return nil, werror.Error("invalid JSON")
		}
		envelope := bytes.NewBuffer(metadataJSON[:idx])
		envelope.Write([]byte(`,"payload":`))
		envelope.Write(p)
		envelope.Write([]byte(`}\n`))
		return envelope.Bytes(), nil
	}
}

func getEnvelopeWithPayload(metadata LogEnvelopeMetadata, payload []byte) LogEnvelopeV1 {
	return LogEnvelopeV1{
		LogEnvelopeMetadata: LogEnvelopeMetadata{
			Type:           "envelope.1",
			Deployment:     metadata.Deployment,
			Environment:    metadata.Environment,
			EnvironmentID:  metadata.EnvironmentID,
			Host:           metadata.Host,
			NodeID:         metadata.NodeID,
			Service:        metadata.Service,
			ServiceID:      metadata.ServiceID,
			Stack:          metadata.Stack,
			StackID:        metadata.StackID,
			Product:        metadata.Product,
			ProductVersion: metadata.ProductVersion,
		},
		Payload: payload,
	}
}
