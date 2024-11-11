// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package standalone

import (
	"context"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/sjy-dv/nnv/gen/protoc/v1/resourceCoordinatorV1"
)

// check profiling, auto-commit
// data_access_layer needs to implement auto-recovery
func profiling() {

}

func systeminfo(ctx context.Context) (*resourceCoordinatorV1.SystemInfo, error) {
	hInfoStat, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	vMemStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	loadAvgStat, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &resourceCoordinatorV1.SystemInfo{
		Uptime:         hInfoStat.Uptime,
		CpuLoad1:       loadAvgStat.Load1,
		CpuLoad5:       loadAvgStat.Load5,
		CpuLoad15:      loadAvgStat.Load15,
		MemTotal:       vMemStat.Total,
		MemAvailable:   vMemStat.Available,
		MemUsed:        vMemStat.Used,
		MemFree:        vMemStat.Free,
		MemUsedPercent: vMemStat.UsedPercent,
	}, nil
}
