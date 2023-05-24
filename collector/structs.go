// Copyright 2017-2018 Mario Trangoni
// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

type queuesInfo struct {
	name   string
	prio   float64
	status string
	maxjob float64
	jl_u   string
	jl_p   string
	jl_h   string
	njobs  float64
	pend   float64
	run    float64
	susp   string
	rsv    string
}

// 以下是bhosts命令的struct
type bhostInfo struct {
	HOST_NAME string  `csv:"HOST_NAME"`
	STATUS    string  `csv:"STATUS"`
	JL_U      string  `csv:"JL/U"`
	MAX       float64 `csv:"MAX"`
	NJOBS     float64 `csv:"NJOBS"`
	RUN       float64 `csv:"RUN"`
	SSUSP     float64 `csv:"SSUSP"`
	USUSP     float64 `csv:"USUSP"`
	RSV       float64 `csv:"RSV"`
}

// 以下是bqueues命令的struct
type bqueuesInfo struct {
	QUEUE_NAME string  `csv:"QUEUE_NAME"`
	PRIO       float64 `csv:"PRIO"`
	STATUS     string  `csv:"STATUS"`
	MAX        string  `csv:"MAX"`
	JL_U       string  `csv:"JL/U"`
	JL_P       string  `csv:"JL/P"`
	JL_H       string  `csv:"JL/H"`
	NJOBS      float64 `csv:"NJOBS"`
	PEND       float64 `csv:"PEND"`
	RUN        float64 `csv:"RUN"`
	SUSP       string  `csv:"SUSP"`
	RSV        string  `csv:"RSV"`
}

// 以下是lsload命令的struct
type lsloadInfo struct {
	Name   string  `csv:"HOST_NAME"`
	STATUS string  `csv:"status"`
	R15S   float64 `csv:"r15s"`
	R1M    float64 `csv:"r1m"`
	R15M   float64 `csv:"r15m"`
	UT     string  `csv:"ut"`
	PG     float64 `csv:"pg"`
	LS     float64 `csv:"ls"`
	IT     float64 `csv:"it"`
	TMP    string  `csv:"tmp"`
	SWP    string  `csv:"swp"`
	MEM    string  `csv:"mem"`
}

// 以下是lshosts命令的struct
type lshostsInfo struct {
	HOST_NAME string `csv:"HOST_NAME"`
	HOST_TYPE string `csv:"type"`
	Model     string `csv:"model"`
	Cpuf      string `csv:"cpuf"`
	Ncpus     string `csv:"ncpus"`
	Maxmem    string `csv:"maxmem"`
	Maxswp    string `csv:"maxswp"`
	Server    string `csv:"server"`
	RESOURCES string `csv:"RESOURCES"`
}
