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
