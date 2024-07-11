package task

import (
	"context"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type ActionFunc func() error
type RollbackFunc func()

type Task struct {
	Name      string
	Logger    *log.Logger
	SubTasks  []*SubTask
	Rollbacks []RollbackFunc
	logFile   *os.File
}

type SubTask struct {
	Name     string
	Retry    int
	Timeout  time.Duration
	Action   ActionFunc
	Rollback RollbackFunc
	Error    error
}

func NewTask(name string, taskType string) (*Task, error) {
	logPath := path.Join(constant.LogDir, taskType)
	//TODO 增加插入到日志表的逻辑
	file, err := os.OpenFile(logPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger := log.New(file, "", log.LstdFlags)
	return &Task{Name: name, logFile: file, Logger: logger}, nil
}

func (t *Task) AddSubTask(name string, action ActionFunc, rollback RollbackFunc) {
	subTask := &SubTask{Name: name, Retry: 0, Timeout: 10 * time.Minute, Action: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (t *Task) AddSubTaskWithOps(name string, action ActionFunc, rollback RollbackFunc, retry int, timeout time.Duration) {
	subTask := &SubTask{Name: name, Retry: retry, Timeout: timeout, Action: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (s *SubTask) Execute(logger *log.Logger) bool {
	logger.Printf(i18n.GetWithName("SubTaskStart", s.Name))
	for i := 0; i < s.Retry+1; i++ {
		if i > 0 {
			logger.Printf(i18n.GetWithName("TaskRetry", strconv.Itoa(i)))
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- s.Action()
		}()

		select {
		case <-ctx.Done():
			logger.Printf(i18n.GetWithName("TaskTimeout", s.Name))
		case err := <-done:
			if err != nil {
				s.Error = err
				logger.Printf(i18n.GetWithNameAndErr("TaskFailed", s.Name, err))
			} else {
				logger.Printf(i18n.GetWithName("TaskSuccess", s.Name))
				return true
			}
		}

		if i == s.Retry {
			if s.Rollback != nil {
				s.Rollback()
			}
		}
		time.Sleep(1 * time.Second)
	}
	if s.Error != nil {
		s.Error = fmt.Errorf(i18n.GetWithName("TaskFailed", s.Name))
	}
	return false
}

func (t *Task) Execute() error {
	t.Logger.Printf(i18n.GetWithName("TaskStart", t.Name))
	var err error
	for _, subTask := range t.SubTasks {
		if subTask.Execute(t.Logger) {
			if subTask.Rollback != nil {
				t.Rollbacks = append(t.Rollbacks, subTask.Rollback)
			}
		} else {
			err = subTask.Error
			for _, rollback := range t.Rollbacks {
				rollback()
			}
			break
		}
	}
	t.Logger.Printf(i18n.GetWithName("TaskEnd", t.Name))
	_ = t.logFile.Close()
	return err
}
