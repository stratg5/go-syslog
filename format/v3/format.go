package format

import (
	"bufio"
	"time"

	syslogparser "github.com/stratg5/go-syslog/parser/v3"
)

type LogParts map[string]interface{}

type LogParser interface {
	Parse() error
	Dump() LogParts
	Location(*time.Location)
}

type Format interface {
	GetParser([]byte) LogParser
	GetSplitFunc() bufio.SplitFunc
}

type parserWrapper struct {
	syslogparser.LogParser
}

func (w *parserWrapper) Dump() LogParts {
	return LogParts(w.LogParser.Dump())
}
