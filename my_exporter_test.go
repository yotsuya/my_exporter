package main

import (
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseStatusLine(t *testing.T) {
	type Output struct {
		key    string
		status string
		pid    string
		group  string
	}

	var tests = []struct {
		input    string
		success  bool
		expected Output
	}{
		{"KEY                       STATUS      PID GROUP", true, Output{"KEY", "STATUS", "PID", "GROUP"}},
		{"OPENIO-account-0          UP         1163 OPENIO,account,0", true, Output{"OPENIO-account-0", "UP", "1163", "OPENIO,account,0"}},
		{"A B C", false, Output{}},
		{"A B C D", true, Output{"A", "B", "C", "D"}},
		{"A B C D E", true, Output{"A", "B", "C", "D"}},
		{" A B C D", true, Output{"A", "B", "C", "D"}},
	}

	for _, test := range tests {
		key, status, pid, group, err := parseStatusLine(test.input)
		if err != nil {
			if test.success {
				t.Errorf("parseStatusLine(%q) => Error: %v", test.input, err)
			} else {
				// pass
			}
			continue
		}

		actual := Output{key, status, pid, group}
		if !test.success {
			t.Errorf("parseStatusLine(%q) => Expected: Error, Actual: %v", test.input, actual)
			continue
		}
		if actual != test.expected {
			t.Errorf("parseStatusLine(%q) => Expected: %v, Actual: %v", test.input, test.expected, actual)
			continue
		}
	}
}

type ProcStatStub struct {
	pid int
}

func (s ProcStatStub) CPUTime() float64            { return float64(s.pid * 10.0) }
func (s ProcStatStub) VirtualMemory() uint         { return uint(s.pid * 100) }
func (s ProcStatStub) ResidentMemory() int         { return s.pid * 1000 }
func (s ProcStatStub) StartTime() (float64, error) { return float64(s.pid * 10000.0), nil }

func TestCollect(t *testing.T) {
	exporter, err := NewExporter(
		log.NewNopLogger(),
		func() string {
			return `KEY                       STATUS      PID GROUP
OPENIO-account-0          UP            1 OPENIO,account,0
OPENIO-beanstalkd-0       UP            2 OPENIO,beanstalkd,0
OPENIO-blob_rebuilder-0   UP            3 OPENIO,oio-blob-rebuilder,0
OPENIO-conscience-0       UP            4 OPENIO,conscience,0
OPENIO-conscienceagent-0  UP            5 OPENIO,conscienceagent,0
OPENIO-memcached-0        UP            6 OPENIO,memcached,0
OPENIO-meta0-0            UP            7 OPENIO,meta0,0
OPENIO-meta1-0            UP            8 OPENIO,meta1,0
OPENIO-meta2-0            UP            9 OPENIO,meta2,0
OPENIO-oio-blob-indexer-0 UP           10 OPENIO,oio-blob-indexer,0
OPENIO-oio-event-agent-0  UP           11 OPENIO,oio-event-agent,0
OPENIO-oioproxy-0         UP           12 OPENIO,oioproxy,0
OPENIO-oioswift-0         UP           13 OPENIO,oioswift,0
OPENIO-rawx-0             UP           14 OPENIO,rawx,0
OPENIO-rdir-1             UP           15 OPENIO,rdir,1
OPENIO-redis-0            UP           16 OPENIO,redis,0`
		},
		func(pid int) (ProcStat, error) {
			return ProcStatStub{pid}, nil
		},
	)
	if err != nil {
		t.Errorf("expected no error but got %q", err)
	}
	prometheus.MustRegister(exporter)

	var tests = []struct {
		name     string
		expected string
	}{
		{
			"up",
			`# HELP openio_up Was the last run of gridinit_cmd successful.
# TYPE openio_up gauge
openio_up 1
`,
		},
		{
			"process_up",
			`# HELP openio_process_up Status of the process (1 = UP, 0 = DOWN).
# TYPE openio_process_up gauge
openio_process_up{group="OPENIO,account,0",pid="1"} 1
openio_process_up{group="OPENIO,beanstalkd,0",pid="2"} 1
openio_process_up{group="OPENIO,conscience,0",pid="4"} 1
openio_process_up{group="OPENIO,conscienceagent,0",pid="5"} 1
openio_process_up{group="OPENIO,memcached,0",pid="6"} 1
openio_process_up{group="OPENIO,meta0,0",pid="7"} 1
openio_process_up{group="OPENIO,meta1,0",pid="8"} 1
openio_process_up{group="OPENIO,meta2,0",pid="9"} 1
openio_process_up{group="OPENIO,oio-blob-indexer,0",pid="10"} 1
openio_process_up{group="OPENIO,oio-blob-rebuilder,0",pid="3"} 1
openio_process_up{group="OPENIO,oio-event-agent,0",pid="11"} 1
openio_process_up{group="OPENIO,oioproxy,0",pid="12"} 1
openio_process_up{group="OPENIO,oioswift,0",pid="13"} 1
openio_process_up{group="OPENIO,rawx,0",pid="14"} 1
openio_process_up{group="OPENIO,rdir,1",pid="15"} 1
openio_process_up{group="OPENIO,redis,0",pid="16"} 1
`,
		},
		{
			"process_cpu_seconds_total",
			`# HELP openio_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE openio_process_cpu_seconds_total counter
openio_process_cpu_seconds_total{group="OPENIO,account,0",pid="1"} 10
openio_process_cpu_seconds_total{group="OPENIO,beanstalkd,0",pid="2"} 20
openio_process_cpu_seconds_total{group="OPENIO,conscience,0",pid="4"} 40
openio_process_cpu_seconds_total{group="OPENIO,conscienceagent,0",pid="5"} 50
openio_process_cpu_seconds_total{group="OPENIO,memcached,0",pid="6"} 60
openio_process_cpu_seconds_total{group="OPENIO,meta0,0",pid="7"} 70
openio_process_cpu_seconds_total{group="OPENIO,meta1,0",pid="8"} 80
openio_process_cpu_seconds_total{group="OPENIO,meta2,0",pid="9"} 90
openio_process_cpu_seconds_total{group="OPENIO,oio-blob-indexer,0",pid="10"} 100
openio_process_cpu_seconds_total{group="OPENIO,oio-blob-rebuilder,0",pid="3"} 30
openio_process_cpu_seconds_total{group="OPENIO,oio-event-agent,0",pid="11"} 110
openio_process_cpu_seconds_total{group="OPENIO,oioproxy,0",pid="12"} 120
openio_process_cpu_seconds_total{group="OPENIO,oioswift,0",pid="13"} 130
openio_process_cpu_seconds_total{group="OPENIO,rawx,0",pid="14"} 140
openio_process_cpu_seconds_total{group="OPENIO,rdir,1",pid="15"} 150
openio_process_cpu_seconds_total{group="OPENIO,redis,0",pid="16"} 160
`,
		},
		{
			"process_virtual_memory_bytes",
			`# HELP openio_process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE openio_process_virtual_memory_bytes gauge
openio_process_virtual_memory_bytes{group="OPENIO,account,0",pid="1"} 100
openio_process_virtual_memory_bytes{group="OPENIO,beanstalkd,0",pid="2"} 200
openio_process_virtual_memory_bytes{group="OPENIO,conscience,0",pid="4"} 400
openio_process_virtual_memory_bytes{group="OPENIO,conscienceagent,0",pid="5"} 500
openio_process_virtual_memory_bytes{group="OPENIO,memcached,0",pid="6"} 600
openio_process_virtual_memory_bytes{group="OPENIO,meta0,0",pid="7"} 700
openio_process_virtual_memory_bytes{group="OPENIO,meta1,0",pid="8"} 800
openio_process_virtual_memory_bytes{group="OPENIO,meta2,0",pid="9"} 900
openio_process_virtual_memory_bytes{group="OPENIO,oio-blob-indexer,0",pid="10"} 1000
openio_process_virtual_memory_bytes{group="OPENIO,oio-blob-rebuilder,0",pid="3"} 300
openio_process_virtual_memory_bytes{group="OPENIO,oio-event-agent,0",pid="11"} 1100
openio_process_virtual_memory_bytes{group="OPENIO,oioproxy,0",pid="12"} 1200
openio_process_virtual_memory_bytes{group="OPENIO,oioswift,0",pid="13"} 1300
openio_process_virtual_memory_bytes{group="OPENIO,rawx,0",pid="14"} 1400
openio_process_virtual_memory_bytes{group="OPENIO,rdir,1",pid="15"} 1500
openio_process_virtual_memory_bytes{group="OPENIO,redis,0",pid="16"} 1600
`,
		},
		{
			"process_resident_memory_bytes",
			`# HELP openio_process_resident_memory_bytes Resident memory size in bytes.
# TYPE openio_process_resident_memory_bytes gauge
openio_process_resident_memory_bytes{group="OPENIO,account,0",pid="1"} 1000
openio_process_resident_memory_bytes{group="OPENIO,beanstalkd,0",pid="2"} 2000
openio_process_resident_memory_bytes{group="OPENIO,conscience,0",pid="4"} 4000
openio_process_resident_memory_bytes{group="OPENIO,conscienceagent,0",pid="5"} 5000
openio_process_resident_memory_bytes{group="OPENIO,memcached,0",pid="6"} 6000
openio_process_resident_memory_bytes{group="OPENIO,meta0,0",pid="7"} 7000
openio_process_resident_memory_bytes{group="OPENIO,meta1,0",pid="8"} 8000
openio_process_resident_memory_bytes{group="OPENIO,meta2,0",pid="9"} 9000
openio_process_resident_memory_bytes{group="OPENIO,oio-blob-indexer,0",pid="10"} 10000
openio_process_resident_memory_bytes{group="OPENIO,oio-blob-rebuilder,0",pid="3"} 3000
openio_process_resident_memory_bytes{group="OPENIO,oio-event-agent,0",pid="11"} 11000
openio_process_resident_memory_bytes{group="OPENIO,oioproxy,0",pid="12"} 12000
openio_process_resident_memory_bytes{group="OPENIO,oioswift,0",pid="13"} 13000
openio_process_resident_memory_bytes{group="OPENIO,rawx,0",pid="14"} 14000
openio_process_resident_memory_bytes{group="OPENIO,rdir,1",pid="15"} 15000
openio_process_resident_memory_bytes{group="OPENIO,redis,0",pid="16"} 16000
`,
		},
		{
			"process_start_time_seconds",
			`# HELP openio_process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE openio_process_start_time_seconds gauge
openio_process_start_time_seconds{group="OPENIO,account,0",pid="1"} 10000
openio_process_start_time_seconds{group="OPENIO,beanstalkd,0",pid="2"} 20000
openio_process_start_time_seconds{group="OPENIO,conscience,0",pid="4"} 40000
openio_process_start_time_seconds{group="OPENIO,conscienceagent,0",pid="5"} 50000
openio_process_start_time_seconds{group="OPENIO,memcached,0",pid="6"} 60000
openio_process_start_time_seconds{group="OPENIO,meta0,0",pid="7"} 70000
openio_process_start_time_seconds{group="OPENIO,meta1,0",pid="8"} 80000
openio_process_start_time_seconds{group="OPENIO,meta2,0",pid="9"} 90000
openio_process_start_time_seconds{group="OPENIO,oio-blob-indexer,0",pid="10"} 100000
openio_process_start_time_seconds{group="OPENIO,oio-blob-rebuilder,0",pid="3"} 30000
openio_process_start_time_seconds{group="OPENIO,oio-event-agent,0",pid="11"} 110000
openio_process_start_time_seconds{group="OPENIO,oioproxy,0",pid="12"} 120000
openio_process_start_time_seconds{group="OPENIO,oioswift,0",pid="13"} 130000
openio_process_start_time_seconds{group="OPENIO,rawx,0",pid="14"} 140000
openio_process_start_time_seconds{group="OPENIO,rdir,1",pid="15"} 150000
openio_process_start_time_seconds{group="OPENIO,redis,0",pid="16"} 160000
`,
		},
	}

	for _, test := range tests {
		err = testutil.CollectAndCompare(exporter, strings.NewReader(test.expected), namespace+"_"+test.name)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}
	}
}
