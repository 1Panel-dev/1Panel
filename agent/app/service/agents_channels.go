package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
)

type openclawPluginPackage struct {
	Version string `json:"version"`
}

func updateHermesChannelConfig(agent *model.Agent, install *model.AppInstall, write func(confDir string) error) error {
	if err := write(path.Dir(agent.ConfigPath)); err != nil {
		return err
	}
	_, err := compose.Restart(install.GetComposePath())
	return err
}

func (a AgentService) GetFeishuConfig(req dto.AgentFeishuConfigReq) (*dto.AgentFeishuConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesFeishuChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractFeishuConfig(conf)
	installed, _ := checkPluginInstalled(install.GetPath(), "feishu")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateFeishuConfig(req dto.AgentFeishuConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesFeishuChannelConfig(confDir, dto.AgentFeishuConfig{
				Enabled:        req.Enabled,
				ThreadSession:  req.ThreadSession,
				ReplyMode:      req.ReplyMode,
				Streaming:      req.Streaming,
				RequireMention: req.RequireMention,
				GroupPolicy:    req.GroupPolicy,
				GroupAllowFrom: req.GroupAllowFrom,
				Domain:         "feishu",
				ConnectionMode: "websocket",
				Bots:           req.Bots,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		config := dto.AgentFeishuConfig{
			Enabled:        req.Enabled,
			ThreadSession:  req.ThreadSession,
			ReplyMode:      req.ReplyMode,
			Streaming:      req.Streaming,
			RequireMention: req.RequireMention,
			GroupPolicy:    req.GroupPolicy,
			GroupAllowFrom: req.GroupAllowFrom,
			Domain:         req.Domain,
			ConnectionMode: req.ConnectionMode,
			Bots:           req.Bots,
		}
		setFeishuConfig(conf, config)
		setFeishuPluginEnabled(conf, config.Enabled && hasEnabledBots(config.Bots))
		return nil
	})
}

func (a AgentService) GetTelegramConfig(req dto.AgentTelegramConfigReq) (*dto.AgentTelegramConfig, error) {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesTelegramChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractTelegramConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateTelegramConfig(req dto.AgentTelegramConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesTelegramChannelConfig(confDir, dto.AgentTelegramConfig{
				Enabled:        req.Enabled,
				DmPolicy:       req.DmPolicy,
				AllowFrom:      req.AllowFrom,
				RequireMention: req.RequireMention,
				GroupPolicy:    req.GroupPolicy,
				GroupAllowFrom: req.GroupAllowFrom,
				Proxy:          req.Proxy,
				Streaming:      req.Streaming,
				DefaultAccount: req.DefaultAccount,
				Bots:           req.Bots,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setTelegramConfig(conf, dto.AgentTelegramConfig{
			Enabled:        req.Enabled,
			DmPolicy:       req.DmPolicy,
			AllowFrom:      req.AllowFrom,
			RequireMention: req.RequireMention,
			GroupPolicy:    req.GroupPolicy,
			GroupAllowFrom: req.GroupAllowFrom,
			Proxy:          req.Proxy,
			Streaming:      req.Streaming,
			DefaultAccount: req.DefaultAccount,
			Bots:           req.Bots,
		})
		return nil
	})
}

func (a AgentService) GetDiscordConfig(req dto.AgentIDReq) (*dto.AgentDiscordConfig, error) {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesDiscordChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractDiscordConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateDiscordConfig(req dto.AgentDiscordConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesDiscordChannelConfig(confDir, dto.AgentDiscordConfig{
				Enabled:        req.Enabled,
				DmPolicy:       req.DmPolicy,
				AllowFrom:      req.AllowFrom,
				RequireMention: req.RequireMention,
				GroupPolicy:    req.GroupPolicy,
				Proxy:          req.Proxy,
				DefaultAccount: req.DefaultAccount,
				Bots:           req.Bots,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setDiscordConfig(conf, dto.AgentDiscordConfig{
			Enabled:        req.Enabled,
			DmPolicy:       req.DmPolicy,
			AllowFrom:      req.AllowFrom,
			RequireMention: req.RequireMention,
			GroupPolicy:    req.GroupPolicy,
			Proxy:          req.Proxy,
			DefaultAccount: req.DefaultAccount,
			Bots:           req.Bots,
		})
		return nil
	})
}

func (a AgentService) GetQQBotConfig(req dto.AgentIDReq) (*dto.AgentQQBotConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesQQBotChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractQQBotConfig(conf)
	installed, _ := checkPluginInstalled(install.GetPath(), "qqbot")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateQQBotConfig(req dto.AgentQQBotConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesQQBotChannelConfig(confDir, dto.AgentQQBotConfig{
				Enabled:        req.Enabled,
				DmPolicy:       req.DmPolicy,
				AllowFrom:      req.AllowFrom,
				GroupPolicy:    req.GroupPolicy,
				GroupAllowFrom: req.GroupAllowFrom,
				Bots:           req.Bots,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setQQBotConfig(conf, dto.AgentQQBotConfig{
			Enabled: req.Enabled,
			Bots:    req.Bots,
		})
		return nil
	})
}

func (a AgentService) GetWecomConfig(req dto.AgentIDReq) (*dto.AgentWecomConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesWecomChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractWecomConfig(conf)
	installed, _ := checkPluginInstalled(install.GetPath(), "wecom")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateWecomConfig(req dto.AgentWecomConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesWecomChannelConfig(confDir, dto.AgentWecomConfig{
				Enabled:        req.Enabled,
				DmPolicy:       req.DmPolicy,
				AllowFrom:      req.AllowFrom,
				GroupPolicy:    req.GroupPolicy,
				GroupAllowFrom: req.GroupAllowFrom,
				BotID:          req.BotID,
				Secret:         req.Secret,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setWecomConfig(conf, dto.AgentWecomConfig{
			Enabled:        req.Enabled,
			DmPolicy:       req.DmPolicy,
			AllowFrom:      req.AllowFrom,
			GroupPolicy:    req.GroupPolicy,
			GroupAllowFrom: req.GroupAllowFrom,
			BotID:          req.BotID,
			Secret:         req.Secret,
		})
		return nil
	})
}

func (a AgentService) GetDingTalkConfig(req dto.AgentIDReq) (*dto.AgentDingTalkConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return readHermesDingTalkChannelConfig(path.Dir(agent.ConfigPath))
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractDingTalkConfig(conf)
	installed, _ := checkPluginInstalled(install.GetPath(), "dingtalk")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateDingTalkConfig(req dto.AgentDingTalkConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return updateHermesChannelConfig(agent, install, func(confDir string) error {
			return writeHermesDingTalkChannelConfig(confDir, dto.AgentDingTalkConfig{
				Enabled:        req.Enabled,
				DmPolicy:       req.DmPolicy,
				AllowFrom:      req.AllowFrom,
				GroupPolicy:    req.GroupPolicy,
				GroupAllowFrom: req.GroupAllowFrom,
				Bots:           req.Bots,
			})
		})
	}
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setDingTalkConfig(conf, dto.AgentDingTalkConfig{
			Enabled:                         req.Enabled,
			DmPolicy:                        req.DmPolicy,
			AllowFrom:                       req.AllowFrom,
			GroupPolicy:                     req.GroupPolicy,
			GroupAllowFrom:                  req.GroupAllowFrom,
			SeparateSessionByConversation:   req.SeparateSessionByConversation,
			GroupSessionScope:               req.GroupSessionScope,
			SharedMemoryAcrossConversations: req.SharedMemoryAcrossConversations,
			AsyncMode:                       req.AsyncMode,
			AckText:                         req.AckText,
			Bots:                            req.Bots,
		})
		return nil
	})
}

func (a AgentService) InstallPlugin(req dto.AgentPluginInstallReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	spec, pluginID, err := resolvePluginMeta(req.Type)
	if err != nil {
		return err
	}
	installTask, err := task.NewTaskWithOps(req.Type, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	installTask.AddSubTask("Install OpenClaw plugin", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		if req.Type == "qqbot" {
			legacyPluginPath := path.Join(openclawPluginBaseDir, "qqbot")
			if err := mgr.RunBashCf("docker exec %s test -d %s", install.ContainerName, legacyPluginPath); err == nil {
				if _, err := mgr.RunWithStdout("docker", "exec", "-i", install.ContainerName, "sh", "-c", buildOpenclawPluginUninstallScript("qqbot")); err != nil {
					return err
				}
				time.Sleep(2 * time.Second)
			}
		}
		if _, err := mgr.RunWithStdout("docker", "exec", install.ContainerName, "sh", "-c", buildOpenclawPluginInstallScript(spec, pluginID)); err != nil {
			return err
		}
		conf, err := readOpenclawConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		appendPluginAllow(conf, pluginID)
		return writeOpenclawConfigRaw(agent.ConfigPath, conf)
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw plugin failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) UpgradePlugin(req dto.AgentPluginUpgradeReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	spec, pluginID, err := resolvePluginMeta(req.Type)
	if err != nil {
		return err
	}
	upgradeTask, err := task.NewTaskWithOps(req.Type, task.TaskUpgrade, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	upgradeTask.AddSubTask("Upgrade OpenClaw plugin", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		if _, err := mgr.RunWithStdout("docker", "exec", "-i", install.ContainerName, "sh", "-c", buildOpenclawPluginUninstallScript(pluginID)); err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
		if _, err := mgr.RunWithStdout("docker", "exec", install.ContainerName, "sh", "-c", buildOpenclawPluginInstallScript(spec, pluginID)); err != nil {
			return err
		}
		conf, err := readOpenclawConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		appendPluginAllow(conf, pluginID)
		return writeOpenclawConfigRaw(agent.ConfigPath, conf)
	}, nil)
	go func() {
		if err := upgradeTask.Execute(); err != nil {
			global.LOG.Errorf("upgrade openclaw plugin failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) UninstallPlugin(req dto.AgentPluginUninstallReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	_, pluginID, err := resolvePluginMeta(req.Type)
	if err != nil {
		return err
	}
	uninstallTask, err := task.NewTaskWithOps(req.Type, task.TaskUninstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	uninstallTask.AddSubTask("Uninstall OpenClaw plugin", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		if _, err := mgr.RunWithStdout("docker", "exec", "-i", install.ContainerName, "sh", "-c", buildOpenclawPluginUninstallScript(pluginID)); err != nil {
			return err
		}
		conf, err := readOpenclawConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		cleanupOpenclawPluginConfig(conf, req.Type)
		return writeOpenclawConfigRaw(agent.ConfigPath, conf)
	}, nil)
	go func() {
		if err := uninstallTask.Execute(); err != nil {
			global.LOG.Errorf("uninstall openclaw plugin failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) LoginWeixinChannel(req dto.AgentWeixinLoginReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	loginTask, err := task.NewTaskWithOps("weixin", task.TaskExec, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	loginTask.AddSubTask("Login Weixin channel", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(30*time.Minute))
		if agent.AgentType == constant.AppHermesAgent {
			return mgr.Run("docker", buildHermesWeixinLoginArgs(install.ContainerName)...)
		}
		return mgr.RunBashCf("docker exec %s openclaw channels login --channel openclaw-weixin", install.ContainerName)
	}, nil)
	if agent.AgentType == constant.AppHermesAgent {
		loginTask.AddSubTask("Restart Hermes-Agent container", func(t *task.Task) error {
			output, err := compose.Restart(install.GetComposePath())
			if output != "" {
				t.Log(output)
			}
			return err
		}, nil)
	}
	go func() {
		if err := loginTask.Execute(); err != nil {
			global.LOG.Errorf("login openclaw weixin channel failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) CheckPlugin(req dto.AgentPluginCheckReq) (*dto.AgentPluginStatus, error) {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	installed, err := checkPluginInstalled(install.GetPath(), req.Type)
	if err != nil {
		return nil, err
	}
	status := &dto.AgentPluginStatus{Installed: installed}
	if !installed {
		return status, nil
	}
	currentVersion, err := loadOpenclawPluginCurrentVersion(install.GetPath(), req.Type)
	if err != nil {
		global.LOG.Errorf("load openclaw plugin current version failed: %v", err)
		return status, nil
	}
	status.CurrentVersion = currentVersion
	if !req.CheckLatest {
		return status, nil
	}
	latestVersion, err := loadOpenclawPluginLatestVersion(install.ContainerName, req.Type)
	if err != nil {
		global.LOG.Errorf("load openclaw plugin latest version failed: %v", err)
		return status, nil
	}
	status.LatestVersion = latestVersion
	if currentVersion != "" && latestVersion != "" {
		status.Upgradable = common.CompareVersion(latestVersion, currentVersion)
	}
	return status, nil
}

func (a AgentService) ApproveChannelPairing(req dto.AgentChannelPairingApproveReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		output, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout(
			"docker",
			buildHermesDockerExecArgs(install.ContainerName, "pairing", "approve", req.Type, req.PairingCode)...,
		)
		if err != nil {
			return err
		}
		return validateHermesPairingApproveOutput(output)
	}
	if req.AccountID != "" {
		return cmd.RunDefaultBashCf(
			"docker exec %s openclaw pairing approve %s %q --account %q",
			install.ContainerName,
			req.Type,
			req.PairingCode,
			req.AccountID,
		)
	}
	return cmd.RunDefaultBashCf(
		"docker exec %s openclaw pairing approve %s %q",
		install.ContainerName,
		req.Type,
		req.PairingCode,
	)
}

func extractFeishuConfig(conf map[string]interface{}) dto.AgentFeishuConfig {
	result := dto.AgentFeishuConfig{
		Enabled:        true,
		ThreadSession:  true,
		ReplyMode:      "auto",
		Streaming:      false,
		RequireMention: "true",
		GroupPolicy:    "open",
		GroupAllowFrom: []string{},
		Bots:           []dto.AgentFeishuBot{defaultFeishuBot()},
	}
	feishu := getChannelConfig(conf, "feishu")
	if len(feishu) == 0 {
		return result
	}
	result.Enabled = extractFeishuPluginEnabled(conf, extractBoolValue(feishu["enabled"], result.Enabled))
	if threadSession, ok := feishu["threadSession"].(bool); ok {
		result.ThreadSession = threadSession
	}
	if replyMode := extractStringValue(feishu["replyMode"]); replyMode != "" {
		result.ReplyMode = replyMode
	}
	if streaming, ok := feishu["streaming"].(bool); ok {
		result.Streaming = streaming
	}
	result.RequireMention = extractRequireMentionValue(feishu["requireMention"], result.RequireMention)
	if groupPolicy := extractStringValue(feishu["groupPolicy"]); groupPolicy != "" {
		result.GroupPolicy = groupPolicy
	}
	result.GroupAllowFrom = extractStringList(feishu["groupAllowFrom"])
	defaultBot := defaultFeishuBot()
	defaultBot.Enabled = extractBoolValue(feishu["enabled"], defaultBot.Enabled)
	defaultBot.Name = extractDisplayName(feishu, "", "default")
	defaultBot.AppID = extractStringValue(feishu["appId"])
	defaultBot.AppSecret = extractStringValue(feishu["appSecret"])
	if dmPolicy := extractStringValue(feishu["dmPolicy"]); dmPolicy != "" {
		defaultBot.DmPolicy = dmPolicy
	}
	if _, ok := feishu["allowFrom"]; ok {
		defaultBot.AllowFrom = extractStringList(feishu["allowFrom"])
	}
	accounts := childMap(feishu, "accounts")
	defaultAccount := childMap(accounts, "default")
	if defaultBot.Name == "Default" {
		defaultBot.Name = extractDisplayName(defaultAccount, extractStringValue(defaultAccount["botName"]), "Default")
	}
	if defaultBot.DmPolicy == "pairing" {
		if dmPolicy := extractStringValue(defaultAccount["dmPolicy"]); dmPolicy != "" {
			defaultBot.DmPolicy = dmPolicy
		}
	}
	if len(defaultBot.AllowFrom) == 0 {
		if _, ok := defaultAccount["allowFrom"]; ok {
			defaultBot.AllowFrom = extractStringList(defaultAccount["allowFrom"])
		}
	}
	baseEnabled := defaultBot.Enabled
	baseDmPolicy := defaultBot.DmPolicy
	baseAllowFrom := append([]string(nil), defaultBot.AllowFrom...)
	bots := []dto.AgentFeishuBot{defaultBot}
	for _, accountID := range sortedChildKeys(accounts) {
		if accountID == "default" {
			continue
		}
		account := childMap(accounts, accountID)
		bot := dto.AgentFeishuBot{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: accountID,
				Name:      extractDisplayName(account, extractStringValue(account["botName"]), accountID),
				Enabled:   extractBoolValue(account["enabled"], baseEnabled),
			},
			AppID:     extractStringValue(account["appId"]),
			AppSecret: extractStringValue(account["appSecret"]),
			DmPolicy:  baseDmPolicy,
			AllowFrom: append([]string(nil), baseAllowFrom...),
		}
		if dmPolicy := extractStringValue(account["dmPolicy"]); dmPolicy != "" {
			bot.DmPolicy = dmPolicy
		}
		if _, ok := account["allowFrom"]; ok {
			bot.AllowFrom = extractStringList(account["allowFrom"])
		}
		bots = append(bots, bot)
	}
	result.Bots = bots
	return result
}

func setFeishuConfig(conf map[string]interface{}, config dto.AgentFeishuConfig) {
	channels := ensureChildMap(conf, "channels")
	feishu := ensureChildMap(channels, "feishu")
	defaultBot := getDefaultFeishuBot(config.Bots)
	feishu["enabled"] = defaultBot.Enabled
	feishu["threadSession"] = config.ThreadSession
	feishu["replyMode"] = config.ReplyMode
	feishu["streaming"] = config.Streaming
	if config.RequireMention == "open" {
		feishu["requireMention"] = "open"
	} else {
		feishu["requireMention"] = config.RequireMention == "true"
	}
	feishu["groupPolicy"] = config.GroupPolicy
	if config.GroupPolicy == "allowlist" {
		feishu["groupAllowFrom"] = append([]string(nil), config.GroupAllowFrom...)
	} else {
		delete(feishu, "groupAllowFrom")
	}
	feishu["appId"] = defaultBot.AppID
	feishu["appSecret"] = defaultBot.AppSecret
	feishu["dmPolicy"] = defaultBot.DmPolicy
	if defaultBot.Name != "" && defaultBot.Name != "Default" {
		feishu["name"] = defaultBot.Name
	} else {
		delete(feishu, "name")
	}
	if defaultBot.DmPolicy == "open" {
		feishu["allowFrom"] = []string{"*"}
	} else if defaultBot.DmPolicy == "allowlist" {
		feishu["allowFrom"] = append([]string(nil), defaultBot.AllowFrom...)
	} else {
		delete(feishu, "allowFrom")
	}
	delete(feishu, "botName")
	delete(feishu, "connectionMode")
	delete(feishu, "domain")
	delete(feishu, "webhookPath")
	delete(feishu, "reactionNotifications")
	delete(feishu, "typingIndicator")
	delete(feishu, "resolveSenderNames")
	delete(feishu, "defaultAccount")
	accounts := make(map[string]interface{}, len(config.Bots))
	for _, bot := range config.Bots {
		if bot.AccountID == "default" || bot.IsDefault {
			continue
		}
		account := map[string]interface{}{
			"enabled":   bot.Enabled,
			"name":      bot.Name,
			"appId":     bot.AppID,
			"appSecret": bot.AppSecret,
		}
		if bot.DmPolicy != "" {
			account["dmPolicy"] = bot.DmPolicy
		}
		if bot.DmPolicy == "open" {
			account["allowFrom"] = []string{"*"}
		} else if bot.DmPolicy == "allowlist" {
			account["allowFrom"] = append([]string(nil), bot.AllowFrom...)
		}
		accounts[bot.AccountID] = account
	}
	if len(accounts) > 0 {
		feishu["accounts"] = accounts
	} else {
		delete(feishu, "accounts")
	}
}

func extractFeishuPluginEnabled(conf map[string]interface{}, defaultValue bool) bool {
	plugins := childMap(conf, "plugins")
	entries := childMap(plugins, "entries")
	lark := childMap(entries, "openclaw-lark")
	return extractBoolValue(lark["enabled"], defaultValue)
}

func setFeishuPluginEnabled(conf map[string]interface{}, enabled bool) {
	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	lark := ensureChildMap(entries, "openclaw-lark")
	lark["enabled"] = enabled
}

func extractTelegramConfig(conf map[string]interface{}) dto.AgentTelegramConfig {
	result := dto.AgentTelegramConfig{
		Enabled:        true,
		DmPolicy:       "pairing",
		AllowFrom:      []string{},
		RequireMention: false,
		GroupPolicy:    "open",
		GroupAllowFrom: []string{},
		Streaming:      "partial",
	}
	telegram := getChannelConfig(conf, "telegram")
	if len(telegram) == 0 {
		return result
	}
	if enabled, ok := telegram["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy := extractStringValue(telegram["dmPolicy"]); dmPolicy != "" {
		result.DmPolicy = dmPolicy
	}
	result.AllowFrom = extractStringList(telegram["allowFrom"])
	if groupPolicy := extractStringValue(telegram["groupPolicy"]); groupPolicy != "" {
		result.GroupPolicy = groupPolicy
	}
	result.RequireMention = result.GroupPolicy == "allowlist"
	result.GroupAllowFrom = extractStringList(telegram["groupAllowFrom"])
	result.Proxy = extractStringValue(telegram["proxy"])
	if streaming := extractStringValue(telegram["streaming"]); streaming != "" {
		result.Streaming = streaming
	}
	accounts := childMap(telegram, "accounts")
	if len(accounts) == 0 {
		botToken := extractStringValue(telegram["botToken"])
		if botToken != "" {
			accounts["default"] = map[string]interface{}{
				"enabled":     extractBoolValue(telegram["enabled"], true),
				"botToken":    botToken,
				"dmPolicy":    result.DmPolicy,
				"groupPolicy": result.GroupPolicy,
				"streaming":   result.Streaming,
			}
		}
	}
	bots := make([]dto.AgentTelegramBot, 0, len(accounts))
	for _, accountID := range sortedChildKeys(accounts) {
		account := childMap(accounts, accountID)
		bots = append(bots, dto.AgentTelegramBot{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: accountID,
				Name:      extractDisplayName(account, accountID, accountID),
				Enabled:   extractBoolValue(account["enabled"], true),
			},
			BotToken:    extractStringValue(account["botToken"]),
			DmPolicy:    extractStringValue(account["dmPolicy"]),
			GroupPolicy: extractStringValue(account["groupPolicy"]),
			Streaming:   extractStringValue(account["streaming"]),
		})
	}
	result.DefaultAccount = normalizeDefaultAccount(extractStringValue(telegram["defaultAccount"]), getTelegramBotAccountIDs(bots))
	setTelegramDefaultFlags(bots, result.DefaultAccount)
	result.Bots = bots
	return result
}

func setTelegramConfig(conf map[string]interface{}, config dto.AgentTelegramConfig) {
	channels := ensureChildMap(conf, "channels")
	telegram := ensureChildMap(channels, "telegram")
	defaultAccount := normalizeDefaultAccount(config.DefaultAccount, getTelegramBotAccountIDs(config.Bots))
	effectiveEnabled := config.Enabled && hasEnabledBots(config.Bots)
	telegram["enabled"] = effectiveEnabled
	telegram["dmPolicy"] = config.DmPolicy
	telegram["groupPolicy"] = config.GroupPolicy
	telegram["defaultAccount"] = defaultAccount
	if config.RequireMention {
		telegram["groupPolicy"] = "allowlist"
	} else if config.GroupPolicy == "allowlist" {
		telegram["groupPolicy"] = "allowlist"
	}
	if config.DmPolicy == "open" {
		telegram["allowFrom"] = []string{"*"}
	} else if config.DmPolicy == "allowlist" {
		telegram["allowFrom"] = append([]string(nil), config.AllowFrom...)
	} else {
		delete(telegram, "allowFrom")
	}
	if config.GroupPolicy == "allowlist" {
		telegram["groupAllowFrom"] = append([]string(nil), config.GroupAllowFrom...)
	} else {
		delete(telegram, "groupAllowFrom")
	}
	if config.Proxy != "" {
		telegram["proxy"] = config.Proxy
	} else {
		delete(telegram, "proxy")
	}
	telegram["streaming"] = config.Streaming
	accounts := make(map[string]interface{}, len(config.Bots))
	for _, bot := range config.Bots {
		account := map[string]interface{}{
			"enabled":     bot.Enabled,
			"name":        bot.Name,
			"botToken":    bot.BotToken,
			"dmPolicy":    bot.DmPolicy,
			"groupPolicy": bot.GroupPolicy,
			"streaming":   bot.Streaming,
		}
		if bot.DmPolicy == "open" {
			account["allowFrom"] = []string{"*"}
		}
		accounts[bot.AccountID] = account
	}
	telegram["accounts"] = accounts
	delete(telegram, "botToken")
}

func extractDiscordConfig(conf map[string]interface{}) dto.AgentDiscordConfig {
	result := dto.AgentDiscordConfig{Enabled: true, DmPolicy: "pairing", AllowFrom: []string{}, RequireMention: false, GroupPolicy: "open"}
	discord := getChannelConfig(conf, "discord")
	if len(discord) == 0 {
		return result
	}
	if enabled, ok := discord["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy := extractStringValue(discord["dmPolicy"]); dmPolicy != "" {
		result.DmPolicy = dmPolicy
	} else if dm := childMap(discord, "dm"); dm != nil {
		if policy := extractStringValue(dm["policy"]); policy != "" {
			result.DmPolicy = policy
		}
	}
	if groupPolicy := extractStringValue(discord["groupPolicy"]); groupPolicy != "" {
		result.GroupPolicy = groupPolicy
	}
	result.RequireMention = result.GroupPolicy == "allowlist"
	result.AllowFrom = extractStringList(discord["allowFrom"])
	result.Proxy = extractStringValue(discord["proxy"])
	accounts := childMap(discord, "accounts")
	if len(accounts) == 0 {
		token := extractStringValue(discord["token"])
		if token != "" {
			accounts["default"] = map[string]interface{}{
				"enabled": extractBoolValue(discord["enabled"], true),
				"token":   token,
			}
		}
	}
	bots := make([]dto.AgentDiscordBot, 0, len(accounts))
	for _, accountID := range sortedChildKeys(accounts) {
		account := childMap(accounts, accountID)
		bots = append(bots, dto.AgentDiscordBot{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: accountID,
				Name:      extractDisplayName(account, accountID, accountID),
				Enabled:   extractBoolValue(account["enabled"], true),
			},
			Token: extractStringValue(account["token"]),
		})
	}
	result.DefaultAccount = normalizeDefaultAccount(extractStringValue(discord["defaultAccount"]), getDiscordBotAccountIDs(bots))
	setDiscordDefaultFlags(bots, result.DefaultAccount)
	result.Bots = bots
	return result
}

func setDiscordConfig(conf map[string]interface{}, config dto.AgentDiscordConfig) {
	channels := ensureChildMap(conf, "channels")
	discord := ensureChildMap(channels, "discord")
	defaultAccount := normalizeDefaultAccount(config.DefaultAccount, getDiscordBotAccountIDs(config.Bots))
	effectiveEnabled := config.Enabled && hasEnabledBots(config.Bots)
	discord["enabled"] = effectiveEnabled
	discord["dmPolicy"] = config.DmPolicy
	discord["groupPolicy"] = config.GroupPolicy
	discord["defaultAccount"] = defaultAccount
	if config.RequireMention {
		discord["groupPolicy"] = "allowlist"
	} else if config.GroupPolicy == "allowlist" {
		discord["groupPolicy"] = "allowlist"
	}
	if config.DmPolicy == "open" {
		discord["allowFrom"] = []string{"*"}
	} else if config.DmPolicy == "allowlist" {
		discord["allowFrom"] = append([]string(nil), config.AllowFrom...)
	} else {
		delete(discord, "allowFrom")
	}
	if config.Proxy != "" {
		discord["proxy"] = config.Proxy
	} else {
		delete(discord, "proxy")
	}
	accounts := make(map[string]interface{}, len(config.Bots))
	for _, bot := range config.Bots {
		accounts[bot.AccountID] = map[string]interface{}{
			"enabled":     bot.Enabled,
			"name":        bot.Name,
			"token":       bot.Token,
			"groupPolicy": config.GroupPolicy,
		}
	}
	discord["accounts"] = accounts
	delete(discord, "token")
	delete(discord, "dm")
}

func extractQQBotConfig(conf map[string]interface{}) dto.AgentQQBotConfig {
	result := dto.AgentQQBotConfig{Enabled: true}
	qqbot := getChannelConfig(conf, "qqbot")
	if len(qqbot) == 0 {
		result.Bots = []dto.AgentQQBotBot{defaultQQBot()}
		return result
	}
	if enabled, ok := qqbot["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	bots := []dto.AgentQQBotBot{
		{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: "default",
				Name:      extractStringValue(qqbot["name"]),
				Enabled:   extractBoolValue(qqbot["enabled"], true),
				IsDefault: true,
			},
			AppID:        extractStringValue(qqbot["appId"]),
			ClientSecret: extractStringValue(qqbot["clientSecret"]),
			AllowFrom:    extractStringList(qqbot["allowFrom"]),
			SystemPrompt: extractStringValue(qqbot["systemPrompt"]),
		},
	}
	if bots[0].Name == "" {
		bots[0].Name = "Default"
	}
	for _, accountID := range sortedChildKeys(childMap(qqbot, "accounts")) {
		account := childMap(childMap(qqbot, "accounts"), accountID)
		bots = append(bots, dto.AgentQQBotBot{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: accountID,
				Name:      extractDisplayName(account, accountID, accountID),
				Enabled:   extractBoolValue(account["enabled"], true),
			},
			AppID:        extractStringValue(account["appId"]),
			ClientSecret: extractStringValue(account["clientSecret"]),
			AllowFrom:    extractStringList(account["allowFrom"]),
			SystemPrompt: extractStringValue(account["systemPrompt"]),
		})
	}
	result.Bots = bots
	return result
}

func extractWecomConfig(conf map[string]interface{}) dto.AgentWecomConfig {
	result := dto.AgentWecomConfig{
		Enabled:        true,
		DmPolicy:       "pairing",
		AllowFrom:      []string{},
		GroupPolicy:    "open",
		GroupAllowFrom: []string{},
	}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	wecom, ok := channels["wecom"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := wecom["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy := extractStringValue(wecom["dmPolicy"]); dmPolicy != "" {
		result.DmPolicy = dmPolicy
	}
	result.AllowFrom = extractStringList(wecom["allowFrom"])
	if groupPolicy := extractStringValue(wecom["groupPolicy"]); groupPolicy != "" {
		result.GroupPolicy = groupPolicy
	}
	result.GroupAllowFrom = extractStringList(wecom["groupAllowFrom"])
	result.BotID = extractStringValue(wecom["botId"])
	result.Secret = extractStringValue(wecom["secret"])
	return result
}

func extractDingTalkConfig(conf map[string]interface{}) dto.AgentDingTalkConfig {
	result := dto.AgentDingTalkConfig{
		Enabled:                         true,
		DmPolicy:                        "open",
		GroupPolicy:                     "disabled",
		AllowFrom:                       []string{},
		GroupAllowFrom:                  []string{},
		SeparateSessionByConversation:   true,
		GroupSessionScope:               "group",
		SharedMemoryAcrossConversations: false,
		AsyncMode:                       false,
		AckText:                         "任务已接收，处理中...",
	}
	dingtalk := getChannelConfig(conf, "dingtalk-connector")
	if len(dingtalk) == 0 {
		return result
	}
	if enabled, ok := dingtalk["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy := extractStringValue(dingtalk["dmPolicy"]); dmPolicy != "" {
		if dmPolicy == "pairing" {
			result.DmPolicy = "open"
		} else {
			result.DmPolicy = dmPolicy
		}
	}
	if groupPolicy := extractStringValue(dingtalk["groupPolicy"]); groupPolicy != "" {
		result.GroupPolicy = groupPolicy
	}
	result.AllowFrom = extractStringList(dingtalk["allowFrom"])
	result.GroupAllowFrom = extractStringList(dingtalk["groupAllowFrom"])
	if separateSessionByConversation, ok := dingtalk["separateSessionByConversation"].(bool); ok {
		result.SeparateSessionByConversation = separateSessionByConversation
	}
	if groupSessionScope := extractStringValue(dingtalk["groupSessionScope"]); groupSessionScope != "" {
		result.GroupSessionScope = groupSessionScope
	}
	if sharedMemoryAcrossConversations, ok := dingtalk["sharedMemoryAcrossConversations"].(bool); ok {
		result.SharedMemoryAcrossConversations = sharedMemoryAcrossConversations
	}
	if asyncMode, ok := dingtalk["asyncMode"].(bool); ok {
		result.AsyncMode = asyncMode
	}
	if ackText := extractStringValue(dingtalk["ackText"]); ackText != "" {
		result.AckText = ackText
	}
	accounts := childMap(dingtalk, "accounts")
	if len(accounts) == 0 {
		clientID := extractStringValue(dingtalk["clientId"])
		clientSecret := extractStringValue(dingtalk["clientSecret"])
		if clientID != "" || clientSecret != "" {
			accounts["default"] = map[string]interface{}{
				"enabled":      extractBoolValue(dingtalk["enabled"], true),
				"clientId":     clientID,
				"clientSecret": clientSecret,
			}
		}
	}
	bots := make([]dto.AgentDingTalkBot, 0, len(accounts))
	for _, accountID := range sortedChildKeys(accounts) {
		account := childMap(accounts, accountID)
		bots = append(bots, dto.AgentDingTalkBot{
			AgentChannelBotBase: dto.AgentChannelBotBase{
				AccountID: accountID,
				Name:      accountID,
				Enabled:   extractBoolValue(account["enabled"], true),
			},
			ClientID:     extractStringValue(account["clientId"]),
			ClientSecret: extractStringValue(account["clientSecret"]),
		})
	}
	result.Bots = bots
	return result
}

func setWecomConfig(conf map[string]interface{}, config dto.AgentWecomConfig) {
	channels := ensureChildMap(conf, "channels")
	wecom := ensureChildMap(channels, "wecom")
	wecom["enabled"] = config.Enabled
	wecom["botId"] = config.BotID
	wecom["secret"] = config.Secret
	wecom["dmPolicy"] = config.DmPolicy
	wecom["groupPolicy"] = config.GroupPolicy
	wecom["sendThinkingMessage"] = true
	if config.DmPolicy == "allowlist" {
		wecom["allowFrom"] = append([]string(nil), config.AllowFrom...)
	} else {
		delete(wecom, "allowFrom")
	}
	if config.GroupPolicy == "allowlist" {
		wecom["groupAllowFrom"] = append([]string(nil), config.GroupAllowFrom...)
	} else {
		delete(wecom, "groupAllowFrom")
	}

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	wecomEntry := ensureChildMap(entries, "wecom-openclaw-plugin")
	wecomEntry["enabled"] = config.Enabled
}

func setDingTalkConfig(conf map[string]interface{}, config dto.AgentDingTalkConfig) {
	channels := ensureChildMap(conf, "channels")
	dingtalk := ensureChildMap(channels, "dingtalk-connector")
	effectiveEnabled := config.Enabled && hasEnabledBots(config.Bots)
	dingtalk["enabled"] = effectiveEnabled
	dingtalk["dmPolicy"] = config.DmPolicy
	dingtalk["groupPolicy"] = config.GroupPolicy
	dingtalk["separateSessionByConversation"] = config.SeparateSessionByConversation
	dingtalk["groupSessionScope"] = config.GroupSessionScope
	dingtalk["sharedMemoryAcrossConversations"] = config.SharedMemoryAcrossConversations
	dingtalk["asyncMode"] = config.AsyncMode
	dingtalk["ackText"] = config.AckText
	delete(dingtalk, "gatewayToken")
	switch config.DmPolicy {
	case "open":
		dingtalk["allowFrom"] = []string{"*"}
	case "allowlist":
		dingtalk["allowFrom"] = append([]string(nil), config.AllowFrom...)
	default:
		delete(dingtalk, "allowFrom")
	}
	switch config.GroupPolicy {
	case "open":
		dingtalk["groupAllowFrom"] = []string{"*"}
	case "allowlist":
		dingtalk["groupAllowFrom"] = append([]string(nil), config.GroupAllowFrom...)
	default:
		delete(dingtalk, "groupAllowFrom")
	}
	accounts := make(map[string]interface{}, len(config.Bots))
	for _, bot := range config.Bots {
		accounts[bot.AccountID] = map[string]interface{}{
			"enabled":      bot.Enabled,
			"clientId":     bot.ClientID,
			"clientSecret": bot.ClientSecret,
		}
	}
	dingtalk["accounts"] = accounts
	delete(dingtalk, "clientId")
	delete(dingtalk, "clientSecret")

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	dingtalkEntry := ensureChildMap(entries, "dingtalk-connector")
	dingtalkEntry["enabled"] = effectiveEnabled

	gateway := ensureChildMap(conf, "gateway")
	httpMap := ensureChildMap(gateway, "http")
	endpoints := ensureChildMap(httpMap, "endpoints")
	chatCompletions := ensureChildMap(endpoints, "chatCompletions")
	chatCompletions["enabled"] = true
}

func setQQBotConfig(conf map[string]interface{}, config dto.AgentQQBotConfig) {
	channels := ensureChildMap(conf, "channels")
	qqbot := ensureChildMap(channels, "qqbot")
	defaultBot := getDefaultQQBot(config.Bots)
	effectiveEnabled := config.Enabled && hasEnabledBots(config.Bots)
	delete(qqbot, "dmPolicy")
	qqbot["enabled"] = effectiveEnabled
	qqbot["appId"] = defaultBot.AppID
	qqbot["clientSecret"] = defaultBot.ClientSecret
	qqbot["name"] = defaultBot.Name
	if len(defaultBot.AllowFrom) > 0 {
		qqbot["allowFrom"] = append([]string(nil), defaultBot.AllowFrom...)
	} else {
		delete(qqbot, "allowFrom")
	}
	if defaultBot.SystemPrompt != "" {
		qqbot["systemPrompt"] = defaultBot.SystemPrompt
	} else {
		delete(qqbot, "systemPrompt")
	}

	accounts := make(map[string]interface{}, len(config.Bots))
	for _, bot := range config.Bots {
		if bot.AccountID == "default" || bot.IsDefault {
			continue
		}
		account := map[string]interface{}{
			"enabled":      bot.Enabled,
			"name":         bot.Name,
			"appId":        bot.AppID,
			"clientSecret": bot.ClientSecret,
		}
		if len(bot.AllowFrom) > 0 {
			account["allowFrom"] = append([]string(nil), bot.AllowFrom...)
		}
		if bot.SystemPrompt != "" {
			account["systemPrompt"] = bot.SystemPrompt
		}
		accounts[bot.AccountID] = account
	}
	qqbot["accounts"] = accounts

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	delete(entries, "qqbot")
	qqbotEntry := ensureChildMap(entries, "openclaw-qqbot")
	qqbotEntry["enabled"] = effectiveEnabled
}

func appendPluginAllow(conf map[string]interface{}, pluginID string) {
	plugins := ensureChildMap(conf, "plugins")
	allow := make([]string, 0, 4)
	seen := map[string]struct{}{}
	switch values := plugins["allow"].(type) {
	case []interface{}:
		for _, value := range values {
			text, ok := value.(string)
			if !ok || text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}
			allow = append(allow, text)
		}
	}
	if _, ok := seen[pluginID]; ok {
		plugins["allow"] = allow
		return
	}
	plugins["allow"] = append(allow, pluginID)
}

func buildOpenclawPluginInstallScript(spec, pluginID string) string {
	return fmt.Sprintf(
		"set -e; workdir=%s/%s; rm -rf \"$workdir\"; mkdir -p \"$workdir\"; cd \"$workdir\"; npm pack --silent %q >/dev/null 2>&1; pkg=$(find \"$workdir\" -maxdepth 1 -type f -name '*.tgz' | head -n 1); printf '%%s\\n' \"$pkg\"; openclaw plugins install \"$pkg\" --dangerously-force-unsafe-install; rm -rf \"$workdir\"",
		openclawPluginPackageTmpDir,
		pluginID,
		spec,
	)
}

func buildOpenclawPluginUninstallScript(pluginID string) string {
	return fmt.Sprintf(
		"set +e; printf 'yes\\n' | openclaw plugins uninstall %s; code=$?; if [ \"$code\" -eq 137 ]; then exit 0; fi; exit \"$code\"",
		pluginID,
	)
}

func resolvePluginMeta(pluginType string) (string, string, error) {
	switch pluginType {
	case "qqbot":
		return "@tencent-connect/openclaw-qqbot", "openclaw-qqbot", nil
	case "feishu":
		return "@larksuite/openclaw-lark", "openclaw-lark", nil
	case "wecom":
		return "@wecom/wecom-openclaw-plugin", "wecom-openclaw-plugin", nil
	case "dingtalk":
		return "@dingtalk-real-ai/dingtalk-connector", "dingtalk-connector", nil
	case "weixin":
		return "@tencent-weixin/openclaw-weixin", "openclaw-weixin", nil
	default:
		return "", "", fmt.Errorf("unsupported plugin type")
	}
}

func checkPluginInstalled(installPath, pluginType string) (bool, error) {
	packagePath, err := resolveOpenclawPluginPackagePath(installPath, pluginType)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(packagePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func loadOpenclawPluginCurrentVersion(installPath, pluginType string) (string, error) {
	packagePath, err := resolveOpenclawPluginPackagePath(installPath, pluginType)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return "", err
	}
	var pkg openclawPluginPackage
	if err := json.Unmarshal(content, &pkg); err != nil {
		return "", err
	}
	return pkg.Version, nil
}

func loadOpenclawPluginLatestVersion(containerName, pluginType string) (string, error) {
	spec, _, err := resolvePluginMeta(pluginType)
	if err != nil {
		return "", err
	}
	output, err := runDockerExecWithStdout(20*time.Second, containerName, "npm", "view", spec, "version", "--json")
	if err != nil {
		return "", err
	}
	var version string
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &version); err == nil {
		return version, nil
	}
	return strings.Trim(strings.TrimSpace(output), `"`), nil
}

func resolveOpenclawPluginPackagePath(installPath, pluginType string) (string, error) {
	_, pluginID, err := resolvePluginMeta(pluginType)
	if err != nil {
		return "", err
	}
	if installPath == "" {
		return "", buserr.New("ErrRecordNotFound")
	}
	return path.Join(installPath, "data", "conf", "extensions", pluginID, "package.json"), nil
}

func cleanupOpenclawPluginConfig(conf map[string]interface{}, pluginType string) {
	channels, _ := conf["channels"].(map[string]interface{})
	plugins, _ := conf["plugins"].(map[string]interface{})
	entries, _ := plugins["entries"].(map[string]interface{})

	switch pluginType {
	case "feishu":
		delete(channels, "feishu")
		delete(entries, "openclaw-lark")
		delete(entries, "feishu")
	case "qqbot":
		delete(channels, "qqbot")
		delete(entries, "openclaw-qqbot")
		delete(entries, "qqbot")
	case "wecom":
		delete(channels, "wecom")
		delete(entries, "wecom-openclaw-plugin")
	case "dingtalk":
		delete(channels, "dingtalk-connector")
		delete(entries, "dingtalk-connector")
		gateway := ensureChildMap(conf, "gateway")
		httpMap := ensureChildMap(gateway, "http")
		endpoints := ensureChildMap(httpMap, "endpoints")
		chatCompletions := ensureChildMap(endpoints, "chatCompletions")
		chatCompletions["enabled"] = false
	case "weixin":
		delete(channels, "weixin")
		delete(entries, "openclaw-weixin")
	}
}

func getChannelConfig(conf map[string]interface{}, channel string) map[string]interface{} {
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return nil
	}
	channelMap, _ := channels[channel].(map[string]interface{})
	return channelMap
}

func childMap(parent map[string]interface{}, key string) map[string]interface{} {
	if parent == nil {
		return map[string]interface{}{}
	}
	if value, ok := parent[key].(map[string]interface{}); ok {
		return value
	}
	return map[string]interface{}{}
}

func extractStringValue(value interface{}) string {
	text, _ := value.(string)
	return text
}

func extractRequireMentionValue(value interface{}, defaultValue string) string {
	switch typed := value.(type) {
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case string:
		if typed != "" {
			return typed
		}
	}
	return defaultValue
}

func extractBoolValue(value interface{}, defaultValue bool) bool {
	result, ok := value.(bool)
	if !ok {
		return defaultValue
	}
	return result
}

func sortedChildKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func extractDisplayName(values map[string]interface{}, fallback string, accountID string) string {
	name := extractStringValue(values["name"])
	if name != "" {
		return name
	}
	if fallback != "" {
		return fallback
	}
	if accountID == "default" {
		return "Default"
	}
	return accountID
}

func normalizeDefaultAccount(defaultAccount string, accountIDs []string) string {
	if len(accountIDs) == 0 {
		return ""
	}
	for _, accountID := range accountIDs {
		if accountID == defaultAccount {
			return defaultAccount
		}
	}
	return accountIDs[0]
}

func getTelegramBotAccountIDs(bots []dto.AgentTelegramBot) []string {
	accountIDs := make([]string, 0, len(bots))
	for _, bot := range bots {
		accountIDs = append(accountIDs, bot.AccountID)
	}
	return accountIDs
}

func getDiscordBotAccountIDs(bots []dto.AgentDiscordBot) []string {
	accountIDs := make([]string, 0, len(bots))
	for _, bot := range bots {
		accountIDs = append(accountIDs, bot.AccountID)
	}
	return accountIDs
}

func setTelegramDefaultFlags(bots []dto.AgentTelegramBot, defaultAccount string) {
	for i := range bots {
		bots[i].IsDefault = bots[i].AccountID == defaultAccount
	}
}

func setDiscordDefaultFlags(bots []dto.AgentDiscordBot, defaultAccount string) {
	for i := range bots {
		bots[i].IsDefault = bots[i].AccountID == defaultAccount
	}
}

type enabledBot interface {
	IsEnabled() bool
}

func hasEnabledBots[T enabledBot](bots []T) bool {
	for _, bot := range bots {
		if bot.IsEnabled() {
			return true
		}
	}
	return false
}

func defaultQQBot() dto.AgentQQBotBot {
	return dto.AgentQQBotBot{
		AgentChannelBotBase: dto.AgentChannelBotBase{
			AccountID: "default",
			Name:      "Default",
			Enabled:   true,
			IsDefault: true,
		},
	}
}

func defaultFeishuBot() dto.AgentFeishuBot {
	return dto.AgentFeishuBot{
		AgentChannelBotBase: dto.AgentChannelBotBase{
			AccountID: "default",
			Name:      "Default",
			Enabled:   true,
			IsDefault: true,
		},
		DmPolicy:  "pairing",
		AllowFrom: []string{},
	}
}

func getDefaultFeishuBot(bots []dto.AgentFeishuBot) dto.AgentFeishuBot {
	for _, bot := range bots {
		if bot.IsDefault || bot.AccountID == "default" {
			bot.IsDefault = true
			if bot.Name == "" {
				bot.Name = "Default"
			}
			return bot
		}
	}
	if len(bots) > 0 {
		bot := bots[0]
		bot.IsDefault = true
		if bot.AccountID == "" {
			bot.AccountID = "default"
		}
		if bot.Name == "" {
			bot.Name = "Default"
		}
		return bot
	}
	return defaultFeishuBot()
}

func getDefaultQQBot(bots []dto.AgentQQBotBot) dto.AgentQQBotBot {
	for _, bot := range bots {
		if bot.IsDefault || bot.AccountID == "default" {
			bot.IsDefault = true
			if bot.Name == "" {
				bot.Name = "Default"
			}
			return bot
		}
	}
	if len(bots) > 0 {
		bot := bots[0]
		bot.IsDefault = true
		if bot.Name == "" {
			bot.Name = "Default"
		}
		return bot
	}
	return defaultQQBot()
}
