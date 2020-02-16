package parser

import "net/url"

type Operation int

const (
	OperationGET Operation = iota
	OperationPOST
	OperationPATCH
	OperationPUT
	OperationDELETE
	OperationHEAD
)

type Option int

const (
	OptionDoNotFollowRedirect Option = iota
)

type Request struct {
	Name      string
	Operation Operation
	RawURL    string
	URL       url.URL
	Headers   map[string]string
	Body      string
	Parts     []RequestPart
	Options   []Option
	Comments  []string
}

func NewRequest(name string) *Request {
	return &Request{
		Name:     name,
		Options:  make([]Option, 0),
		Comments: make([]string, 0),
		Headers:  make(map[string]string),
	}
}

type RequestPart struct {
	Name    string
	Headers map[string]string
	Body    string
}
