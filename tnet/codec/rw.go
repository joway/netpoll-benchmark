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

package codec

import (
	"encoding/binary"
	"reflect"
	"unsafe"

	"trpc.group/trpc-go/tnet"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

// Encode include flush
func Encode(conn tnet.Conn, msg *runner.Message) (err error) {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(msg.Message)))
	//conn.SetSafeWrite(true)
	_, err = conn.Writev(header, unsafeStringToSlice(msg.Message))
	return err
}

func Decode(conn tnet.Conn, msg *runner.Message) (err error) {
	bLen, err := conn.Next(4)
	if err != nil {
		return err
	}
	l := int(binary.BigEndian.Uint32(bLen))
	// We cannot use Zero-Copy API like .Next, because it used to decode binary into server request,
	// and we don't know when user will not use request, if we reuse memory, it will cause panic
	buf, err := conn.ReadN(l)
	if err != nil {
		return err
	}
	msg.Message = unsafeSliceToString(buf)
	conn.Release()
	return nil
}

// zero-copy slice convert to string
func unsafeSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// zero-copy slice convert to string
func unsafeStringToSlice(s string) (b []byte) {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Data = uintptr(p)
	hdr.Cap = len(s)
	hdr.Len = len(s)
	return b
}
