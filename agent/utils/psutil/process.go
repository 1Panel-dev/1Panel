package psutil

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/tklauser/go-sysconf"
)

const defaultClockTicks = 100
const maxProcessCreateTimeSkew = time.Minute

type ProcessCreateTimeResolver struct {
	procRoot   string
	bootTime   int64
	clockTicks uint64
}

func NewProcessCreateTimeResolver() *ProcessCreateTimeResolver {
	procRoot := os.Getenv("HOST_PROC")
	if procRoot == "" {
		procRoot = "/proc"
	}

	resolver := &ProcessCreateTimeResolver{
		procRoot:   procRoot,
		clockTicks: defaultClockTicks,
	}
	if clockTicks, err := sysconf.Sysconf(sysconf.SC_CLK_TCK); err == nil && clockTicks > 0 {
		resolver.clockTicks = uint64(clockTicks)
	}
	if bootTime, err := readBootTime(filepath.Join(procRoot, "stat")); err == nil {
		resolver.bootTime = bootTime
	}
	return resolver
}

// CreateTime returns the process start time in milliseconds since Unix epoch.
//
// On some LXC systems, gopsutil combines a container-relative /proc/uptime with
// the host-relative starttime from /proc/<pid>/stat, which can place the start
// time in the future. In that case, reading btime and starttime from /proc keeps
// both values on the same clock base, matching the calculation used by ps.
func (r *ProcessCreateTimeResolver) CreateTime(proc *process.Process) (int64, error) {
	now := time.Now()
	createTime, createTimeErr := proc.CreateTime()
	if createTimeErr == nil && isValidProcessCreateTime(createTime, now) {
		return createTime, nil
	}

	if r.bootTime > 0 && r.clockTicks > 0 {
		statPath := filepath.Join(r.procRoot, strconv.Itoa(int(proc.Pid)), "stat")
		if content, err := os.ReadFile(statPath); err == nil {
			if startTicks, err := parseProcessStartTicks(string(content)); err == nil {
				startMillis := startTicks * 1000 / r.clockTicks
				fallbackTime := r.bootTime*1000 + int64(startMillis)
				if isValidProcessCreateTime(fallbackTime, now) {
					return fallbackTime, nil
				}
			}
		}
	}
	if createTimeErr != nil {
		return 0, fmt.Errorf("resolve process create time: %w", createTimeErr)
	}
	return 0, fmt.Errorf("invalid process create time: %d", createTime)
}

func isValidProcessCreateTime(createTime int64, now time.Time) bool {
	return createTime > 0 && createTime <= now.Add(maxProcessCreateTimeSkew).UnixMilli()
}

func readBootTime(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || fields[0] != "btime" {
			continue
		}
		bootTime, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse btime: %w", err)
		}
		if bootTime <= 0 {
			return 0, errors.New("invalid btime")
		}
		return bootTime, nil
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, errors.New("btime not found")
}

func parseProcessStartTicks(stat string) (uint64, error) {
	// The second field (comm) is wrapped in parentheses and may contain spaces
	// or parentheses, so split only after its closing parenthesis. The remaining
	// fields begin at field 3 (state), making starttime (field 22) index 19.
	commEnd := strings.LastIndex(stat, ")")
	if commEnd == -1 || commEnd+1 >= len(stat) {
		return 0, errors.New("invalid process stat")
	}
	fields := strings.Fields(stat[commEnd+1:])
	const startTimeIndex = 19
	if len(fields) <= startTimeIndex {
		return 0, errors.New("process stat has too few fields")
	}
	return strconv.ParseUint(fields[startTimeIndex], 10, 64)
}
