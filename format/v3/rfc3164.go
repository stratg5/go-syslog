package format

import (
	"bufio"

	"github.com/stratg5/go-syslog/rfc3164/v3"
)

type RFC3164 struct{}

func (f *RFC3164) GetParser(line []byte) LogParser {
	return &parserWrapper{rfc3164.NewParser(line)}
}

func (f *RFC3164) GetSplitFunc() bufio.SplitFunc {
	// Return nil so bufio.Scanner uses its default ScanLines split function,
	// which splits on newline delimiters. This is the correct behavior for
	// TCP syslog using non-transparent framing (newline-delimited messages).
	return nil
}
