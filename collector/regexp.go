package collector

import "regexp"

var (
	// Regexp to parse lsid.
	ClusterNameRegex = regexp.MustCompile(`My\s+cluster\s+name\s+is\s+(?P<cluster_name>[^\s]+)`)
	MasterNameRegex  = regexp.MustCompile(`My\s+master\s+name\s+is\s+(?P<master_name>[^\s]+)`)
	LSFVersionRegex  = regexp.MustCompile(`(?P<lsf_version>\d+.\d+.\d+.\d+)`)

	ResourceRegex = regexp.MustCompile(`\((?P<resource_type>.+)\)`)
)
