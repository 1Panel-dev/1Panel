package service

import (
	"database/sql"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

const hermesSessionListLimit = 100

func (a AgentService) GetHermesChatSessions(req dto.AgentIDReq) ([]dto.AgentHermesChatSessionItem, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType != constant.AppHermesAgent {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	return listHermesChatSessionsFromStateDB(filepath.Join(install.GetPath(), "data", "state.db"))
}

func (a AgentService) RenameHermesChatSession(req dto.AgentHermesChatSessionRenameReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType != constant.AppHermesAgent {
		return fmt.Errorf("%s does not support", agent.AgentType)
	}

	_, err = cmd.RunDefaultWithStdoutBashC(buildHermesSessionRenameCommand(install.ContainerName, req.ID, req.Title))
	return err
}

func listHermesChatSessionsFromStateDB(stateDBPath string) ([]dto.AgentHermesChatSessionItem, error) {
	if !files.NewFileOp().Stat(stateDBPath) {
		return []dto.AgentHermesChatSessionItem{}, nil
	}

	db, err := sql.Open("sqlite", stateDBPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
SELECT
    s.id,
    COALESCE(NULLIF(TRIM(s.title), ''), s.id) AS title,
    COALESCE(s.model, '') AS model,
    COALESCE(s.message_count, 0) AS message_count,
    s.started_at,
    COALESCE(MAX(m.timestamp), s.started_at) AS last_active
FROM sessions s
LEFT JOIN messages m ON m.session_id = s.id
WHERE s.source = 'cli'
GROUP BY s.id, title, model, s.message_count, s.started_at
ORDER BY last_active DESC
LIMIT ?
`, hermesSessionListLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.AgentHermesChatSessionItem, 0, 8)
	for rows.Next() {
		var item dto.AgentHermesChatSessionItem
		var title sql.NullString
		var model sql.NullString
		var startedAt sql.NullFloat64
		var lastActive sql.NullFloat64
		if err := rows.Scan(&item.ID, &title, &model, &item.MessageCount, &startedAt, &lastActive); err != nil {
			return nil, err
		}
		item.Title = strings.TrimSpace(title.String)
		if item.Title == "" {
			item.Title = item.ID
		}
		item.Model = strings.TrimSpace(model.String)
		item.StartedAt = formatHermesSessionTimestamp(startedAt)
		item.LastActive = formatHermesSessionTimestamp(lastActive)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func formatHermesSessionTimestamp(value sql.NullFloat64) string {
	if !value.Valid || value.Float64 <= 0 {
		return ""
	}

	seconds, fraction := math.Modf(value.Float64)
	return time.Unix(int64(seconds), int64(fraction*float64(time.Second))).UTC().Format(time.RFC3339)
}

func buildHermesSessionRenameCommand(containerName string, sessionID string, title string) string {
	return fmt.Sprintf(
		"docker exec -u hermes -e HOME=/opt/data/home -e HERMES_HOME=/opt/data %s hermes sessions rename %q %q",
		containerName,
		sessionID,
		title,
	)
}
