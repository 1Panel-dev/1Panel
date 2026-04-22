package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	fileShareCodePattern        = regexp.MustCompile(`^[A-Za-z0-9]{10,16}$`)
	fileShareLimiterCleanupLock sync.Mutex
	fileShareLimiterCleanupAt   time.Time
	publicIPLimiters            sync.Map
	publicCodeLimiters          sync.Map
)

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func FileSharePublicAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := strings.TrimSpace(c.Query("code"))

		if code != "" && !fileShareCodePattern.MatchString(code) {
			helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrFileShareInvalid", buserr.New("ErrFileShareInvalid"))
			return
		}

		ip := c.ClientIP()
		if !allowLimiter(&publicIPLimiters, "ip:"+ip, rate.Every(time.Second), 20) {
			helper.ErrorWithDetail(c, http.StatusTooManyRequests, "ErrFileShareRateLimit", buserr.New("ErrFileShareRateLimit"))
			return
		}
		if code != "" && (strings.HasSuffix(c.Request.URL.Path, "/share/check") || strings.HasSuffix(c.Request.URL.Path, "/share/download")) {
			if !allowLimiter(&publicCodeLimiters, "code:"+ip+":"+code, rate.Every(5*time.Second), 4) {
				helper.ErrorWithDetail(c, http.StatusTooManyRequests, "ErrFileShareRateLimit", buserr.New("ErrFileShareRateLimit"))
				return
			}
		}

		maybeCleanupLimiters()
		c.Next()
	}
}

func allowLimiter(store *sync.Map, key string, refill rate.Limit, burst int) bool {
	now := time.Now()
	value, _ := store.LoadOrStore(key, &visitorLimiter{
		limiter:  rate.NewLimiter(refill, burst),
		lastSeen: now,
	})
	item := value.(*visitorLimiter)
	item.lastSeen = now
	return item.limiter.Allow()
}

func maybeCleanupLimiters() {
	fileShareLimiterCleanupLock.Lock()
	defer fileShareLimiterCleanupLock.Unlock()

	now := time.Now()
	if !fileShareLimiterCleanupAt.IsZero() && now.Sub(fileShareLimiterCleanupAt) < 10*time.Minute {
		return
	}
	fileShareLimiterCleanupAt = now
	cleanupLimiterMap(&publicIPLimiters, now)
	cleanupLimiterMap(&publicCodeLimiters, now)
}

func cleanupLimiterMap(store *sync.Map, now time.Time) {
	store.Range(func(key, value any) bool {
		item, ok := value.(*visitorLimiter)
		if ok && now.Sub(item.lastSeen) > 30*time.Minute {
			store.Delete(key)
		}
		return true
	})
}
