package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	stdpprof "runtime/pprof"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	profile "github.com/google/pprof/profile"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	diagnosticsDefaultDuration    = 15
	diagnosticsMaxProfileSize     = 64 * 1024 * 1024
	diagnosticsMaxSnapshotSize    = 16 * 1024 * 1024
	diagnosticsDetailedGoroutines = 10_000
	diagnosticsBlockProfileRate   = 1_000_000
	diagnosticsMaxGroups          = 200
)

var (
	runtimeDiagnosticsInstance = &RuntimeDiagnosticsService{}
	goroutineHeaderPattern     = regexp.MustCompile(`^goroutine \d+ \[([^]]+)\]:$`)
	goroutineArgPattern        = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	errProfileSizeLimit        = errors.New("runtime profile exceeds the 64 MiB size limit")
)

type cappedWriter struct {
	writer    io.Writer
	remaining int64
	exceeded  bool
}

func (w *cappedWriter) Write(data []byte) (int, error) {
	if int64(len(data)) <= w.remaining {
		n, err := w.writer.Write(data)
		w.remaining -= int64(n)
		return n, err
	}
	w.exceeded = true
	if w.remaining <= 0 {
		return 0, errProfileSizeLimit
	}
	allowed := int(w.remaining)
	n, err := w.writer.Write(data[:allowed])
	w.remaining -= int64(n)
	if err != nil {
		return n, err
	}
	return n, errProfileSizeLimit
}

type IRuntimeDiagnosticsService interface {
	Summary() (dto.RuntimeDiagnosticsSummary, error)
	Goroutines() (dto.RuntimeGoroutineSnapshot, error)
	CreateProfile(req dto.RuntimeProfileCreate) (RuntimeProfileResult, error)
}

type RuntimeProfileResult struct {
	Path string
	Name string
}

type RuntimeDiagnosticsService struct {
	captureMu sync.Mutex
	processMu sync.Mutex
	process   *process.Process
}

func NewIRuntimeDiagnosticsService() IRuntimeDiagnosticsService {
	return runtimeDiagnosticsInstance
}

func (s *RuntimeDiagnosticsService) Summary() (dto.RuntimeDiagnosticsSummary, error) {
	rss, err := s.processRSS()
	if err != nil {
		return dto.RuntimeDiagnosticsSummary{}, err
	}
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return dto.RuntimeDiagnosticsSummary{
		RSS:         rss,
		HeapAlloc:   stats.HeapAlloc,
		HeapObjects: stats.HeapObjects,
		Goroutines:  runtime.NumGoroutine(),
	}, nil
}

func (s *RuntimeDiagnosticsService) Goroutines() (dto.RuntimeGoroutineSnapshot, error) {
	total := runtime.NumGoroutine()
	if total > diagnosticsDetailedGoroutines {
		groups, truncated := compactGoroutineSnapshot(total)
		return dto.RuntimeGoroutineSnapshot{
			Total: total, GroupCount: len(groups), Truncated: truncated, CapturedAt: time.Now(), Goroutines: groups,
		}, nil
	}
	var data bytes.Buffer
	writer := &cappedWriter{writer: &data, remaining: diagnosticsMaxSnapshotSize}
	if err := stdpprof.Lookup("goroutine").WriteTo(writer, 2); err != nil {
		groups, truncated := compactGoroutineSnapshot(total)
		return dto.RuntimeGoroutineSnapshot{
			Total: total, GroupCount: len(groups), Truncated: truncated, CapturedAt: time.Now(), Goroutines: groups,
		}, nil
	}
	groups, truncated := parseGoroutineDump(&data, diagnosticsMaxGroups)
	result := dto.RuntimeGoroutineSnapshot{
		Total:      total,
		GroupCount: len(groups),
		Truncated:  truncated,
		CapturedAt: time.Now(),
	}
	result.Goroutines = groups
	return result, nil
}

func (s *RuntimeDiagnosticsService) CreateProfile(req dto.RuntimeProfileCreate) (RuntimeProfileResult, error) {
	if !s.captureMu.TryLock() {
		return RuntimeProfileResult{}, errors.New("another runtime profile is being captured")
	}
	defer s.captureMu.Unlock()
	return captureRuntimeProfile(req)
}

func captureRuntimeProfile(req dto.RuntimeProfileCreate) (result RuntimeProfileResult, err error) {
	duration := req.Duration
	if duration == 0 {
		duration = diagnosticsDefaultDuration
	}
	if duration < 5 || duration > 30 {
		return result, errors.New("profile duration must be between 5 and 30 seconds")
	}
	if req.Type == "heap" || req.Type == "goroutine" {
		duration = 0
	}

	name := fmt.Sprintf("%s-%s-%ds.pb.gz", req.Type, newRuntimeEventID(), duration)
	file, err := os.CreateTemp("", "1panel-runtime-profile-*.tmp")
	if err != nil {
		return result, err
	}
	writer := &cappedWriter{writer: file, remaining: diagnosticsMaxProfileSize}
	removeOnError := true
	defer func() {
		_ = file.Close()
		if removeOnError {
			_ = os.Remove(file.Name())
		}
	}()

	switch req.Type {
	case "cpu":
		if err = stdpprof.StartCPUProfile(writer); err != nil {
			return result, err
		}
		time.Sleep(time.Duration(duration) * time.Second)
		stdpprof.StopCPUProfile()
	case "heap":
		err = stdpprof.Lookup("heap").WriteTo(writer, 0)
	case "goroutine":
		err = stdpprof.Lookup("goroutine").WriteTo(writer, 0)
	case "mutex":
		err = captureWindowedRuntimeProfile("mutex", duration, writer, func() func() {
			previous := runtime.SetMutexProfileFraction(5)
			return func() { runtime.SetMutexProfileFraction(previous) }
		})
	case "block":
		err = captureWindowedRuntimeProfile("block", duration, writer, func() func() {
			runtime.SetBlockProfileRate(diagnosticsBlockProfileRate)
			return func() { runtime.SetBlockProfileRate(0) }
		})
	default:
		return result, errors.New("unsupported runtime profile type")
	}
	if err != nil {
		return result, err
	}
	if writer.exceeded {
		return result, errProfileSizeLimit
	}
	if err = file.Close(); err != nil {
		return result, err
	}
	removeOnError = false
	return RuntimeProfileResult{Path: file.Name(), Name: name}, nil
}

func captureWindowedRuntimeProfile(name string, duration int, writer io.Writer, enable func() func()) error {
	restore := enable()
	sampling := true
	defer func() {
		if sampling {
			restore()
		}
	}()

	before, err := readRuntimeProfile(name)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	time.Sleep(time.Duration(duration) * time.Second)
	restore()
	sampling = false
	after, err := readRuntimeProfile(name)
	if err != nil {
		return err
	}
	delta, err := diffRuntimeProfiles(before, after, startedAt, time.Duration(duration)*time.Second)
	if err != nil {
		return err
	}
	return delta.Write(writer)
}

func readRuntimeProfile(name string) (*profile.Profile, error) {
	var data bytes.Buffer
	writer := &cappedWriter{writer: &data, remaining: diagnosticsMaxProfileSize}
	if err := stdpprof.Lookup(name).WriteTo(writer, 0); err != nil {
		return nil, err
	}
	if writer.exceeded {
		return nil, errProfileSizeLimit
	}
	return profile.Parse(&data)
}

func diffRuntimeProfiles(before, after *profile.Profile, startedAt time.Time, duration time.Duration) (*profile.Profile, error) {
	baseline := before.Copy()
	baseline.Scale(-1)
	delta, err := profile.Merge([]*profile.Profile{after, baseline})
	if err != nil {
		return nil, err
	}
	delta.TimeNanos = startedAt.UnixNano()
	delta.DurationNanos = duration.Nanoseconds()
	return delta, nil
}

func (s *RuntimeDiagnosticsService) processRSS() (uint64, error) {
	s.processMu.Lock()
	defer s.processMu.Unlock()
	if s.process == nil {
		proc, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			return 0, err
		}
		s.process = proc
	}
	memoryInfo, err := s.process.MemoryInfo()
	if err != nil {
		return 0, err
	}
	if memoryInfo == nil {
		return 0, errors.New("process memory information is unavailable")
	}
	return memoryInfo.RSS, nil
}

func newRuntimeEventID() string {
	return time.Now().Format("20060102-150405.000000000")
}

func parseGoroutineDump(reader io.Reader, maxGroups int) ([]dto.RuntimeGoroutineGroup, bool) {
	type groupValue struct {
		state string
		top   string
		stack []string
		count int
	}
	groups := make(map[string]*groupValue)
	truncated := false
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var block []string
	flush := func() {
		if len(block) == 0 {
			return
		}
		matches := goroutineHeaderPattern.FindStringSubmatch(block[0])
		state := "unknown"
		if len(matches) == 2 {
			state = matches[1]
		}
		stack := append([]string(nil), block[1:]...)
		if len(stack) > 40 {
			stack = stack[:40]
		}
		top := "runtime"
		functionLines := make([]string, 0, len(stack)/2)
		for _, line := range stack {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "created by ") {
				continue
			}
			normalized := goroutineArgPattern.ReplaceAllString(trimmed, "0x…")
			functionLines = append(functionLines, normalized)
			if top == "runtime" {
				top = strings.Split(normalized, "(")[0]
			}
		}
		signature := state + "\n" + strings.Join(functionLines, "\n")
		if existing, ok := groups[signature]; ok {
			existing.count++
			return
		}
		if len(groups) >= maxGroups {
			truncated = true
			return
		}
		groups[signature] = &groupValue{state: state, top: top, stack: stack, count: 1}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "goroutine ") && strings.HasSuffix(line, "]:") {
			flush()
			block = []string{line}
			continue
		}
		if len(block) > 0 {
			block = append(block, line)
		}
	}
	if scanner.Err() != nil {
		truncated = true
	}
	flush()

	result := make([]dto.RuntimeGoroutineGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, dto.RuntimeGoroutineGroup{State: group.state, Top: group.top, Count: group.count, Stack: group.stack})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Top < result[j].Top
		}
		return result[i].Count > result[j].Count
	})
	return result, truncated
}

func compactGoroutineSnapshot(initialSize int) ([]dto.RuntimeGoroutineGroup, bool) {
	records := make([]runtime.StackRecord, initialSize+32)
	count, ok := runtime.GoroutineProfile(records)
	if !ok {
		records = make([]runtime.StackRecord, count+32)
		count, ok = runtime.GoroutineProfile(records)
	}
	truncated := !ok
	if count > len(records) {
		count = len(records)
		truncated = true
	}
	records = records[:count]
	type compactGroup struct {
		top   string
		stack []string
		count int
	}
	groups := make(map[string]*compactGroup)
	for _, record := range records {
		frames := runtime.CallersFrames(record.Stack())
		stack := make([]string, 0, 16)
		functions := make([]string, 0, 8)
		top := "runtime"
		for {
			frame, more := frames.Next()
			if frame.Function != "" {
				if top == "runtime" {
					top = frame.Function
				}
				functions = append(functions, frame.Function)
				stack = append(stack, frame.Function, fmt.Sprintf("\t%s:%d", frame.File, frame.Line))
			}
			if !more || len(functions) >= 20 {
				break
			}
		}
		signature := strings.Join(functions, "\n")
		if existing, exists := groups[signature]; exists {
			existing.count++
			continue
		}
		if len(groups) >= diagnosticsMaxGroups {
			truncated = true
			continue
		}
		groups[signature] = &compactGroup{top: top, stack: stack, count: 1}
	}
	result := make([]dto.RuntimeGoroutineGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, dto.RuntimeGoroutineGroup{State: "profiled", Top: group.top, Count: group.count, Stack: group.stack})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Top < result[j].Top
		}
		return result[i].Count > result[j].Count
	})
	return result, truncated
}
