package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
)

// Keep panel configuration writes out of the backup/activation/recovery window.
var feishuSwitches sync.Map

func isOpenclawBundledFeishuVersion(version string) bool {
	return common.CompareAppVersion(version, "2026.10.7")
}

func hasBundledFeishu(container string) bool {
	image, err := feishuContainerImage(container)
	if err != nil {
		return false
	}
	_, err = bundledFeishuImage(image)
	return err == nil
}

func usesBundledFeishu(conf map[string]interface{}) bool {
	f := getChannelConfig(conf, "feishu")
	for _, key := range []string{"threadSession", "replyMode"} {
		if _, old := f[key]; old {
			return false
		}
	}
	if _, old := f["streaming"].(bool); old {
		return false
	}
	entries := childMap(childMap(conf, "plugins"), "entries")
	if extractBoolValue(childMap(entries, "openclaw-lark")["enabled"], false) {
		return false
	}
	_, streaming := f["streaming"].(map[string]interface{})
	_, selected := entries["feishu"]
	return selected && (extractBoolValue(childMap(entries, "feishu")["enabled"], false) || f["groupSessionScope"] != nil || streaming)
}

func hasLegacyFeishu(conf map[string]interface{}, installPath string) bool {
	legacy, _, err := classifyLegacyFeishu(conf, installPath)
	return err != nil || legacy
}

func selectBundledFeishu(conf map[string]interface{}, enabled bool) {
	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	if legacy, ok := entries["openclaw-lark"]; ok {
		if entry, ok := legacy.(map[string]interface{}); ok {
			entry["enabled"] = false
		}
	}
	ensureChildMap(entries, "feishu")["enabled"] = enabled
	// Preserve an absent allowlist; creating one would disable unrelated plugins.
	if allow, ok := plugins["allow"]; ok && len(extractStringList(allow)) > 0 {
		ids := slices.DeleteFunc(extractStringList(allow), func(id string) bool { return id == "openclaw-lark" })
		if !slices.Contains(ids, "feishu") {
			ids = append(ids, "feishu")
		}
		plugins["allow"] = ids
	}
	if deny, ok := plugins["deny"]; ok {
		plugins["deny"] = slices.DeleteFunc(extractStringList(deny), func(id string) bool { return id == "feishu" })
	}
}

func feishuConfigError(field string) error { return buserr.WithDetail("ErrFeishuConfig", field, nil) }

func convertLegacyFeishu(conf map[string]interface{}, req dto.AgentFeishuSwitchReq) error {
	f := getChannelConfig(conf, "feishu")
	if len(f) == 0 {
		if raw, exists := childMap(conf, "channels")["feishu"]; exists && raw != nil {
			if _, ok := raw.(map[string]interface{}); !ok {
				return feishuConfigError("channels.feishu")
			}
		}
		f = ensureChildMap(ensureChildMap(conf, "channels"), "feishu")
		f["enabled"], f["dmPolicy"], f["groupPolicy"] = false, "pairing", "allowlist"
	}
	if err := checkLegacyFeishuFields(f, "channels.feishu", false); err != nil {
		return err
	}
	if err := convertFeishuAccess(f, "channels.feishu"); err != nil {
		return err
	}
	delete(f, "threadSession")
	delete(f, "replyMode")
	f["streaming"] = map[string]interface{}{"mode": "off"}
	if req.Streaming {
		f["streaming"] = map[string]interface{}{"mode": "partial"}
	}
	f["groupSessionScope"] = req.GroupSessionScope
	f["requireMention"] = req.RequireMention == "true"
	delete(f, "topicSessionMode")
	// Preserve legacy defaults explicitly; bundled defaults are more restrictive.
	if _, ok := f["groupPolicy"]; !ok {
		f["groupPolicy"] = "open"
	}
	entries := ensureChildMap(ensureChildMap(conf, "plugins"), "entries")
	ensureChildMap(entries, "openclaw-lark")["enabled"] = false
	selectBundledFeishu(conf, extractBoolValue(f["enabled"], true))
	return nil
}

// Limit automatic conversion to the shared channel fields; custom plugin options require review.
func checkLegacyFeishuFields(f map[string]interface{}, field string, group bool) error {
	allowed := strings.Fields("enabled name appId appSecret encryptKey verificationToken domain connectionMode webhookPath defaultAccount dmPolicy allowFrom groupPolicy groupAllowFrom groupSenderAllowFrom requireMention accounts groups streaming threadSession replyMode groupSessionScope topicSessionMode replyInThread")
	if group {
		allowed = strings.Fields("enabled requireMention allowFrom groupSessionScope topicSessionMode replyInThread")
	}
	for _, key := range sortedChildKeys(f) {
		if !slices.Contains(allowed, key) {
			return feishuConfigError(field + "." + key)
		}
	}
	if enabled, exists := f["enabled"]; exists {
		if _, ok := enabled.(bool); !ok {
			return feishuConfigError(field + ".enabled")
		}
	}
	for _, key := range []string{"dmPolicy", "groupPolicy"} {
		if raw, exists := f[key]; exists {
			value, ok := raw.(string)
			options := []string{"open", "allowlist", "disabled"}
			if key == "dmPolicy" {
				options = append(options, "pairing")
			}
			if !ok || !slices.Contains(options, value) {
				return feishuConfigError(field + "." + key)
			}
		}
	}
	for _, key := range []string{"allowFrom", "groupSenderAllowFrom"} {
		if raw, exists := f[key]; exists {
			values, ok := raw.([]interface{})
			if !ok {
				return feishuConfigError(field + "." + key)
			}
			for _, value := range values {
				if _, ok := value.(string); !ok {
					return feishuConfigError(field + "." + key)
				}
			}
		}
	}
	for _, section := range []string{"accounts", "groups"} {
		raw, exists := f[section]
		if !exists {
			continue
		}
		entries, ok := raw.(map[string]interface{})
		if !ok {
			return feishuConfigError(field + "." + section)
		}
		for _, id := range sortedChildKeys(entries) {
			entry, ok := entries[id].(map[string]interface{})
			name := field + "." + section + "." + id
			if !ok {
				return feishuConfigError(name)
			}
			if err := checkLegacyFeishuFields(entry, name, section == "groups"); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertFeishuAccess(f map[string]interface{}, field string) error {
	if raw, ok := f["groupAllowFrom"]; ok {
		values, ok := raw.([]interface{})
		if !ok {
			return feishuConfigError(field + ".groupAllowFrom")
		}
		groups, senders := []string{}, []string{}
		for _, value := range values {
			id, ok := value.(string)
			if !ok {
				return feishuConfigError(field + ".groupAllowFrom")
			}
			switch {
			case strings.HasPrefix(id, "oc_"):
				groups = append(groups, id)
			case strings.HasPrefix(id, "ou_"):
				senders = append(senders, id)
			default:
				return feishuConfigError(field + ".groupAllowFrom")
			}
		}
		// A sender-only legacy allowlist has different group admission semantics.
		if len(senders) > 0 && (len(groups) == 0 || f["groupSenderAllowFrom"] != nil) {
			return feishuConfigError(field + ".groupAllowFrom")
		}
		f["groupAllowFrom"] = groups
		if len(senders) > 0 {
			f["groupSenderAllowFrom"] = senders
		}
	}
	for _, section := range []string{"accounts", "groups"} {
		for id, value := range childMap(f, section) {
			entry, ok := value.(map[string]interface{})
			name := field + "." + section + "." + id
			if !ok {
				return feishuConfigError(name)
			}
			// Account/group overrides must be resolved explicitly, not flattened.
			for _, key := range []string{"threadSession", "replyMode"} {
				if _, ok := entry[key]; ok {
					return feishuConfigError(name + "." + key)
				}
			}
			if mention, ok := entry["requireMention"]; ok {
				if _, ok := mention.(bool); !ok {
					return feishuConfigError(name + ".requireMention")
				}
			}
			if raw, exists := entry["streaming"]; exists {
				streaming, ok := raw.(bool)
				if !ok || streaming {
					return feishuConfigError(name + ".streaming")
				}
				entry["streaming"] = map[string]interface{}{"mode": "off"}
			}
			if err := convertFeishuAccess(entry, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractBundledFeishuConfig(conf map[string]interface{}) dto.AgentFeishuConfig {
	result := extractFeishuConfig(conf)
	f := getChannelConfig(conf, "feishu")
	result.Enabled = extractBoolValue(f["enabled"], true) && extractBoolValue(childMap(childMap(childMap(conf, "plugins"), "entries"), "feishu")["enabled"], true)
	result.GroupSessionScope = extractStringValue(f["groupSessionScope"])
	if result.GroupSessionScope == "" {
		result.GroupSessionScope = "group"
	}
	result.Streaming = extractStringValue(childMap(f, "streaming")["mode"]) != "off"
	result.GroupPolicy = extractStringValue(f["groupPolicy"])
	if result.GroupPolicy == "" {
		result.GroupPolicy = "allowlist"
	}
	if len(f) == 0 {
		result.Streaming = false
		result.RequireMention = "true"
	} else if _, ok := f["requireMention"]; !ok {
		result.RequireMention = "open"
	}
	accounts := childMap(f, "accounts")
	ids := sortedChildKeys(accounts)
	if (f["appId"] != nil || f["appSecret"] != nil) && !slices.Contains(ids, "default") {
		ids = append([]string{"default"}, ids...)
	}
	result.Bots = []dto.AgentFeishuBot{}
	for _, id := range ids {
		effective := make(map[string]interface{}, len(f))
		for k, v := range f {
			effective[k] = v
		}
		for k, v := range childMap(accounts, id) {
			effective[k] = v
		}
		bot := defaultFeishuBot()
		bot.AccountID, bot.IsDefault = id, id == "default"
		bot.Name = extractDisplayName(childMap(accounts, id), "", id)
		bot.Enabled = extractBoolValue(childMap(accounts, id)["enabled"], true)
		bot.AppID, bot.AppSecret = extractStringValue(effective["appId"]), extractStringValue(effective["appSecret"])
		bot.DmPolicy = extractStringValue(effective["dmPolicy"])
		if bot.DmPolicy == "" {
			bot.DmPolicy = "pairing"
		}
		bot.AllowFrom = extractStringList(effective["allowFrom"])
		result.Bots = append(result.Bots, bot)
	}
	return result
}

func setBundledFeishuConfig(conf map[string]interface{}, req dto.AgentFeishuConfigUpdateReq) error {
	if !slices.Contains([]string{"group", "group_sender", "group_topic", "group_topic_sender"}, req.GroupSessionScope) {
		return feishuConfigError("groupSessionScope")
	}
	f := ensureChildMap(ensureChildMap(conf, "channels"), "feishu")
	accounts := childMap(f, "accounts")
	seen := map[string]bool{}
	appIDs := map[string]bool{}
	for _, bot := range req.Bots {
		if bot.AccountID == "" || bot.AccountID != strings.ToLower(strings.TrimSpace(bot.AccountID)) || seen[bot.AccountID] || appIDs[bot.AppID] || bot.AppID == "" || !slices.Contains([]string{"pairing", "open", "allowlist", "disabled"}, bot.DmPolicy) || (bot.DmPolicy == "allowlist" && len(bot.AllowFrom) == 0) {
			return feishuConfigError("accounts")
		}
		seen[bot.AccountID] = true
		appIDs[bot.AppID] = true
		entry := childMap(accounts, bot.AccountID)
		if _, exists := entry["appSecret"]; !exists && bot.AppSecret == "" {
			entry["appSecret"] = f["appSecret"]
		}
		// SecretRef credentials cannot be represented by the form: keep them intact.
		if _, secretRef := entry["appSecret"].(map[string]interface{}); !secretRef && bot.AppSecret == "" {
			return feishuConfigError("accounts." + bot.AccountID + ".appSecret")
		}
		entry["appId"], entry["name"], entry["enabled"], entry["dmPolicy"] = bot.AppID, bot.Name, bot.Enabled, bot.DmPolicy
		if bot.AppSecret != "" {
			entry["appSecret"] = bot.AppSecret
		}
		switch bot.DmPolicy {
		case "open":
			entry["allowFrom"] = []string{"*"}
		case "allowlist":
			entry["allowFrom"] = bot.AllowFrom
		default:
			if len(bot.AllowFrom) > 0 {
				entry["allowFrom"] = bot.AllowFrom
			} else {
				entry["allowFrom"] = []string{}
			}
		}
		accounts[bot.AccountID] = entry
	}
	for id := range accounts {
		if !seen[id] {
			delete(accounts, id)
		}
	}
	if !seen["default"] {
		delete(f, "appId")
		delete(f, "appSecret")
	}
	if id := extractStringValue(f["defaultAccount"]); id != "" && !seen[id] {
		return feishuConfigError("defaultAccount")
	}
	if len(accounts) > 0 {
		f["accounts"] = accounts
	} else {
		delete(f, "accounts")
	}
	f["enabled"], f["groupSessionScope"], f["groupPolicy"] = req.Enabled, req.GroupSessionScope, req.GroupPolicy
	if req.RequireMention == "open" {
		delete(f, "requireMention")
	} else {
		f["requireMention"] = req.RequireMention == "true"
	}
	if req.GroupPolicy == "allowlist" {
		f["groupAllowFrom"] = req.GroupAllowFrom
	} else {
		delete(f, "groupAllowFrom")
	}
	streaming := ensureChildMap(f, "streaming")
	streaming["mode"] = "off"
	if req.Streaming {
		streaming["mode"] = "partial"
	}
	selectBundledFeishu(conf, req.Enabled && hasEnabledBots(req.Bots))
	return nil
}

// Write beside the original so validation and rename preserve the mounted filesystem and owner.
func stageFeishuConfig(configPath string, payload []byte) (string, error) {
	info, err := os.Lstat(configPath)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", feishuConfigError("openclaw.json")
	}
	f, err := os.CreateTemp(filepath.Dir(configPath), ".feishu-*.json")
	if err != nil {
		return "", err
	}
	name := f.Name()
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(name)
		}
	}()
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		err = f.Chown(int(stat.Uid), int(stat.Gid))
	}
	if err == nil {
		err = f.Chmod(info.Mode().Perm())
	}
	if err == nil {
		_, err = f.Write(payload)
	}
	if err == nil {
		err = f.Sync()
	}
	if err == nil {
		err = f.Close()
	}
	return name, err
}

func validateFeishuConfig(container, configPath string) error {
	out, err := cmd.RunDockerExecWithStdout(time.Minute, container, "env", "OPENCLAW_CONFIG_PATH=/home/node/.openclaw/"+filepath.Base(configPath), "openclaw", "config", "validate", "--json")
	return feishuValidationResult(out, err)
}

func replaceFeishuConfig(configPath string, expected, payload []byte) error {
	tmp, err := stageFeishuConfig(configPath, payload)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	current, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, expected) {
		return buserr.New("ErrFeishuConfigChanged")
	}
	return os.Rename(tmp, configPath)
}

func (a AgentService) SwitchFeishuPlugin(req dto.AgentFeishuSwitchReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if !isOpenclawBundledFeishuVersion(install.Version) {
		return buserr.New("ErrFeishuImageRequired")
	}
	if _, busy := feishuSwitches.LoadOrStore(agent.ConfigPath, true); busy {
		return buserr.New("TaskIsExecuting")
	}
	started := false
	defer func() {
		if !started {
			feishuSwitches.Delete(agent.ConfigPath)
		}
	}()
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	if !hasBundledFeishu(install.ContainerName) {
		return buserr.New("ErrFeishuImageRequired")
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	legacy, configured, err := classifyLegacyFeishu(conf, install.GetPath())
	if err != nil {
		return err
	}
	if !legacy {
		return buserr.New("ErrFeishuAlreadySwitched")
	}
	if err := prepareLegacyFeishu(conf, configured, &dto.AgentFeishuTransition{Confirmed: true, Streaming: req.Streaming, GroupSessionScope: req.GroupSessionScope, RequireMention: req.RequireMention}); err != nil {
		return err
	}
	switchTask, err := task.NewTaskWithOps("Feishu", task.TaskUpdate, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	switchTask.AddSubTaskWithAlias("AgentFeishuSwitch", func(t *task.Task) error { return a.runFeishuSwitch(agent, install, req, t) }, nil)
	started = true
	go func() { defer feishuSwitches.Delete(agent.ConfigPath); executeAgentPluginTask(switchTask) }()
	return nil
}

func (a AgentService) runFeishuSwitch(agent *model.Agent, install *model.AppInstall, req dto.AgentFeishuSwitchReq, t *task.Task) (result error) {
	original, conf, err := readFeishuConfigBytes(agent.ConfigPath)
	if err != nil {
		return err
	}
	legacy, configured, err := classifyLegacyFeishu(conf, install.GetPath())
	if err != nil {
		return err
	}
	if !legacy {
		return buserr.New("ErrFeishuAlreadySwitched")
	}
	image, err := feishuContainerImage(install.ContainerName)
	if err != nil {
		return err
	}
	image, err = bundledFeishuImage(image)
	if err != nil {
		return err
	}
	if err = prepareLegacyFeishu(conf, configured, &dto.AgentFeishuTransition{Confirmed: true, Streaming: req.Streaming, GroupSessionScope: req.GroupSessionScope, RequireMention: req.RequireMention}); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	recovery, err := feishuRecoveryPayload(payload, configured)
	if err != nil {
		return err
	}
	// Validate both candidates without invoking Doctor or mutating retained state.
	for _, candidate := range [][]byte{payload, recovery} {
		staged, stageErr := stageFeishuConfig(agent.ConfigPath, candidate)
		if stageErr != nil {
			return stageErr
		}
		validateErr := validateFeishuImageConfig(image, staged)
		os.Remove(staged)
		if validateErr != nil {
			return validateErr
		}
	}
	if _, err = compose.Stop(install.GetComposePath()); err != nil {
		return err
	}
	replaced := false
	defer func() {
		if !replaced {
			if _, startErr := compose.Operate(install.GetComposePath(), "start"); startErr != nil {
				result = errors.Join(result, buserr.WithErr("ErrFeishuRestore", startErr))
			}
			return
		}
		if result == nil {
			return
		}
		if _, stopErr := compose.Stop(install.GetComposePath()); stopErr != nil {
			result = errors.Join(result, buserr.WithErr("ErrFeishuRestore", stopErr))
			return
		}
		if restoreErr := replaceFeishuConfig(agent.ConfigPath, payload, recovery); restoreErr != nil {
			result = errors.Join(result, buserr.WithErr("ErrFeishuRestore", restoreErr))
			return
		}
		if _, startErr := compose.Operate(install.GetComposePath(), "start"); startErr != nil {
			result = errors.Join(result, buserr.WithErr("ErrFeishuRestore", startErr))
			return
		}
		if healthErr := verifyFeishuActivation(install.ContainerName, false); healthErr != nil {
			result = errors.Join(result, buserr.WithErr("ErrFeishuRestore", healthErr))
			return
		}
		t.Log(i18n.GetMsgByKey("AgentFeishuRecoveryDisabled"))
	}()
	current, err := os.ReadFile(agent.ConfigPath)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, original) {
		return buserr.New("ErrFeishuConfigChanged")
	}
	backup, err := backupFeishuConfig(agent.ConfigPath, original)
	if err != nil {
		return err
	}
	t.Log(i18n.GetWithName("AgentFeishuBackup", filepath.Base(backup)))
	if err = replaceFeishuConfig(agent.ConfigPath, original, payload); err != nil {
		return err
	}
	replaced = true
	if _, err = compose.Operate(install.GetComposePath(), "start"); err != nil {
		return err
	}
	enabled := configured && extractBoolValue(getChannelConfig(conf, "feishu")["enabled"], true)
	if err = verifyFeishuActivation(install.ContainerName, enabled); err != nil {
		return err
	}
	if enabled {
		probeFeishuConnection(install.ContainerName, t)
	}
	if configured {
		t.Log(i18n.GetMsgByKey("AgentFeishuVerifyMessages"))
	} else {
		t.Log(i18n.GetMsgByKey("AgentFeishuRecoveryDisabled"))
	}
	return nil
}

func saveBundledFeishuConfig(agent *model.Agent, install *model.AppInstall, req dto.AgentFeishuConfigUpdateReq) error {
	if !isOpenclawBundledFeishuVersion(install.Version) {
		return buserr.New("ErrFeishuImageRequired")
	}
	if _, busy := feishuSwitches.LoadOrStore(agent.ConfigPath, true); busy {
		return buserr.New("TaskIsExecuting")
	}
	defer feishuSwitches.Delete(agent.ConfigPath)
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	if !hasBundledFeishu(install.ContainerName) {
		return buserr.New("ErrFeishuImageRequired")
	}
	original, err := os.ReadFile(agent.ConfigPath)
	if err != nil {
		return err
	}
	var conf map[string]interface{}
	if err = json.Unmarshal(original, &conf); err != nil {
		return err
	}
	if hasLegacyFeishu(conf, install.GetPath()) {
		return buserr.New("ErrFeishuSwitchRequired")
	}
	if err = setBundledFeishuConfig(conf, req); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	staged, err := stageFeishuConfig(agent.ConfigPath, payload)
	if err != nil {
		return err
	}
	defer os.Remove(staged)
	if err = validateFeishuConfig(install.ContainerName, staged); err != nil {
		return err
	}
	return replaceFeishuConfig(agent.ConfigPath, original, payload)
}

func feishuProbeOK(output string) bool {
	raw, err := extractEmbeddedJSON(output)
	if err != nil {
		return false
	}
	var status struct {
		ChannelAccounts map[string][]struct {
			Enabled    bool `json:"enabled"`
			Configured bool `json:"configured"`
			Running    bool `json:"running"`
			Probe      struct {
				OK bool `json:"ok"`
			} `json:"probe"`
		} `json:"channelAccounts"`
	}
	if json.Unmarshal(raw, &status) != nil {
		return false
	}
	accounts, ok := status.ChannelAccounts["feishu"]
	if !ok {
		return false
	}
	for _, account := range accounts {
		if account.Enabled && (!account.Configured || !account.Running || !account.Probe.OK) {
			return false
		}
	}
	return true
}

func feishuGatewayReady(output string, enabled bool) bool {
	raw, err := extractEmbeddedJSON(output)
	if err != nil {
		return false
	}
	var health struct {
		OK      bool `json:"ok"`
		Plugins struct {
			Loaded []string `json:"loaded"`
		} `json:"plugins"`
	}
	return json.Unmarshal(raw, &health) == nil && health.OK && !slices.Contains(health.Plugins.Loaded, "openclaw-lark") && (!enabled || slices.Contains(health.Plugins.Loaded, "feishu"))
}
