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
	"fmt"
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/tnet"

	"github.com/cloudwego/netpoll-benchmark/runner"
	"github.com/cloudwego/netpoll-benchmark/tnet/codec"
)

func NewClientWithMux(network, address string, size int) runner.Client {
	cli := &muxclient{}
	cli.network = network
	cli.address = address
	cli.size = uint64(size)
	cli.conns = make([]*muxConn, size)
	for i := range cli.conns {
		cn, err := cli.DialTimeout(network, address, time.Second)
		if err != nil {
			panic(fmt.Errorf("mux dial conn failed: %s", err.Error()))
		}
		mc := newMuxConn(cn.(tnet.Conn))
		cli.conns[i] = mc
	}
	return cli
}

var _ runner.Client = &muxclient{}

type muxclient struct {
	network string
	address string
	conns   []*muxConn
	size    uint64
	cursor  uint64
}

func (cli *muxclient) DialTimeout(network, address string, timeout time.Duration) (runner.Conn, error) {
	return tnet.DialTCP(network, address, timeout)
}

func (cli *muxclient) Echo(req *runner.Message) (resp *runner.Message, err error) {
	// get conn & codec
	mc := cli.conns[atomic.AddUint64(&cli.cursor, 1)%cli.size]

	// encode
	err = codec.Encode(mc.conn, req)
	if err != nil {
		return nil, err
	}

	// decode
	resp = <-mc.rch

	// reporter
	runner.ProcessResponse(resp)
	return resp, nil
}

func newMuxConn(conn tnet.Conn) *muxConn {
	mc := &muxConn{}
	mc.conn = conn
	mc.rch = make(chan *runner.Message)
	// loop read
	conn.SetOnRequest(func(conn tnet.Conn) error {
		// decode
		resp := &runner.Message{}
		codec.Decode(conn, resp)
		mc.rch <- resp
		return nil
	})
	return mc
}

type muxConn struct {
	conn tnet.Conn
	rch  chan *runner.Message
}
