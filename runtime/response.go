package runtime

import (
	"github.com/dustin/go-humanize"
)

type Response struct {
	HTTPVersion string
	ReturnCode  int
	Header      map[string]string
	Content     []byte
}
