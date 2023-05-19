package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type QueuesCollector struct {
	QueuesPendingCount *prometheus.Desc
	QueuesRuningCount  *prometheus.Desc
	QueuesMaxJobCount  *prometheus.Desc
	QueuesStatus       *prometheus.Desc
	logger             log.Logger
}

func init() {
	registerCollector("bqueues", defaultEnabled, NewLSFQueuesCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFQueuesCollector(logger log.Logger) (Collector, error) {

	return &QueuesCollector{
		QueuesRuningCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "runing_count"),
			"how many runing of jobs in this queues",
			[]string{"queues_name"}, nil,
		),
		QueuesPendingCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "pending_count"),
			"how many pending of jobs in this queues",
			[]string{"queues_name"}, nil,
		),
		QueuesStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "queues", "status"),
			"this queues status",
			[]string{"queues_name"}, nil,
		),
		QueuesMaxJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "queues", "max_jobcount"),
			"this queues max job",
			[]string{"queues_name"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *QueuesCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.parseQueuesJobCount(ch)
	if err != nil {
		return fmt.Errorf("couldn't get queues infomation: %w", err)
	}

	return nil
}

func lsfOutput(logger log.Logger, exe_file string, args ...string) ([]byte, error) {
	_, err := os.Stat(*LSF_BINDIR)
	if os.IsNotExist(err) {
		level.Error(logger).Log("err", *LSF_BINDIR, "missing")
		os.Exit(1)
	}

	_, err = os.Stat(*LSF_SERVERDIR)
	if os.IsNotExist(err) {
		level.Error(logger).Log("err", *LSF_SERVERDIR, "missing")
		os.Exit(1)
	}

	_, err = os.Stat(*LSF_ENVDIR)
	if os.IsNotExist(err) {
		level.Error(logger).Log("err", *LSF_ENVDIR, "missing")
		os.Exit(1)
	}

	//设置环境变量
	PATH := os.Getenv("PATH")

	NEW_PATH := PATH + ":" + *LSF_BINDIR

	os.Setenv("PATH", NEW_PATH)
	os.Setenv("LSF_BINDIR", *LSF_BINDIR)
	os.Setenv("LSF_ENVDIR", *LSF_ENVDIR)
	os.Setenv("LSF_LIBDIR", *LSF_LIBDIR)
	os.Setenv("LSF_SERVERDIR", *LSF_SERVERDIR)

	cmd := exec.Command(exe_file, args...)

	out, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("error while calling '%s %s': %v:'unknown error'",
			exe_file, strings.Join(args, " "), err)
	}

	return out, nil
}

func QueuesSplite(lsfOutput []byte, logger log.Logger) (map[int]*queuesInfo, error) {
	r := csv.NewReader(bytes.NewReader(lsfOutput))
	r.LazyQuotes = true

	result, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not parse lsf_tools output: %w", err)
	}

	var index = 0
	var queuesInfos = make(map[int]*queuesInfo)

	//去除header列，并且遍历后存放到struct中
	for _, v := range result[1:] {
		var max float64
		level.Debug(logger).Log("当前获取到的字符串是", v[0])
		r := regexp.MustCompile(`[^\s]+`)
		arr := r.FindAllString(v[0], -1)

		prio, _ := strconv.Atoi(arr[1])
		njobs, _ := strconv.Atoi(arr[7])
		pend, _ := strconv.Atoi(arr[8])
		run, _ := strconv.Atoi(arr[9])
		if arr[3] == "-" {
			max, _ = strconv.ParseFloat("-1", 64)
		} else {
			max, _ = strconv.ParseFloat(arr[3], 64)

		}
		queuesInfos[index] = &queuesInfo{
			name:   arr[0],
			prio:   float64(prio),
			status: arr[2],
			maxjob: float64(max),
			jl_u:   arr[4],
			jl_p:   arr[5],
			jl_h:   arr[6],
			njobs:  float64(njobs),
			pend:   float64(pend),
			run:    float64(run),
			susp:   arr[10],
			rsv:    arr[11],
		}
		index++
	}

	return queuesInfos, nil
}

func FormatQueusStatus(status string, logger log.Logger) float64 {
	state := strings.ToLower(status)
	level.Debug(logger).Log("当前获取到的值是", status, "转换后的值是", state)
	switch {
	case state == "open:active":
		return float64(1)
	case state == "open:inact_win":
		return float64(2)
	case state == "closed:active":
		return float64(3)
	case state == "closed:inact_win":
		return float64(4)
	default:
		return float64(0)
	}
}

func (c *QueuesCollector) parseQueuesJobCount(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "bqueues", "-w")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	queues, err := QueuesSplite(output, c.logger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, q := range queues {
		ch <- prometheus.MustNewConstMetric(c.QueuesRuningCount, prometheus.GaugeValue, q.run, q.name)
		ch <- prometheus.MustNewConstMetric(c.QueuesPendingCount, prometheus.GaugeValue, q.pend, q.name)
		ch <- prometheus.MustNewConstMetric(c.QueuesMaxJobCount, prometheus.GaugeValue, q.maxjob, q.name)
		ch <- prometheus.MustNewConstMetric(c.QueuesStatus, prometheus.GaugeValue, FormatQueusStatus(q.status, c.logger), q.name)
	}

	return nil
}
