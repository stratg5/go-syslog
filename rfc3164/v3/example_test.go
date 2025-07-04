package rfc3164_test

import (
	"fmt"

	"github.com/stratg5/go-syslog/rfc3164/v3"
)

func ExampleNewParser() {
	b := "<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8"
	buff := []byte(b)

	p := rfc3164.NewParser(buff)
	err := p.Parse()
	if err != nil {
		panic(err)
	}

	fmt.Println(p.Dump())
}
