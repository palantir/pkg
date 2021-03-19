// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/palantir/pkg/metrics"
	"github.com/palantir/witchcraft-go-logging/conjure/witchcraft/api/logging"
	"github.com/palantir/witchcraft-go-logging/wlog"
	"github.com/palantir/witchcraft-go-logging/wlog/svclog/svc1log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncWriter(t *testing.T) {
	out := &bytes.Buffer{}
	w := StartAsyncWriter(out, metrics.DefaultMetricsRegistry)
	for i := 0; i < 5; i++ {
		str := strconv.Itoa(i)
		go func() {
			_, _ = w.Write([]byte(str))
		}()
	}
	time.Sleep(time.Millisecond)

	written := out.String()
	t.Log(written)
	assert.Len(t, written, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, written, strconv.Itoa(i))
	}

	t.Run("fails when closed", func(t *testing.T) {
		require.NoError(t, w.Close())
		_, err := w.Write([]byte("will fail!"))
		require.EqualError(t, err, "write to closed asyncWriter")
	})
}

// TestAsyncWriteWithSvc1log verifies the that svc1log lines are properly written to the output
// when using the async writer. This also ensures the original input bytes are added to the buffered channel
// correctly and the underlying byte slice is not stored directly which would cause the output to be malformed.
func TestAsyncWriteWithSvc1log(t *testing.T) {
	provider := &bufferedConnProvider{}
	tcpWriter := NewTCPWriter(testMetadata, provider)
	asyncTCPWriter := StartAsyncWriter(tcpWriter, metrics.DefaultMetricsRegistry)
	defer func() {
		_ = asyncTCPWriter.Close()
	}()
	logger := svc1log.NewFromCreator(asyncTCPWriter, wlog.DebugLevel, wlog.NewJSONMarshalLoggerProvider().NewLeveledLogger)

	// Write log lines with deterministic messages to verify later
	const totalLogLines = 100
	for i := 0; i < totalLogLines; i++ {
		logger.Debug(strconv.Itoa(i))
	}

	// Wait for write count to match the total log lines, otherwise fail the test after a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timed out waiting to receive all log lines")
		default:
		}
		if atomic.LoadInt32(&provider.writeCount) == int32(totalLogLines) {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// verify all log lines are received and well formed
	logLines := bytes.SplitN(provider.buffer.Bytes(), []byte("\n"), totalLogLines)
	assert.Equal(t, totalLogLines, len(logLines))

	for i, logLine := range logLines {
		var gotEnvelope LogEnvelopeV1
		err := json.Unmarshal(logLine, &gotEnvelope)
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
		assert.Equal(t, strconv.Itoa(i), gotPayload.Message)
		assert.Equal(t, logging.New_LogLevel(logging.LogLevel_DEBUG), gotPayload.Level)
	}
}
