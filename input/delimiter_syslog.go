package input

import (
	"regexp"
	"strings"
)

const (
	// @see https://github.com/hpcugent/logstash-patterns/blob/master/files/grok-patterns
	MONTH           = `\b(?:Jan(?:uary|uar)?|Feb(?:ruary|ruar)?|M(?:a|ä)?r(?:ch|z)?|Apr(?:il)?|Ma(?:y|i)?|Jun(?:e|i)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|O(?:c|k)?t(?:ober)?|Nov(?:ember)?|De(?:c|z)(?:ember)?)\b`
	MONTHDAY        = `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`
	HOUR            = `(?:2[0123]|[01]?[0-9])`
	MINUTE          = `(?:[0-5][0-9])`
	SECOND          = `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`
	TIME            = HOUR + `:` + MINUTE + `(?::` + SECOND + `)`
	SYSLOGTIMESTAMP = MONTH + ` +` + MONTHDAY + ` ` + TIME
	// SYSLOG_DELIMITER indicates the start of a syslog line
	SYSLOG_DELIMITER = MONTH + `\s`
)

var syslogRegex *regexp.Regexp
var startRegex *regexp.Regexp
var runRegex *regexp.Regexp

func init() {
	syslogRegex = regexp.MustCompile(SYSLOG_DELIMITER)
	startRegex = regexp.MustCompile(SYSLOG_DELIMITER + `$`)
	runRegex = regexp.MustCompile(`\n` + SYSLOG_DELIMITER)
}

// A SyslogDelimiter detects when Syslog lines start.
type SyslogDelimiter struct {
	buffer []byte
	regex  *regexp.Regexp
}

// NewSyslogDelimiter returns an initialized SyslogDelimiter.
func NewSyslogDelimiter(maxSize int) *SyslogDelimiter {
	s := &SyslogDelimiter{}
	s.buffer = make([]byte, 0, maxSize)
	s.regex = startRegex
	return s
}

// Push a byte into the SyslogDelimiter. If the byte results in a
// a new Syslog message, it'll be flagged via the bool.
func (s *SyslogDelimiter) Push(b byte) (string, bool) {
	s.buffer = append(s.buffer, b)
	delimiter := s.regex.FindIndex(s.buffer)
	if delimiter == nil {
		return "", false
	}

	if s.regex == startRegex {
		// First match -- switch to the regex for embedded lines, and
		// drop any leading characters.
		s.buffer = s.buffer[delimiter[0]:]
		s.regex = runRegex
		return "", false
	}

	dispatch := strings.TrimRight(string(s.buffer[:delimiter[0]]), "\r")
	s.buffer = s.buffer[delimiter[0]+1:]
	return dispatch, true
}

// Vestige returns the bytes which have been pushed to SyslogDelimiter, since
// the last Syslog message was returned, but only if the buffer appears
// to be a valid syslog message.
func (s *SyslogDelimiter) Vestige() (string, bool) {
	delimiter := syslogRegex.FindIndex(s.buffer)
	if delimiter == nil {
		s.buffer = nil
		return "", false
	}
	dispatch := strings.TrimRight(string(s.buffer), "\r\n")
	s.buffer = nil
	return dispatch, true
}
