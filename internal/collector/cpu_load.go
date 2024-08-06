package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type CPULoadSnapshot struct {
	User   float64
	System float64
	Idle   float64
}

type CPULoad struct{}

func NewCPULoad() CPULoad {
	return CPULoad{}
}

func (CPULoad) Collect(ctx context.Context) (CPULoadSnapshot, error) {
	path, err := exec.LookPath("top")
	if err != nil && !errors.Is(err, exec.ErrDot) {
		return CPULoadSnapshot{}, fmt.Errorf("%w: %w", ErrCommandNotFound, err)
	}

	output := bytes.NewBuffer(make([]byte, 0, 1024*512)) // 512 KiB

	cmd := exec.CommandContext(ctx, path, "-b", "-n", "1")
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return CPULoadSnapshot{}, fmt.Errorf("%w: %w", ErrCommandExec, err)
	}

	var snapshot CPULoadSnapshot
	parts := strings.Split(output.String(), "\n")

	cpuLine := parts[2]
	cpuInfo := strings.Split(strings.TrimLeft(cpuLine, "%Cpu(s): "), ",")

	for _, info := range cpuInfo {
		suffix := info[len(info)-2:]
		switch suffix {
		case "us":
			value, err := strconv.ParseFloat(strings.TrimSpace(info[:len(info)-2]), 64)
			if err != nil {
				return CPULoadSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidParseResult, err)
			}
			snapshot.User = value
		case "sy":
			value, err := strconv.ParseFloat(strings.TrimSpace(info[:len(info)-2]), 64)
			if err != nil {
				return CPULoadSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidParseResult, err)
			}
			snapshot.System = value
		case "id":
			value, err := strconv.ParseFloat(strings.TrimSpace(info[:len(info)-2]), 64)
			if err != nil {
				return CPULoadSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidParseResult, err)
			}
			snapshot.Idle = value
		}
	}

	return snapshot, nil
}
