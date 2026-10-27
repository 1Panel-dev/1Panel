package service

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"gopkg.in/yaml.v3"
)

func classifyLegacyFeishu(conf map[string]interface{}, installPath string) (legacy, configured bool, err error) {
	for _, field := range []string{"channels", "channels.feishu", "plugins", "plugins.entries", "plugins.entries.openclaw-lark", "plugins.entries.feishu"} {
		parts := strings.Split(field, ".")
		parent := conf
		for _, part := range parts[:len(parts)-1] {
			parent = childMap(parent, part)
		}
		if raw, exists := parent[parts[len(parts)-1]]; exists {
			if _, ok := raw.(map[string]interface{}); !ok {
				return false, false, feishuConfigError(field)
			}
		}
	}
	plugins := childMap(conf, "plugins")
	for _, key := range []string{"allow", "deny"} {
		if raw, exists := plugins[key]; exists {
			if _, ok := raw.([]string); ok {
				continue
			}
			ids, ok := raw.([]interface{})
			if !ok {
				return false, false, feishuConfigError("plugins." + key)
			}
			for _, id := range ids {
				if _, ok := id.(string); !ok {
					return false, false, feishuConfigError("plugins." + key)
				}
			}
		}
	}
	for _, id := range []string{"openclaw-lark", "feishu"} {
		if raw, exists := childMap(childMap(plugins, "entries"), id)["enabled"]; exists {
			if _, ok := raw.(bool); !ok {
				return false, false, feishuConfigError("plugins.entries." + id + ".enabled")
			}
		}
	}
	if usesBundledFeishu(conf) {
		return false, false, nil
	}
	f := getChannelConfig(conf, "feishu")
	configured = len(f) > 0
	if raw, exists := conf["bindings"]; exists {
		bindings, ok := raw.([]interface{})
		if !ok {
			return false, false, feishuConfigError("bindings")
		}
		for _, raw := range bindings {
			binding, ok := raw.(map[string]interface{})
			if !ok {
				return false, false, feishuConfigError("bindings")
			}
			if match, exists := binding["match"]; exists {
				if _, ok := match.(map[string]interface{}); !ok {
					return false, false, feishuConfigError("bindings.match")
				}
			}
			if extractStringValue(childMap(binding, "match")["channel"]) == "feishu" {
				configured = true
			}
		}
	}
	entries := childMap(childMap(conf, "plugins"), "entries")
	entry, selected := entries["openclaw-lark"]
	if selected {
		old, ok := entry.(map[string]interface{})
		if !ok {
			return false, configured, feishuConfigError("plugins.entries.openclaw-lark")
		}
		// A completed disable-only repair must not repeat on every upgrade.
		if enabled, ok := old["enabled"].(bool); ok && !enabled && !configured {
			return false, false, nil
		}
	}
	installed, err := checkPluginInstalled(installPath, "feishu")
	if err != nil {
		return false, configured, err
	}
	_, oldThread := f["threadSession"]
	_, oldReply := f["replyMode"]
	_, oldStreaming := f["streaming"].(bool)
	return selected || installed || oldThread || oldReply || oldStreaming, configured, nil
}

func prepareLegacyFeishu(conf map[string]interface{}, configured bool, transition *dto.AgentFeishuTransition) error {
	if !configured {
		ensureChildMap(ensureChildMap(ensureChildMap(conf, "plugins"), "entries"), "openclaw-lark")["enabled"] = false
		return nil
	}
	if transition == nil || !transition.Confirmed {
		return buserr.New("ErrFeishuUpgradeConfirm")
	}
	if !validFeishuTransition(transition.GroupSessionScope, transition.RequireMention) {
		return feishuConfigError("feishuTransition")
	}
	return convertLegacyFeishu(conf, dto.AgentFeishuSwitchReq{Streaming: transition.Streaming, GroupSessionScope: transition.GroupSessionScope, RequireMention: transition.RequireMention})
}

func validFeishuTransition(scope, mention string) bool {
	return (scope == "group" || scope == "group_sender" || scope == "group_topic" || scope == "group_topic_sender") && (mention == "true" || mention == "false")
}

func feishuRecoveryPayload(candidate []byte, configured bool) ([]byte, error) {
	var conf map[string]interface{}
	if err := json.Unmarshal(candidate, &conf); err != nil {
		return nil, err
	}
	if configured {
		ensureChildMap(ensureChildMap(conf, "channels"), "feishu")["enabled"] = false
		selectBundledFeishu(conf, false)
	}
	ensureChildMap(ensureChildMap(ensureChildMap(conf, "plugins"), "entries"), "openclaw-lark")["enabled"] = false
	return json.MarshalIndent(conf, "", "  ")
}

func (u *appUpgradeContext) prepareFeishuUpgrade() error {
	if u.original.App.Key != constant.AppOpenclaw || !isOpenclawBundledFeishuVersion(u.candidate.Version) {
		return nil
	}
	source := filepath.Join(u.original.GetPath(), "data/conf/openclaw.json")
	staged := filepath.Join(u.stageDir, "data/conf/openclaw.json")
	// The staged copy predates protocol migration and is the exact cutover comparison source.
	original := u.openclawOriginal
	current, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if !bytes.Equal(original, current) {
		return buserr.New("ErrFeishuConfigChanged")
	}
	_, conf, err := readFeishuConfigBytes(staged)
	if err != nil {
		return err
	}
	legacy, configured, err := classifyLegacyFeishu(conf, u.original.GetPath())
	if err != nil || !legacy {
		return err
	}
	if !u.req.Backup {
		return buserr.New("ErrFeishuUpgradeBackup")
	}
	if err = prepareLegacyFeishu(conf, configured, u.req.FeishuTransition); err != nil {
		return err
	}
	project, err := docker.GetComposeProject(u.original.Name, u.original.GetPath(), []byte(u.candidate.DockerCompose), u.envContent, false)
	if err != nil {
		return err
	}
	if len(project.Services) != 1 {
		return buserr.New("ErrFeishuUpgradeMounts")
	}
	for _, service := range project.Services {
		if service.Build != nil {
			return buserr.New("ErrFeishuImageRequired")
		}
		u.feishuImage, err = bundledFeishuImage(service.Image)
		if err != nil {
			return err
		}
		configMounted := false
		for _, v := range service.Volumes {
			if v.Target == "/etc/localtime" {
				continue
			}
			if v.Type != "bind" {
				return buserr.New("ErrFeishuUpgradeMounts")
			}
			if err = feishuBackupPath(u.original.GetPath(), v.Source); err != nil {
				return err
			}
			if v.Target == "/home/node/.openclaw" {
				configMounted = filepath.Clean(v.Source) == filepath.Dir(source)
			}
		}
		if !configMounted {
			return buserr.New("ErrFeishuUpgradeMounts")
		}
	}
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	info, err := cli.ContainerInspect(ctx, u.original.ContainerName)
	if err != nil {
		return err
	}
	u.feishuOldImage = info.Image
	for _, m := range info.Mounts {
		if m.Destination == "/etc/localtime" {
			continue
		}
		if string(m.Type) != "bind" {
			return buserr.New("ErrFeishuUpgradeMounts")
		}
		if err = feishuBackupPath(u.original.GetPath(), m.Source); err != nil {
			return err
		}
	}
	// Tar preserves symlinks: an external target would not be in the application backup.
	if err = filepath.WalkDir(filepath.Join(u.original.GetPath(), "data"), func(file string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return feishuBackupPath(u.original.GetPath(), file)
		}
		return nil
	}); err != nil {
		return err
	}
	u.feishuOriginal = original
	u.feishuEnabled = configured && extractBoolValue(getChannelConfig(conf, "feishu")["enabled"], true)
	payload, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	before, err := os.ReadFile(staged)
	if err != nil {
		return err
	}
	if err = replaceFeishuConfig(staged, before, payload); err != nil {
		return err
	}
	targetCompose := filepath.Join(u.stageDir, "docker-compose.yml")
	if err = pinFeishuRollbackImage(targetCompose, u.feishuImage); err != nil {
		return err
	}
	pinned, err := os.ReadFile(targetCompose)
	if err != nil {
		return err
	}
	u.candidate.DockerCompose = string(pinned)
	return nil
}

func feishuBackupPath(root, source string) error {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(source)
	if err != nil {
		return buserr.New("ErrFeishuUpgradeMounts")
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return buserr.New("ErrFeishuUpgradeMounts")
	}
	return nil
}

func (u *appUpgradeContext) checkFeishuUpgradeSource() error {
	if u.feishuOriginal == nil {
		return nil
	}
	current, err := os.ReadFile(filepath.Join(u.original.GetPath(), "data/conf/openclaw.json"))
	if err != nil {
		return err
	}
	if !bytes.Equal(current, u.feishuOriginal) {
		return buserr.New("ErrFeishuConfigChanged")
	}
	return nil
}

func (u *appUpgradeContext) verifyFeishuUpgrade(t *task.Task) error {
	if u.feishuOriginal == nil {
		return nil
	}
	if err := verifyFeishuActivation(u.candidate.ContainerName, u.feishuEnabled); err != nil {
		return err
	}
	if u.feishuEnabled {
		probeFeishuConnection(u.candidate.ContainerName, t)
	}
	return nil
}

func probeFeishuConnection(name string, t *task.Task) {
	output, err := cmd.RunDockerExecWithStdout(30*time.Second, name, "openclaw", "channels", "status", "--channel", "feishu", "--probe", "--json")
	if err != nil || !feishuProbeOK(output) {
		t.Log(i18n.GetMsgByKey("AgentFeishuProbeWarning"))
	}
}

// Pin the restored service to the actual old image, even if its registry tag moved.
func pinFeishuRollbackImage(composeFile string, image string) error {
	raw, err := os.ReadFile(composeFile)
	if err != nil {
		return err
	}
	var conf map[string]interface{}
	if err = yaml.Unmarshal(raw, &conf); err != nil {
		return err
	}
	services := childMap(conf, "services")
	if len(services) != 1 {
		return buserr.New("ErrFeishuUpgradeMounts")
	}
	for _, raw := range services {
		service, ok := raw.(map[string]interface{})
		if !ok {
			return buserr.New("ErrFeishuUpgradeMounts")
		}
		service["image"] = image
	}
	payload, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return writeUpgradeFile(composeFile, payload, constant.FilePerm)
}
