package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service struct {
	Name    string
	apisets []*APISet
}

func (e *Service) queryHandler(op string, version string) (*APISet, Handler, error) {
	for _, apiset := range e.apisets {
		if apiset.Version == version {
			handler, ok := apiset.QueryHandler(op)
			if ok {
				return apiset, handler, nil
			}
		}
	}
	if version == "" {
		version = "NO_VERSION_SPECIFIED"
	}
	return nil, nil, &SenderFault{
		Code_:    "InvalidAction",
		Message_: fmt.Sprintf("Could not find operation %s for version %s", op, version),
	}
}

func (e *Service) buildMetadata(apiset *APISet) aws.Metadata {
	return aws.Metadata{
		ServiceName: e.Name,
		APIVersion:  apiset.Version,
	}
}

func (e *Service) renderResponse(w http.ResponseWriter, req *http.Request, requestId string) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	action := req.Form.Get("Action")
	apiVersion := req.Form.Get("Version")

	log.Debug().Str("action", action).Str("apiVersion", apiVersion)

	apiset, handler, err := e.queryHandler(action, apiVersion)
	if err != nil {
		return err
	}

	params, err := handler.UnmarshalParams(req)
	if err != nil {
		return err
	}

	awsReq := &aws.Request{
		HTTPRequest: req,
		Params:      params,
		Metadata:    e.buildMetadata(apiset),
		Operation: &aws.Operation{
			Name:       action,
			HTTPMethod: req.Method,
			HTTPPath:   req.URL.Path,
		},
	}

	awsResp, err := handler.Handle(awsReq)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}
	// _, _ = b.WriteString(xml.Header)
	enc := xml.NewEncoder(b)
	err = marshal(enc, awsReq.Operation, requestId, apiset.Namespace, awsResp.Request.Data)
	if err != nil {
		return err
	}
	err = enc.Flush()
	if err != nil {
		return err
	}
	{
		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.WriteHeader(200)
		w.Write(b.Bytes())
	}
	return nil
}

func (e *Service) handleInner(w http.ResponseWriter, req *http.Request) error {
	requestId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	requestIdStr := requestId.String()
	w.Header().Set("x-amzn-RequestId", requestIdStr)

	err = e.renderResponse(w, req, requestIdStr)
	if err != nil {
		if _err, ok := err.(Fault); ok {
			err := renderFaultResponse(w, requestIdStr, _err)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (e *Service) Handle(w http.ResponseWriter, req *http.Request) {
	err := e.handleInner(w, req)
	if err != nil {
		log.Error().Err(err).Str("url", req.URL.String())
		http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), 500)
	}
}

func (e *Service) AddAPISet(apiset *APISet) {
	e.apisets = append(e.apisets, apiset)
}

func getAccountId(req *aws.Request) string {
	return "000000000000" //TODO
}
