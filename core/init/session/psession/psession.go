package psession

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

type SessionUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type sessionItem struct {
	CreatedAt time.Time
	CSRFToken string
	User      SessionUser
	ExpiredAt time.Time
}

type PSession struct {
	mu              sync.RWMutex
	sessions        map[string]sessionItem
	cleanupCursor   uint64
	lastFullCleanup time.Time
}

const maxSessionEntries = 64

func NewPSession(_ string) *PSession {
	return &PSession{
		sessions: make(map[string]sessionItem),
	}
}

func (p *PSession) Get(c *gin.Context) (SessionUser, error) {
	var result SessionUser

	sessionID, err := c.Cookie(constant.SessionName)
	if err != nil || sessionID == "" {
		return result, errors.New("ErrSessionDataNotFound")
	}

	p.mu.RLock()
	item, ok := p.sessions[sessionID]
	p.mu.RUnlock()
	if !ok {
		return result, errors.New("ErrSessionDataNotFound")
	}
	if !item.ExpiredAt.IsZero() && time.Now().After(item.ExpiredAt) {
		p.mu.Lock()
		delete(p.sessions, sessionID)
		p.mu.Unlock()
		return result, errors.New("ErrSessionDataNotFound")
	}
	return item.User, nil
}

func (p *PSession) Set(c *gin.Context, user SessionUser, secure bool, ttlSeconds int) error {
	return p.set(c, user, secure, ttlSeconds, false)
}

func (p *PSession) SetFresh(c *gin.Context, user SessionUser, secure bool, ttlSeconds int) error {
	return p.set(c, user, secure, ttlSeconds, true)
}

func (p *PSession) set(c *gin.Context, user SessionUser, secure bool, ttlSeconds int, forceNew bool) error {
	sessionID, err := c.Cookie(constant.SessionName)
	if forceNew {
		if err == nil && sessionID != "" {
			p.mu.Lock()
			delete(p.sessions, sessionID)
			p.mu.Unlock()
		}
		sessionID = ""
	}
	if err != nil || sessionID == "" {
		sessionID, err = generateSessionID()
		if err != nil {
			return err
		}
	}

	expiredAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	createdAt := time.Now()
	csrfToken := ""

	p.mu.Lock()
	if existing, ok := p.sessions[sessionID]; ok {
		if !existing.CreatedAt.IsZero() {
			createdAt = existing.CreatedAt
		}
		csrfToken = existing.CSRFToken
	}
	if csrfToken == "" {
		csrfToken, err = generateSessionID()
		if err != nil {
			p.mu.Unlock()
			return err
		}
	}
	p.sessions[sessionID] = sessionItem{
		CreatedAt: createdAt,
		CSRFToken: csrfToken,
		User:      user,
		ExpiredAt: expiredAt,
	}
	p.evictOverflowLocked(sessionID)
	p.mu.Unlock()
	p.cleanupExpiredOnWrite()

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.SessionName, sessionID, ttlSeconds, "/", "", secure, true)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.CSRFTokenName, csrfToken, ttlSeconds, "/", "", secure, false)
	return nil
}

func (p *PSession) evictOverflowLocked(currentSessionID string) {
	if maxSessionEntries <= 0 || len(p.sessions) <= maxSessionEntries {
		return
	}

	for len(p.sessions) > maxSessionEntries {
		oldestID := ""
		var oldestItem sessionItem
		for sessionID, item := range p.sessions {
			if sessionID == currentSessionID {
				continue
			}
			if oldestID == "" || item.CreatedAt.Before(oldestItem.CreatedAt) {
				oldestID = sessionID
				oldestItem = item
			}
		}
		if oldestID == "" {
			return
		}
		delete(p.sessions, oldestID)
	}
}

func (p *PSession) RefreshIfNeeded(c *gin.Context, user SessionUser, secure bool, ttlSeconds int) (bool, error) {
	sessionID, err := c.Cookie(constant.SessionName)
	if err != nil || sessionID == "" {
		return false, p.Set(c, user, secure, ttlSeconds)
	}

	p.mu.RLock()
	item, ok := p.sessions[sessionID]
	p.mu.RUnlock()
	if !ok {
		return false, p.Set(c, user, secure, ttlSeconds)
	}
	if !item.ExpiredAt.IsZero() && time.Now().After(item.ExpiredAt) {
		p.mu.Lock()
		delete(p.sessions, sessionID)
		p.mu.Unlock()
		return false, errors.New("ErrSessionDataNotFound")
	}
	return true, p.Set(c, user, secure, ttlSeconds)
}

func (p *PSession) Delete(c *gin.Context) error {
	sessionID, err := c.Cookie(constant.SessionName)
	if err == nil && sessionID != "" {
		p.mu.Lock()
		delete(p.sessions, sessionID)
		p.mu.Unlock()
	}
	return nil
}

func (p *PSession) CheckCSRFToken(c *gin.Context, token string) bool {
	sessionID, err := c.Cookie(constant.SessionName)
	if err != nil || sessionID == "" || token == "" {
		return false
	}

	p.mu.RLock()
	item, ok := p.sessions[sessionID]
	p.mu.RUnlock()
	if !ok {
		return false
	}
	if !item.ExpiredAt.IsZero() && time.Now().After(item.ExpiredAt) {
		p.mu.Lock()
		delete(p.sessions, sessionID)
		p.mu.Unlock()
		return false
	}
	return item.CSRFToken == token
}

func (p *PSession) Clean() error {
	p.mu.Lock()
	p.sessions = make(map[string]sessionItem)
	p.lastFullCleanup = time.Time{}
	p.mu.Unlock()
	return nil
}

func generateSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (p *PSession) cleanupExpiredOnWrite() {
	const (
		sampleSize             = 32
		fullCleanupThreshold   = 1024
		fullCleanupMinInterval = time.Minute
	)

	now := time.Now()

	p.mu.RLock()
	size := len(p.sessions)
	lastFullCleanup := p.lastFullCleanup
	p.mu.RUnlock()

	if size == 0 {
		return
	}
	if size >= fullCleanupThreshold && now.Sub(lastFullCleanup) >= fullCleanupMinInterval {
		p.cleanupExpiredAll(now)
		return
	}
	p.cleanupExpiredSample(now, sampleSize)
}

func (p *PSession) cleanupExpiredSample(now time.Time, limit int) {
	if limit <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	total := len(p.sessions)
	if total == 0 {
		return
	}
	start := int(atomic.AddUint64(&p.cleanupCursor, uint64(limit)) % uint64(total))
	checked := 0

	idx := 0
	for key, item := range p.sessions {
		if idx < start {
			idx++
			continue
		}
		if !item.ExpiredAt.IsZero() && now.After(item.ExpiredAt) {
			delete(p.sessions, key)
		}
		checked++
		idx++
		if checked >= limit {
			break
		}
	}
	if checked < limit {
		for key, item := range p.sessions {
			if checked >= limit {
				break
			}
			if !item.ExpiredAt.IsZero() && now.After(item.ExpiredAt) {
				delete(p.sessions, key)
			}
			checked++
		}
	}
}

func (p *PSession) cleanupExpiredAll(now time.Time) {
	p.mu.Lock()
	for key, item := range p.sessions {
		if !item.ExpiredAt.IsZero() && now.After(item.ExpiredAt) {
			delete(p.sessions, key)
		}
	}
	p.lastFullCleanup = now
	p.mu.Unlock()
}
