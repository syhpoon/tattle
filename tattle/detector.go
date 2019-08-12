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
	"time"

	"github.com/pkg/errors"
)

var (
	// Returned when no initial peers were provided in params
	ErrNoPeers = errors.New("no peers provided")
)

type Detector struct {
	DetectorParams
}

// Create a new  Detector instance
func NewDetector(params DetectorParams) (*Detector, error) {
	if len(params.Peers) == 0 {
		return nil, errors.WithStack(ErrNoPeers)
	}

	return &Detector{
		DetectorParams: params,
	}, nil
}

// Run main detector loop
func (d *Detector) Run() error {
	if d.WaitGroup != nil {
		defer d.WaitGroup.Done()
	}

	timer := time.NewTimer(d.PingInterval)
	idx := 0

	// Process incoming requests
	go d.processIncoming()

	for {
		select {
		case <-d.Ctx.Done():
			return errors.WithStack(context.Canceled)

		case <-timer.C:
			if idx == 0 {
				d.shufflePeers()
			}

			// Need to ping one of the peers
			peer := d.Peers[idx]

			go d.pingPeer(peer)

			idx = (idx + 1) % len(d.Peers)
		}
	}
}

func (d *Detector) processIncoming() {
	for {
		select {
		case <-d.Ctx.Done():
			return

		case inReq := <-d.Transport.IncomingRequests():
			d.Logger.Info("Incoming request: %v", inReq)

			// TODO
		}
	}
}

// Send a direct ping request to the peer
func (d *Detector) pingPeer(peer Peer) {
	// TODO: Actual updates
	updates := []UpdateEvent{
		{
			Peer:       peer,
			UpdateType: UpdateTypePeerAlive,
			SeqNum:     1,
		},
	}

	req := &RequestDirectPing{
		Updates: updates,
	}

	resp, err := d.Transport.Rpc(peer, req, d.PingTimeout)

	if err != nil {
		d.markDead(peer)
	}

	// TODO:
	_ = resp
}

// Shuffle the peers
func (d *Detector) shufflePeers() {
	d.Rnd.Shuffle(len(d.Peers), func(i, j int) {
		d.Peers[i], d.Peers[j] = d.Peers[j], d.Peers[i]
	})
}

func (d *Detector) markDead(peer Peer) {
	d.Logger.Info("TODO: Marking peer %v dead\n", peer)
}
