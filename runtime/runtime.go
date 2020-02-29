package runtime

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/resty.v1"
	"intelirest-cli/parser"
)

const DefaultMaxSimultaneousConnections = 4

var QueryJoinCharacter = ", "

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
	//for key, vals := range req.URL.Query() {
	//	restReq.SetQueryParam(key, strings.Join(vals, QueryJoinCharacter))
	//}

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
	case parser.OperationPOST, parser.OperationPATCH:
		var body []byte
		if req.FileLoad != "" {
			var err error
			body, err = ioutil.ReadFile(req.FileLoad)
			if err != nil {
				return nil, err
			}
		} else {
			body = []byte(req.Body)
		}

		var resp *resty.Response
		var err error
		if req.Operation == parser.OperationPATCH {
			resp, err = restReq.SetBody(body).Patch(req.URL.String())
		} else {
			resp, err = restReq.SetBody(body).Post(req.URL.String())
		}
		if err != nil {
			return nil, err
		}
		return respFromResty(resp)
	case parser.OperationDELETE:
		resp, err := restReq.Delete(req.URL.String())
		if err != nil {
			return nil, err
		}
		return respFromResty(resp)
	case parser.OperationHEAD:
		resp, err := restReq.Head(req.URL.String())
		if err != nil {
			return nil, err
		}
		return respFromResty(resp)
	default:
		return nil, fmt.Errorf("operation %s is not supported by this runtime", req.Operation)
	}
}

func respFromResty(restyResp *resty.Response) (*Response, error) {
	resp := &Response{
		Header: make(map[string]string),
	}
	for key, value := range restyResp.Header() {
		resp.Header[key] = strings.Join(value, QueryJoinCharacter)
	}

	resp.Content = restyResp.Body()

	resp.HTTPVersion = restyResp.RawResponse.Proto

	resp.ReturnCode = restyResp.StatusCode()

	return resp, nil
}
