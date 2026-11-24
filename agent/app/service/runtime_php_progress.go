package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/distribution/reference"
)

var (
	phpMetadataLine = regexp.MustCompile(`^#([0-9]+) \[(?:[\w./-]+ )?internal\] load metadata for (\S+)$`)
	phpStepLine     = regexp.MustCompile(`^#([0-9]+) \[ *(?:([\w./-]{1,96}(?: [\w./-]{1,96})?) +)?([0-9]{1,5}/[0-9]{1,5})\] [A-Z]+(?: |$)`)
	phpLegacyStep   = regexp.MustCompile(`^Step ([0-9]+/[0-9]+) : [A-Z]+(?: |$)`)
	phpVertexError  = regexp.MustCompile(`^#([0-9]+) ERROR: `)
	phpRequestURL   = regexp.MustCompile(`(?i)\b(?:head|get|post) "(https?://[^"]+)"`)
	phpKnownCause   = regexp.MustCompile(`(?i)\b(unexpected EOF|EOF|not found|manifest unknown|unauthorized|denied|connection reset by peer|connection refused|no such host|network is unreachable|i/o timeout|TLS handshake timeout|context deadline exceeded|no space left on device|port is already allocated)\s*$`)
)

// One instance belongs to one build. Compose can report different vertices on both streams.
type phpBuildProgress struct {
	task   *task.Task
	detail io.Writer
	phase  string

	mu          sync.Mutex
	seen        map[string]bool
	vertices    map[string]string
	failedStage string
	cause       string
}

func (p *phpBuildProgress) run(phase string, timeout time.Duration, argv ...string) error {
	return p.runWithExtraStdout(phase, timeout, nil, argv...)
}

func (p *phpBuildProgress) runWithStdout(phase string, timeout time.Duration, argv ...string) (string, error) {
	var stdout bytes.Buffer
	err := p.runWithExtraStdout(phase, timeout, &stdout, argv...)
	return stdout.String(), err
}

func (p *phpBuildProgress) runWithExtraStdout(phase string, timeout time.Duration, extraStdout io.Writer, argv ...string) error {
	if p.phase != phase && p.task != nil {
		p.task.Log(i18n.Get(phase))
	}
	p.phase, p.failedStage, p.cause = phase, "", ""
	options := []cmd.Option{cmd.WithTimeout(timeout)}
	if p.task != nil {
		options = append(options, cmd.WithContext(p.task.TaskCtx))
	}
	if phase == "PHPBuildImage" && p.task != nil {
		options = append(options, cmd.WithEnv("BUILDKIT_PROGRESS=plain", "COMPOSE_PROGRESS=plain"))
		p.seen, p.vertices = make(map[string]bool), make(map[string]string)
	}
	stdout, stderr := &phpProgressStream{report: p.observe}, &phpProgressStream{report: p.observe}
	stdoutWriter := io.MultiWriter(p.detail, stdout)
	if extraStdout != nil {
		stdoutWriter = io.MultiWriter(p.detail, stdout, extraStdout)
	}
	err := cmd.NewCommandMgr(options...).RunWithWriters(stdoutWriter, io.MultiWriter(p.detail, stderr), argv[0], argv[1:]...)
	stdout.flush()
	stderr.flush()
	if err == nil {
		return nil
	}
	fmt.Fprintln(p.detail, err)
	stage := p.failedStage
	if stage == "" {
		stage = i18n.Get(phase)
	}
	var exitErr *exec.ExitError
	var businessErr buserr.BusinessError
	switch {
	case p.task != nil && p.task.TaskCtx != nil && errors.Is(p.task.TaskCtx.Err(), context.Canceled):
		return fmt.Errorf("%s: %w", stage, context.Canceled)
	case errors.As(err, &businessErr):
		return fmt.Errorf("%s: %w", stage, err)
	case p.cause != "":
		return fmt.Errorf("%s: %s", stage, p.cause)
	case errors.As(err, &exitErr):
		return fmt.Errorf("%s: %w", stage, exitErr)
	case errors.Is(err, exec.ErrNotFound):
		return fmt.Errorf("%s: %w", stage, exec.ErrNotFound)
	default:
		return errors.New(stage + ": " + i18n.Get("PHPBuildSeeLog"))
	}
}

func (p *phpBuildProgress) observe(line string) {
	line = strings.TrimSpace(line)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.phase == "PHPBuildImage" && p.task != nil {
		if m := phpMetadataLine.FindStringSubmatch(line); m != nil {
			if image, err := reference.ParseNormalizedNamed(m[2]); err == nil {
				p.step(m[1], "PHPBuildBaseImage", reference.FamiliarString(image))
			}
			return
		}
		if m := phpStepLine.FindStringSubmatch(line); m != nil {
			p.step(m[1], "PHPBuildStep", strings.TrimSpace(m[2]+" "+m[3]))
			return
		}
		if m := phpLegacyStep.FindStringSubmatch(line); m != nil {
			p.step("", "PHPBuildStep", m[1])
			return
		}
	}
	vertex := phpVertexError.FindStringSubmatch(line)
	if vertex == nil && !strings.HasPrefix(line, "failed to solve:") && !strings.HasPrefix(line, "ERROR:") && !strings.HasPrefix(line, "Error response from daemon:") {
		return
	}
	// Only publish known diagnostics and the request host. The complete command and URL stay in build.log.
	var stage string
	if p.phase == "PHPBuildImage" && vertex != nil {
		stage = p.vertices[vertex[1]]
		if p.cause == "" && stage != "" {
			p.failedStage = stage
		}
	}
	if cause := phpKnownCause.FindStringSubmatch(line); cause != nil {
		reason := cause[1]
		if m := phpRequestURL.FindStringSubmatch(line); m != nil {
			if endpoint, err := url.Parse(m[1]); err == nil && endpoint.Host != "" && len(endpoint.Host) <= 260 {
				p.cause = endpoint.Host + ": " + reason
				if stage != "" {
					p.failedStage = stage
				}
				return
			}
		}
		if p.cause == "" {
			p.cause = reason
		}
	}
}

func (p *phpBuildProgress) step(vertex, key, value string) {
	message := i18n.GetWithName(key, value)
	if vertex != "" {
		p.vertices[vertex] = message
	}
	if !p.seen[message] {
		p.seen[message] = true
		p.task.Log(message)
	}
}

// Each stream owns its partial line. Oversized lines still go to the detailed log.
type phpProgressStream struct {
	report func(string)
	line   []byte
	skip   bool
}

func (s *phpProgressStream) Write(data []byte) (int, error) {
	n := len(data)
	for len(data) > 0 {
		end := bytes.IndexAny(data, "\r\n")
		chunk := data
		if end >= 0 {
			chunk = data[:end]
		}
		if len(s.line)+len(chunk) > 4096 {
			s.skip = true
		}
		if !s.skip {
			s.line = append(s.line, chunk...)
		}
		if end < 0 {
			break
		}
		s.flush()
		data = data[end+1:]
	}
	return n, nil
}

func (s *phpProgressStream) flush() {
	if !s.skip && len(s.line) > 0 {
		s.report(string(s.line))
	}
	s.line, s.skip = s.line[:0], false
}
