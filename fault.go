// Copyright (c) 2020 Moriyoshi Koizumi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.
package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

const awsFaultNamespaceUrl = "http://webservices.amazon.com/AWSFault/2005-15-09"

type Fault interface {
	Type() string
	Code() string
	Message() string
}

type ErrorResponsePayload struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

func renderFaultResponse(w http.ResponseWriter, requestId string, err Fault) error {
	w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	enc := xml.NewEncoder(w)
	return enc.EncodeElement(
		struct {
			ErrorResponsePayload ErrorResponsePayload `xml:"Error"`
			RequestId            string               `xml:"RequestId"`
		}{
			ErrorResponsePayload: ErrorResponsePayload{
				Type:    err.Type(),
				Code:    err.Code(),
				Message: err.Message(),
			},
			RequestId: requestId,
		},
		xml.StartElement{
			Name: xml.Name{Space: awsFaultNamespaceUrl, Local: "ErrorResponse"},
		},
	)
}

type SenderFault struct {
	Code_    string
	Message_ string
}

func (fault *SenderFault) Type() string {
	return "Sender"
}

func (fault *SenderFault) Code() string {
	return fault.Code_
}

func (fault *SenderFault) Message() string {
	return fault.Message_
}

func (fault *SenderFault) Error() string {
	return fmt.Sprintf("SenderFault: Code=%s, Message=%s", fault.Code_, fault.Message_)
}
