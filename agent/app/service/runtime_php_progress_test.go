package service

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/sirupsen/logrus"
)

func setupPHPProgressTest(t *testing.T) {
	t.Helper()
	db := appCacheTestDB(t)
	if err := db.AutoMigrate(&model.Task{}, &model.Runtime{}); err != nil {
		t.Fatal(err)
	}
	oldTaskDB, oldLang := global.TaskDB, global.I18n
	t.Cleanup(func() { global.TaskDB, global.I18n = oldTaskDB, oldLang })
	global.TaskDB = db
	global.Dir.RuntimeDir, global.Dir.TaskDir = t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(global.Dir.RuntimeDir, constant.RuntimePHP), 0755); err != nil {
		t.Fatal(err)
	}
	i18n.Init()
	if err := ConfigurePanelMCPTasks(); err != nil {
		t.Fatal(err)
	}
}

func TestPHPCreationTaskProgress(t *testing.T) {
	setupPHPProgressTest(t)
	template := t.TempDir()
	appCacheTestWrite(t, filepath.Join(template, "docker-compose.yml"), "services:\n  php:\n    image: php-test\n")
	appCacheTestWrite(t, filepath.Join(template, "data.yml"), "{}")
	bin := t.TempDir()
	// The test releases the process only after reading its progress through the task log service.
	fakeDocker := `#!/bin/sh
set -eu
case " $* " in
  *" build ")
    echo '#3 [internal] load metadata for docker.io/1panel/php:8.5.13-fpm'
    echo '#6 [2/5] RUN private-build-command' >&2
    echo 'package-output-only-in-build.log' >&2
    while [ ! -f "$PHP_TEST_DIR/release" ]; do sleep 0.02; done
    if [ "$PHP_TEST_FAILURE" = manifest ]; then
      echo '#3 ERROR: failed to do request: Head "https://user:password@mirror.example/v2/manifest?token=private": EOF' >&2
      exit 1
    fi ;;
  *" install-ext "*)
    echo "extension-output"
    if [ "$PHP_TEST_FAILURE" = extension ]; then exit 7; fi ;;
  *" commit "*) echo "commit-output" ;;
  *" up -d ")
    if [ "$PHP_TEST_FAILURE" = start ] && [ -f "$PHP_TEST_DIR/started" ]; then exit 8; fi
    touch "$PHP_TEST_DIR/started" ;;
esac
`
	for _, name := range []string{"docker", "docker-compose"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(fakeDocker), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := global.DB.Create(&model.Setting{Key: "DockerSockPath", Value: "unix://" + filepath.Join(bin, "unused.sock")}).Error; err != nil {
		t.Fatal(err)
	}
	oldCompose := global.CONF.DockerConfig.Command
	t.Cleanup(func() { global.CONF.DockerConfig.Command = oldCompose })
	for _, tc := range []struct{ name, extensions, failure, result string }{
		{"plain", "", "", ""},
		{"extensions", "redis", "", ""},
		{"manifest", "", "manifest", "mirror.example: EOF"},
		{"extension-error", "redis", "extension", "exit status 7"},
		{"final-start-error", "", "start", "exit status 8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			global.CONF.DockerConfig.Command = "docker compose"
			if tc.name == "extensions" {
				global.CONF.DockerConfig.Command = "docker-compose"
			}
			dir := t.TempDir()
			t.Setenv("PHP_TEST_DIR", dir)
			t.Setenv("PHP_TEST_FAILURE", tc.failure)
			req := request.RuntimeCreate{Name: tc.name, TaskID: tc.name, Type: constant.RuntimePHP, Resource: constant.ResourceAppstore,
				Params: map[string]interface{}{"PHP_EXTENSIONS": tc.extensions, "PHP_VERSION": "8.5.13"}}
			rt := &model.Runtime{Name: tc.name, Type: req.Type, Resource: req.Resource, Image: "php-test", ContainerName: "php-test"}
			createTask, err := task.NewTaskWithOps(tc.name, task.TaskCreate, task.TaskScopeRuntime, tc.name, 0)
			if err != nil {
				t.Fatal(err)
			}
			createTask.AddSubTask(createTask.Name, func(item *task.Task) error {
				return executeRuntimeCreateTask(req, rt, model.App{Resource: constant.ResourceLocal}, model.AppDetail{}, template, item)
			}, nil)
			done := make(chan struct{})
			go func() { defer close(done); _ = createTask.Execute() }()
			release := func() { appCacheTestWrite(t, filepath.Join(dir, "release"), "") }
			t.Cleanup(func() { release(); <-done })
			read := request.TaskLogReadReq{TaskReq: request.TaskReq{TaskID: tc.name}, Page: 1, PageSize: 100}
			logs := NewITaskService()
			deadline := time.Now().Add(5 * time.Second)
			for {
				entry, err := logs.ReadByLine(read)
				if err == nil && entry.TaskStatus != constant.StatusExecuting {
					t.Fatalf("task finished before the release signal: %+v", entry)
				}
				if err == nil {
					lines := strings.Join(entry.Lines, "\n")
					if strings.Contains(lines, "step 2/5") && strings.Contains(lines, "Resolving base image 1panel/php:8.5.13-fpm") {
						break
					}
				}
				if time.Now().After(deadline) {
					t.Fatalf("no live progress: %+v / %v", entry, err)
				}
				time.Sleep(10 * time.Millisecond)
			}
			release()
			<-done
			entry, err := logs.ReadByLine(read)
			if err != nil {
				t.Fatal(err)
			}
			summary := strings.Join(entry.Lines, "\n")
			wantStatus := constant.StatusSuccess
			if tc.failure != "" {
				wantStatus = constant.StatusFailed
			}
			if entry.TaskStatus != wantStatus || !strings.Contains(summary, "[TASK-END]") || !strings.Contains(summary, tc.result) {
				t.Fatalf("task %s: %s", entry.TaskStatus, summary)
			}
			for _, private := range []string{"private-build-command", "package-output", "extension-output", "password", "token=private"} {
				if strings.Contains(summary, private) {
					t.Fatalf("unexpected detail %q in %s", private, summary)
				}
			}
			detail, err := os.ReadFile(rt.GetLogPath())
			if err != nil || !bytes.Contains(detail, []byte("package-output-only-in-build.log")) {
				t.Fatalf("build.log: %s / %v", detail, err)
			}
			if (tc.extensions != "") != strings.Contains(summary, "Installing PHP extensions") {
				t.Fatal(summary)
			}
			if tc.name == "extensions" && (!bytes.Contains(detail, []byte("extension-output")) || !bytes.Contains(detail, []byte("commit-output"))) {
				t.Fatalf("incomplete build.log: %s", detail)
			}
			if tc.failure == "" && rt.Status != constant.StatusRunning {
				t.Fatal(rt.Status)
			}
		})
	}
	t.Run("local registration", func(t *testing.T) {
		rt := &model.Runtime{Name: "local", Type: constant.RuntimePHP, Resource: constant.ResourceLocal}
		req := request.RuntimeCreate{Type: rt.Type, Resource: rt.Resource}
		if err := executeRuntimeCreateTask(req, rt, model.App{}, model.AppDetail{}, "", nil); err != nil || rt.Status != constant.StatusNormal {
			t.Fatalf("local PHP: %v, %s", err, rt.Status)
		}
	})
	t.Run("rebuild without a create task", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("PHP_TEST_DIR", dir)
		t.Setenv("PHP_TEST_FAILURE", "")
		appCacheTestWrite(t, filepath.Join(dir, "release"), "")
		rt := &model.Runtime{Name: "rebuild", Type: constant.RuntimePHP}
		appCacheTestWrite(t, rt.GetComposePath(), "services: {}")
		buildRuntime(rt, "", "", true)
		detail, err := os.ReadFile(rt.GetLogPath())
		if err != nil || rt.Status != constant.StatusRunning || !bytes.Contains(detail, []byte("package-output")) {
			t.Fatalf("PHP rebuild: %v, %s, %s", err, rt.Status, detail)
		}
	})
}

func TestPHPBuildSummaryFormats(t *testing.T) {
	setupPHPProgressTest(t)
	for range 2 {
		var output bytes.Buffer
		logger := logrus.New()
		logger.SetOutput(&output)
		p := phpBuildProgress{task: &task.Task{Logger: logger}, phase: "PHPBuildImage", seen: make(map[string]bool), vertices: make(map[string]string)}
		s := phpProgressStream{report: p.observe}
		_, _ = io.WriteString(&s, strings.Repeat("x", 5000)+"\n#2 [internal] load meta")
		_, _ = io.WriteString(&s, "data for php:8\n#2 CACHED\n#2 [internal] load metadata for php:8\n#4 [build 1/4] RUN secret\n#7 [serve 1/4] COPY file\nStep 2/3 : RUN secret")
		s.flush()
		p.observe("#9 [linux/arm64 compile  2/12] RUN secret")
		p.observe("#10 [ 3/12] COPY file")
		p.observe("#2 ERROR: docker.io/library/php:8: not found")
		for _, want := range []string{"Resolving base image php:8", "build 1/4", "serve 1/4", "step 2/3", "linux/arm64 compile 2/12", "step 3/12"} {
			if strings.Count(output.String(), want) != 1 {
				t.Fatal(output.String())
			}
		}
		if strings.Contains(output.String(), "secret") || p.cause != "not found" || !strings.Contains(p.failedStage, "php:8") {
			t.Fatalf("%s / %s / %s", output.String(), p.failedStage, p.cause)
		}
	}
	p := phpBuildProgress{detail: io.Discard}
	if err := p.run("PHPBuildImage", 50*time.Millisecond, "sleep", "3"); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("timeout: %v", err)
	}
}
