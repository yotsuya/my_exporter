package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/procfs"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "openio"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last run of gridinit_cmd successful.",
		nil, nil,
	)
	procUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "process_up"),
		"Status of the process (1 = UP, 0 = DOWN).",
		[]string{"pid", "group"}, nil,
	)
	vsize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "process_virtual_memory_bytes"),
		"Virtual memory size in bytes.",
		[]string{"pid", "group"}, nil,
	)
	rss = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "process_resident_memory_bytes"),
		"Resident memory size in bytes.",
		[]string{"pid", "group"}, nil,
	)
)

// Exporter collects OpenIO stats from the `*** TBD ***` and exports them using
// the prometheus metrics package.
type Exporter struct {
	mutex  sync.RWMutex
	logger log.Logger
}

// Describe describes all the metrics exported by the OpenIO exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- vsize
	ch <- rss
}

// Collect fetches the stats from /proc/{pid}/stat and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	upVal := 0.0

	scanner := bufio.NewScanner(strings.NewReader(runGridInitCmd()))
	for scanner.Scan() {
		line := scanner.Text()
		level.Debug(e.logger).Log("line", line)

		key, status, pid, group, err := parseStatusLine(line)
		if err != nil {
			level.Warn(e.logger).Log("msg", "Invalid status line", "err", err, "line", line)
			continue
		}
		if key == "KEY" {
			// skip header line
			continue
		}

		// If at least one process information can be obtained, the result of gridinit_cmd
		// is considered successful.
		upVal = 1.0

		if status == "UP" {
			ch <- prometheus.MustNewConstMetric(procUp, prometheus.GaugeValue, 1.0, pid, group)
		} else {
			ch <- prometheus.MustNewConstMetric(procUp, prometheus.GaugeValue, 0.0, pid, group)
			if status != "DOWN" {
				level.Warn(e.logger).Log("msg", "Unknown process status", "line", line)
			}
		}

		nPid, err := strconv.Atoi(pid)
		if err != nil {
			level.Warn(e.logger).Log("msg", "Invalid PID", "line", line)
			continue
		}
		if nPid < 1 {
			// If the STATUS is "DOWN", the PID will be -1 or 0.
			continue
		}

		proc, err := procfs.NewProc(nPid)
		if err != nil {
			level.Warn(e.logger).Log("msg", "Error `procfs.NewProc()`", "err", err, "pid", pid)
			continue
		}

		stat, err := proc.Stat()
		if err != nil {
			level.Warn(e.logger).Log("msg", "Error `proc.Stat()`", "err", err, "pid", pid)
			continue
		}

		ch <- prometheus.MustNewConstMetric(vsize, prometheus.GaugeValue, float64(stat.VirtualMemory()), pid, group)
		ch <- prometheus.MustNewConstMetric(rss, prometheus.GaugeValue, float64(stat.ResidentMemory()), pid, group)
	}

	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, upVal)
}

func runGridInitCmd() string {
	output, _ := exec.Command("gridinit_cmd", "status").Output()
	// Since the exit code of gridinit_cmd may return 0 even if execution fails,
	// it does not make sense to check for errors.
	// ```
	// $ gridinit_cmd status
	// Connection to UNIX socket [/run/gridinit/gridinit.sock] failed : Permission denied
	// KEY  STATUS      PID GROUP
	// $ echo $?
	// 0
	// ```
	return string(output)
}

func parseStatusLine(line string) (key, status, pid, group string, err error) {
	_, err = fmt.Sscan(line, &key, &status, &pid, &group)
	return key, status, pid, group, err
}

// NewExporter returns an initialized exporter.
func NewExporter(logger log.Logger) (*Exporter, error) {
	return &Exporter{
		logger: logger,
	}, nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("openio_exporter"))
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":11010").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("openio_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting openio_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	exporter, err := NewExporter(logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Haproxy Exporter</title></head>
			<body>
			<h1>OpenIO Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
