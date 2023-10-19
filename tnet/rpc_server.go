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

	"trpc.group/trpc-go/tnet"

	"github.com/cloudwego/netpoll-benchmark/runner"
	"github.com/cloudwego/netpoll-benchmark/tnet/codec"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct{}

func (s *rpcServer) Run(network, address string) error {
	// new listener
	listener, err := tnet.Listen(network, address)
	if err != nil {
		panic(err)
	}

	// new server
	svr, err := tnet.NewTCPService(
		listener,
		s.handler,
		// In fact, the blocking mode of tnet have non-blocking mode... Don't know why
		// By default nonblocking is false
		//tnet.WithNonBlocking(false),
	)
	if err != nil {
		panic(err)
	}
	// start listen loop ...
	return svr.Serve(context.Background())
}

func (s *rpcServer) handler(conn tnet.Conn) (err error) {
	// decode
	req := &runner.Message{}
	err = codec.Decode(conn, req)
	if err != nil {
		return err
	}

	// handler
	resp := runner.ProcessRequest(reporter, req)

	// encode
	err = codec.Encode(conn, resp)
	if err != nil {
		return err
	}
	return nil
}
