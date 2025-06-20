package rfc3164

import (
	"bytes"
	"testing"
	"time"

	syslogparser "github.com/stratg5/go-syslog/parser/v3"
	. "gopkg.in/check.v1"
)

// Hooks up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type Rfc3164TestSuite struct {
}

var (
	_ = Suite(&Rfc3164TestSuite{})

	// XXX : corresponds to the length of the last tried timestamp format
	// XXX : Jan  2 15:04:05
	lastTriedTimestampLen = 15
)

func (s *Rfc3164TestSuite) TestParser_Valid(c *C) {
	buff := []byte("<34>Oct 11 22:14:15 mymachine very.large.syslog.message.tag: 'su root' failed for lonvick on /dev/pts/8")

	p := NewParser(buff)
	expectedP := &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.Parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.Dump()
	expected := syslogparser.LogParts{
		"timestamp": time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		"hostname":  "mymachine",
		"tag":       "very.large.syslog.message.tag",
		"content":   "'su root' failed for lonvick on /dev/pts/8",
		"priority":  34,
		"facility":  4,
		"severity":  2,
	}

	c.Assert(obtained, DeepEquals, expected)
}

func (s *Rfc3164TestSuite) TestParser_ValidNoTag(c *C) {
	buff := []byte("<34>Oct 11 22:14:15 mymachine singleword")

	p := NewParser(buff)
	expectedP := &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.Parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.Dump()
	expected := syslogparser.LogParts{
		"timestamp": time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		"hostname":  "mymachine",
		"tag":       "",
		"content":   "singleword",
		"priority":  34,
		"facility":  4,
		"severity":  2,
	}

	c.Assert(obtained, DeepEquals, expected)
}

// RFC 3164 section 4.3.2
func (s *Rfc3164TestSuite) TestParser_NoTimestamp(c *C) {
	buff := []byte("<14>INFO     leaving (1) step postscripts")

	p := NewParser(buff)
	expectedP := &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.Parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.Dump()

	obtainedTime := obtained["timestamp"].(time.Time)
	s.assertTimeIsCloseToNow(c, obtainedTime)

	obtained["timestamp"] = now // XXX: Need to mock out time to test this fully
	expected := syslogparser.LogParts{
		"timestamp": now,
		"hostname":  "",
		"tag":       "",
		"content":   "INFO     leaving (1) step postscripts",
		"priority":  14,
		"facility":  1,
		"severity":  6,
	}

	c.Assert(obtained, DeepEquals, expected)
}

// RFC 3164 section 4.3.3
func (s *Rfc3164TestSuite) TestParser_NoPriority(c *C) {
	buff := []byte("Oct 11 22:14:15 Testing no priority")

	p := NewParser(buff)
	expectedP := &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.Parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.Dump()
	obtainedTime := obtained["timestamp"].(time.Time)
	s.assertTimeIsCloseToNow(c, obtainedTime)

	obtained["timestamp"] = now // XXX: Need to mock out time to test this fully
	expected := syslogparser.LogParts{
		"timestamp": now,
		"hostname":  "",
		"tag":       "",
		"content":   "Oct 11 22:14:15 Testing no priority",
		"priority":  13,
		"facility":  1,
		"severity":  5,
	}

	c.Assert(obtained, DeepEquals, expected)
}

func (s *Rfc3164TestSuite) TestParseHeader_Valid(c *C) {
	buff := []byte("Oct 11 22:14:15 mymachine ")
	now := time.Now()
	hdr := header{
		timestamp: time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		hostname:  "mymachine",
	}

	s.assertRfc3164Header(c, hdr, buff, 25, nil)

	// expected header for next two tests
	hdr = header{
		timestamp: time.Date(now.Year(), time.October, 1, 22, 14, 15, 0, time.UTC),
		hostname:  "mymachine",
	}
	// day with leading zero
	buff = []byte("Oct 01 22:14:15 mymachine ")
	s.assertRfc3164Header(c, hdr, buff, 25, nil)
	// day with leading space
	buff = []byte("Oct  1 22:14:15 mymachine ")
	s.assertRfc3164Header(c, hdr, buff, 25, nil)

}

func (s *Rfc3164TestSuite) TestParseHeader_RFC3339Timestamp(c *C) {
	buff := []byte("2018-01-12T22:14:15+00:00 mymachine app[101]: msg")
	hdr := header{
		timestamp: time.Date(2018, time.January, 12, 22, 14, 15, 0, time.UTC),
		hostname:  "mymachine",
	}
	s.assertRfc3164Header(c, hdr, buff, 35, nil)
}

func (s *Rfc3164TestSuite) TestParser_ValidRFC3339Timestamp(c *C) {
	buff := []byte("<34>2018-01-12T22:14:15+00:00 mymachine app[101]: msg")
	p := NewParser(buff)
	err := p.Parse()
	c.Assert(err, IsNil)
	obtained := p.Dump()
	expected := syslogparser.LogParts{
		"timestamp": time.Date(2018, time.January, 12, 22, 14, 15, 0, time.UTC),
		"hostname":  "mymachine",
		"tag":       "app",
		"content":   "msg",
		"priority":  34,
		"facility":  4,
		"severity":  2,
	}
	c.Assert(obtained, DeepEquals, expected)
}

func (s *Rfc3164TestSuite) TestParseHeader_InvalidTimestamp(c *C) {
	buff := []byte("Oct 34 32:72:82 mymachine ")
	hdr := header{}

	s.assertRfc3164Header(c, hdr, buff, lastTriedTimestampLen+1, syslogparser.ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParsemessage_Valid(c *C) {
	content := "foo bar baz blah quux"
	buff := []byte("sometag[123]: " + content)
	hdr := rfc3164message{
		tag:     "sometag",
		content: content,
	}

	s.assertRfc3164message(c, hdr, buff, len(buff), syslogparser.ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Invalid(c *C) {
	buff := []byte("Oct 34 32:72:82")
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, lastTriedTimestampLen, syslogparser.ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_TrailingSpace(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15 ")

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_OneDigitForMonths(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct  1 22:14:15")

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 1, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15")

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_Pid(c *C) {
	buff := []byte("apache2[10]:")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_NoPid(c *C) {
	buff := []byte("apache2:")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TrailingSpace(c *C) {
	buff := []byte("apache2: ")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_NoTag(c *C) {
	buff := []byte("apache2")
	tag := ""

	s.assertTag(c, tag, buff, 0, nil)
}

func (s *Rfc3164TestSuite) TestParseContent_Valid(c *C) {
	buff := []byte(" foo bar baz quux ")
	content := string(bytes.Trim(buff, " "))

	p := NewParser(buff)
	obtained, err := p.parseContent()
	c.Assert(err, Equals, syslogparser.ErrEOL)
	c.Assert(obtained, Equals, content)
	c.Assert(p.cursor, Equals, len(content))
}

func (s *Rfc3164TestSuite) BenchmarkParseTimestamp(c *C) {
	buff := []byte("Oct 11 22:14:15")

	p := NewParser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseTimestamp()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseHostname(c *C) {
	buff := []byte("gimli.local")

	p := NewParser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseHostname()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseTag(c *C) {
	buff := []byte("apache2[10]:")

	p := NewParser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseTag()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseHeader(c *C) {
	buff := []byte("Oct 11 22:14:15 mymachine ")

	p := NewParser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseHeader()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParsemessage(c *C) {
	buff := []byte("sometag[123]: foo bar baz blah quux")

	p := NewParser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parsemessage()
		if err != syslogparser.ErrEOL {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) assertTimestamp(c *C, ts time.Time, b []byte, expC int, e error) {
	p := NewParser(b)
	obtained, err := p.parseTimestamp()
	c.Assert(obtained, Equals, ts)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertTag(c *C, t string, b []byte, expC int, e error) {
	p := NewParser(b)
	obtained, err := p.parseTag()
	c.Assert(obtained, Equals, t)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertRfc3164Header(c *C, hdr header, b []byte, expC int, e error) {
	p := NewParser(b)
	obtained, err := p.parseHeader()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, hdr)
	c.Assert(p.cursor, Equals, expC)
}

func (s *Rfc3164TestSuite) assertRfc3164message(c *C, msg rfc3164message, b []byte, expC int, e error) {
	p := NewParser(b)
	obtained, err := p.parsemessage()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, msg)
	c.Assert(p.cursor, Equals, expC)
}

func (s *Rfc3164TestSuite) assertTimeIsCloseToNow(c *C, obtainedTime time.Time) {
	now := time.Now()
	timeStart := now.Add(-(time.Second * 5))
	timeEnd := now.Add(time.Second)
	c.Assert(obtainedTime.After(timeStart), Equals, true)
	c.Assert(obtainedTime.Before(timeEnd), Equals, true)
}
