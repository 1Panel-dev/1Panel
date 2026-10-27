package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
)

// Inspect a stopped disposable container; its entrypoint never touches retained state.
func bundledFeishuImage(image string) (string, error) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	info, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return "", buserr.New("ErrFeishuImageRequired")
	}
	created, err := cli.ContainerCreate(ctx, &container.Config{Image: info.ID, Entrypoint: []string{"/bin/true"}}, &container.HostConfig{NetworkMode: "none"}, nil, nil, "")
	if err != nil {
		return "", err
	}
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		_ = cli.ContainerRemove(cleanup, created.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	}()
	if _, err = cli.ContainerStatPath(ctx, created.ID, "/app/.1panel-feishu-bundled-v1"); err != nil {
		return "", buserr.New("ErrFeishuImageRequired")
	}
	return info.ID, nil
}

func feishuContainerImage(name string) (string, error) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	info, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", err
	}
	return info.Image, nil
}

// The original state is read-only and the normal Doctor entrypoint is bypassed.
func validateFeishuImageConfig(image, configPath string) error {
	info, err := os.Stat(configPath)
	if err != nil {
		return err
	}
	uid, gid := uint32(1000), uint32(1000)
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid, gid = stat.Uid, stat.Gid
	}
	user := fmt.Sprintf("%d:%d", uid, gid)
	tmpfs := fmt.Sprintf("/tmp:rw,mode=0700,uid=%d,gid=%d", uid, gid)
	name := "1panel-feishu-" + strings.TrimSuffix(filepath.Base(configPath), ".json")
	mgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Minute))
	defer func() { _ = cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).Run("docker", "rm", "-f", "-v", name) }()
	out, runErr := mgr.RunWithStdout("docker", "run", "--name", name, "--rm", "--network", "none", "--read-only", "--user", user,
		"--tmpfs", tmpfs,
		"--mount", "type=bind,src="+filepath.Dir(configPath)+",dst=/home/node/.openclaw,readonly",
		"--env", "HOME=/home/node", "--env", "TMPDIR=/tmp", "--env", "OPENCLAW_STATE_DIR=/tmp/state",
		"--env", "OPENCLAW_CONFIG_PATH=/home/node/.openclaw/"+filepath.Base(configPath),
		"--entrypoint", "sh", image, "-c", "mkdir -m 700 /tmp/state && ln -s /home/node/.openclaw/extensions /tmp/state/extensions && exec node /app/openclaw.mjs config validate --json")
	return feishuValidationResult(out, runErr)
}

func feishuValidationResult(out string, runErr error) error {
	raw, _ := extractEmbeddedJSON(out)
	var result struct {
		Valid  bool `json:"valid"`
		Issues []struct {
			Path string `json:"path"`
		} `json:"issues"`
	}
	if json.Unmarshal(raw, &result) == nil && result.Valid && runErr == nil {
		return nil
	}
	fields := []string{}
	for _, issue := range result.Issues {
		if issue.Path != "" {
			fields = append(fields, issue.Path)
		}
	}
	if len(fields) == 0 {
		fields = append(fields, "openclaw.json")
	}
	return feishuConfigError(strings.Join(fields, ", "))
}

func backupFeishuConfig(configPath string, original []byte) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(configPath), ".feishu-backup-*.json")
	if err != nil {
		return "", err
	}
	if _, err = f.Write(original); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	return f.Name(), nil
}

// Doctor can spend 94 seconds waiting for the previous container's owner lease.
func waitOpenclawReady(name string) error {
	script := `const fs=require('fs');const c=JSON.parse(fs.readFileSync('/home/node/.openclaw/openclaw.json','utf8'));const tls=c.gateway?.tls?.enabled===true;const h=require(tls?'https':'http');function check(p){const r=h.get({hostname:'127.0.0.1',port:c.gateway?.port||18789,path:p,rejectUnauthorized:false,timeout:3000},s=>{s.resume();if(s.statusCode===404&&p==='/readyz')return check('/healthz');process.exitCode=s.statusCode===200?0:1});r.on('timeout',()=>r.destroy());r.on('error',()=>process.exitCode=1)}check('/readyz')`
	deadline := time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		if cmd.RunDockerExec(8*time.Second, name, "node", "-e", script) == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return buserr.New("ErrFeishuActivation")
}

func verifyFeishuActivation(name string, enabled bool) error {
	if err := waitOpenclawReady(name); err != nil {
		return err
	}
	output, err := cmd.RunDockerExecWithStdout(time.Minute, name, "openclaw", "plugins", "list", "--json")
	if err != nil {
		return buserr.New("ErrFeishuActivation")
	}
	raw, _ := extractEmbeddedJSON(output)
	var result struct {
		Plugins []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Origin string `json:"origin"`
		} `json:"plugins"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return buserr.New("ErrFeishuActivation")
	}
	found := false
	for _, p := range result.Plugins {
		if p.ID == "openclaw-lark" && p.Status != "disabled" {
			return buserr.New("ErrFeishuActivation")
		}
		if p.ID == "feishu" && p.Origin == "bundled" && (p.Status == "loaded" || (!enabled && p.Status == "disabled")) {
			found = true
		}
	}
	if !found {
		return buserr.New("ErrFeishuActivation")
	}
	health, err := cmd.RunDockerExecWithStdout(20*time.Second, name, "openclaw", "gateway", "health", "--json")
	if err != nil || !feishuGatewayReady(health, enabled) {
		return buserr.New("ErrFeishuActivation")
	}
	return nil
}

func readFeishuConfigBytes(file string) ([]byte, map[string]interface{}, error) {
	info, err := os.Lstat(file)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, nil, feishuConfigError("openclaw.json")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}
	var conf map[string]interface{}
	if err = json.Unmarshal(raw, &conf); err != nil || conf == nil {
		return nil, nil, feishuConfigError("openclaw.json")
	}
	return raw, conf, nil
}
