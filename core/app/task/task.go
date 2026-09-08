package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ActionFunc func(*Task) error
type ContextActionFunc func(context.Context, *Task) error
type RollbackFunc func(*Task)

var ErrExecutionUnconfirmed = errors.New("task execution termination could not be confirmed")

type Task struct {
	TaskCtx           context.Context
	Name              string
	TaskID            string
	Logger            *logrus.Logger
	SubTasks          []*SubTask
	Rollbacks         []RollbackFunc
	logFile           *os.File
	taskRepo          repo.ITaskRepo
	Task              *model.Task
	ParentID          string
	CancelWhenTimeout bool
}

type SubTask struct {
	RootTask          *Task
	Name              string
	StepAlias         string
	Retry             int
	Timeout           time.Duration
	Action            ActionFunc
	ContextAction     ContextActionFunc
	ShouldRetry       func(error) bool
	RetryBackoff      time.Duration
	Rollback          RollbackFunc
	Error             error
	IgnoreErr         bool
	CancelWhenTimeout bool
}

const (
	TaskInstall        = "TaskInstall"
	TaskCreate         = "TaskCreate"
	TaskUpdate         = "TaskUpdate"
	TaskDelete         = "TaskDelete"
	TaskUpgrade        = "TaskUpgrade"
	TaskAddNode        = "TaskAddNode"
	TaskSync           = "TaskSync"
	TaskSyncForNode    = "TaskSyncForNode"
	TaskRsync          = "TaskRsync"
	TaskInstallCluster = "TaskInstallCluster"
	TaskCreateCluster  = "TaskCreateCluster"
	TaskBackup         = "TaskBackup"
	TaskPush           = "TaskPush"
	TaskExec           = "TaskExec"
)

const (
	TaskScopeSystem     = "System"
	TaskScopeScript     = "ScriptLibrary"
	TaskScopeNodeFile   = "NodeFile"
	TaskScopeNodeSSL    = "NodeSSL"
	TaskScopeAppBackup  = "AppBackup"
	TaskScopeCluster    = "Cluster"
	TaskScopeAppInstall = "AppInstallTask"
	TaskScopeAI         = "AI"
	TaskScopeVm         = "VirtualMachine"
)

func GetTaskName(resourceName, operate, scope string) string {
	return fmt.Sprintf("%s%s [%s]", i18n.GetMsgByKey(operate), i18n.GetMsgByKey(scope), resourceName)
}

func NewTaskWithOps(resourceName, operate, scope, taskID string, resourceID uint) (*Task, error) {
	return NewTask(GetTaskName(resourceName, operate, scope), operate, scope, taskID, resourceID)
}

func NewTask(name, operate, taskScope, taskID string, resourceID uint) (*Task, error) {
	if taskID == "" {
		taskID = uuid.New().String()
	}
	logItem := path.Join(global.CONF.Base.InstallDir, "1panel/log/task")
	logDir := path.Join(logItem, taskScope)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err = os.MkdirAll(logDir, constant.DirPerm); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	logPath := path.Join(logItem, taskScope, taskID+".log")
	logger := logrus.New()
	logger.SetFormatter(&SimpleFormatter{})
	logFile, err := os.OpenFile(logPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, constant.FilePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger.SetOutput(logFile)
	taskModel := &model.Task{
		ID:         taskID,
		Name:       name,
		Type:       taskScope,
		LogFile:    logPath,
		Status:     constant.StatusExecuting,
		ResourceID: resourceID,
		Operate:    operate,
	}
	taskRepo := repo.NewITaskRepo()
	task := &Task{TaskID: taskID, Name: name, logFile: logFile, Logger: logger, taskRepo: taskRepo, Task: taskModel}
	return task, nil
}

func (t *Task) AddSubTask(name string, action ActionFunc, rollback RollbackFunc) {
	subTask := &SubTask{RootTask: t, Name: name, Retry: 0, Timeout: 10 * time.Minute, Action: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (t *Task) AddSubTaskWithAlias(key string, action ActionFunc, rollback RollbackFunc) {
	subTask := &SubTask{RootTask: t, Name: i18n.GetMsgByKey(key), StepAlias: key, Retry: 0, Timeout: 10 * time.Minute, Action: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (t *Task) AddSubTaskWithOps(name string, action ActionFunc, rollback RollbackFunc, retry int, timeout time.Duration) {
	subTask := &SubTask{RootTask: t, Name: name, Retry: retry, Timeout: timeout, Action: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (t *Task) AddSubTaskWithContext(name string, action ContextActionFunc, rollback RollbackFunc, retry int, timeout time.Duration) *SubTask {
	subTask := &SubTask{RootTask: t, Name: name, Retry: retry, Timeout: timeout, ContextAction: action, Rollback: rollback}
	t.SubTasks = append(t.SubTasks, subTask)
	return subTask
}

func (t *Task) Context() context.Context {
	if t.TaskCtx != nil {
		return t.TaskCtx
	}
	return context.Background()
}

func (t *Task) AddSubTaskWithIgnoreErr(name string, action ActionFunc) {
	subTask := &SubTask{RootTask: t, Name: name, Retry: 0, Timeout: 10 * time.Minute, Action: action, Rollback: nil, IgnoreErr: true}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (s *SubTask) Execute() error {
	if s.Timeout < 0 || s.Retry < 0 || (s.Action == nil && s.ContextAction == nil) {
		return fmt.Errorf("invalid subtask execution options")
	}
	subTaskName := s.Name
	if s.Name == "" {
		subTaskName = i18n.GetMsgByKey("SubTask")
	}
	s.RootTask.LogStart(subTaskName)
	var err error
	attempted := false
	for i := 0; i < s.Retry+1; i++ {
		if err = s.RootTask.Context().Err(); err != nil {
			break
		}
		if i > 0 {
			s.RootTask.Log(i18n.GetWithName("TaskRetry", strconv.Itoa(i)))
		}
		var started bool
		err, started = s.executeAttempt(subTaskName)
		attempted = attempted || started
		if err == nil {
			s.RootTask.Log(i18n.GetWithName("SubTaskSuccess", subTaskName))
			return nil
		}
		s.RootTask.Log(i18n.GetWithNameAndErr("SubTaskFailed", subTaskName, err))
		if i == s.Retry || !s.canRetry(err) {
			break
		}
		timer := time.NewTimer(s.retryDelay(i))
		select {
		case <-s.RootTask.Context().Done():
			timer.Stop()
			err = s.RootTask.Context().Err()
		case <-timer.C:
		}
	}
	if parentErr := s.RootTask.Context().Err(); parentErr != nil {
		err = errors.Join(parentErr, err)
	}
	if attempted && s.Rollback != nil && !errors.Is(err, ErrExecutionUnconfirmed) {
		s.Rollback(s.RootTask)
	}
	return err
}

func (s *SubTask) executeAttempt(name string) (error, bool) {
	parent := s.RootTask.Context()
	var ctx context.Context
	var cancel context.CancelFunc
	if s.Timeout == 0 {
		ctx, cancel = context.WithCancel(parent)
	} else {
		ctx, cancel = context.WithTimeout(parent, s.Timeout)
	}
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err, false
	}
	done := make(chan error, 1)
	go func() {
		if s.ContextAction != nil {
			done <- s.ContextAction(ctx, s.RootTask)
		} else {
			done <- s.Action(s.RootTask)
		}
	}()
	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			s.RootTask.Log(i18n.GetWithName("TaskTimeout", name))
		}
		err = <-done
	}
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return errors.Join(fmt.Errorf("%s: %w", i18n.GetWithName("TaskTimeout", name), ctx.Err()), err), true
		}
		return errors.Join(ctx.Err(), err), true
	}
	return err, true
}

func (s *SubTask) canRetry(err error) bool {
	if s.RootTask.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, ErrExecutionUnconfirmed) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) && (s.CancelWhenTimeout || s.ShouldRetry == nil) {
		return false
	}
	return s.ShouldRetry == nil || s.ShouldRetry(err)
}

func (s *SubTask) retryDelay(attempt int) time.Duration {
	if s.RetryBackoff <= 0 {
		return time.Second
	}
	delay := min(s.RetryBackoff, 30*time.Second)
	for i := 0; i < attempt && delay < 30*time.Second; i++ {
		delay = min(delay*2, 30*time.Second)
	}
	return delay
}

func (t *Task) updateTask(task *model.Task) {
	_ = t.taskRepo.Update(context.Background(), task)
}

func (t *Task) Execute() error {
	if err := t.taskRepo.Save(context.Background(), t.Task); err != nil {
		if t.logFile != nil {
			_ = t.logFile.Close()
		}
		return err
	}
	var err error
	t.Log(i18n.GetWithName("TaskStart", t.Name))
	for _, subTask := range t.SubTasks {
		subTask.CancelWhenTimeout = t.CancelWhenTimeout
		t.Task.CurrentStep = subTask.StepAlias
		t.updateTask(t.Task)
		if err = subTask.Execute(); err == nil {
			if subTask.Rollback != nil {
				t.Rollbacks = append(t.Rollbacks, subTask.Rollback)
			}
		} else {
			if subTask.IgnoreErr && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, ErrExecutionUnconfirmed) {
				err = nil
				continue
			}
			t.Task.ErrorMsg = err.Error()
			t.Task.Status = constant.StatusFailed
			if errors.Is(err, context.Canceled) && !errors.Is(err, ErrExecutionUnconfirmed) {
				t.Task.Status = constant.StatusCanceled
			}
			if !errors.Is(err, ErrExecutionUnconfirmed) {
				for _, rollback := range t.Rollbacks {
					rollback(t)
				}
			}
			t.updateTask(t.Task)
			break
		}
	}
	if t.Task.Status == constant.StatusExecuting {
		t.Task.Status = constant.StatusSuccess
		t.Log(i18n.GetWithName("TaskSuccess", t.Name))
	} else {
		t.Log(i18n.GetWithName("TaskFailed", t.Name))
	}
	t.Log("[TASK-END]")
	t.Task.EndAt = time.Now()
	t.updateTask(t.Task)
	if t.logFile != nil {
		_ = t.logFile.Close()
	}
	return err
}

func (t *Task) DeleteLogFile() {
	_ = os.Remove(t.Task.LogFile)
}

func (t *Task) LogWithStatus(msg string, err error) {
	if err != nil {
		t.Logger.Print(i18n.GetWithNameAndErr("FailedStatus", msg, err))
	} else {
		t.Logger.Print(i18n.GetWithName("SuccessStatus", msg))
	}
}

func (t *Task) Log(msg string) {
	t.Logger.Print(msg)
}

func (t *Task) Logf(format string, v ...any) {
	t.Logger.Printf(format, v...)
}

func (t *Task) LogFailed(msg string) {
	t.Logger.Print(msg + i18n.GetMsgByKey("Failed"))
}

func (t *Task) LogFailedWithErr(msg string, err error) {
	t.Logger.Printf("%s %s : %s", msg, i18n.GetMsgByKey("Failed"), err.Error())
}

func (t *Task) LogSuccess(msg string) {
	t.Logger.Print(msg + i18n.GetMsgByKey("Success"))
}
func (t *Task) LogSuccessF(format string, v ...any) {
	t.Logger.Print(fmt.Sprintf(format, v...) + i18n.GetMsgByKey("Success"))
}

func (t *Task) LogStart(msg string) {
	t.Logger.Printf("%s%s", i18n.GetMsgByKey("Start"), msg)
}

func (t *Task) LogWithOps(operate, msg string) {
	t.Logger.Printf("%s%s", i18n.GetMsgByKey(operate), msg)
}

func (t *Task) LogSuccessWithOps(operate, msg string) {
	t.Logger.Printf("%s%s%s", i18n.GetMsgByKey(operate), msg, i18n.GetMsgByKey("Success"))
}

func (t *Task) LogFailedWithOps(operate, msg string, err error) {
	t.Logger.Printf("%s%s%s : %s ", i18n.GetMsgByKey(operate), msg, i18n.GetMsgByKey("Failed"), err.Error())
}

type SimpleFormatter struct{}

func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	message := fmt.Sprintf("%s %s\n", timestamp, entry.Message)
	return []byte(message), nil
}
