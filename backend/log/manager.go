package log

import (
	"github.com/robfig/cron/v3"
	"path"
	"sync"
	"time"
)

type manager struct {
	startAt time.Time
	fire    chan string
	cr      *cron.Cron
	context chan int
	wg      sync.WaitGroup
	lock    sync.Mutex
}

func (m *manager) Fire() chan string {
	return m.fire
}

func (m *manager) Close() {
	close(m.context)
	m.cr.Stop()
}

func NewManager(c *Config) (Manager, error) {
	m := &manager{
		startAt: time.Now(),
		cr:      cron.New(),
		fire:    make(chan string),
		context: make(chan int),
		wg:      sync.WaitGroup{},
	}

	if _, err := m.cr.AddFunc(c.RollingTimePattern, func() {
		m.fire <- m.GenLogFileName(c)
	}); err != nil {
		return nil, err
	}
	m.cr.Start()

	return m, nil
}

func (m *manager) GenLogFileName(c *Config) (filename string) {
	m.lock.Lock()
	filename = path.Join(c.LogPath, c.FileName+"-"+m.startAt.Format(c.TimeTagFormat)) + c.LogSuffix
	m.startAt = time.Now()
	m.lock.Unlock()
	return
}
