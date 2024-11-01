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
