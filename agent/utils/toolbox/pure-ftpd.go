package toolbox

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/toolbox/helper"
)

type Ftp struct{}

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

var ErrFtpNotInitialized = fmt.Errorf(
	"FTP identity %d:%d is not initialized",
	constant.FTPUid,
	constant.FTPGid,
)

var ErrFtpUnsafePath = errors.New("FTP root path is unsafe")

var ftpUnsafeRootPaths = map[string]struct{}{
	"/":              {},
	"/bin":           {},
	"/sbin":          {},
	"/usr/bin":       {},
	"/usr/sbin":      {},
	"/usr/local/bin": {},
	"/etc":           {},
	"/lib":           {},
	"/lib64":         {},
	"/usr/lib":       {},
	"/home":          {},
	"/tmp":           {},
	"/var":           {},
	"/dev":           {},
	"/proc":          {},
	"/sys":           {},
}

var ftpInitMu sync.Mutex

func IsFtpInitialized() (bool, error) {
	userExists, groupExists, err := loadFtpIdentityStatus()
	if err != nil {
		return false, err
	}
	return userExists && groupExists, nil
}

func InitFtp() error {
	ftpInitMu.Lock()
	defer ftpInitMu.Unlock()

	uid := strconv.Itoa(constant.FTPUid)
	gid := strconv.Itoa(constant.FTPGid)
	userExists, groupExists, err := loadFtpIdentityStatus()
	if err != nil {
		return err
	}

	if !userExists {
		userItem, err := user.Lookup(constant.FTPUser)
		if err == nil {
			if userItem.Uid != uid {
				return fmt.Errorf("user %s already exists with UID %s", constant.FTPUser, userItem.Uid)
			}
			userExists = true
		} else {
			var unknownUser user.UnknownUserError
			if !errors.As(err, &unknownUser) {
				return err
			}
		}
	}
	if !groupExists {
		groupItem, err := user.LookupGroup(constant.FTPUser)
		if err == nil {
			if groupItem.Gid != gid {
				return fmt.Errorf("group %s already exists with GID %s", constant.FTPUser, groupItem.Gid)
			}
			groupExists = true
		} else {
			var unknownGroup user.UnknownGroupError
			if !errors.As(err, &unknownGroup) {
				return err
			}
		}
	}

	cmdMgr := cmd.NewCommandMgr()
	if !groupExists {
		if err := cmdMgr.Run("groupadd", "-g", gid, constant.FTPUser); err != nil {
			return err
		}
	}
	if !userExists {
		noLoginShell := "/bin/false"
		for _, item := range []string{"/usr/sbin/nologin", "/sbin/nologin"} {
			if _, err := os.Stat(item); err == nil {
				noLoginShell = item
				break
			}
		}
		if err := cmdMgr.Run(
			"useradd",
			"-u", uid,
			"-g", gid,
			"-M",
			"-s", noLoginShell,
			constant.FTPUser,
		); err != nil {
			return err
		}
	}
	isInitialized, err := IsFtpInitialized()
	if err != nil {
		return err
	}
	if !isInitialized {
		return ErrFtpNotInitialized
	}
	return nil
}

func loadFtpIdentityStatus() (bool, bool, error) {
	uid := strconv.Itoa(constant.FTPUid)
	_, userErr := user.LookupId(uid)
	if userErr != nil {
		var unknownUser user.UnknownUserIdError
		if !errors.As(userErr, &unknownUser) {
			return false, false, userErr
		}
	}

	gid := strconv.Itoa(constant.FTPGid)
	_, groupErr := user.LookupGroupId(gid)
	if groupErr != nil {
		var unknownGroup user.UnknownGroupIdError
		if !errors.As(groupErr, &unknownGroup) {
			return false, false, groupErr
		}
	}
	return userErr == nil, groupErr == nil, nil
}

func NewFtpClient() (*Ftp, error) {
	isInitialized, err := IsFtpInitialized()
	if err != nil {
		return nil, err
	}
	if !isInitialized {
		return nil, ErrFtpNotInitialized
	}
	return &Ftp{}, nil
}

func FtpStatus() (bool, bool) {
	isActive, _ := controller.CheckActive("pure-ftpd.service")
	isExist, _ := controller.CheckExist("pure-ftpd.service")

	return isActive, isExist
}

func (f *Ftp) Status() (bool, bool) {
	return FtpStatus()
}

func (f *Ftp) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop":
		if err := controller.Handle(operate, "pure-ftpd.service"); err != nil {
			return fmt.Errorf("%s the pure-ftpd.service failed, err: %v", operate, err)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Ftp) UserAdd(username, passwd, path string) error {
	if cmd.CheckIllegal(username, path) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := ValidateFtpRootPath(path); err != nil {
		return err
	}
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
	owner := fmt.Sprintf("%d:%d", constant.FTPUid, constant.FTPGid)
	if err := cmd.NewCommandMgr().Run("chown", "-R", owner, "--", path); err != nil {
		return err
	}
	return nil
}

func (f *Ftp) UserDel(username string) error {
	if cmd.CheckIllegal(username) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := cmd.NewCommandMgr().Run("pure-pw", "userdel", username); err != nil {
		return err
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
	if cmd.CheckIllegal(username, path) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := ValidateFtpRootPath(path); err != nil {
		return err
	}
	if err := cmd.NewCommandMgr().Run("pure-pw", "usermod", username, "-d", path); err != nil {
		return err
	}
	owner := fmt.Sprintf("%d:%d", constant.FTPUid, constant.FTPGid)
	if err := cmd.NewCommandMgr().Run("chown", "-R", owner, "--", path); err != nil {
		return err
	}
	return nil
}

func ValidateFtpRootPath(rootPath string) error {
	if strings.TrimSpace(rootPath) == "" {
		return fmt.Errorf("%w: path is required", ErrFtpUnsafePath)
	}
	if !filepath.IsAbs(rootPath) {
		return fmt.Errorf("%w: path must be absolute", ErrFtpUnsafePath)
	}

	cleanedPath := filepath.Clean(rootPath)
	if isUnsafeFtpRootPath(cleanedPath) {
		return fmt.Errorf("%w: %s", ErrFtpUnsafePath, cleanedPath)
	}
	if realPath, err := filepath.EvalSymlinks(cleanedPath); err == nil {
		cleanedPath = filepath.Clean(realPath)
	}
	if isUnsafeFtpRootPath(cleanedPath) {
		return fmt.Errorf("%w: %s", ErrFtpUnsafePath, cleanedPath)
	}
	return nil
}

func isUnsafeFtpRootPath(rootPath string) bool {
	_, unsafe := ftpUnsafeRootPaths[rootPath]
	return unsafe
}

func (f *Ftp) SetStatus(username, status string) error {
	if cmd.CheckIllegal(username, status) {
		return buserr.New("ErrCmdIllegal")
	}
	statusItem := ""
	if status == constant.StatusDisable {
		statusItem = "1"
	}
	if err := cmd.NewCommandMgr().Run("pure-pw", "usermod", username, "-r", statusItem); err != nil {
		return err
	}
	return nil
}

func (f *Ftp) LoadList() ([]FtpList, error) {
	std, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("pure-pw", "list")
	if err != nil {
		return nil, err
	}
	var lists []FtpList
	lines := strings.Split(std, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		std2, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("pure-pw", "show", parts[0])
		if err != nil {
			global.LOG.Errorf("handle pure-pw show %s failed, %v", parts[0], err)
			continue
		}
		status := constant.StatusDisable
		allowedLine := ""
		for _, line := range strings.Split(std2, "\n") {
			if strings.Contains(line, "Allowed client IPs :") {
				allowedLine = line
				break
			}
		}
		if len(strings.TrimSpace(strings.ReplaceAll(allowedLine, "Allowed client IPs :", ""))) == 0 {
			status = constant.StatusEnable
		}
		lists = append(lists, FtpList{User: parts[0], Path: strings.ReplaceAll(parts[1], "/./", ""), Status: status})
	}
	return lists, nil
}

func (f *Ftp) Reload() error {
	if err := cmd.NewCommandMgr().Run("pure-pw", "mkdb"); err != nil {
		return err
	}
	return nil
}

func (f *Ftp) LoadLogs(user, operation string) ([]FtpLog, error) {
	var logs []FtpLog
	logItem := ""
	if _, err := os.Stat("/etc/pure-ftpd/conf"); err != nil && os.IsNotExist(err) {
		logItem = "/var/log/pureftpd.log"
		data, readErr := os.ReadFile("/etc/pure-ftpd/pure-ftpd.conf")
		if readErr == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "#") || !strings.Contains(line, "AltLog") || !strings.Contains(line, "clf:") {
					continue
				}
				logItem = line
				break
			}
		}
	} else {
		if err != nil {
			return logs, err
		}
		logItem = "/var/log/pure-ftpd/transfer.log"
		data, readErr := os.ReadFile("/etc/pure-ftpd/conf/AltLog")
		if readErr == nil {
			std := string(data)
			if !strings.HasPrefix(strings.TrimSpace(std), "#") {
				logItem = std
			}
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
	if err := cmd.NewCommandMgr().Run("gunzip", path); err != nil {
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
	return fmt.Sprintf(
		"%s:%s:%d:%d::%s/./::::::::::::",
		username,
		passwdAfterSha512,
		constant.FTPUid,
		constant.FTPGid,
		path,
	), nil
}
