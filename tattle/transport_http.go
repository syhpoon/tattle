/*
 MIT License

 Copyright (c) 2019 Max Kuznetsov <syhpoon@syhpoon.ca>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

package tattle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Parameters for TransportHttp instance
type TransportHttpParams struct {
	// Listener for Http server
	Listener     net.Listener
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	// RPC client-side timeout
	RpcTimeout             time.Duration
	DetectorInjectTimeout  time.Duration
	DetectorProcessTimeout time.Duration
	TLSCertFile            string
	TLSKeyFile             string
	Logger                 Logger
	IncomingBufferSize     int
	HttpClient             http.Client
	Codec                  Codec
	Ctx                    context.Context
}

// Peer for Http transport
type HttpPeer struct {
	Id       string
	Host     string
	Port     uint16
	Protocol string
}

// Implement Peer interface
func (p HttpPeer) IsTattlePeer() {}

// TransportHttp uses HTTP protocol to exchange messages between peers
type TransportHttp struct {
	TransportHttpParams

	router *mux.Router
	server http.Server
	inChan chan IncomingRequest
}

// Create default parameters for Http transport
func DefaulTransportHttpParams() TransportHttpParams {
	return TransportHttpParams{
		WriteTimeout:           5 * time.Minute,
		ReadTimeout:            5 * time.Minute,
		DetectorInjectTimeout:  1 * time.Second,
		DetectorProcessTimeout: 5 * time.Second,
		RpcTimeout:             5 * time.Second,
		IncomingBufferSize:     100,
		HttpClient:             http.Client{},
		Ctx:                    context.Background(),
	}
}

// Create a new Http Transport instance
func NewTransportHttp(params TransportHttpParams) *TransportHttp {
	router := mux.NewRouter()

	return &TransportHttp{
		router: router,
		server: http.Server{
			Handler:      router,
			WriteTimeout: params.WriteTimeout,
			ReadTimeout:  params.ReadTimeout,
		},
		TransportHttpParams: params,
		inChan:              make(chan IncomingRequest, params.IncomingBufferSize),
	}
}

// Run main loop
func (t *TransportHttp) Run(errch chan<- error) {
	// API endpoints
	t.setupHandlers()

	ctx, cancel := context.WithCancel(t.Ctx)
	defer cancel()

	// Shutdown http server upon global signal
	go func() {
		<-ctx.Done()
		_ = t.server.Shutdown(ctx)
	}()

	useTls := t.TLSCertFile != "" && t.TLSKeyFile != ""
	mode := ""

	if useTls {
		mode = " [TLS]"
	}

	t.Logger.Info("starting HTTP transport %s at %s",
		mode, t.Listener.Addr().String())

	var err error

	if useTls {
		err = t.server.ServeTLS(t.Listener, t.TLSCertFile, t.TLSKeyFile)
	} else {
		err = t.server.Serve(t.Listener)
	}

	if err != nil {
		errch <- errors.WithStack(err)
	}
}

func (t *TransportHttp) setupHandlers() {
	// POST /v1/ping/direct - Direct ping
	t.router.HandleFunc("/v1/ping/direct", t.pingDirectHandler).
		Methods(http.MethodPost)
}

func (t *TransportHttp) pingDirectHandler(
	w http.ResponseWriter,
	req *http.Request,
) {
	var resp Response

	//noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	preq := RequestDirectPing{}

	if err := t.Codec.DecodeRequest(req.Body, &preq); err != nil {
		t.Logger.Error("error decoding request body: %s", err)

		t.apiResponse(w, http.StatusBadRequest,
			"error decoding request body", resp)

		return
	}

	respChan := make(chan Response)

	inReq := IncomingRequest{
		Request:      preq,
		ResponseChan: respChan,
	}

	opTimer := time.NewTimer(t.DetectorInjectTimeout)

	// Try injecting the incoming req into detector
	select {
	case t.inChan <- inReq:
	case <-opTimer.C:
		NumberOfInjectTimeouts.WithLabelValues("direct_ping").Inc()

		// This can happen if detector loop is overloaded
		t.apiResponse(w, http.StatusServiceUnavailable,
			"timeout injecting a request, detector is likely overloaded", resp)
	}

	// Now wait for response
	if !opTimer.Stop() {
		<-opTimer.C
	}
	opTimer.Reset(t.DetectorProcessTimeout)

	select {
	case <-t.Ctx.Done():
		t.apiResponse(w, http.StatusServiceUnavailable,
			"context was cancelled", resp)

	case resp = <-respChan:
		t.apiResponse(w, http.StatusOK, "", resp)

	case <-opTimer.C:
		NumberOfProcessTimeouts.WithLabelValues("direct_ping").Inc()

		t.apiResponse(w, http.StatusServiceUnavailable,
			"timeout waiting for a response, detector is likely overloaded",
			resp)
	}
}

// Send a request to a remote peer
func (t *TransportHttp) Rpc(
	peer Peer,
	req Request,
	timeout time.Duration,
) (Response, error) {
	resp := Response{}

	httpPeer, ok := peer.(HttpPeer)

	if !ok {
		return resp, errors.Errorf(
			"invalid peer type, expected HttpPeer but got %T", peer)
	}

	hdr := http.Header{}
	hdr.Set("Content-Type", "application/json")

	httpReq := &http.Request{
		Close:  true,
		Header: hdr,
		Method: http.MethodPost,
	}

	rawUrl := fmt.Sprintf("%s://%s:%d",
		httpPeer.Protocol, httpPeer.Host, httpPeer.Port)

	switch req.(type) {
	case RequestDirectPing:
		rawUrl += "/v1/ping/direct"
	case RequestIndirectPing:
		rawUrl += "/v1/ping/indirect"
	default:
		return resp, errors.Errorf("unexpected request type: %T", req)
	}

	uri, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return resp, errors.Wrapf(err, "error parsing url: %s", rawUrl)
	}

	httpReq.URL = uri

	t.Logger.Debug("about to make rpc to %s", rawUrl)

	// Now prepare body
	buf := new(bytes.Buffer)

	if err := t.Codec.EncodeRequest(req, buf); err != nil {
		return resp, errors.Wrap(err, "error encoding rpc")
	}

	httpReq.Body = ioutil.NopCloser(buf)
	ctx, cancel := context.WithTimeout(t.Ctx, t.RpcTimeout)
	defer cancel()

	httpResp, err := t.HttpClient.Do(httpReq.WithContext(ctx))

	if err != nil {
		return resp, errors.Wrapf(err, "rpc error to %s", rawUrl)
	}

	//noinspection GoUnhandledErrorResult
	defer httpResp.Body.Close()

	if err = t.Codec.DecodeResponse(httpResp.Body, &resp); err != nil {
		return resp, errors.Wrap(err, "error decoding response")
	}

	return resp, nil
}

// Return a channel of incoming requests from other peers
func (t *TransportHttp) IncomingRequests() <-chan IncomingRequest {
	return t.inChan
}

func (t *TransportHttp) apiResponse(
	w http.ResponseWriter,
	code int,
	errorMsg string,
	resp Response) {

	w.WriteHeader(code)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error

	if len(errorMsg) > 0 {
		_, err = io.WriteString(w, errorMsg)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		err = t.Codec.EncodeResponse(resp, w)
	}

	if err != nil {
		t.Logger.Error("Error sending API response: %s", err)
	}
}
