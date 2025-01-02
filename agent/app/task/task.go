package task

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/google/uuid"
)

type ActionFunc func(*Task) error
type RollbackFunc func(*Task)

type Task struct {
	Name      string
	TaskID    string
	Logger    *log.Logger
	SubTasks  []*SubTask
	Rollbacks []RollbackFunc
	logFile   *os.File
	taskRepo  repo.ITaskRepo
	Task      *model.Task
	ParentID  string
}

type SubTask struct {
	RootTask  *Task
	Name      string
	StepAlias string
	Retry     int
	Timeout   time.Duration
	Action    ActionFunc
	Rollback  RollbackFunc
	Error     error
	IgnoreErr bool
}

const (
	TaskInstall   = "TaskInstall"
	TaskUninstall = "TaskUninstall"
	TaskCreate    = "TaskCreate"
	TaskDelete    = "TaskDelete"
	TaskUpgrade   = "TaskUpgrade"
	TaskUpdate    = "TaskUpdate"
	TaskRestart   = "TaskRestart"
	TaskBackup    = "TaskBackup"
	TaskRecover   = "TaskRecover"
	TaskRollback  = "TaskRollback"
	TaskSync      = "TaskSync"
	TaskBuild     = "TaskBuild"
	TaskPull      = "TaskPull"
	TaskPush      = "TaskPush"
	TaskHandle    = "TaskHandle"
)

const (
	TaskScopeWebsite          = "Website"
	TaskScopeApp              = "App"
	TaskScopeRuntime          = "Runtime"
	TaskScopeDatabase         = "Database"
	TaskScopeCronjob          = "Cronjob"
	TaskScopeSystem           = "System"
	TaskScopeAppStore         = "AppStore"
	TaskScopeSnapshot         = "Snapshot"
	TaskScopeContainer        = "Container"
	TaskScopeCompose          = "Compose"
	TaskScopeImage            = "Image"
	TaskScopeRuntimeExtension = "RuntimeExtension"
)

func GetTaskName(resourceName, operate, scope string) string {
	return fmt.Sprintf("%s%s [%s]", i18n.GetMsgByKey(operate), i18n.GetMsgByKey(scope), resourceName)
}

func NewTaskWithOps(resourceName, operate, scope, taskID string, resourceID uint) (*Task, error) {
	return NewTask(GetTaskName(resourceName, operate, scope), operate, scope, taskID, resourceID)
}

func CheckTaskIsExecuting(name string) error {
	taskRepo := repo.NewITaskRepo()
	task, _ := taskRepo.GetFirst(taskRepo.WithByStatus(constant.StatusExecuting), repo.WithByName(name))
	if task.ID != "" {
		return buserr.New("TaskIsExecuting")
	}
	return nil
}

func NewTask(name, operate, taskScope, taskID string, resourceID uint) (*Task, error) {
	if taskID == "" {
		taskID = uuid.New().String()
	}
	logDir := path.Join(constant.LogDir, taskScope)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err = os.MkdirAll(logDir, constant.DirPerm); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	logPath := path.Join(constant.LogDir, taskScope, taskID+".log")
	file, err := os.OpenFile(logPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, constant.FilePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger := log.New(file, "", log.LstdFlags)
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
	task := &Task{Name: name, logFile: file, Logger: logger, taskRepo: taskRepo, Task: taskModel}
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

func (t *Task) AddSubTaskWithIgnoreErr(name string, action ActionFunc) {
	subTask := &SubTask{RootTask: t, Name: name, Retry: 0, Timeout: 10 * time.Minute, Action: action, Rollback: nil, IgnoreErr: true}
	t.SubTasks = append(t.SubTasks, subTask)
}

func (s *SubTask) Execute() error {
	subTaskName := s.Name
	if s.Name == "" {
		subTaskName = i18n.GetMsgByKey("SubTask")
	}
	var err error
	for i := 0; i < s.Retry+1; i++ {
		if i > 0 {
			s.RootTask.Log(i18n.GetWithName("TaskRetry", strconv.Itoa(i)))
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- s.Action(s.RootTask)
		}()

		select {
		case <-ctx.Done():
			s.RootTask.Log(i18n.GetWithName("TaskTimeout", subTaskName))
		case err = <-done:
			if err != nil {
				s.RootTask.Log(i18n.GetWithNameAndErr("SubTaskFailed", subTaskName, err))
			} else {
				s.RootTask.Log(i18n.GetWithName("SubTaskSuccess", subTaskName))
				return nil
			}
		}

		if i == s.Retry {
			if s.Rollback != nil {
				s.Rollback(s.RootTask)
			}
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

func (t *Task) updateTask(task *model.Task) {
	_ = t.taskRepo.Update(context.Background(), task)
}

func (t *Task) Execute() error {
	if err := t.taskRepo.Save(context.Background(), t.Task); err != nil {
		return err
	}
	var err error
	t.Log(i18n.GetWithName("TaskStart", t.Name))
	for _, subTask := range t.SubTasks {
		t.Task.CurrentStep = subTask.StepAlias
		t.updateTask(t.Task)
		if err = subTask.Execute(); err == nil {
			if subTask.Rollback != nil {
				t.Rollbacks = append(t.Rollbacks, subTask.Rollback)
			}
		} else {
			if subTask.IgnoreErr {
				err = nil
				continue
			}
			t.Task.ErrorMsg = err.Error()
			t.Task.Status = constant.StatusFailed
			for _, rollback := range t.Rollbacks {
				rollback(t)
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
	_ = t.logFile.Close()
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
	t.Logger.Println(msg + i18n.GetMsgByKey("Failed"))
}

func (t *Task) LogFailedWithErr(msg string, err error) {
	t.Logger.Println(fmt.Sprintf("%s %s : %s", msg, i18n.GetMsgByKey("Failed"), err.Error()))
}

func (t *Task) LogSuccess(msg string) {
	t.Logger.Println(msg + i18n.GetMsgByKey("Success"))
}
func (t *Task) LogSuccessF(format string, v ...any) {
	t.Logger.Println(fmt.Sprintf(format, v...) + i18n.GetMsgByKey("Success"))
}

func (t *Task) LogStart(msg string) {
	t.Logger.Println(fmt.Sprintf("%s%s", i18n.GetMsgByKey("Start"), msg))
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
