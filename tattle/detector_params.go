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
	"context"
	"math/rand"
	"sync"
	"time"
)

type DetectorParams struct {
	Transport         Transport
	Peers             []Peer
	PingInterval      time.Duration
	PingTimeout       time.Duration
	IndirectPingPeers int
	Rnd               *rand.Rand
	Logger            Logger
	WaitGroup         *sync.WaitGroup
	Ctx               context.Context
}

// Return default detector parameters
func DefaultDetectorParams() DetectorParams {
	return DetectorParams{
		Transport:         nil,
		Peers:             nil,
		PingInterval:      3 * time.Second,
		PingTimeout:       1 * time.Second,
		IndirectPingPeers: 1,
		Rnd:               rand.New(rand.NewSource(time.Now().UnixNano())),
		Logger:            &LoggerPrintf{},
		Ctx:               context.Background(),
	}
}
