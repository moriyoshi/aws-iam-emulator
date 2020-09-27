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
