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

type lshostsCollector struct {
	HostMaxMem *prometheus.Desc
	HostMaxSWP *prometheus.Desc
	HostNCpus  *prometheus.Desc
	HostCpuf   *prometheus.Desc
	logger     log.Logger
}

func init() {
	registerCollector("lshosts", defaultEnabled, NewLSFlshostCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLSFlshostCollector(logger log.Logger) (Collector, error) {

	return &lshostsCollector{
		HostMaxMem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lshosts", "max_mem"),
			"The maximum amount of physical memory available for user processes.     By default, the amount is displayed in KB. The amount can appear in MB depending on the actual system memory. Use the LSF_UNIT_FOR_LIMITS parameter in the lsf.conf file to specify a larger unit for the limit (GB, TB, PB, or EB).",
			[]string{"host_name", "host_type", "host_model", "server_type", "resource_type"}, nil,
		),
		HostMaxSWP: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lshosts", "max_swp"),
			"The total available swap space.  By default, the amount is displayed in KB. The amount can appear in MB depending on the actual system swap space. Use the LSF_UNIT_FOR_LIMITS parameter in the lsf.conf file to specify a larger unit for the limit (GB, TB, PB, or EB).",
			[]string{"host_name", "host_type", "host_model", "server_type", "resource_type"}, nil,
		),
		HostNCpus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lshosts", "ncpus"),
			"The number of processors on this host. If the LSF_ENABLE_DUALCORE=Y parameter is specified in the lsf.conf file for multi-core CPU hosts, displays the number of cores instead of physical CPUs.",
			[]string{"host_name", "host_type", "host_model", "server_type", "resource_type"}, nil,
		),
		HostCpuf: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lshosts", "cpuf"),
			"The relative CPU performance factor. The CPU factor is used to scale the CPU load value so that differences in CPU speeds are considered. The faster the CPU, the larger the CPU factor.The default CPU factor of a host with an host type is 1.0. unknown",
			[]string{"host_name", "host_type", "host_model", "server_type", "resource_type"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *lshostsCollector) Update(ch chan<- prometheus.Metric) error {
	// err := c.getLmstatInfo(ch)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get lmstat version information: %w", err)
	// }

	err := c.parselshostsCount(ch)

	if err != nil {
		return fmt.Errorf("couldn't get bhosts infomation: %w", err)
	}

	return nil
}

func lshosts_CsvtoStruct(lsfOutput []byte, logger log.Logger) ([]lshostsInfo, error) {
	csv_out := csv.NewReader(TrimReader{bytes.NewReader(lsfOutput)})
	csv_out.LazyQuotes = true
	csv_out.Comma = ' '
	csv_out.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csv_out)
	if err != nil {
		level.Error(logger).Log("err=", err)
		return nil, nil
	}

	var lshostsInfos []lshostsInfo

	for {
		var u lshostsInfo
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			level.Error(logger).Log("err=", err)
			return nil, nil
		}

		lshostsInfos = append(lshostsInfos, u)
	}
	return lshostsInfos, nil

}

func FormatlshostsUnit(size float64, unit string) float64 {
	//单位换算  KB，MB，GB, TB, PB, or EB -> KB
	switch unit {
	case "K":
		return size
	case "M":
		return size * 1024
	case "G":
		return size * 1024 * 1024
	case "T":
		return size * 1024 * 1024 * 1024
	case "P":
		return size * 1024 * 1024 * 1024 * 1024
	case "E":
		return size * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return -1
	}
}

// 转义服务器类型
func ConvertServerType(server_type string) string {
	switch strings.ToLower(server_type) {
	case "yes":
		return "servers"
	case "no":
		return "client"
	case "dyn":
		return "dynamic"
	default:
		return "unknown"
	}
}

// 去除字符串中的()
func ConvertresourceType(resource_type string) string {
	resource_type_1 := strings.ReplaceAll(resource_type, "(", "")
	resource_type_new := strings.ReplaceAll(resource_type_1, ")", "")
	return resource_type_new
}

func (c *lshostsCollector) parselshostsCount(ch chan<- prometheus.Metric) error {
	output, err := lsfOutput(c.logger, "lshosts", "-w")
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}
	lshosts, err := lshosts_CsvtoStruct(output, c.logger)
	if err != nil {
		level.Error(c.logger).Log("err: ", err)
		return nil
	}

	for _, lshost := range lshosts {

		Ncpus, err := strconv.ParseFloat(lshost.Ncpus, 64)
		if err != nil {
			Ncpus = -1
		}

		Cpuf, err := strconv.ParseFloat(lshost.Cpuf, 64)
		if err != nil {
			Cpuf = -1
		}

		var dataSize float64
		var dataUnit string
		fmt.Sscanf(lshost.Maxmem, "%f%s", &dataSize, &dataUnit)
		ch <- prometheus.MustNewConstMetric(c.HostMaxMem, prometheus.GaugeValue, FormatlshostsUnit(dataSize, dataUnit), lshost.HOST_NAME, lshost.HOST_TYPE, lshost.Model, ConvertServerType(lshost.Server), ConvertresourceType(lshost.RESOURCES))

		fmt.Sscanf(lshost.Maxswp, "%f%s", &dataSize, &dataUnit)
		ch <- prometheus.MustNewConstMetric(c.HostMaxSWP, prometheus.GaugeValue, FormatlshostsUnit(dataSize, dataUnit), lshost.HOST_NAME, lshost.HOST_TYPE, lshost.Model, ConvertServerType(lshost.Server), ConvertresourceType(lshost.RESOURCES))

		ch <- prometheus.MustNewConstMetric(c.HostNCpus, prometheus.GaugeValue, Ncpus, lshost.HOST_NAME, lshost.HOST_TYPE, lshost.Model, ConvertServerType(lshost.Server), ConvertresourceType(lshost.RESOURCES))
		ch <- prometheus.MustNewConstMetric(c.HostCpuf, prometheus.GaugeValue, Cpuf, lshost.HOST_NAME, lshost.HOST_TYPE, lshost.Model, ConvertServerType(lshost.Server), ConvertresourceType(lshost.RESOURCES))
	}

	return nil
}
