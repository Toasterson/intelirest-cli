package parser

import (
	"bytes"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseMacrosFromLine(t *testing.T) {
	tc := []struct {
		input  string
		tokens []token
	}{
		{
			input:  "https://httpbin.org/ip",
			tokens: []token{{Token: "https://httpbin.org/ip", IsMacro: false}},
		},
		{
			input: "{{host}}/get?show_env={{show_env}}",
			tokens: []token{
				{Token: "host", IsMacro: true},
				{Token: "/get?show_env=", IsMacro: false},
				{Token: "show_env", IsMacro: true},
			},
		},
		{
			input: "http://httpbin.org/anything?id={{$uuid}}&ts={{$timestamp}}",
			tokens: []token{
				{Token: "http://httpbin.org/anything?id=", IsMacro: false},
				{Token: "$uuid", IsMacro: true},
				{Token: "&ts=", IsMacro: false},
				{Token: "$timestamp", IsMacro: true},
			},
		},
	}
	for _, c := range tc {
		tok := ParseMacrosFromLine(c.input)
		assert.Equal(t, c.tokens, tok)
	}
}

func TestParseFile(t *testing.T) {
	id := uuid.NewV4()
	timestamp := time.Now().Format(time.RFC3339)
	tc := []struct {
		input     string
		variables map[string]string
		output    []Request
	}{
		{
			input: `### GET request with a header
GET https://httpbin.org/ip
Accept: application/json
`,
			output: []Request{
				{
					Name:      "GET request with a header",
					Operation: OperationGET,
					RawURL:    "https://httpbin.org/ip",
					URL: url.URL{
						Scheme: "https",
						Host:   "httpbin.org",
						Path:   "/ip",
					},
					Headers: map[string]string{
						"Accept": "application/json",
					},
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### GET request with parameter
GET https://httpbin.org/get?show_env=1
Accept: application/json
`,
			output: []Request{
				{
					Name:      "GET request with parameter",
					Operation: OperationGET,
					RawURL:    "https://httpbin.org/get?show_env=1",
					URL: url.URL{
						Scheme:   "https",
						Host:     "httpbin.org",
						Path:     "/get",
						RawQuery: "show_env=1",
					},
					Headers: map[string]string{
						"Accept": "application/json",
					},
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### GET request with environment variables
GET {{host}}/get?show_env={{show_env}}
Accept: application/json
`,
			variables: map[string]string{
				"host":     "http://httpbin.org",
				"show_env": "1",
			},
			output: []Request{
				{
					Name:      "GET request with environment variables",
					Operation: OperationGET,
					RawURL:    "http://httpbin.org/get?show_env=1",
					URL: url.URL{
						Scheme:   "http",
						Host:     "httpbin.org",
						Path:     "/get",
						RawQuery: "show_env=1",
					},
					Headers: map[string]string{
						"Accept": "application/json",
					},
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### GET request with disabled redirects
# @no-redirect
GET http://httpbin.org/status/301
`,
			output: []Request{
				{
					Name:      "GET request with disabled redirects",
					Operation: OperationGET,
					RawURL:    "http://httpbin.org/status/301",
					URL: url.URL{
						Scheme: "http",
						Host:   "httpbin.org",
						Path:   "/status/301",
					},
					Headers: make(map[string]string),
					Options: []Option{
						OptionDoNotFollowRedirect,
					},
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### GET request with dynamic variables
GET http://httpbin.org/anything?id={{$uuid}}&ts={{$timestamp}}

###
`,
			variables: map[string]string{
				"$uuid":      id.String(),
				"$timestamp": timestamp,
			},
			output: []Request{
				{
					Name:      "GET request with dynamic variables",
					Operation: OperationGET,
					RawURL:    "http://httpbin.org/anything?id=" + id.String() + "&ts=" + timestamp,
					URL: url.URL{
						Scheme:   "http",
						Host:     "httpbin.org",
						Path:     "/anything",
						RawQuery: "id=" + id.String() + "&ts=" + timestamp,
					},
					Headers:  make(map[string]string),
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
	}

	for _, c := range tc {
		buff := bytes.NewBufferString(c.input)
		p, err := NewReader(buff, c.variables)
		assert.NoError(t, err)
		requests, err := p.Parse()
		assert.NoError(t, err)
		assert.Equal(t, c.output, requests)
	}
}

func TestPost(t *testing.T) {
	id := uuid.NewV4()
	timestamp := time.Now().Format(time.RFC3339)
	itoa := strconv.Itoa(30)
	tc := []struct {
		input     string
		variables map[string]string
		output    []Request
	}{
		{
			input: `### Send POST request with json body
POST https://httpbin.org/post
Content-Type: application/json

{
  "id": 999,
  "value": "content"
}
`,
			output: []Request{
				{
					Name:      "Send POST request with json body",
					Operation: OperationPOST,
					RawURL:    "https://httpbin.org/post",
					URL: url.URL{
						Scheme: "https",
						Host:   "httpbin.org",
						Path:   "/post",
					},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body:     `{"id":999,"value":"content"}`,
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### Send POST request with body as parameters
POST https://httpbin.org/post
Content-Type: application/x-www-form-urlencoded

id=999&value=content
`,
			output: []Request{
				{
					Name:      "Send POST request with body as parameters",
					Operation: OperationPOST,
					RawURL:    "https://httpbin.org/post",
					URL: url.URL{
						Scheme: "https",
						Host:   "httpbin.org",
						Path:   "/post",
					},
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body:     `id=999&value=content`,
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### Send a form with the text and file fields
POST https://httpbin.org/post
Content-Type: multipart/form-data; boundary=WebAppBoundary

--WebAppBoundary
Content-Disposition: form-data; name="element-name"
Content-Type: text/plain

Name
--WebAppBoundary
Content-Disposition: form-data; name="data"; filename="data.json"
Content-Type: application/json

< ./request-form-data.json
--WebAppBoundary--
`,
			output: []Request{
				{
					Name:      "Send a form with the text and file fields",
					Operation: OperationPOST,
					RawURL:    "https://httpbin.org/post",
					URL: url.URL{
						Scheme: "https",
						Host:   "httpbin.org",
						Path:   "/post",
					},
					Headers: map[string]string{
						"Content-Type": "multipart/form-data; boundary=WebAppBoundary",
					},
					Parts: []RequestPart{
						{
							Name: "element-name",
							Headers: map[string]string{
								"Content-Disposition": "form-data; name=\"element-name\"",
								"Content-Type":        "text/plain",
							},
							Body: `Name`,
						},
						{
							Name: "data",
							Headers: map[string]string{
								"Content-Disposition": "form-data; name=\"data\"; filename=\"data.json\"",
								"Content-Type":        "application/json",
							},
							FileLoad: "./request-form-data.json",
						},
					},
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
		{
			input: `### Send request with dynamic variables in request's body
POST https://httpbin.org/post
Content-Type: application/json

{
  "id": {{$uuid}},
  "price": {{$randomInt}},
  "ts": {{$timestamp}},
  "value": "content"
}

###`,
			variables: map[string]string{
				"$uuid":      id.String(),
				"$randomInt": itoa,
				"$timestamp": timestamp,
			},
			output: []Request{
				{
					Name:      "Send request with dynamic variables in request's body",
					Operation: OperationPOST,
					RawURL:    "https://httpbin.org/post",
					URL: url.URL{
						Scheme: "https",
						Host:   "httpbin.org",
						Path:   "/post",
					},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body:     `{"id":"` + id.String() + `","price":` + itoa + `,"ts":"` + timestamp + `","value":"content"}`,
					Options:  make([]Option, 0),
					Comments: make([]string, 0),
				},
			},
		},
	}

	for i, c := range tc {
		buff := bytes.NewBufferString(c.input)
		p, err := NewReader(buff, c.variables)
		assert.NoErrorf(t, err, "Test %d failed", i)
		requests, err := p.Parse()
		assert.NoErrorf(t, err, "Test %d failed", i)
		assert.Equal(t, c.output, requests, "Test %d failed", i)
	}
}
