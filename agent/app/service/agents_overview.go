package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	openclawCronJobsRelativePath     = "data/conf/cron/jobs.json"
	openclawSessionStoreRelativeGlob = "data/conf/agents/*/sessions/sessions.json"
	openclawSessionFilesRelativeGlob = "data/conf/agents/*/sessions/*.jsonl"
)

func (a AgentService) GetOverview(req dto.AgentOverviewReq) (*dto.AgentOverview, error) {
	agent, install, conf, err := a.loadOpenclawAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
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

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		skillCount, err := loadOpenclawOverviewSkillStats(install.ContainerName)
		if err == nil {
			overview.Snapshot.SkillCount = skillCount
		}
	}()
	go func() {
		defer wg.Done()
		sessionCount, err := loadOpenclawOverviewSessionCount(install.GetPath())
		if err == nil {
			overview.Snapshot.SessionCount = sessionCount
		}
	}()
	go func() {
		defer wg.Done()
		jobCount, err := loadOpenclawOverviewJobCount(install.GetPath())
		if err == nil {
			overview.Snapshot.JobCount = jobCount
		}
	}()
	wg.Wait()

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
		"docker exec %s openclaw gateway call skills.status --json 2>&1",
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

func loadOpenclawOverviewSessionCount(installPath string) (int, error) {
	sessionStorePaths, err := filepath.Glob(filepath.Join(installPath, openclawSessionStoreRelativeGlob))
	if err != nil {
		return 0, err
	}
	total := 0
	for _, sessionStorePath := range sessionStorePaths {
		count, err := loadOpenclawSessionCountFromFile(sessionStorePath)
		if err != nil {
			continue
		}
		total += count
	}
	if total > 0 {
		return total, nil
	}
	sessionFiles, err := filepath.Glob(filepath.Join(installPath, openclawSessionFilesRelativeGlob))
	if err != nil {
		return 0, err
	}
	return len(sessionFiles), nil
}

func loadOpenclawSessionCountFromFile(sessionStorePath string) (int, error) {
	content, err := os.ReadFile(sessionStorePath)
	if err != nil {
		return 0, err
	}
	return parseOpenclawSessionCount(string(content))
}

func loadOpenclawOverviewJobCount(installPath string) (int, error) {
	content, err := os.ReadFile(filepath.Join(installPath, openclawCronJobsRelativePath))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return parseOpenclawCronCount(string(content))
}

func parseOpenclawSessionCount(output string) (int, error) {
	if strings.TrimSpace(output) == "" {
		return 0, nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &payload); err != nil {
		return 0, err
	}
	return countOpenclawSessions(payload), nil
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

func countOpenclawSessions(payload interface{}) int {
	switch value := payload.(type) {
	case []interface{}:
		return len(value)
	case map[string]interface{}:
		if count, ok := value["count"].(float64); ok {
			return int(count)
		}
		if sessions, ok := value["sessions"]; ok {
			return countOpenclawSessions(sessions)
		}
		count := 0
		for _, item := range value {
			switch typed := item.(type) {
			case map[string]interface{}:
				if looksLikeOpenclawSession(typed) {
					count++
				}
			case string:
				if strings.HasSuffix(strings.TrimSpace(typed), ".jsonl") {
					count++
				}
			}
		}
		return count
	default:
		return 0
	}
}

func looksLikeOpenclawSession(item map[string]interface{}) bool {
	if _, ok := item["sessionFile"]; ok {
		return true
	}
	if _, ok := item["agentId"]; ok {
		return true
	}
	if _, ok := item["lastActivityTs"]; ok {
		return true
	}
	if _, ok := item["updatedAt"]; ok {
		return true
	}
	return false
}
