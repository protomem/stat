package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type DiskLoadSnapshot struct {
	Devices []struct {
		Name  string
		TPS   float64
		Read  float64
		Write float64
	}
}

type DiskLoad struct{}

func NewDiskLoad() DiskLoad {
	return DiskLoad{}
}

func (DiskLoad) Collect(ctx context.Context) (DiskLoadSnapshot, error) {
	path, err := exec.LookPath("iostat")
	if err != nil && !errors.Is(err, exec.ErrDot) {
		return DiskLoadSnapshot{}, fmt.Errorf("%w: %w", ErrCommandNotFound, err)
	}

	output := bytes.NewBuffer(make([]byte, 0, 1024*512)) // 512 KiB

	cmd := exec.CommandContext(ctx, path, "-d", "-k", "-o", "JSON")
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return DiskLoadSnapshot{}, fmt.Errorf("%w: %w", ErrCommandExec, err)
	}

	var stats iostatStatistics
	if err := json.NewDecoder(output).Decode(&stats); err != nil {
		return DiskLoadSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidParseResult, err)
	}

	var snapshot DiskLoadSnapshot
	for _, host := range stats.Systat.Hosts {
		for _, stat := range host.Statistics {
			for _, disk := range stat.Disk {
				snapshot.Devices = append(snapshot.Devices, struct {
					Name  string
					TPS   float64
					Read  float64
					Write float64
				}{
					Name:  disk.Device,
					TPS:   disk.TPS,
					Read:  disk.Read,
					Write: disk.Write,
				})
			}
		}
	}

	return snapshot, nil
}

type iostatStatistics struct {
	Systat struct {
		Hosts []struct {
			Statistics []struct {
				Disk []struct {
					Device string  `json:"disk_device"`
					TPS    float64 `json:"tps"`
					Read   float64 `json:"kB_read/s"`
					Write  float64 `json:"kB_wrtn/s"`
				} `json:"disk"`
			} `json:"statistics"`
		} `json:"hosts"`
	} `json:"sysstat"`
}
