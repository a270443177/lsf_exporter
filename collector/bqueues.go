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

type QueuesCollector struct {
	QueuesRuningJobCount  *prometheus.Desc
	QueuesPendingJobCount *prometheus.Desc
	QueuesMaxJobCount     *prometheus.Desc
	queuesPriority        *prometheus.Desc
	QueuesStatus          *prometheus.Desc
	logger                log.Logger
}

func init() {
	registerCollector("bqueues", defaultEnabled, NewLSFQueuesCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFQueuesCollector(logger log.Logger) (Collector, error) {

	return &QueuesCollector{
		QueuesRuningJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bqueues", "runingjob_count"),
			"The total number of tasks for all running jobs in the queue. If the -alloc option is used, the total is allocated slots for the jobs in the queue.",
			[]string{"queues_name"}, nil,
		),
		QueuesPendingJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bqueues", "pendingjob_count"),
			"The total number of tasks for all pending jobs in the queue. If used with the -alloc option, total is zero.",
			[]string{"queues_name"}, nil,
		),
		QueuesStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bqueues", "status"),
			"The status of the queue. The following values are supported:	1-Open:Active、 2-Open:Inact_Win、 3-Closed:Active 4、Closed:Inact_Win	0-UnKnow	",
			[]string{"queues_name"}, nil,
		),
		queuesPriority: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bqueues", "priority"),
			"The priority of the queue. The larger the value, the higher the priority. If job priority is not configured, determines the queue search order at job dispatch, suspend, and resume time. Contrary to usual order of UNIX process priority, jobs from higher priority queues are dispatched first and jobs from lower priority queues are suspended first when hosts are overloaded.",
			[]string{"queues_name"}, nil,
		),
		QueuesMaxJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bqueues", "maxjob_count"),
			"The maximum number of job slots that can be used by the jobs from the queue. These job slots are used by dispatched jobs that are not yet finished, and by pending jobs that reserve slots.			",
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

func bqueues_CsvtoStruct(lsfOutput []byte, logger log.Logger) ([]bqueuesInfo, error) {
	csv_out := csv.NewReader(TrimReader{bytes.NewReader(lsfOutput)})
	csv_out.LazyQuotes = true
	csv_out.Comma = ' '
	csv_out.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csv_out)
	if err != nil {
		level.Error(logger).Log("err=", err)
		return nil, nil
	}

	var bqueuesInfos []bqueuesInfo

	for {
		var u bqueuesInfo
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			level.Error(logger).Log("err=", err)
			return nil, nil
		}

		bqueuesInfos = append(bqueuesInfos, u)
	}
	return bqueuesInfos, nil

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
		level.Error(c.logger).Log("err=", err)
		return nil
	}
	queues, err := bqueues_CsvtoStruct(output, c.logger)
	if err != nil {
		level.Error(c.logger).Log("err=", err)
		return nil
	}

	for _, q := range queues {
		MAXCount, err := strconv.ParseFloat(q.MAX, 64)
		if err != nil {
			MAXCount = -1
		}
		ch <- prometheus.MustNewConstMetric(c.QueuesRuningJobCount, prometheus.GaugeValue, q.RUN, q.QUEUE_NAME)
		ch <- prometheus.MustNewConstMetric(c.QueuesPendingJobCount, prometheus.GaugeValue, q.PEND, q.QUEUE_NAME)
		ch <- prometheus.MustNewConstMetric(c.QueuesMaxJobCount, prometheus.GaugeValue, MAXCount, q.QUEUE_NAME)
		ch <- prometheus.MustNewConstMetric(c.queuesPriority, prometheus.GaugeValue, q.PRIO, q.QUEUE_NAME)
		ch <- prometheus.MustNewConstMetric(c.QueuesStatus, prometheus.GaugeValue, FormatQueusStatus(q.STATUS, c.logger), q.QUEUE_NAME)
	}

	return nil
}
