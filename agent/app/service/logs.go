package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type LogService struct{}

const maxSystemLogCursorSkip = 100
const maxFileSystemLogLines = 1000

var (
	syslogPriorityKeywords = []struct {
		keyword  string
		priority string
	}{
		{"emerg", "0"}, {"panic", "0"}, {"alert", "1"}, {"crit", "2"}, {"fatal", "2"},
		{"error", "3"}, {"err", "3"}, {"warn", "4"}, {"notice", "5"}, {"debug", "7"}, {"info", "6"},
	}
)

type systemLogCursor struct {
	EndTime int64 `json:"endTime"`
	Skipped int   `json:"skipped"`
}

type ILogService interface {
	ListSystemLogFile() ([]string, error)
	ReadSystemLog(req dto.SystemLogReq) (dto.SystemLogRes, error)
	ListRunningServices() ([]string, error)
}

func (u *LogService) ReadSystemLog(req dto.SystemLogReq) (dto.SystemLogRes, error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 100
	}
	now := time.Now()
	startTime := req.StartTime
	if startTime.IsZero() {
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	endTime := req.EndTime
	if endTime.IsZero() {
		endTime = now
	}
	cursor, err := decodeSystemLogCursor(req.Cursor)
	if err != nil {
		return dto.SystemLogRes{}, err
	}
	if cursor != nil {
		endTime = time.Unix(0, cursor.EndTime*int64(time.Microsecond))
	}
	journalctl, err := exec.LookPath("journalctl")
	if err != nil {
		return u.readFileSystemLog(req, startTime, endTime, pageSize, cursor)
	}
	args := []string{
		"--no-pager", "--reverse", "--output=json",
		"--since", formatJournalQueryTime(startTime),
		"--until", formatJournalQueryTime(endTime),
	}
	if service := strings.TrimSpace(req.Service); service != "" {
		args = append(args, "-u", service)
	}
	if priority := strings.TrimSpace(req.Priority); priority != "" {
		args = append(args, "--priority", priority)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		args = append(args, "--grep", keyword)
	}
	queryLines := pageSize + 1
	if cursor != nil {
		queryLines += cursor.Skipped
	}
	queryArgs := append(args, "--lines="+strconv.Itoa(queryLines))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	output, err := exec.CommandContext(ctx, journalctl, queryArgs...).CombinedOutput()
	cancel()
	if err != nil {
		return dto.SystemLogRes{}, fmt.Errorf("read host system logs failed: %s", strings.TrimSpace(string(output)))
	}
	content := strings.TrimSpace(string(output))
	if content == "" || strings.HasPrefix(content, "-- No entries --") {
		return dto.SystemLogRes{Source: "journalctl", Items: []dto.SystemLogItem{}}, nil
	}
	items := parseJournalLogItems(content)
	if cursor != nil {
		items, err = skipSystemLogCursorItems(items, *cursor)
		if err != nil {
			return dto.SystemLogRes{}, err
		}
	}
	return buildSystemLogResponse("journalctl", items, pageSize, cursor)
}

func (u *LogService) readFileSystemLog(req dto.SystemLogReq, startTime, endTime time.Time, pageSize int, cursor *systemLogCursor) (dto.SystemLogRes, error) {
	items := make([]dto.SystemLogItem, 0)
	for _, logFile := range []string{"/var/log/syslog", "/var/log/messages", "/var/log/system.log"} {
		if _, err := os.Stat(logFile); err != nil {
			continue
		}
		content, err := readLastSystemLogLines(logFile, maxFileSystemLogLines)
		if err != nil {
			return dto.SystemLogRes{}, err
		}
		for _, line := range strings.Split(content, "\n") {
			item, ok := parseFileSystemLogItem(line)
			if !ok || !matchSystemLogItem(item, req, startTime, endTime) {
				continue
			}
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Timestamp > items[j].Timestamp })
	if cursor != nil {
		var err error
		items, err = skipSystemLogCursorItems(items, *cursor)
		if err != nil {
			return dto.SystemLogRes{}, err
		}
	}
	return buildSystemLogResponse("file", items, pageSize, cursor)
}

func readLastSystemLogLines(logFile string, lines int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "tail", "-n", strconv.Itoa(lines), logFile).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func buildSystemLogResponse(source string, items []dto.SystemLogItem, pageSize int, cursor *systemLogCursor) (dto.SystemLogRes, error) {
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	res := dto.SystemLogRes{Source: source, Items: items, HasMore: hasMore}
	if !hasMore || len(items) == 0 {
		return res, nil
	}
	lastItem := items[len(items)-1]
	skipped := countSystemLogTimestamp(items, lastItem.Timestamp)
	if cursor != nil && cursor.EndTime == lastItem.Timestamp {
		skipped += cursor.Skipped
	}
	nextCursor, err := encodeSystemLogCursor(systemLogCursor{EndTime: lastItem.Timestamp, Skipped: skipped})
	if err != nil {
		return dto.SystemLogRes{}, err
	}
	res.NextCursor = nextCursor
	return res, nil
}

func decodeSystemLogCursor(value string) (*systemLogCursor, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid system log cursor")
	}
	var cursor systemLogCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.EndTime <= 0 || cursor.Skipped < 0 || cursor.Skipped > maxSystemLogCursorSkip {
		return nil, fmt.Errorf("invalid system log cursor")
	}
	return &cursor, nil
}

func encodeSystemLogCursor(cursor systemLogCursor) (string, error) {
	value, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func skipSystemLogCursorItems(items []dto.SystemLogItem, cursor systemLogCursor) ([]dto.SystemLogItem, error) {
	skipped := 0
	for len(items) > 0 && skipped < cursor.Skipped && items[0].Timestamp == cursor.EndTime {
		items = items[1:]
		skipped++
	}
	if skipped != cursor.Skipped {
		return nil, fmt.Errorf("system log cursor has expired")
	}
	return items, nil
}

func countSystemLogTimestamp(items []dto.SystemLogItem, timestamp int64) int {
	count := 0
	for _, item := range items {
		if item.Timestamp == timestamp {
			count++
		}
	}
	return count
}

func formatJournalQueryTime(value time.Time) string {
	return value.In(time.Local).Format("2006-01-02 15:04:05.000000")
}

func parseFileSystemLogItem(line string) (dto.SystemLogItem, bool) {
	raw := strings.TrimSpace(line)
	if raw == "" {
		return dto.SystemLogItem{}, false
	}
	priority := "6"
	if matches := re.GetRegex(re.SyslogPRIPattern).FindStringSubmatch(raw); len(matches) == 2 {
		value, _ := strconv.Atoi(matches[1])
		priority = strconv.Itoa(value % 8)
		raw = strings.TrimSpace(strings.TrimPrefix(raw, matches[0]))
	}
	logTime, rest, ok := parseFileSystemLogTime(raw)
	if !ok {
		return dto.SystemLogItem{}, false
	}
	service, message := parseFileSystemLogService(rest)
	if priority == "6" {
		priority = detectFileSystemLogPriority(message)
	}
	return dto.SystemLogItem{
		Timestamp: logTime.UnixMicro(),
		Time:      logTime.Format("2006-01-02 15:04:05"),
		Priority:  priority,
		Service:   service,
		Message:   message,
		Raw:       line,
	}, true
}

func parseFileSystemLogTime(value string) (time.Time, string, bool) {
	if matches := re.GetRegex(re.SyslogRFC3339Pattern).FindStringSubmatch(value); len(matches) == 3 {
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999Z07:00"} {
			if parsed, err := time.Parse(layout, matches[1]); err == nil {
				return parsed.Local(), matches[2], true
			}
		}
		for _, layout := range []string{"2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05.999999999"} {
			if parsed, err := time.ParseInLocation(layout, matches[1], time.Local); err == nil {
				return parsed, matches[2], true
			}
		}
	}
	if matches := re.GetRegex(re.SyslogRFC3164Pattern).FindStringSubmatch(value); len(matches) == 3 {
		parsed, err := time.ParseInLocation("Jan _2 15:04:05", matches[1], time.Local)
		if err != nil {
			return time.Time{}, "", false
		}
		now := time.Now()
		parsed = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.Local)
		if parsed.After(now.Add(24 * time.Hour)) {
			parsed = parsed.AddDate(-1, 0, 0)
		}
		return parsed, matches[2], true
	}
	return time.Time{}, "", false
}

func parseFileSystemLogService(value string) (string, string) {
	if matches := re.GetRegex(re.SyslogServicePattern).FindStringSubmatch(value); len(matches) == 3 {
		return matches[1], matches[2]
	}
	// RFC5424 records have the form HOST APP-NAME PROCID MSGID STRUCTURED-DATA MSG.
	fields := strings.Fields(value)
	if len(fields) >= 6 {
		return fields[1], strings.Join(fields[5:], " ")
	}
	return "", value
}

func detectFileSystemLogPriority(message string) string {
	lowerMessage := strings.ToLower(message)
	for _, item := range syslogPriorityKeywords {
		if strings.Contains(lowerMessage, item.keyword) {
			return item.priority
		}
	}
	return "6"
}

func matchSystemLogItem(item dto.SystemLogItem, req dto.SystemLogReq, startTime, endTime time.Time) bool {
	itemTime := time.UnixMicro(item.Timestamp)
	if itemTime.Before(startTime) || itemTime.After(endTime) {
		return false
	}
	if priority := strings.TrimSpace(req.Priority); priority != "" && item.Priority != priority {
		return false
	}
	if service := strings.TrimSpace(req.Service); service != "" && !matchFileSystemLogService(item.Service, service) {
		return false
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" && !strings.Contains(strings.ToLower(item.Raw), strings.ToLower(keyword)) {
		return false
	}
	return true
}

func matchFileSystemLogService(itemService, requestedService string) bool {
	itemService = strings.TrimSpace(itemService)
	requestedService = strings.TrimSpace(requestedService)
	if itemService == "" || requestedService == "" {
		return itemService == requestedService
	}
	return strings.EqualFold(strings.TrimSuffix(itemService, ".service"), strings.TrimSuffix(requestedService, ".service"))
}

func (u *LogService) ListRunningServices() ([]string, error) {
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		return []string{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	args := []string{"list-units", "--type=service", "--state=running", "--no-legend", "--no-pager", "--plain"}
	output, err := exec.CommandContext(ctx, systemctl, args...).Output()
	if err != nil {
		return []string{}, nil
	}

	services := make([]string, 0)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasSuffix(fields[0], ".service") {
			continue
		}
		services = append(services, fields[0])
	}
	sort.Strings(services)
	return services, nil
}

func parseJournalLogItems(content string) []dto.SystemLogItem {
	entries := strings.Split(content, "\n")
	items := make([]dto.SystemLogItem, 0, len(entries))
	for _, entry := range entries {
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(entry), &fields); err != nil {
			continue
		}
		items = append(items, dto.SystemLogItem{
			Timestamp: journalFieldMicroseconds(fields),
			Time:      formatJournalTimestamp(journalFieldString(fields, "__REALTIME_TIMESTAMP")),
			Priority:  journalFieldString(fields, "PRIORITY"),
			Service:   journalFieldString(fields, "_SYSTEMD_UNIT"),
			Message:   journalFieldString(fields, "MESSAGE"),
			Raw:       entry,
		})
	}
	return items
}

func journalFieldMicroseconds(fields map[string]interface{}) int64 {
	value, _ := strconv.ParseInt(journalFieldString(fields, "__REALTIME_TIMESTAMP"), 10, 64)
	return value
}

func journalFieldString(fields map[string]interface{}, key string) string {
	value, ok := fields[key]
	if !ok || value == nil {
		return ""
	}
	switch item := value.(type) {
	case string:
		return item
	case float64:
		return strconv.FormatInt(int64(item), 10)
	case []interface{}:
		values := make([]string, 0, len(item))
		for _, value := range item {
			values = append(values, fmt.Sprint(value))
		}
		return strings.Join(values, " ")
	default:
		return fmt.Sprint(item)
	}
}

func formatJournalTimestamp(value string) string {
	microseconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return value
	}
	return time.Unix(0, microseconds*int64(time.Microsecond)).Local().Format("2006-01-02 15:04:05")
}

func NewILogService() ILogService {
	return &LogService{}
}

func (u *LogService) ListSystemLogFile() ([]string, error) {
	var listFile []string
	files, err := os.ReadDir(global.Dir.LogDir)
	if err != nil {
		return nil, err
	}
	listMap := make(map[string]struct{})
	for _, item := range files {
		if item.IsDir() || !strings.HasPrefix(item.Name(), "1Panel") {
			continue
		}
		if item.Name() == "1Panel.log" || item.Name() == "1Panel-Core.log" {
			itemName := time.Now().Format("2006-01-02")
			if _, ok := listMap[itemName]; ok {
				continue
			}
			listMap[itemName] = struct{}{}
			listFile = append(listFile, itemName)
			continue
		}
		itemFileName := strings.TrimPrefix(item.Name(), "1Panel-Core-")
		itemFileName = strings.TrimPrefix(itemFileName, "1Panel-")
		itemFileName = strings.TrimSuffix(itemFileName, ".gz")
		itemFileName = strings.TrimSuffix(itemFileName, ".log")
		if len(itemFileName) == 0 {
			continue
		}
		if _, ok := listMap[itemFileName]; ok {
			continue
		}
		listMap[itemFileName] = struct{}{}
		listFile = append(listFile, itemFileName)
	}
	if len(listFile) < 2 {
		return listFile, nil
	}
	sort.Slice(listFile, func(i, j int) bool {
		return listFile[i] > listFile[j]
	})

	return listFile, nil
}
