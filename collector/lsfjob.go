package collector

import (
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type Job struct {
	ID            string
	User          string
	Status        string
	Queue         string
	FromHost      string
	ExecutionHost string
	JobName       string
	SubmitTime    int64
}

type JobCollector struct {
	JobInfo *prometheus.Desc
	logger  log.Logger
}

func init() {
	registerCollector("lsfjob", defaultEnabled, NewLSFJobCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFJobCollector(logger log.Logger) (Collector, error) {

	return &JobCollector{
		JobInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bjobs", "status"),
			"这边是测试的测试的 测试的测试的",
			[]string{"ID", "User", "Status", "Queue", "FromHost", "ExecutionHost", "JobName"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *JobCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.getJobStatus(ch)
	if err != nil {
		return fmt.Errorf("couldn't get queues infomation: %w", err)
	}

	return nil
}

func parseJJobStatus() Job {
	now := time.Now().Unix()
	return Job{
		ID:            "1",
		User:          "t01",
		Status:        "RUN",
		Queue:         "normal",
		FromHost:      "master01",
		ExecutionHost: "master01",
		JobName:       "sleep 10",
		SubmitTime:    now,
	}

}

func (c *JobCollector) getJobStatus(ch chan<- prometheus.Metric) error {
	parseJobStatus := parseJJobStatus()
	ch <- prometheus.MustNewConstMetric(c.JobInfo, prometheus.GaugeValue, float64(parseJobStatus.SubmitTime), parseJobStatus.ID, parseJobStatus.User, parseJobStatus.Status, parseJobStatus.Queue, parseJobStatus.FromHost, parseJobStatus.ExecutionHost, parseJobStatus.JobName)
	return nil

}
