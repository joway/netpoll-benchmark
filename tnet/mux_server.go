/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"

	"trpc.group/trpc-go/tnet"

	"github.com/cloudwego/netpoll-benchmark/runner"
	"github.com/cloudwego/netpoll-benchmark/tnet/codec"
)

func NewMuxServer() runner.Server {
	return &muxServer{}
}

var _ runner.Server = &muxServer{}

type muxServer struct{}

func (s *muxServer) Run(network, address string) error {
	// new listener
	listener, err := tnet.Listen(network, address)
	if err != nil {
		panic(err)
	}

	// new server
	svr, err := tnet.NewTCPService(
		listener,
		s.handler,
		tnet.WithNonBlocking(true),
	)
	if err != nil {
		panic(err)
	}
	// start listen loop ...
	return svr.Serve(context.Background())
}

func (s *muxServer) handler(conn tnet.Conn) (err error) {
	var respCh = make(chan *runner.Message)
	go func() {
		for resp := range respCh {
			err = codec.Encode(conn, resp)
			if err != nil {
				panic(fmt.Errorf("tnet encode failed: %s", err.Error()))
			}
		}
	}()
	for {
		req := &runner.Message{}
		err = codec.Decode(conn, req)
		if err != nil {
			return err
		}

		// handler must use another goroutine
		go func() {
			resp := runner.ProcessRequest(reporter, req)
			respCh <- resp
		}()
	}
}
