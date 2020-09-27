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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/private/protocol/xml/xmlutil"
)

func marshal(e *xml.Encoder, op *aws.Operation, requestId, ns string, result interface{}) error {
	var err error
	outerEnvelope := xml.StartElement{Name: xml.Name{Space: ns, Local: op.Name + "Response"}}
	err = e.EncodeToken(outerEnvelope)
	if err != nil {
		return err
	}
	innerEnvelope := xml.StartElement{Name: xml.Name{Local: op.Name + "Result"}}
	err = e.EncodeToken(innerEnvelope)
	if err != nil {
		return err
	}
	err = xmlutil.BuildXML(result, e)
	if err != nil {
		return err
	}
	err = e.EncodeToken(innerEnvelope.End())
	if err != nil {
		return err
	}
	{
		requestMetadataElem := xml.StartElement{Name: xml.Name{Local: "RequestMetadata"}}
		err = e.EncodeToken(requestMetadataElem)
		if err != nil {
			return err
		}
		err = e.EncodeElement(requestId, xml.StartElement{Name: xml.Name{Local: "RequestId"}})
		if err != nil {
			return err
		}
		err = e.EncodeToken(requestMetadataElem.End())
		if err != nil {
			return err
		}
	}
	return e.EncodeToken(outerEnvelope.End())
}
