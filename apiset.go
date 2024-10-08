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
	"log/slog"
	"net/http"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type APISet struct {
	Version   string
	Namespace string
	handlers  map[string]Handler
}

type HandlerFunc func(*aws.Request) (*aws.Response, error)

type UnmarshalParamsFunc func(interface{}, *http.Request) error

type Handler interface {
	Name() string
	UnmarshalParams(req *http.Request) (interface{}, error)
	Handle(*aws.Request) (*aws.Response, error)
}

func (s *APISet) RegisterHandler(handler Handler) {
	s.handlers[handler.Name()] = handler
}

func (s *APISet) QueryHandler(op string) (Handler, bool) {
	h, ok := s.handlers[op]
	return h, ok
}

func NewAPISet(version, namespace string) *APISet {
	return &APISet{
		Version:   version,
		Namespace: namespace,
		handlers:  make(map[string]Handler),
	}
}

type QueryOperationHandler struct {
	Name_              string
	ParamsUnmarshaller UnmarshalParamsFunc
	Proto              interface{}
	IsEC2              bool
	Handler            HandlerFunc
}

func (h *QueryOperationHandler) Name() string {
	return h.Name_
}

func (h *QueryOperationHandler) UnmarshalParams(req *http.Request) (interface{}, error) {
	v := reflect.New(reflect.TypeOf(h.Proto))
	err := UnmarshalParams(v.Interface(), req.Form, h.IsEC2)
	if err != nil {
		logger.Info("error occurred", slog.String("error", err.Error()))
		return nil, &SenderFault{
			Code_:    "InvalidParameterValue",
			Message_: "invalid parameter",
		}
	}
	return v.Interface(), nil
}

func (h *QueryOperationHandler) Handle(req *aws.Request) (*aws.Response, error) {
	return h.Handler(req)
}
