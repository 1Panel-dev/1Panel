package convert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/global"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type FormatOption struct {
	Type  string
	Codec string
}

var FormatMap = map[string]FormatOption{
	// images
	"png":  {Type: "image", Codec: "png"},
	"jpg":  {Type: "image", Codec: "mjpeg"},
	"jpeg": {Type: "image", Codec: "mjpeg"},
	"webp": {Type: "image", Codec: "libwebp"},
	"gif":  {Type: "image", Codec: "gif"},
	"bmp":  {Type: "image", Codec: "bmp"},
	"tiff": {Type: "image", Codec: "tiff"},

	// videos
	"mp4": {Type: "video", Codec: "libx264"},
	"avi": {Type: "video", Codec: "libx264"},
	"mov": {Type: "video", Codec: "libx264"},
	"mkv": {Type: "video", Codec: "libx264"},

	// audios
	"mp3":  {Type: "audio", Codec: "libmp3lame"},
	"wav":  {Type: "audio", Codec: "pcm_s16le"},
	"flac": {Type: "audio", Codec: "flac"},
	"aac":  {Type: "audio", Codec: "aac"},
}

func hasFfmpeg() (string, bool) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	return ffmpegPath, err == nil
}

func MediaFile(inputFile, outputFile, outputFormat string, deleteSource bool) (state string, err error) {
	status := "FAILED"
	msg := ""
	ffmpegPath, flag := hasFfmpeg()
	if !flag {
		return status, fmt.Errorf("ffmpeg not found, cannot convert file")
	}
	logFile, logErr := os.OpenFile(filepath.Join(global.Dir.ConvertLogDir, "convert.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	allLogFile, allErr := os.OpenFile(filepath.Join(global.Dir.ConvertLogDir, "convert-all.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if logErr != nil || allErr != nil {
		return status, fmt.Errorf("cannot open log file: %w", err)
	}
	defer logFile.Close()
	args, fileType, err := buildFFmpegArgs(inputFile, outputFile, outputFormat)
	if err != nil {
		return status, fmt.Errorf("FFmpeg args failed: %w", err)
	}
	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmdr := exec.CommandContext(ctx, ffmpegPath, args...)
	cmdr.Env = append(os.Environ(),
		"PATH="+filepath.Dir(ffmpegPath)+":"+os.Getenv("PATH"),
		"LD_LIBRARY_PATH=/usr/local/lib:"+os.Getenv("LD_LIBRARY_PATH"),
	)
	var buf bytes.Buffer
	cmdr.Stdout = &buf
	cmdr.Stderr = &buf
	err = cmdr.Run()
	logStr := buf.String()

	stat, statErr := os.Stat(outputFile)
	if err != nil || statErr != nil || stat.Size() == 0 {
		status = "FAILED"
		msg = extractFFmpegError(logStr)
		_ = os.Remove(outputFile)
	} else {
		status = "SUCCESS"
		msg = "SUCCESS"
	}

	entry := response.FileConvertLog{
		Date:    time.Now().Format("2006-01-02 15:04:05"),
		Type:    fileType,
		Log:     fmt.Sprintf("%s -> %s", inputFile, outputFile),
		Status:  status,
		Message: msg,
	}
	_ = appendJSONLog(logFile, entry)

	allLogEntry := fmt.Sprintf("[%s] %s %s -> %s [%s]: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		fileType, inputFile, outputFile, status, logStr)
	_ = appendLog(allLogFile, allLogEntry)
	if err == nil && deleteSource {
		_ = os.Remove(inputFile)
	}
	return status, nil
}

func buildFFmpegArgs(inputFile, outputFile, outputFormat string) ([]string, string, error) {
	args := []string{"-y", "-i", inputFile}
	opt, ok := FormatMap[outputFormat]
	if !ok {
		return nil, "", fmt.Errorf("unsupported format: %s", outputFormat)
	}

	switch opt.Type {
	case "image":
		switch outputFormat {
		case "webp":
			args = append(args, "-c:v", "libwebp", "-lossless", "0", "-q:v", "75")
		case "png", "gif", "jpg", "jpeg", "bmp", "tiff":
			args = append(args, "-c:v", opt.Codec)
		}

	case "video":
		args = append(args, "-c:v", opt.Codec, "-preset", "fast", "-crf", "23", "-c:a", "aac", "-b:a", "192k")

	case "audio":
		args = append(args, "-c:a", opt.Codec, "-b:a", "192k")

	default:
		return nil, opt.Type, fmt.Errorf("unsupported media type: %s", opt.Type)
	}

	args = append(args, outputFile)
	return args, opt.Type, nil
}

func appendLog(f *os.File, content string) error {
	_, err := f.WriteString(content)
	return err
}

func appendJSONLog(f *os.File, entry response.FileConvertLog) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return err
	}
	return nil
}

func extractFFmpegError(logStr string) string {
	priority := []string{"Error", "Invalid", "failed", "No "}
	matches := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(logStr), "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		for _, kw := range priority {
			if _, ok := matches[kw]; !ok && strings.Contains(line, kw) {
				matches[kw] = line
			}
		}
	}

	for _, kw := range priority {
		if line, ok := matches[kw]; ok {
			return line
		}
	}

	if len(lines) > 0 {
		return lines[len(lines)-1]
	}
	return ""
}
