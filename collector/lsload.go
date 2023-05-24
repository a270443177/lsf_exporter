package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jszwec/csvutil"
	"github.com/prometheus/client_golang/prometheus"
)

type lsLoadCollector struct {
	LsLoadR15s       *prometheus.Desc
	LsLoadR1m        *prometheus.Desc
	LsLoadR15m       *prometheus.Desc
	LsLoadut         *prometheus.Desc
	LsLoadls         *prometheus.Desc
	LsLoadHostStatus *prometheus.Desc
	logger           log.Logger
}

func init() {
	registerCollector("lsload", defaultEnabled, NewLSFlsLoadCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFlsLoadCollector(logger log.Logger) (Collector, error) {

	return &lsLoadCollector{
		LsLoadR15s: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "r15s"),
			"The 15 second exponentially averaged CPU run queue length.",
			[]string{"host_name"}, nil,
		),
		LsLoadR1m: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "r1m"),
			"The 1 minute exponentially averaged CPU run queue length.",
			[]string{"host_name"}, nil,
		),
		LsLoadR15m: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "r15m"),
			"The 15 minute exponentially averaged CPU run queue length.",
			[]string{"host_name"}, nil,
		),
		LsLoadut: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "ut"),
			"The CPU utilization exponentially averaged over the last minute, 0 - 1.",
			[]string{"host_name"}, nil,
		),
		LsLoadls: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "login_usersCount"),
			"The number of current login users.",
			[]string{"host_name"}, nil,
		),
		LsLoadHostStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lsload", "host_status"),
			"The status of the host and the sbatchd daemon. Batch jobs can be dispatched only to hosts with an ok status. Host status has the following, 0:Unknow, 1:ok, 2:unavail, 3:unreach, 4:closed/closed_full, 5:closed_cu_excl",
			[]string{"host_name"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *lsLoadCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.parselsLoad(ch)

	if err != nil {
		return fmt.Errorf("couldn't get bhosts infomation: %w", err)
	}

	return nil
}

func lsload_CsvtoStruct(lsfOutput []byte, logger log.Logger) ([]lsloadInfo, error) {
	csv_out := csv.NewReader(TrimReader{bytes.NewReader(lsfOutput)})
	csv_out.LazyQuotes = true
	csv_out.Comma = ' '
	csv_out.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csv_out)
	if err != nil {
		level.Error(logger).Log("err=", err)
		return nil, nil
	}

	var lsloadInfos []lsloadInfo

	for {
		var u lsloadInfo
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			level.Error(logger).Log("err=", err)
			return nil, nil
		}

		lsloadInfos = append(lsloadInfos, u)
	}
	return lsloadInfos, nil

}

func FormatlsLoadStatus(status string, logger log.Logger) float64 {
	state := strings.ToLower(status)
	level.Debug(logger).Log("当前获取到的值是", status, "转换后的值是", state)
	switch {
	case state == "ok":
		return float64(1)
	case state == "-ok":
		return float64(2)
	case state == "busy":
		return float64(3)
	case state == "lockw":
		return float64(4)
	case state == "locku":
		return float64(5)
	case state == "unavail":
		return float64(6)
	default:
		return float64(0)
	}
}

func ConvertUT(data string, logger log.Logger) float64 {
	data_new := strings.ReplaceAll(data, "%", "")

	fl, err := strconv.ParseFloat(data_new, 64)
	if err != nil {
		level.Error(logger).Log("err: ", err)
		return -1
	}
	return fl
}

func (c *lsLoadCollector) parselsLoad(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "lsload", "-w")
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}
	lsloads, err := lsload_CsvtoStruct(output, c.logger)
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}

	for _, lsload := range lsloads {
		ch <- prometheus.MustNewConstMetric(c.LsLoadR15s, prometheus.GaugeValue, lsload.R15S, lsload.Name)
		ch <- prometheus.MustNewConstMetric(c.LsLoadR1m, prometheus.GaugeValue, lsload.R1M, lsload.Name)
		ch <- prometheus.MustNewConstMetric(c.LsLoadR15m, prometheus.GaugeValue, lsload.R15M, lsload.Name)
		ch <- prometheus.MustNewConstMetric(c.LsLoadut, prometheus.GaugeValue, ConvertUT(lsload.UT, c.logger), lsload.Name)
		ch <- prometheus.MustNewConstMetric(c.LsLoadls, prometheus.GaugeValue, lsload.LS, lsload.Name)
		ch <- prometheus.MustNewConstMetric(c.LsLoadHostStatus, prometheus.GaugeValue, FormatbhostsStatus(lsload.STATUS, c.logger), lsload.Name)
		// ch <- prometheus.MustNewConstMetric(c.JobRuningCount, prometheus.GaugeValue, bhost.RUN, bhost.HOST_NAME)
		// ch <- prometheus.MustNewConstMetric(c.JobMaxJobCount, prometheus.GaugeValue, bhost.MAX, bhost.HOST_NAME)
		// ch <- prometheus.MustNewConstMetric(c.JobSSUSPJobCount, prometheus.GaugeValue, bhost.SSUSP, bhost.HOST_NAME)
		// ch <- prometheus.MustNewConstMetric(c.JobUSUSPJobCount, prometheus.GaugeValue, bhost.USUSP, bhost.HOST_NAME)

	}

	return nil
}
