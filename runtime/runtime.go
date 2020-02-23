package runtime

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/resty.v1"
	"intelirest-cli/parser"
)

const DefaultMaxSimultaneousConnections = 4

type Client struct {
	client  *resty.Client
	maxconn int
}

func New(maxSimulataneousConnections int) *Client {
	if maxSimulataneousConnections == 0 {
		maxSimulataneousConnections = DefaultMaxSimultaneousConnections
	}

	return &Client{
		client:  resty.New(),
		maxconn: maxSimulataneousConnections,
	}
}

func (c *Client) Do(requests []parser.Request) ([]Response, error) {
	var rErr *multierror.Error
	responses := make([]Response, 0)
	for i, request := range requests {
		resp, err := c.ExecuteRequest(request)
		if err != nil {
			rErr = multierror.Append(fmt.Errorf("error executing request %d: %e", i, err))
		}
		responses = append(responses, *resp)
	}

	return responses, rErr
}

func (c *Client) ExecuteRequest(req parser.Request) (*Response, error) {
	restReq := c.client.R()
	restReq.SetHeaders(req.Headers)
	if req.IsMultiPart() {
		//for _, part := range req.Parts {
		//	restReq.SetMultipartField(part.Name)
		//}
	}
	switch req.Operation {
	case parser.OperationGET:
		resp, err := restReq.Get(req.URL.String())
		if err != nil {
			return nil, err
		}
		return respFromResty(resp)
	}
}

func respFromResty(restyResp *resty.Response) (*Response, error) {
	hdrBuffer := bytes.NewBuffer(nil)
	if err := restyResp.Header().Write(hdrBuffer); err != nil {
		return nil, fmt.Errorf("could not write header to internal buffer: %e", err)
	}

}
