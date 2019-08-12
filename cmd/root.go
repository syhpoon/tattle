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

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/syhpoon/tattle"
)

var flagCodec string
var flagTransport string
var flagHttpListen string

var RootCmd = &cobra.Command{
	Use:   "tattle",
	Short: "tattle command",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		logger := &tattle.LoggerPrintf{}

		params := tattle.DefaultDetectorParams()
		params.Ctx = ctx
		params.Logger = logger

		var codec tattle.Codec

		// Prepare codec
		switch flagCodec {
		case "json":
			codec = tattle.NewCodecJson()
		default:
			logger.Error("Invalid codec: %s", flagCodec)

			os.Exit(1)
		}

		// Prepare transport
		switch flagTransport {
		case "http":
			httpParams := tattle.DefaulTransportHttpParams()

			listener, err := net.Listen("tcp", flagHttpListen)

			if err != nil {
				logger.Error("unable to start TCP listener: %s", err)

				os.Exit(1)
			}

			httpParams.Ctx = ctx
			httpParams.Codec = codec
			httpParams.Logger = logger
			httpParams.Listener = listener

			params.Transport = tattle.NewTransportHttp(httpParams)
		default:
			logger.Error("invalid transport: %s", flagTransport)

			os.Exit(1)
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		params.WaitGroup = wg

		detector, err := tattle.NewDetector(params)

		if err != nil {
			logger.Error("error creating detector: %+v", err)

			os.Exit(1)
		}

		go func() {
			if err := detector.Run(); err != nil {
				logger.Error("error running detector: %+v", err)

				cancel()
			}
		}()

		wait(ctx, cancel, wg)
	},
}

func wait(ctx context.Context, cancel func(), wg *sync.WaitGroup) {
	c := make(chan os.Signal, 1)

	signal.Notify(c,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGPIPE,
		syscall.SIGBUS,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGQUIT)

	select {
	case sig := <-c:
		switch sig {
		case os.Interrupt, syscall.SIGTERM, syscall.SIGABRT,
			syscall.SIGPIPE, syscall.SIGBUS, syscall.SIGUSR1, syscall.SIGUSR2,
			syscall.SIGQUIT:
			cancel()
		}

	case <-ctx.Done():
	}

	wg.Wait()
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&flagCodec,
		"codec", "c", "json", "Codec to use. Possible values: json")

	RootCmd.Flags().StringVarP(&flagTransport,
		"transport", "t", "http", "Transport to use. Possible values: http")

	RootCmd.Flags().StringVar(&flagTransport,
		"http-listen", ":9000", "Listen address for http transport")
}
