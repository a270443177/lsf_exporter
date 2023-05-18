package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type bHostsCollector struct {
	LsfInfo          *prometheus.Desc
	JobNJobsCount    *prometheus.Desc
	JobRuningCount   *prometheus.Desc
	JobMaxJobCount   *prometheus.Desc
	JobSSUSPJobCount *prometheus.Desc
	JobUSUSPJobCount *prometheus.Desc
	bhostStatus      *prometheus.Desc
	logger           log.Logger
}

const (
	notFound = "not found"
)

func init() {
	registerCollector("bhosts", defaultEnabled, NewLSFbHostCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFbHostCollector(logger log.Logger) (Collector, error) {

	return &bHostsCollector{
		JobRuningCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "runing_count"),
			"The number of tasks for all running jobs on the host.",
			[]string{"host_name"}, nil,
		),
		JobNJobsCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "njobs_count"),
			"The number of tasks for all jobs that are dispatched to the host. The NJOBS value includes running, suspended, and chunk jobs.",
			[]string{"host_name"}, nil,
		),
		JobMaxJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "maxjobs_count"),
			"The maximum number of job slots available. A dash (-1) indicates no limit.",
			[]string{"host_name"}, nil,
		),
		JobSSUSPJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "ssusp_count"),
			"The number of tasks for all system suspended jobs on the host.",
			[]string{"host_name"}, nil,
		),
		JobUSUSPJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "ususp_count"),
			"The number of tasks for all user suspended jobs on the host. Jobs can be suspended by the user or by the LSF administrator.",
			[]string{"host_name"}, nil,
		),
		bhostStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "host_status"),
			"The status of the host and the sbatchd daemon. Batch jobs can be dispatched only to hosts with an ok status. Host status has the following, 0:Unknow, 1:ok, 2:unavail, 3:unreach, 4:closed/closed_full, 5:closed_cu_excl",
			[]string{"host_name"}, nil,
		),
		LsfInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "info"),
			"A metric with a constant '1' value labeled by ClusterName, MasterName and Version of the IBM Spectrum LSF .",
			[]string{"clustername", "mastername", "version"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *bHostsCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.parsebHostJobCount(ch)

	if err != nil {
		return fmt.Errorf("couldn't get bhosts infomation: %w", err)
	}

	err = c.parsebLsfClusterInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get bhosts infomation: %w", err)
	}

	return nil
}

func bhostSplite(lsfOutput []byte, logger log.Logger) (map[int]*bhostInfo, error) {
	r := csv.NewReader(bytes.NewReader(lsfOutput))
	r.LazyQuotes = true

	result, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not parse lsf_tools output: %w", err)
	}

	var index = 0
	var bhostInfos = make(map[int]*bhostInfo)

	//去除header列，并且遍历后存放到struct中
	for _, v := range result[1:] {
		var max float64
		level.Debug(logger).Log("当前获取到的字符串是", v[0])
		r := regexp.MustCompile(`[^\s]+`)
		arr := r.FindAllString(v[0], -1)

		njobs, _ := strconv.ParseFloat(arr[4], 64)
		run, _ := strconv.ParseFloat(arr[5], 64)
		ssusp, _ := strconv.ParseFloat(arr[6], 64)
		ususp, _ := strconv.ParseFloat(arr[7], 64)
		rsv, _ := strconv.ParseFloat(arr[8], 64)
		if arr[3] == "-" {
			max, _ = strconv.ParseFloat("-1", 64)
		} else {
			max, _ = strconv.ParseFloat(arr[3], 64)

		}
		bhostInfos[index] = &bhostInfo{
			name:   arr[0],
			status: arr[1],
			jl_u:   arr[2],
			maxjob: max,
			njobs:  njobs,
			run:    run,
			susp:   ssusp,
			uusp:   ususp,
			rsv:    rsv,
		}
		index++
	}

	return bhostInfos, nil
}

func FormatbhostsStatus(status string, logger log.Logger) float64 {
	state := strings.ToLower(status)
	level.Debug(logger).Log("当前获取到的值是", status, "转换后的值是", state)
	switch {
	case state == "ok":
		return float64(1)
	case state == "unavail":
		return float64(2)
	case state == "unreach":
		return float64(3)
	case state == "closed":
		return float64(4)
	case state == "closed_full":
		return float64(4)
	case state == "closed_cu_excl":
		return float64(5)
	default:
		return float64(0)
	}
}

func (c *bHostsCollector) parsebHostJobCount(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "bhosts", "-w")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	bhost, err := bhostSplite(output, c.logger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, q := range bhost {
		ch <- prometheus.MustNewConstMetric(c.JobNJobsCount, prometheus.GaugeValue, q.njobs, q.name)
		ch <- prometheus.MustNewConstMetric(c.JobRuningCount, prometheus.GaugeValue, q.run, q.name)
		ch <- prometheus.MustNewConstMetric(c.JobMaxJobCount, prometheus.GaugeValue, q.maxjob, q.name)
		ch <- prometheus.MustNewConstMetric(c.JobSSUSPJobCount, prometheus.GaugeValue, q.susp, q.name)
		ch <- prometheus.MustNewConstMetric(c.JobUSUSPJobCount, prometheus.GaugeValue, q.uusp, q.name)
		ch <- prometheus.MustNewConstMetric(c.bhostStatus, prometheus.GaugeValue, FormatbhostsStatus(q.status, c.logger), q.name)
	}

	return nil
}

func (c *bHostsCollector) parsebLsfClusterInfo(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "lsid", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	lsf_summary := string(output)
	md := map[string]string{}
	if ClusterNameRegex.MatchString(lsf_summary) {
		names := ClusterNameRegex.SubexpNames()
		matches := ClusterNameRegex.FindAllStringSubmatch(lsf_summary, -1)[0]
		for i, n := range matches {
			md[names[i]] = n
		}
	}

	if MasterNameRegex.MatchString(lsf_summary) {
		names := MasterNameRegex.SubexpNames()
		matches := MasterNameRegex.FindAllStringSubmatch(lsf_summary, -1)[0]
		for i, n := range matches {
			md[names[i]] = n
		}
	}

	if LSFVersionRegex.MatchString(lsf_summary) {
		names := LSFVersionRegex.SubexpNames()
		matches := LSFVersionRegex.FindAllStringSubmatch(lsf_summary, -1)[0]
		for i, n := range matches {
			md[names[i]] = n
		}
	}

	level.Debug(c.logger).Log("当前集群名称：", md["cluster_name"], ",当前的master节点名是:", md["master_name"], ",版本是:", md["lsf_version"])
	ch <- prometheus.MustNewConstMetric(c.LsfInfo, prometheus.GaugeValue, 1.0, md["cluster_name"], md["master_name"], md["lsf_version"])

	return nil
}
