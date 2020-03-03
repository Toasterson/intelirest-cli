package runtime

import "encoding/json"

// MaybeJSON allows marshaling of text that my be json or not
// returns a string if it is not
type MaybeJSON []byte

// MarshallJSON Implements the Marshaling interface of JSON
func (m MaybeJSON) MarshalJSON() ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}

	if m[0] == '[' || m[0] == '{' || m[0] == '"' {
		return json.Marshal(json.RawMessage(m))
	}

	return json.Marshal(string(m))
}

// Response is the struct which client populates from the answers from the Rest calls.
type Response struct {
	HTTPVersion string
	ReturnCode  int
	Header      map[string]string
	Content     MaybeJSON
}
