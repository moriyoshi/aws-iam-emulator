package main

import (
	"net/http"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rs/zerolog/log"
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
		log.Info().Err(err)
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
