package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/terminal"
)

const (
	settingTerminalSessionKeepAlive = "TerminalSessionKeepAlive"
	settingTerminalSessionMaxPinned = "TerminalSessionMaxPinned"
	settingTerminalSessionBuffer    = "TerminalSessionBuffer"
)

// KeepAlive is stored in minutes, Buffer in KB.
const (
	defaultTerminalSessionKeepAlive = 30
	minTerminalSessionKeepAlive     = 0
	maxTerminalSessionKeepAlive     = 1440

	defaultTerminalSessionMaxPinned = 10
	minTerminalSessionMaxPinned     = 1
	maxTerminalSessionMaxPinned     = 50

	defaultTerminalSessionBuffer = 256
	minTerminalSessionBuffer     = 64
	maxTerminalSessionBuffer     = 4096
)

type TerminalSessionService struct{}

type ITerminalSessionService interface {
	List(owner string) ([]dto.TerminalSessionInfo, error)
	Pin(owner string, req dto.TerminalSessionPin) error
	Close(owner, id string) error
}

// NewITerminalSessionService wires the session manager to the settings table.
func NewITerminalSessionService() ITerminalSessionService {
	terminal.DefaultManager.SetConfigProvider(loadTerminalSessionConfig)
	return &TerminalSessionService{}
}

func (u *TerminalSessionService) List(owner string) ([]dto.TerminalSessionInfo, error) {
	infos := terminal.DefaultManager.List(owner)

	list := make([]dto.TerminalSessionInfo, 0, len(infos))
	for _, info := range infos {
		item := dto.TerminalSessionInfo{
			ID:           info.ID,
			Kind:         info.Kind,
			HostID:       info.HostID,
			Title:        info.Title,
			Pinned:       info.Pinned,
			Attached:     info.Attached,
			CreatedAt:    info.CreatedAt,
			LastActiveAt: info.LastActiveAt,
		}
		if !info.Attached && !info.DetachedAt.IsZero() {
			item.DetachedAt = new(info.DetachedAt)
			if info.Pinned && !info.ExpiresAt.IsZero() {
				item.ExpiresAt = new(info.ExpiresAt)
			}
		}
		list = append(list, item)
	}
	return list, nil
}

func (u *TerminalSessionService) Pin(owner string, req dto.TerminalSessionPin) error {
	return terminalSessionErr(terminal.DefaultManager.Pin(strings.TrimSpace(req.ID), owner, req.Pinned))
}

func (u *TerminalSessionService) Close(owner, id string) error {
	return terminalSessionErr(terminal.DefaultManager.Close(strings.TrimSpace(id), owner))
}

// terminalSessionErr translates manager errors into business errors.
func terminalSessionErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, terminal.ErrSessionNotFound):
		return buserr.New("ErrTerminalSessionNotFound")
	case errors.Is(err, terminal.ErrPinDisabled):
		return buserr.New("ErrTerminalSessionDisabled")
	case errors.Is(err, terminal.ErrPinLimit):
		return buserr.WithDetail("ErrTerminalSessionLimit", terminal.DefaultManager.Config().MaxPinned, err)
	}
	return err
}

// loadTerminalSessionConfig reads keep-alive settings, clamping to accepted ranges.
func loadTerminalSessionConfig() terminal.Config {
	values, err := settingRepo.GetValuesByKeys([]string{
		settingTerminalSessionKeepAlive,
		settingTerminalSessionMaxPinned,
		settingTerminalSessionBuffer,
	})
	if err != nil {
		values = nil
	}
	keepAlive := terminalSessionSetting(values, settingTerminalSessionKeepAlive, defaultTerminalSessionKeepAlive, minTerminalSessionKeepAlive, maxTerminalSessionKeepAlive)
	maxPinned := terminalSessionSetting(values, settingTerminalSessionMaxPinned, defaultTerminalSessionMaxPinned, minTerminalSessionMaxPinned, maxTerminalSessionMaxPinned)
	buffer := terminalSessionSetting(values, settingTerminalSessionBuffer, defaultTerminalSessionBuffer, minTerminalSessionBuffer, maxTerminalSessionBuffer)
	return terminal.Config{
		KeepAlive: time.Duration(keepAlive) * time.Minute,
		MaxPinned: maxPinned,
		RingSize:  buffer * 1024,
	}
}

func terminalSessionSetting(values map[string]string, key string, fallback, lower, upper int) int {
	number, err := strconv.Atoi(strings.TrimSpace(values[key]))
	if err != nil {
		return fallback
	}
	return min(max(number, lower), upper)
}
