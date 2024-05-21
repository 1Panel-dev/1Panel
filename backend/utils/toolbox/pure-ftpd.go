package toolbox

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
)

type Ftp struct {
	DefaultUser  string
	DefaultGroup string
}

type FtpClient interface {
	Status() (bool, bool)
	Operate(operate string) error
	LoadList() ([]FtpList, error)
	UserAdd(username, path, passwd string) error
	UserDel(username string) error
	SetPasswd(username, passwd string) error
	Reload() error
	LoadLogs() ([]FtpLog, error)
}

func NewFtpClient() (*Ftp, error) {
	userItem, err := user.LookupId("1000")
	if err == nil {
		groupItem, err := user.LookupGroupId(userItem.Gid)
		if err != nil {
			return nil, err
		}
		return &Ftp{DefaultUser: userItem.Username, DefaultGroup: groupItem.Name}, err
	}
	if err.Error() != user.UnknownUserIdError(1000).Error() {
		return nil, err
	}

	groupItem, err := user.LookupGroupId("1000")
	if err == nil {
		stdout2, err := cmd.Execf("useradd -u 1000 -g %s %s", groupItem.Name, "1panel")
		if err != nil {
			return nil, errors.New(stdout2)
		}
		return &Ftp{DefaultUser: "1panel", DefaultGroup: groupItem.Name}, nil
	}
	if err.Error() != user.UnknownGroupIdError("1000").Error() {
		return nil, err
	}
	stdout, err := cmd.Exec("groupadd -g 1000 1panel")
	if err != nil {
		return nil, errors.New(string(stdout))
	}
	stdout2, err := cmd.Execf("useradd -u 1000 -g 1panel %s", userItem.Username)
	if err != nil {
		return nil, errors.New(stdout2)
	}
	return &Ftp{DefaultUser: "1panel", DefaultGroup: "1panel"}, nil
}

func (f *Ftp) Status() (bool, bool) {
	isActive, _ := systemctl.IsActive("pure-ftpd.service")
	isExist, _ := systemctl.IsExist("pure-ftpd.service")

	return isActive, isExist
}

func (f *Ftp) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop":
		stdout, err := cmd.Execf("systemctl %s pure-ftpd.service", operate)
		if err != nil {
			return fmt.Errorf("%s the pure-ftpd.service failed, err: %s", operate, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Ftp) UserAdd(username, passwd, path string) error {
	std, err := cmd.Execf("pure-pw useradd %s -u %s -d %s <<EOF \n%s\n%s\nEOF", username, f.DefaultUser, path, passwd, passwd)
	if err != nil {
		return errors.New(std)
	}
	_ = f.Reload()
	std2, err := cmd.Execf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path)
	if err != nil {
		return errors.New(std2)
	}
	return nil
}

func (f *Ftp) UserDel(username string) error {
	std, err := cmd.Execf("pure-pw userdel %s", username)
	if err != nil {
		return errors.New(std)
	}
	_ = f.Reload()
	return nil
}

func (f *Ftp) SetPasswd(username, passwd string) error {
	std, err := cmd.Execf("pure-pw passwd %s <<EOF \n%s\n%s\nEOF", username, passwd, passwd)
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) SetPath(username, path string) error {
	std, err := cmd.Execf("pure-pw usermod %s -d %s", username, path)
	if err != nil {
		return errors.New(std)
	}
	std2, err := cmd.Execf("chown %s %s", f.DefaultUser, path)
	if err != nil {
		return errors.New(std2)
	}
	return nil
}

func (f *Ftp) SetStatus(username, status string) error {
	statusItem := "''"
	if status == constant.StatusDisable {
		statusItem = "1"
	}
	std, err := cmd.Execf("pure-pw usermod %s -r %s", username, statusItem)
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) LoadList() ([]FtpList, error) {
	std, err := cmd.Exec("pure-pw list")
	if err != nil {
		return nil, errors.New(std)
	}
	var lists []FtpList
	lines := strings.Split(std, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		lists = append(lists, FtpList{User: parts[0], Path: strings.ReplaceAll(parts[1], "/./", "")})
	}
	return lists, nil
}

type FtpList struct {
	User string
	Path string
}

func (f *Ftp) Reload() error {
	std, err := cmd.Exec("pure-pw mkdb")
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) LoadLogs(user, operation string) ([]FtpLog, error) {
	var logs []FtpLog
	logItem := ""
	if _, err := os.Stat("/etc/pure-ftpd/conf"); err != nil && os.IsNotExist(err) {
		std, err := cmd.Exec("cat /etc/pure-ftpd/pure-ftpd.conf | grep AltLog | grep clf:")
		logItem = "/var/log/pureftpd.log"
		if err == nil && !strings.HasPrefix(logItem, "#") {
			logItem = std
		}
	} else {
		if err != nil {
			return logs, err
		}
		std, err := cmd.Exec("cat /etc/pure-ftpd/conf/AltLog")
		logItem = "/var/log/pure-ftpd/transfer.log"
		if err != nil && !strings.HasPrefix(logItem, "#") {
			logItem = std
		}
	}

	logItem = strings.ReplaceAll(logItem, "AltLog", "")
	logItem = strings.ReplaceAll(logItem, "clf:", "")
	logItem = strings.ReplaceAll(logItem, "\n", "")
	logPath := strings.Trim(logItem, " ")

	fileName := path.Base(logPath)
	var fileList []string
	if err := filepath.Walk(path.Dir(logPath), func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), fileName) {
			fileList = append(fileList, pathItem)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	logs = loadLogsByFiles(fileList, user, operation)
	return logs, nil
}

func loadLogsByFiles(fileList []string, user, operation string) []FtpLog {
	var logs []FtpLog
	layout := "02/Jan/2006:15:04:05-0700"
	for _, file := range fileList {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) < 9 {
				continue
			}
			if (len(user) != 0 && parts[2] != user) || (len(operation) != 0 && parts[5] != fmt.Sprintf("\"%s", operation)) {
				continue
			}
			timeStr := parts[3] + parts[4]
			timeStr = strings.ReplaceAll(timeStr, "[", "")
			timeStr = strings.ReplaceAll(timeStr, "]", "")
			timeItem, err := time.Parse(layout, timeStr)
			if err == nil {
				timeStr = timeItem.Format("2006-01-02 15:04:05")
			}
			operateStr := parts[5] + parts[6]
			logs = append(logs, FtpLog{
				IP:        parts[0],
				User:      parts[2],
				Time:      timeStr,
				Operation: operateStr,
				Status:    parts[7],
				Size:      parts[8],
			})
		}
	}
	return logs
}

type FtpLog struct {
	IP        string `json:"ip"`
	User      string `json:"user"`
	Time      string `json:"time"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Size      string `json:"size"`
}
