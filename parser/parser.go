package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	file        *os.File
	reader      io.Reader
	environment map[string]string
}

func New(name string, env map[string]string) (*Parser, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could open file %s for parsing: %e", name, err)
	}
	return NewReader(f, env)
}

func NewReader(reader io.Reader, env map[string]string) (*Parser, error) {
	if f, ok := reader.(*os.File); ok {
		return &Parser{reader: reader, file: f, environment: env}, nil
	}
	return &Parser{reader: reader, environment: env}, nil
}

func (p *Parser) Close() error {
	if p.file == nil {
		return nil
	}
	return p.file.Close()
}

func ParseFile(name string, env map[string]string) ([]Request, error) {
	p, err := New(name, env)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	return p.Parse()
}

func (p *Parser) Parse() ([]Request, error) {
	type ParserState int
	const (
		ParserStateURL ParserState = iota
		ParserStateHeader
		ParserStateBody
	)
	state := ParserStateURL
	scanner := bufio.NewScanner(p.reader)
	var req *Request
	requests := make([]Request, 0)
	lineIter := 0
	partBoundary := ""
	partIdx := 0
	for scanner.Scan() {
		text := scanner.Text()
		// Split text by unicode.IsSpace
		tokens := strings.Fields(text)
		// Used to make error more informative
		lineIter++
		if len(tokens) == 0 {
			if state == ParserStateHeader {
				state = ParserStateBody
			}
			continue
		}
		switch {
		case tokens[0] == "###":
			// Parse start of Request with name

			// If we have something in the Request we need to finish it first
			// and initialise a new one
			if req != nil {
				if err := finishRequest(req, p.environment); err != nil {
					return nil, err
				}
				requests = append(requests, *req)
				req = nil
			}

			// Switch state to URL as we expect the URL to be next
			state = ParserStateURL
			// Initialise new Request with the part after the ### as Name of the Request
			req = NewRequest(strings.Join(tokens[1:], " "))
			continue
		case tokens[0] == "#":
			if req == nil {
				return nil, requestNotInitializedError(lineIter)
			}
			// Parse comment or add request option
			if strings.HasPrefix(tokens[1], "@") {
				// Option
				switch tokens[1] {
				case "@no-redirect":
					req.Options = append(req.Options, OptionDoNotFollowRedirect)
				}
			} else {
				// Comment
				req.Comments = append(req.Comments, strings.TrimPrefix(text, "#"))
			}
			continue
		case strings.HasPrefix(tokens[0], "#"):
			if req == nil {
				return nil, requestNotInitializedError(lineIter)
			}
			// Comment without space after # symbol
			req.Comments = append(req.Comments, strings.TrimPrefix(text, "#"))
			continue
		case strings.TrimSpace(text) == "":
			if state == ParserStateHeader {
				state = ParserStateBody
			}
		}

		if req == nil {
			return nil, requestNotInitializedError(lineIter)
		}

		switch state {
		case ParserStateURL:
			switch tokens[0] {
			case "GET":
				req.Operation = OperationGET
			case "POST":
				req.Operation = OperationPOST
			case "PATCH":
				req.Operation = OperationPATCH
			case "PUT":
				req.Operation = OperationPUT
			case "DELETE":
				req.Operation = OperationDELETE
			case "HEAD":
				req.Operation = OperationHEAD
			}
			// Replace any environment variables or macros in the URL before parsing
			req.RawURL = macroReplace(p.environment, tokens[1])

			if u, err := url.Parse(req.RawURL); err != nil {
				return nil, fmt.Errorf("error on line %d: could not parse string \"%s\" as URL: %e", lineIter, req.RawURL, err)
			} else {
				req.URL = *u
			}
			state = ParserStateHeader
			continue
		case ParserStateHeader:
			hdrName := strings.TrimSuffix(tokens[0], ":")
			hdr := strings.Join(tokens[1:], " ")
			if partIdx > 0 {
				if idx := strings.Index(hdr, "name="); idx != -1 {
					p := hdr[idx+5:]
					startIndex := strings.Index(p, "\"")
					endIndex := strings.Index(p[startIndex+1:], "\"")
					req.Parts[partIdx-1].Name = p[startIndex+1 : endIndex+1]
				}

				req.Parts[partIdx-1].Headers[hdrName] = hdr
			} else {
				if strings.Contains(hdr, "multipart/form-data;") {
					partBoundary = strings.Split(strings.TrimSpace(strings.TrimPrefix(hdr, "multipart/form-data;")), "=")[1]
					req.Parts = make([]RequestPart, 0)
				}

				req.Headers[hdrName] = hdr
			}

		case ParserStateBody:
			if strings.Contains(text, "--"+partBoundary) && !strings.HasSuffix(text, "--") {
				partIdx++
				req.Parts = append(req.Parts, RequestPart{
					Headers: make(map[string]string),
				})
				state = ParserStateHeader
				continue
			}

			if partIdx > 0 {
				if !strings.Contains(text, "--"+partBoundary) {
					req.Parts[partIdx-1].Body += text
				}
			} else {
				req.Body += text
			}
		}
	}

	// Finish any dangling Requests and don't add empty ones to the return value
	if req != nil && req.RawURL != "" {
		if err := finishRequest(req, p.environment); err != nil {
			return nil, err
		}
		requests = append(requests, *req)
		req = nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning error on line %d: %e", lineIter, err)
	}

	return requests, nil
}

func requestNotInitializedError(line int) error {
	return fmt.Errorf("error in line %d: request is not initialised did you forget the ### $NAME line at the beginning", line)
}

func finishRequest(req *Request, vars map[string]string) error {
	if contentType, ok := req.Headers["Content-Type"]; ok && contentType == "application/json" {
		req.Body = strings.ReplaceAll(strings.ReplaceAll(req.Body, "{{", "\"{{"), "}}", "}}\"")
		bodyBlob, err := processJSON([]byte(req.Body), vars)
		if err != nil {
			return err
		}
		req.Body = string(bodyBlob)
	}
	req.Body = strings.TrimSpace(req.Body)
	if strings.HasPrefix(req.Body, "< ") {
		req.FileLoad = strings.ReplaceAll(req.Body, "< ", "")
		req.Body = ""
	}
	if req.Parts != nil {
		for i, part := range req.Parts {
			if contentType, ok := req.Headers["Content-Type"]; ok && contentType == "application/json" {
				part.Body = strings.ReplaceAll(strings.ReplaceAll(part.Body, "{{", "\"{{"), "}}", "}}\"")
				bodyBlob, err := processJSON([]byte(part.Body), vars)
				if err != nil {
					return err
				}
				part.Body = string(bodyBlob)
			}
			part.Body = strings.TrimSpace(part.Body)
			if strings.HasPrefix(part.Body, "< ") {
				part.FileLoad = strings.ReplaceAll(part.Body, "< ", "")
				part.Body = ""
			}
			req.Parts[i] = part
		}
	}
	return nil
}

func processJSON(blob []byte, vars map[string]string) ([]byte, error) {
	tmpJSON := make(map[string]json.RawMessage)
	if err := json.Unmarshal(blob, &tmpJSON); err != nil {
		return nil, err
	}

	for key, value := range tmpJSON {
		str := string(value)
		if strings.Contains(str, "{{") && strings.Contains(str, "}}") {
			str = macroReplace(vars, strings.ReplaceAll(str, "\"", ""))
			if _, err := strconv.Atoi(str); err == nil {
				tmpJSON[key] = json.RawMessage(str)
				continue
			}

			if str == "null" {
				tmpJSON[key] = json.RawMessage(str)
				continue
			}

			tmpJSON[key] = json.RawMessage("\"" + str + "\"")
			continue
		}

		if strings.HasPrefix(str, "{") {
			v, err := processJSON(value, vars)
			if err != nil {
				return nil, err
			}
			tmpJSON[key] = v
			continue
		}
	}
	return json.Marshal(tmpJSON)
}

func macroReplace(vars map[string]string, text string) string {
	s := ""
	tokens := ParseMacrosFromLine(text)
	for _, tok := range tokens {
		result := tok.Token
		if tok.IsMacro {
			for key, value := range vars {
				if key == tok.Token {
					result = value
				}
			}
		}
		s += result
	}

	return s
}

type token struct {
	IsMacro bool
	Token   string
}

func ParseMacrosFromLine(s string) []token {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		isMacro bool
		start   int
		end     int
	}
	spans := make([]span, 0, 32)

	// Find the field start and end indices.
	startElementCounter := 0
	endElementCounter := 0
	macroStartIndex := 0
	macroEndIndex := 0
	for i, r := range s {
		if r == '{' {
			startElementCounter++
			if startElementCounter == 2 {
				macroStartIndex = i + 1
				if i-1 != 0 {
					if macroEndIndex == 0 {
						spans = append(spans, span{start: macroEndIndex, end: macroStartIndex - 2, isMacro: false})
					} else {
						spans = append(spans, span{start: macroEndIndex + 2, end: macroStartIndex - 2, isMacro: false})
					}
				}
			}
		} else if r == '}' {
			endElementCounter++
			if endElementCounter == 2 {
				macroEndIndex = i - 1
				spans = append(spans, span{start: macroStartIndex, end: macroEndIndex, isMacro: true})
				endElementCounter = 0
				startElementCounter = 0
			}
		}
	}

	if macroStartIndex == 0 {
		spans = append(spans, span{
			isMacro: false,
			start:   0,
			end:     len(s),
		})
	} else {
		spans = append(spans, span{
			isMacro: false,
			start:   macroEndIndex + 2,
			end:     len(s),
		})
	}

	// Create strings from recorded field indices.
	a := make([]token, len(spans))
	for i, span := range spans {
		a[i] = token{Token: s[span.start:span.end], IsMacro: span.isMacro}
	}

	return a
}
