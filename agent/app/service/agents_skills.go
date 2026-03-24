package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type openclawSkillsList struct {
	Skills []openclawSkillListItem `json:"skills"`
}

type openclawSkillListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Bundled     bool   `json:"bundled"`
	Disabled    bool   `json:"disabled"`
}

type openclawSkillInfo struct {
	SkillKey string `json:"skillKey"`
}

func (a AgentService) ListSkills(req dto.AgentSkillsReq) ([]dto.AgentSkillItem, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("copaw does not support skills")
	}
	status, err := checkContainerStatus(install.ContainerName)
	if err != nil {
		return nil, err
	}
	if status != "running" {
		return nil, fmt.Errorf("container %s is not running, please check and retry", install.ContainerName)
	}
	output, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(
		"docker exec %s openclaw skills list --json 2>&1",
		30*time.Second,
		install.ContainerName,
	)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	return parseOpenclawSkillsList(output)
}

func (a AgentService) UpdateSkill(req dto.AgentSkillUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType != constant.AppOpenclaw {
		return fmt.Errorf("copaw does not support skills")
	}
	status, err := checkContainerStatus(install.ContainerName)
	if err != nil {
		return err
	}
	if status != "running" {
		return fmt.Errorf("container %s is not running, please check and retry", install.ContainerName)
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	skillKey, err := getOpenclawSkillKey(install.ContainerName, req.Name)
	if err != nil {
		return err
	}
	setOpenclawSkillEnabled(conf, skillKey, req.Enabled)
	return writeOpenclawConfigRaw(agent.ConfigPath, conf)
}

func parseOpenclawSkillsList(output string) ([]dto.AgentSkillItem, error) {
	var payload openclawSkillsList
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &payload); err != nil {
		return nil, err
	}
	items := make([]dto.AgentSkillItem, 0, len(payload.Skills))
	for _, item := range payload.Skills {
		items = append(items, dto.AgentSkillItem{
			Name:        item.Name,
			Description: item.Description,
			Source:      item.Source,
			Bundled:     item.Bundled,
			Disabled:    item.Disabled,
		})
	}
	return items, nil
}

func getOpenclawSkillKey(containerName, name string) (string, error) {
	output, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(
		"docker exec %s openclaw skills info %q --json 2>&1",
		30*time.Second,
		containerName,
		name,
	)
	if err != nil {
		return "", err
	}
	return parseOpenclawSkillKey(name, output)
}

func parseOpenclawSkillKey(name, output string) (string, error) {
	var payload openclawSkillInfo
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &payload); err != nil {
		return "", err
	}
	if payload.SkillKey == "" {
		return "", fmt.Errorf("skill %s does not have a skillKey", name)
	}
	return payload.SkillKey, nil
}

func setOpenclawSkillEnabled(conf map[string]interface{}, skillKey string, enabled bool) {
	skills := ensureChildMap(conf, "skills")
	entries := ensureChildMap(skills, "entries")
	entry := ensureChildMap(entries, skillKey)
	entry["enabled"] = enabled
}
