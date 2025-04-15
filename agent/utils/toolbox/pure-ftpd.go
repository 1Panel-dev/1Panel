package toolbox

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/systemctl"
	"github.com/1Panel-dev/1Panel/agent/utils/toolbox/helper"
)

type Ftp struct {
	DefaultUser  string
	DefaultGroup string
}

type FtpList struct {
	User   string
	Path   string
	Status string
}

type FtpLog struct {
	IP        string `json:"ip"`
	User      string `json:"user"`
	Time      string `json:"time"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Size      string `json:"size"`
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
		stdout2, err := cmd.RunDefaultWithStdoutBashCf("useradd -u 1000 -g %s %s", groupItem.Name, "1panel")
		if err != nil {
			return nil, errors.New(stdout2)
		}
		return &Ftp{DefaultUser: "1panel", DefaultGroup: groupItem.Name}, nil
	}
	if err.Error() != user.UnknownGroupIdError("1000").Error() {
		return nil, err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashC("groupadd -g 1000 1panel")
	if err != nil {
		return nil, errors.New(string(stdout))
	}
	stdout2, err := cmd.RunDefaultWithStdoutBashC("useradd -u 1000 -g 1panel 1panel")
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
		stdout, err := cmd.RunDefaultWithStdoutBashCf("systemctl %s pure-ftpd.service", operate)
		if err != nil {
			return fmt.Errorf("%s the pure-ftpd.service failed, err: %s", operate, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Ftp) UserAdd(username, passwd, path string) error {
	entry, err := generatePureFtpEntrySimple(username, passwd, path)
	if err != nil {
		return fmt.Errorf("generate pure-ftpd entry failed, err: %v", err)
	}
	pwdFile, err := os.OpenFile("/etc/pure-ftpd/pureftpd.passwd", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer pwdFile.Close()

	_, err = pwdFile.WriteString("\n" + entry + "\n")
	if err != nil {
		return err
	}
	_ = f.Reload()
	std2, err := cmd.RunDefaultWithStdoutBashCf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path)
	if err != nil {
		return errors.New(std2)
	}
	return nil
}

func (f *Ftp) UserDel(username string) error {
	std, err := cmd.RunDefaultWithStdoutBashCf("pure-pw userdel %s", username)
	if err != nil {
		return errors.New(std)
	}
	_ = f.Reload()
	return nil
}

func (f *Ftp) SetPasswd(username, passwd string) error {
	hashedPassword, err := helper.Generate([]byte(passwd))
	if err != nil {
		return err
	}
	pwdFile, err := os.Open("/etc/pure-ftpd/pureftpd.passwd")
	if err != nil {
		return err
	}
	defer pwdFile.Close()

	var entrys []string
	scanner := bufio.NewScanner(pwdFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		userEntry := strings.Split(line, ":")
		if len(userEntry) < 2 {
			continue
		}
		if userEntry[0] == username {
			userEntry[1] = string(hashedPassword)
			line = strings.Join(userEntry, ":")
		}
		entrys = append(entrys, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	pwdFile.Close()

	pwdFile, err = os.Create("/etc/pure-ftpd/pureftpd.passwd")
	if err != nil {
		return err
	}
	defer pwdFile.Close()

	for _, entry := range entrys {
		_, err := pwdFile.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Ftp) SetPath(username, path string) error {
	std, err := cmd.RunDefaultWithStdoutBashCf("pure-pw usermod %s -d %s", username, path)
	if err != nil {
		return errors.New(std)
	}
	std2, err := cmd.RunDefaultWithStdoutBashCf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path)
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
	std, err := cmd.RunDefaultWithStdoutBashCf("pure-pw usermod %s -r %s", username, statusItem)
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) LoadList() ([]FtpList, error) {
	std, err := cmd.RunDefaultWithStdoutBashC("pure-pw list")
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
		std2, err := cmd.RunDefaultWithStdoutBashCf("pure-pw  show %s | grep 'Allowed client IPs :'", parts[0])
		if err != nil {
			global.LOG.Errorf("handle pure-pw show %s failed, err: %v", parts[0], std2)
			continue
		}
		status := constant.StatusDisable
		itemStd := strings.ReplaceAll(std2, "\n", "")
		if len(strings.TrimSpace(strings.ReplaceAll(itemStd, "Allowed client IPs :", ""))) == 0 {
			status = constant.StatusEnable
		}
		lists = append(lists, FtpList{User: parts[0], Path: strings.ReplaceAll(parts[1], "/./", ""), Status: status})
	}
	return lists, nil
}

func (f *Ftp) Reload() error {
	std, err := cmd.RunDefaultWithStdoutBashC("pure-pw mkdb")
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) LoadLogs(user, operation string) ([]FtpLog, error) {
	var logs []FtpLog
	logItem := ""
	if _, err := os.Stat("/etc/pure-ftpd/conf"); err != nil && os.IsNotExist(err) {
		std, err := cmd.RunDefaultWithStdoutBashC("cat /etc/pure-ftpd/pure-ftpd.conf | grep AltLog | grep clf:")
		logItem = "/var/log/pureftpd.log"
		if err == nil && !strings.HasPrefix(std, "#") {
			logItem = std
		}
	} else {
		if err != nil {
			return logs, err
		}
		std, err := cmd.RunDefaultWithStdoutBashC("cat /etc/pure-ftpd/conf/AltLog")
		logItem = "/var/log/pure-ftpd/transfer.log"
		if err != nil && !strings.HasPrefix(std, "#") {
			logItem = std
		}
	}

	logItem = strings.ReplaceAll(logItem, "AltLog", "")
	logItem = strings.ReplaceAll(logItem, "clf:", "")
	logItem = strings.ReplaceAll(logItem, "\n", "")
	logPath := strings.Trim(logItem, " ")

	logDir := path.Dir(logPath)
	filesItem, err := os.ReadDir(logDir)
	if err != nil {
		return logs, err
	}
	var fileList []string
	for i := 0; i < len(filesItem); i++ {
		if filesItem[i].IsDir() {
			continue
		}
		itemPath := path.Join(logDir, filesItem[i].Name())
		if !strings.HasSuffix(itemPath, ".gz") {
			fileList = append(fileList, itemPath)
			continue
		}
		itemFileName := strings.TrimSuffix(itemPath, ".gz")
		if _, err := os.Stat(itemFileName); err != nil && os.IsNotExist(err) {
			if err := handleGunzip(itemPath); err == nil {
				fileList = append(fileList, itemFileName)
			}
		}
	}
	logs = loadLogsByFiles(fileList, user, operation)
	return logs, nil
}

func handleGunzip(path string) error {
	if _, err := cmd.RunDefaultWithStdoutBashCf("gunzip %s", path); err != nil {
		return err
	}
	return nil
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
				timeStr = timeItem.Format(constant.DateTimeLayout)
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

func generatePureFtpEntrySimple(username, password, path string) (string, error) {
	passwdAfterSha512, err := helper.Generate([]byte(password))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:1000:1000::%s/./::::::::::::", username, passwdAfterSha512, path), nil
}
