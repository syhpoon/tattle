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

import "time"

type UpdateType int

const (
	UpdateTypePeerAlive      UpdateType = 1
	UpdateTypePeerSuspicious UpdateType = 2
	UpdateTypePeerDead       UpdateType = 3
)

type UpdateEvent struct {
	Peer       Peer
	UpdateType UpdateType
	SeqNum     uint64
}

type Request interface {
	IsTattleTransportRequest()
}

type RequestDirectPing struct {
	Updates []UpdateEvent
}

func (r RequestDirectPing) IsTattleTransportRequest() {}

type RequestIndirectPing struct {
	Updates    []UpdateEvent
	TargetPeer Peer
}

func (r RequestIndirectPing) IsTattleTransportRequest() {}

type Response struct {
	Updates []UpdateEvent
}

type IncomingRequest struct {
	Request      Request
	ResponseChan chan<- Response
}

// Transport is a general abstraction responsible for sending messages
// between peers
type Transport interface {
	Rpc(peer Peer, req Request, timeout time.Duration) (Response, error)
	IncomingRequests() <-chan IncomingRequest
}