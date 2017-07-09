package input

import (
	"regexp"
	"strconv"
)

// YYFmt represents a parser for YYFmt-compliant log messages
type YYFmt struct {
	matcher *regexp.Regexp
}

func (p *Parser) newYYFmtParser() {
	p.yyfmt = &YYFmt{}
	p.yyfmt.compileMatcher()
}

func (s *YYFmt) compileMatcher() {
	leading := `(?s)`
	timestamp := `(` + SYSLOGTIMESTAMP + `)`
	pri := `(\w+)`
	app := `([^ \[]+)`
	pid := `\[(-|[0-9]{1,5})\]:`
	msg := `(.+$)`
	s.matcher = regexp.MustCompile(leading + timestamp + `\s` + pri + `\s` + app + pid + `\s` + msg)
}

func (s *YYFmt) parse(raw []byte, result *map[string]interface{}) {
	m := s.matcher.FindStringSubmatch(string(raw))
	if m == nil || len(m) != 6 {
		stats.Add("yyfmtUnparsed", 1)
		return
	}
	stats.Add("yyfmtParsed", 1)
	pri := m[2]
	var pid int
	if m[4] != "-" {
		pid, _ = strconv.Atoi(m[4])
	}
	*result = map[string]interface{}{
		"priority":  pri,
		"timestamp": m[1],
		"app":       m[3],
		"pid":       pid,
		"message":   m[5],
	}
}
