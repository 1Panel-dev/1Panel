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

const (
	openclawCronJobsPath = "/home/node/.openclaw/cron/jobs.json"
)

func (a AgentService) GetOverview(req dto.AgentOverviewReq) (*dto.AgentOverview, error) {
	agent, install, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("copaw does not support overview")
	}

	overview := &dto.AgentOverview{
		Snapshot: dto.AgentOverviewSnapshot{
			ContainerStatus: install.Status,
			AppVersion:      install.Version,
			DefaultModel:    extractOpenclawDefaultModel(conf),
			ChannelCount:    countOpenclawConfiguredChannels(conf),
		},
	}
	if overview.Snapshot.DefaultModel == "" {
		overview.Snapshot.DefaultModel = agent.Model
	}
	if install.Status != constant.StatusRunning {
		return overview, nil
	}

	skillCount, err := loadOpenclawOverviewSkillStats(install.ContainerName)
	if err == nil {
		overview.Snapshot.SkillCount = skillCount
	}

	sessionCount, err := loadOpenclawOverviewSessionCount(install.ContainerName)
	if err == nil {
		overview.Snapshot.SessionCount = sessionCount
	}

	jobCount, err := loadOpenclawOverviewJobCount(install.ContainerName)
	if err == nil {
		overview.Snapshot.JobCount = jobCount
	}

	return overview, nil
}

func extractOpenclawDefaultModel(conf map[string]interface{}) string {
	agents, ok := conf["agents"].(map[string]interface{})
	if !ok {
		return ""
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		return ""
	}
	model, ok := defaults["model"].(map[string]interface{})
	if !ok {
		return ""
	}
	primary, _ := model["primary"].(string)
	return strings.TrimSpace(primary)
}

func countOpenclawConfiguredChannels(conf map[string]interface{}) int {
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return 0
	}
	count := 0
	for _, value := range channels {
		channel, ok := value.(map[string]interface{})
		if !ok || len(channel) == 0 {
			continue
		}
		count++
	}
	return count
}

func loadOpenclawOverviewSkillStats(containerName string) (int, error) {
	output, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(
		"docker exec %s openclaw skills list --json 2>&1",
		30*time.Second,
		containerName,
	)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(output) == "" {
		return 0, nil
	}
	skills, err := parseOpenclawSkillsList(output)
	if err != nil {
		return 0, err
	}
	return len(skills), nil
}

func loadOpenclawOverviewSessionCount(containerName string) (int, error) {
	output, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(
		"docker exec %s openclaw sessions --all-agents --json",
		20*time.Second,
		containerName,
	)
	if err != nil {
		return 0, err
	}
	return parseOpenclawSessionCount(output)
}

func loadOpenclawOverviewJobCount(containerName string) (int, error) {
	script := fmt.Sprintf(`if [ -f %q ]; then cat %q; fi`, openclawCronJobsPath, openclawCronJobsPath)
	output, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(
		"docker exec %s sh -c %q",
		20*time.Second,
		containerName,
		script,
	)
	if err != nil {
		return 0, err
	}
	return parseOpenclawCronCount(output)
}

func parseOpenclawSessionCount(output string) (int, error) {
	if strings.TrimSpace(output) == "" {
		return 0, nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &payload); err != nil {
		return 0, err
	}
	switch value := payload.(type) {
	case []interface{}:
		return len(value), nil
	case map[string]interface{}:
		if count, ok := value["count"].(float64); ok {
			return int(count), nil
		}
		if sessions, ok := value["sessions"].([]interface{}); ok {
			return len(sessions), nil
		}
	}
	return 0, nil
}

func parseOpenclawCronCount(output string) (int, error) {
	if strings.TrimSpace(output) == "" {
		return 0, nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &payload); err != nil {
		return 0, err
	}
	switch value := payload.(type) {
	case []interface{}:
		return len(value), nil
	case map[string]interface{}:
		if total, ok := value["total"].(float64); ok {
			return int(total), nil
		}
		if jobs, ok := value["jobs"].([]interface{}); ok {
			return len(jobs), nil
		}
		return len(value), nil
	default:
		return 0, nil
	}
}
