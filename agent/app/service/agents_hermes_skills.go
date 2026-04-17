package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type hermesSkillsListEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type hermesSkillsListPayload struct {
	Skills []hermesSkillsListEntry `json:"skills"`
}

func listHermesSkills(containerName string) ([]dto.AgentSkillItem, error) {
	output, err := runHermesSkillsCommandWithStdout(2*time.Minute, containerName, "list", "--source", "all")
	if err != nil {
		return nil, err
	}
	items, err := parseHermesSkillsListOutput(output)
	if err != nil {
		return nil, err
	}
	metadata, err := readHermesSkillsListMetadata(containerName)
	if err != nil {
		return nil, err
	}
	return mergeHermesSkillsWithMetadata(items, metadata), nil
}

func searchHermesSkills(containerName, source, keyword string) ([]dto.AgentSkillSearchItem, error) {
	if source != "official" && source != "skills-sh" {
		return nil, fmt.Errorf("unsupported hermes skill source: %s", source)
	}
	output, err := runHermesSkillsCommandWithStdout(
		2*time.Minute,
		containerName,
		"search",
		keyword,
		"--source",
		source,
		"--limit",
		"20",
	)
	if err != nil {
		return nil, err
	}
	return parseHermesSkillSearchOutput(output)
}

func runHermesSkillsCommandWithStdout(timeout time.Duration, containerName string, hermesArgs ...string) (string, error) {
	args := []string{"exec", "-e", "COLUMNS=240", "-u", "hermes", containerName, "hermes", "skills"}
	args = append(args, hermesArgs...)
	return cmd.NewCommandMgr(cmd.WithTimeout(timeout)).RunWithStdout("docker", args...)
}

func readHermesSkillsListMetadata(containerName string) (map[string]hermesSkillsListEntry, error) {
	output, err := cmd.NewCommandMgr(cmd.WithTimeout(2*time.Minute)).RunWithStdout(
		"docker",
		"exec",
		"-u",
		"hermes",
		containerName,
		"python",
		"-c",
		"import sys; sys.path.insert(0, '/opt/hermes'); from tools.skills_tool import skills_list; print(skills_list())",
	)
	if err != nil {
		return nil, err
	}
	return parseHermesSkillsListMetadataOutput(output)
}

func parseHermesSkillsListOutput(output string) ([]dto.AgentSkillItem, error) {
	headers, rows, err := parseHermesTableOutput(output)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AgentSkillItem, 0, len(rows))
	for _, row := range rows {
		item := dto.AgentSkillItem{
			Name:          row[headers["Name"]],
			Category:      row[headers["Category"]],
			Source:        row[headers["Source"]],
			Trust:         row[headers["Trust"]],
			Uninstallable: row[headers["Source"]] != "builtin" && row[headers["Source"]] != "local",
		}
		items = append(items, item)
	}
	return items, nil
}

func parseHermesSkillsListMetadataOutput(output string) (map[string]hermesSkillsListEntry, error) {
	if strings.TrimSpace(output) == "" {
		return map[string]hermesSkillsListEntry{}, nil
	}
	var payload hermesSkillsListPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return nil, err
	}
	metadata := make(map[string]hermesSkillsListEntry, len(payload.Skills))
	for _, skill := range payload.Skills {
		if skill.Name == "" {
			continue
		}
		metadata[skill.Name] = skill
	}
	return metadata, nil
}

func mergeHermesSkillsWithMetadata(items []dto.AgentSkillItem, metadata map[string]hermesSkillsListEntry) []dto.AgentSkillItem {
	for i := range items {
		entry, ok := metadata[items[i].Name]
		if !ok {
			continue
		}
		items[i].Description = entry.Description
		if items[i].Category == "" && entry.Category != "" {
			items[i].Category = entry.Category
		}
	}
	return items
}

func parseHermesSkillSearchOutput(output string) ([]dto.AgentSkillSearchItem, error) {
	if strings.Contains(output, "No skills found matching your query.") {
		return []dto.AgentSkillSearchItem{}, nil
	}
	headers, rows, err := parseHermesTableOutput(output)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AgentSkillSearchItem, 0, len(rows))
	for _, row := range rows {
		identifier := row[headers["Identifier"]]
		items = append(items, dto.AgentSkillSearchItem{
			Slug:        identifier,
			Identifier:  identifier,
			Name:        row[headers["Name"]],
			Description: row[headers["Description"]],
			Source:      row[headers["Source"]],
			Trust:       row[headers["Trust"]],
		})
	}
	return items, nil
}

func parseHermesTableOutput(output string) (map[string]int, [][]string, error) {
	lines := strings.Split(strings.TrimSpace(ansiEscapePattern.ReplaceAllString(output, "")), "\n")
	var headers []string
	rows := make([][]string, 0)
	var current []string

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if (!strings.HasPrefix(line, "│") || !strings.HasSuffix(line, "│")) &&
			(!strings.HasPrefix(line, "┃") || !strings.HasSuffix(line, "┃")) {
			continue
		}
		line = strings.ReplaceAll(line, "┃", "│")
		parts := strings.Split(line, "│")
		if len(parts) < 3 {
			continue
		}
		cols := make([]string, 0, len(parts)-2)
		for _, part := range parts[1 : len(parts)-1] {
			cols = append(cols, strings.TrimSpace(part))
		}
		if len(headers) == 0 {
			headers = cols
			continue
		}
		if cols[0] != "" {
			if current != nil {
				rows = append(rows, current)
			}
			current = cols
			continue
		}
		if current == nil {
			continue
		}
		for i := range cols {
			if cols[i] == "" {
				continue
			}
			if current[i] == "" {
				current[i] = cols[i]
				continue
			}
			current[i] = current[i] + " " + cols[i]
		}
	}

	if current != nil {
		rows = append(rows, current)
	}
	if len(headers) == 0 {
		return nil, nil, fmt.Errorf("hermes skills table not found")
	}

	headerIndex := make(map[string]int, len(headers))
	for i, header := range headers {
		headerIndex[header] = i
	}
	return headerIndex, rows, nil
}
