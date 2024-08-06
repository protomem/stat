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

type LoadAverageSnapshot struct {
	Values [3]float64 // 1, 5, 15
}

type LoadAverage struct{}

func NewLoadAverage() LoadAverage {
	return LoadAverage{}
}

func (LoadAverage) Collect(ctx context.Context) (LoadAverageSnapshot, error) {
	path, err := exec.LookPath("cat")
	if err != nil && !errors.Is(err, exec.ErrDot) {
		return LoadAverageSnapshot{}, fmt.Errorf("%w: %w", ErrCommandNotFound, err)
	}

	output := bytes.NewBuffer(make([]byte, 0, 30))

	cmd := exec.CommandContext(ctx, path, "/proc/loadavg")
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return LoadAverageSnapshot{}, fmt.Errorf("%w: %w", ErrCommandExec, err)
	}

	var snapshot LoadAverageSnapshot
	parts := strings.Split(output.String(), " ")

	for i := range snapshot.Values {
		snapshot.Values[i], err = strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return LoadAverageSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidParseResult, err)
		}
	}

	return snapshot, nil
}
