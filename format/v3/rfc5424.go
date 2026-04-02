package format

import (
	"bufio"

	"github.com/stratg5/go-syslog/rfc5424/v3"
)

type RFC5424 struct{}

func (f *RFC5424) GetParser(line []byte) LogParser {
	return &parserWrapper{rfc5424.NewParser(line)}
}

func (f *RFC5424) GetSplitFunc() bufio.SplitFunc {
	// Return nil so bufio.Scanner uses its default ScanLines split function,
	// which splits on newline delimiters. This is the correct behavior for
	// TCP syslog using non-transparent framing (newline-delimited messages).
	return nil
}
