package parser

import (
	"bytes"
	"net/url"
	"testing"

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
	tc := []struct {
		input  string
		output []Request
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
	}

	for _, c := range tc {
		buff := bytes.NewBufferString(c.input)
		p, err := NewReader(buff, nil)
		assert.NoError(t, err)
		requests, err := p.Parse()
		assert.NoError(t, err)
		assert.Equal(t, c.output, requests)
	}
}
