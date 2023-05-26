package collector

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type InformationCollector struct {
	LsfInformation *prometheus.Desc
	logger         log.Logger
}

func init() {
	registerCollector("lsf_information", defaultEnabled, NewLSFInformationCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFInformationCollector(logger log.Logger) (Collector, error) {

	return &InformationCollector{
		LsfInformation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "info"),
			"A metric with a constant '1' value labeled by ClusterName, MasterName and Version of the IBM Spectrum LSF .",
			[]string{"clustername", "mastername", "version"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *InformationCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.parsebLsfClusterInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get queues infomation: %w", err)
	}

	return nil
}

func lsfOutput(logger log.Logger, exe_file string, args ...string) ([]byte, error) {
	// _, err := os.Stat(*LSF_BINDIR)
	// if os.IsNotExist(err) {
	// 	level.Error(logger).Log("err", *LSF_BINDIR, "missing")
	// 	os.Exit(1)
	// }

	// _, err = os.Stat(*LSF_SERVERDIR)
	// if os.IsNotExist(err) {
	// 	level.Error(logger).Log("err", *LSF_SERVERDIR, "missing")
	// 	os.Exit(1)
	// }

	// _, err = os.Stat(*LSF_ENVDIR)
	// if os.IsNotExist(err) {
	// 	level.Error(logger).Log("err", *LSF_ENVDIR, "missing")
	// 	os.Exit(1)
	// }

	cmd := exec.Command(exe_file, args...)

	out, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("error while calling '%s %s': %v:'unknown error'",
			exe_file, strings.Join(args, " "), err)
	}

	return out, nil
}

func (c *InformationCollector) parsebLsfClusterInfo(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "lsid", "")
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
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
	ch <- prometheus.MustNewConstMetric(c.LsfInformation, prometheus.GaugeValue, 1.0, md["cluster_name"], md["master_name"], md["lsf_version"])

	return nil
}
