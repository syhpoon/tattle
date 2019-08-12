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
	"encoding/json"
	"io"
)

// Json codec
type CodecJson struct{}

func NewCodecJson() *CodecJson {
	return &CodecJson{}
}

// Encode a request
func (c *CodecJson) EncodeRequest(req Request, w io.Writer) error {
	return c.enc(w, req)
}

// Encode a response
func (c *CodecJson) EncodeResponse(resp Response, w io.Writer) error {
	return c.enc(w, resp)
}

// Decode a request
func (c *CodecJson) DecodeRequest(reader io.Reader, req Request) error {
	return c.dec(reader, req)
}

// Decode a response
func (c *CodecJson) DecodeResponse(reader io.Reader, resp *Response) error {
	return c.dec(reader, resp)
}

func (c *CodecJson) enc(w io.Writer, obj interface{}) error {
	enc := json.NewEncoder(w)

	return enc.Encode(obj)
}

func (c *CodecJson) dec(reader io.Reader, dst interface{}) error {
	dec := json.NewDecoder(reader)

	return dec.Decode(dst)
}
