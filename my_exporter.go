package main

import (
	"net/http"
	"os"
	"strconv"
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
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":11010").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()

	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last check of OpenIO successful.",
		nil, nil,
	)
	vsize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "process_virtual_memory_bytes"),
		"Virtual memory size in bytes.",
		[]string{"pid"}, nil,
	)
	rss = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "process_resident_memory_bytes"),
		"Resident memory size in bytes.",
		[]string{"pid"}, nil,
	)
)

// Exporter collects OpenIO stats from the `*** TBD ***` and exports them using
// the prometheus metrics package.
type Exporter struct {
	// TODO: プロセスを特定するための何かを追加
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

	allProcs, err := procfs.AllProcs()
	if err != nil {
		level.Error(e.logger).Log("msg", "AllProcs()", "err", err)
	}
	// TODO: OIOプロセスのみに限定する

	for _, p := range allProcs {
		// p, err := procfs.NewProc(proc.PID)
		// if err != nil {
		// 	// c.reportError(ch, nil, err)
		// 	level.Error(e.logger).Log("msg", "SOMETHING BAD!", "err", err)
		// 	return
		// }

		if stat, err := p.Stat(); err == nil {
			ch <- prometheus.MustNewConstMetric(vsize, prometheus.GaugeValue, float64(stat.VirtualMemory()), strconv.Itoa(p.PID))
			ch <- prometheus.MustNewConstMetric(rss, prometheus.GaugeValue, float64(stat.ResidentMemory()), strconv.Itoa(p.PID))
		} else {
			level.Error(e.logger).Log("msg", "Stat()", "err", err)
		}
	}

	// TODO: OIOプロセスが見つかったかどうかを返すようにする
	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1)
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
