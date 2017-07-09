package input

import (
	"testing"
)

/*
 * SyslogDelimiter tests.
 */

func Test_SyslogDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected []string
	}{
		{
			name:     "simple",
			line:     "Jan sshd is down\nFeb sshd is up\nMar password accepted",
			expected: []string{"Jan sshd is down", "Feb sshd is up"},
		},
		{
			name:     "leading",
			line:     "password accepted for user rootApr sshd is down\nDec sshd is up\nMar password accepted",
			expected: []string{"Dec sshd is up"},
		},
		{
			name:     "CRLF",
			line:     "Apr sshd is down\r\nDec sshd is up\r\nMar password accepted",
			expected: []string{"Apr sshd is down", "Dec sshd is up"},
		},
		{
			name:     "stacktrace",
			line:     "Apr sshd is down\nDec OOM on line 42, dummy.java\n\tclass_loader.jar\nMar password accepted",
			expected: []string{"Apr sshd is down", "Dec OOM on line 42, dummy.java\n\tclass_loader.jar"},
		},
		{
			name:     "embedded",
			line:     "Apr sshd is <down>\nDec sshd is upNov\nMar password accepted",
			expected: []string{"Apr sshd is <down>", "Dec sshd is upNov"},
		},
	}

	for _, tt := range tests {
		d := NewSyslogDelimiter(256)
		events := []string{}

		for _, b := range tt.line {
			event, match := d.Push(byte(b))
			if match {
				events = append(events, event)
			}
		}

		if len(events) != len(tt.expected) {
			t.Errorf("test %s: failed to delimit '%s' as expected", tt.name, tt.line)
		} else {
			for i := 0; i < len(events); i++ {
				if events[i] != tt.expected[i] {
					t.Errorf("test %s: failed to delimit '%s', got %s, expected %s", tt.name, tt.line, events[i], tt.expected)
				}
			}
		}
	}
}

func TestSyslogDelimiter_Vestige(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedEvent string
		expectedMatch bool
	}{
		{
			name:          "vestige zero",
			line:          "",
			expectedEvent: "",
			expectedMatch: false,
		},
		{
			name:          "vestige no match",
			line:          "Ja\n",
			expectedEvent: "",
			expectedMatch: false,
		},
		{
			name:          "vestige match",
			line:          "Oct ",
			expectedEvent: "Oct ",
			expectedMatch: true,
		},
		{
			name:          "vestige rich match",
			line:          "Dec OOM on line 42, dummy.java\n\tclass_loader.jar",
			expectedEvent: "Dec OOM on line 42, dummy.java\n\tclass_loader.jar",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		d := NewSyslogDelimiter(256)
		for _, c := range tt.line {
			d.Push(byte(c))
		}
		e, m := d.Vestige()
		if e != tt.expectedEvent || m != tt.expectedMatch {
			t.Errorf("test %s: vestige test failed, got %s %v, expected %s %v", tt.name, e, m, tt.expectedEvent, tt.expectedMatch)
		}
	}
}
