// Copyright 2017 Mario Trangoni
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

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// The path of the Flexlm binaries.
var lmutilPath = kingpin.Flag("path.lmutil", "FLEXlm `lmutil` path.").Default("E:\\DEV\\flexlm_exporter\\flexnet\\lmutil.exe").String()
