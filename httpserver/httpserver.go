// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httpserver provides functions that are useful for HTTP servers.
package httpserver

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AvailablePort returns a best-effort determination of an available port. Does so by opening a TCP listener on
// localhost, determining the port used by that listener, closing the listener and returning the address that was used
// by the listener. This is best-effort because there is no way to guarantee that another process will not take the port
// between the time when the listener is closed and the returned port is used by the caller.
func AvailablePort() (port int, rErr error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	defer func() {
		if err := l.Close(); err != nil && rErr == nil {
			rErr = err
		}
	}()
	if err != nil {
		return 0, err
	}

	addrString := l.Addr().String()
	port, err = strconv.Atoi(addrString[strings.LastIndex(addrString, ":")+1:])
	if err != nil {
		return 0, err
	}

	return port, nil
}

// URLReady returns a channel that is sent "true" when an http.Get executed against the provided URL returns a response
// with status code http.StatusOK. This is a convenience function that calls Ready with a readyCall that consists of
// sending a GET request to the provided URL and a readyResp that returns true on a 200 status.
func URLReady(url string, timeout time.Duration) <-chan bool {
	return Ready(func() (*http.Response, error) {
		return http.Get(url)
	}, func(resp *http.Response) bool {
		return resp.StatusCode == http.StatusOK
	}, timeout)
}

// Ready returns a channel that is sent "true" when the provided readyCall returns a nil error. The readyCall is invoked
// once every 100ms until it either returns a nil error or the provided timeout duration is reached, in which case
// "false" is sent on the channel.
//
// Note that any call that returns a nil error for readyCall is interpreted to mean that the server is ready. For most
// HTTP clients, any response (including a response with an error code) will result in a nil error: the error is
// typically only non-nil if a transport-level failure occurs.
//
// readyCall should by a function that returns quickly. At most one readyCall will be running at a particular time.
//
// Example:
//
//   Ready(func() (*http.Response, error) {
//     return http.Get(fmt.Sprintf("http://localhost:%d/example/ready", port))
//   }, func(resp *http.Response) bool {
//     return resp.StatusCode == http.StatusOK
//   }, 5 * time.Second)
func Ready(readyCall func() (*http.Response, error), readyResp func(*http.Response) bool, timeout time.Duration) <-chan bool {
	once := &sync.Once{}
	done := make(chan struct{})

	ready := make(chan bool)
	go func() {
		timeout := time.NewTimer(timeout)
		defer timeout.Stop()

		// start a separate goroutine with the ticker. Done so that the possibly expensive action will not
		// block the timeout.
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if resp, err := readyCall(); err == nil && readyResp(resp) {
						once.Do(func() {
							ready <- true
							close(done)
						})
					}
				}
			}
		}()

		// timeout channel
		for {
			select {
			case <-done:
				return
			case <-timeout.C:
				once.Do(func() {
					ready <- false
					close(done)
				})
				return
			}
		}
	}()
	return ready
}
