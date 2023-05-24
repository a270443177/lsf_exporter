package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jszwec/csvutil"
	"github.com/prometheus/client_golang/prometheus"
)

type bHostsCollector struct {
	HostRuningJobCount *prometheus.Desc
	HostNJobsCount     *prometheus.Desc

	HostMaxJobCount   *prometheus.Desc
	HostSSUSPJobCount *prometheus.Desc
	HostUSUSPJobCount *prometheus.Desc
	HostStatus        *prometheus.Desc
	logger            log.Logger
}

func init() {
	registerCollector("bhosts", defaultEnabled, NewLSFbHostCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFbHostCollector(logger log.Logger) (Collector, error) {

	return &bHostsCollector{
		HostRuningJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "runingjob_count"),
			"The number of tasks for all running jobs on the host.",
			[]string{"host_name"}, nil,
		),
		HostNJobsCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "njobs_count"),
			"The number of tasks for all jobs that are dispatched to the host. The NJOBS value includes running, suspended, and chunk jobs.",
			[]string{"host_name"}, nil,
		),
		HostMaxJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "maxjob_count"),
			"The maximum number of job slots available. A dash (-1) indicates no limit.",
			[]string{"host_name"}, nil,
		),
		HostSSUSPJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "ssuspjob_count"),
			"The number of tasks for all system suspended jobs on the host.",
			[]string{"host_name"}, nil,
		),
		HostUSUSPJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "ususpjob_count"),
			"The number of tasks for all user suspended jobs on the host. Jobs can be suspended by the user or by the LSF administrator.",
			[]string{"host_name"}, nil,
		),
		HostStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bhost", "host_status"),
			"The status of the host and the sbatchd daemon. Batch jobs can be dispatched only to hosts with an ok status. Host status has the following, 0:Unknow, 1:ok, 2:unavail, 3:unreach, 4:closed/closed_full, 5:closed_cu_excl",
			[]string{"host_name"}, nil,
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

	return nil
}

type TrimReader struct{ io.Reader }

var trailingws = regexp.MustCompile(` +\r?\n`)

func (tr TrimReader) Read(bs []byte) (int, error) {
	// Perform the requested read on the given reader.
	n, err := tr.Reader.Read(bs)
	if err != nil {
		return n, err
	}

	// Remove trailing whitespace from each line.
	lines := string(bs[:n])
	trimmed := []byte(trailingws.ReplaceAllString(lines, "\n"))
	copy(bs, trimmed)
	return len(trimmed), nil
}

func bhost_CsvtoStruct(lsfOutput []byte, logger log.Logger) ([]bhostInfo, error) {
	csv_out := csv.NewReader(TrimReader{bytes.NewReader(lsfOutput)})
	csv_out.LazyQuotes = true
	csv_out.Comma = ' '
	csv_out.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csv_out)
	if err != nil {
		level.Error(logger).Log("err=", err)
		return nil, nil
	}

	var bhostInfos []bhostInfo

	for {
		var u bhostInfo
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			level.Error(logger).Log("err=", err)
			return nil, nil
		}

		bhostInfos = append(bhostInfos, u)
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
	default:
		return float64(0)
	}
}

func (c *bHostsCollector) parsebHostJobCount(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "bhosts", "-w")
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}
	bhosts, err := bhost_CsvtoStruct(output, c.logger)
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}

	for _, bhost := range bhosts {
		ch <- prometheus.MustNewConstMetric(c.HostNJobsCount, prometheus.GaugeValue, bhost.NJOBS, bhost.HOST_NAME)
		ch <- prometheus.MustNewConstMetric(c.HostRuningJobCount, prometheus.GaugeValue, bhost.RUN, bhost.HOST_NAME)
		ch <- prometheus.MustNewConstMetric(c.HostMaxJobCount, prometheus.GaugeValue, bhost.MAX, bhost.HOST_NAME)
		ch <- prometheus.MustNewConstMetric(c.HostSSUSPJobCount, prometheus.GaugeValue, bhost.SSUSP, bhost.HOST_NAME)
		ch <- prometheus.MustNewConstMetric(c.HostUSUSPJobCount, prometheus.GaugeValue, bhost.USUSP, bhost.HOST_NAME)
		ch <- prometheus.MustNewConstMetric(c.HostStatus, prometheus.GaugeValue, FormatbhostsStatus(bhost.STATUS, c.logger), bhost.HOST_NAME)
	}

	return nil
}
