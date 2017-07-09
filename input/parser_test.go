package input

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_Formats(t *testing.T) {
	var p *Parser
	mismatched := func(rtrnd string, intnd string, intndA string) {
		if intndA != "" {
			t.Fatalf("Parser format %v does not match the intended format %v.\n", rtrnd, intnd)
		}
		t.Fatalf("Parser format %v does not match the intended format %v (same as: %v).\n", rtrnd, intndA, intnd)
	}
	for i, f := range fmtsByName {
		p, _ = NewParser(f)
		if p.fmt != fmtsByStandard[i] {
			mismatched(p.fmt, f, fmtsByStandard[i])
		}
	}
	for _, f := range fmtsByStandard {
		p, _ = NewParser(f)
		if p.fmt != f {
			mismatched(p.fmt, f, "")
		}
	}
	p, err := NewParser("unknown-format")
	if err == nil {
		t.Fatalf("parser successfully created with invalid format")
	}
}

func Test_Parsing(t *testing.T) {
	tests := []struct {
		fmt      string
		message  string
		expected map[string]interface{}
		fail     bool
	}{
		{
			fmt:     "syslog",
			message: `Aug 24 05:14:15 info sshd[1999]: password accepted`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "Aug 24 05:14:15",
				"app":       "sshd",
				"pid":       1999,
				"message":   "password accepted",
			},
		},
		{
			fmt:     "syslog",
			message: `Apr 12 23:20:50 warn cron[304]: password accepted`,
			expected: map[string]interface{}{
				"priority":  "warn",
				"timestamp": "Apr 12 23:20:50",
				"app":       "cron",
				"pid":       304,
				"message":   "password accepted",
			},
		},
		{
			fmt:     "syslog",
			message: `Apr 12 19:20:50 error cron[65535]: password accepted`,
			expected: map[string]interface{}{
				"priority":  "error",
				"timestamp": "Apr 12 19:20:50",
				"app":       "cron",
				"pid":       65535,
				"message":   "password accepted",
			},
		},
		{
			fmt:     "syslog",
			message: `Oct 11 22:14:15 debug cron[65535]: password accepted`,
			expected: map[string]interface{}{
				"priority":  "debug",
				"timestamp": "Oct 11 22:14:15",
				"app":       "cron",
				"pid":       65535,
				"message":   "password accepted",
			},
		},
		{
			fmt:     "syslog",
			message: `Aug 24 05:14:15 some cron[65535]: JVM NPE\nsome_file.java:48\n\tsome_other_file.java:902`,
			expected: map[string]interface{}{
				"priority":  "some",
				"timestamp": "Aug 24 05:14:15",
				"app":       "cron",
				"pid":       65535,
				"message":   `JVM NPE\nsome_file.java:48\n\tsome_other_file.java:902`,
			},
		},
		{
			fmt:     "syslog",
			message: `Mar  2 22:53:45 any puppet-agent[5334]: mirrorurls.extend(list(self.metalink_data.urls()))`,
			expected: map[string]interface{}{
				"priority":  "any",
				"timestamp": "Mar  2 22:53:45",
				"app":       "puppet-agent",
				"pid":       5334,
				"message":   "mirrorurls.extend(list(self.metalink_data.urls()))",
			},
		},
		{
			fmt:     "syslog",
			message: `Mar  3 06:49:08 123 puppet-agent[51564]: (/Stage[main]/Users_prd/Ssh_authorized_key[1063-username]) Dependency Group[group] has failures: true`,
			expected: map[string]interface{}{
				"priority":  "123",
				"timestamp": "Mar  3 06:49:08",
				"app":       "puppet-agent",
				"pid":       51564,
				"message":   "(/Stage[main]/Users_prd/Ssh_authorized_key[1063-username]) Dependency Group[group] has failures: true",
			},
		},
		{
			fmt:     "syslog",
			message: `Mar  2 22:23:07 142 Keepalived_vrrp[21125]: VRRP_Instance(VI_1) ignoring received advertisement...`,
			expected: map[string]interface{}{
				"priority":  "142",
				"timestamp": "Mar  2 22:23:07",
				"app":       "Keepalived_vrrp",
				"pid":       21125,
				"message":   "VRRP_Instance(VI_1) ignoring received advertisement...",
			},
		},
		{
			fmt:     "syslog",
			message: `Mar  2 22:23:07 info Keepalived_vrrp[21125]: HEAD /wp-login.php HTTP/1.1" 200 167 "http://www.philipotoole.com/" "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.97 Safari/537.11`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "Mar  2 22:23:07",
				"app":       "Keepalived_vrrp",
				"pid":       21125,
				"message":   `HEAD /wp-login.php HTTP/1.1" 200 167 "http://www.philipotoole.com/" "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.97 Safari/537.11`,
			},
		},
		{
			fmt:     "syslog",
			message: `May  5 21:20:00 info apache-access[-]: 173.247.206.174 - - [05/May/2015:21:19:52 +0000] "GET /2013/11/ HTTP/1.1" 200 22056 "http://www.philipotoole.com/" "Wget/1.15 (linux-gnu)"`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "May  5 21:20:00",
				"app":       "apache-access",
				"pid":       0,
				"message":   `173.247.206.174 - - [05/May/2015:21:19:52 +0000] "GET /2013/11/ HTTP/1.1" 200 22056 "http://www.philipotoole.com/" "Wget/1.15 (linux-gnu)"`,
			},
		},
		{
			fmt:     "syslog",
			message: `Jun  4 14:09:13 info filterlog[-]: 67,,,0,vtnet0,match,pass,out,4,0x0,,127,3328,0,DF,6,tcp,366,192.168.1.66,31.13.86.4,50800,443,326,PA,1912507082:1912507408,2077294259,257,,`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "Jun  4 14:09:13",
				"app":       "filterlog",
				"pid":       0,
				"message":   `67,,,0,vtnet0,match,pass,out,4,0x0,,127,3328,0,DF,6,tcp,366,192.168.1.66,31.13.86.4,50800,443,326,PA,1912507082:1912507408,2077294259,257,,`,
			},
		},
		{
			fmt:     "syslog",
			message: `<134> 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 - password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 $ password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 - - password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `<33>7 2013-09-04T10:25:52.618085 test.com cron not_a_pid - password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `5:52.618085 test.com cron 65535 - password accepted`,
			fail:    true,
		},
		{
			fmt:     "syslog",
			message: `Jul  9 13:32:42 info music_userlist_m.t589375.0609.r[24681]: [src/channelIssues.cpp:198]: <26252>: onPGetOnlineUserListReqV2:uid 1756403961, topcid 96508020, subcid 96508020,userip 3722735630, terminate [android_6_4_1]`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "Jul  9 13:32:42",
				"app":       "music_userlist_m.t589375.0609.r",
				"pid":       24681,
				"message":   `[src/channelIssues.cpp:198]: <26252>: onPGetOnlineUserListReqV2:uid 1756403961, topcid 96508020, subcid 96508020,userip 3722735630, terminate [android_6_4_1]`,
			},
		},
		{
			fmt:     "syslog",
			message: `Jul  9 13:34:00 info music_dbgate_m.t537913.0124.r[29985]: [μs:588508][src/dbgateSystem.cpp:498]: <30044>: procDbAndRedis: feip: 101.226.24.7:59027 connId: 535 To execute SQL dbName: AttentionList_4, SourceExe: music_ssdb2mysql_m.t596386.0621.r, hashKey:  : INSERT INTO  attention_list_414    (uid, object, attr, updatetime, appdata) VALUES ( 578348414 ,  503755 , '1', '2017-07-09 13:34:00'  , 'shenqu.yy.com'  );, size: 0 result:0 usetime: 0`,
			expected: map[string]interface{}{
				"priority":  "info",
				"timestamp": "Jul  9 13:34:00",
				"app":       "music_dbgate_m.t537913.0124.r",
				"pid":       29985,
				"message":   `[μs:588508][src/dbgateSystem.cpp:498]: <30044>: procDbAndRedis: feip: 101.226.24.7:59027 connId: 535 To execute SQL dbName: AttentionList_4, SourceExe: music_ssdb2mysql_m.t596386.0621.r, hashKey:  : INSERT INTO  attention_list_414    (uid, object, attr, updatetime, appdata) VALUES ( 578348414 ,  503755 , '1', '2017-07-09 13:34:00'  , 'shenqu.yy.com'  );, size: 0 result:0 usetime: 0`,
			},
		},
	}

	for i, tt := range tests {
		p, _ := NewParser(tt.fmt)
		t.Logf("using %d\n", i+1)
		ok := p.Parse(bytes.NewBufferString(tt.message).Bytes())
		if tt.fail {
			if ok {
				t.Error("\n\nParser should fail.\n")
			}
		} else {
			if !ok {
				t.Error("\n\nParser should succeed.\n")
			}
		}
		if !tt.fail && !reflect.DeepEqual(tt.expected, p.Result) {
			t.Logf("%v", p.Result)
			t.Logf("%v", tt.expected)
			t.Error("\n\nParser result does not match expected result.\n")
		}
	}
}

func Benchmark_Parsing(b *testing.B) {
	p, _ := NewParser("syslog")
	for n := 0; n < b.N; n++ {
		ok := p.Parse(bytes.NewBufferString(`May  5 21:20:00 info apache-access[-]: 173.247.206.174 - - [05/May/2015:21:19:52 +0000] "GET /2013/11/ HTTP/1.  1" 200 22056 "http://www.philipotoole.com/" "Wget/1.15 (linux-gnu)"`).Bytes())
		if !ok {
			panic("message failed to parse during benchmarking")
		}
	}
}
