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
	UID    uint
	GID    uint
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
	UserAdd(username, passwd, path string, uid, gid uint) error
	UserDel(username string) error
	SetPasswd(username, passwd string) error
	SetPath(username, path string, uid, gid uint) error
	Reload() error
	LoadLogs() ([]FtpLog, error)
}

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

var ftpIdentityMu sync.Mutex

const (
	standaloneFTPMinID = 10000
	standaloneFTPMaxID = 60000
)

func EnsureStandaloneFtpIdentity() (uint, uint, error) {
	ftpIdentityMu.Lock()
	defer ftpIdentityMu.Unlock()

	userItem, userErr := user.Lookup(constant.FTPUser)
	if userErr != nil && !isUnknownUser(userErr) {
		return 0, 0, userErr
	}
	groupItem, groupErr := user.LookupGroup(constant.FTPUser)
	if groupErr != nil && !isUnknownGroup(groupErr) {
		return 0, 0, groupErr
	}

	if groupErr == nil {
		gid, err := strconv.ParseUint(groupItem.Gid, 10, 32)
		if err != nil {
			return 0, 0, err
		}
		if gid < standaloneFTPMinID || gid > standaloneFTPMaxID {
			return 0, 0, fmt.Errorf(
				"FTP group %s must use a GID between %d and %d, got %d",
				constant.FTPUser,
				standaloneFTPMinID,
				standaloneFTPMaxID,
				gid,
			)
		}
	}
	if groupErr != nil {
		groupID := ""
		if userErr == nil {
			uid, err := strconv.ParseUint(userItem.Uid, 10, 32)
			if err != nil {
				return 0, 0, err
			}
			gid, err := strconv.ParseUint(userItem.Gid, 10, 32)
			if err != nil {
				return 0, 0, err
			}
			if uid < standaloneFTPMinID || uid > standaloneFTPMaxID ||
				gid < standaloneFTPMinID || gid > standaloneFTPMaxID {
				return 0, 0, fmt.Errorf(
					"FTP user %s must use UID and GID between %d and %d, got %d:%d",
					constant.FTPUser,
					standaloneFTPMinID,
					standaloneFTPMaxID,
					uid,
					gid,
				)
			}
			groupByID, err := user.LookupGroupId(userItem.Gid)
			if err == nil {
				return 0, 0, fmt.Errorf(
					"FTP user %s uses GID %s owned by group %s",
					constant.FTPUser,
					userItem.Gid,
					groupByID.Name,
				)
			}
			var unknownGroup user.UnknownGroupIdError
			if !errors.As(err, &unknownGroup) {
				return 0, 0, err
			}
			groupID = userItem.Gid
		} else {
			identityID, err := findAvailableFtpIdentityID(true, true)
			if err != nil {
				return 0, 0, err
			}
			groupID = identityID
		}
		if err := cmd.NewCommandMgr().Run("groupadd", "-g", groupID, constant.FTPUser); err != nil {
			return 0, 0, err
		}
	}
	if userErr != nil {
		uid, err := findAvailableFtpIdentityID(true, false)
		if err != nil {
			return 0, 0, err
		}
		noLoginShell := "/bin/false"
		for _, item := range []string{"/usr/sbin/nologin", "/sbin/nologin"} {
			if _, err := os.Stat(item); err == nil {
				noLoginShell = item
				break
			}
		}
		if err := cmd.NewCommandMgr().Run(
			"useradd",
			"-u", uid,
			"-g", constant.FTPUser,
			"-M",
			"-d", "/nonexistent",
			"-s", noLoginShell,
			constant.FTPUser,
		); err != nil {
			return 0, 0, err
		}
	}

	userItem, err := user.Lookup(constant.FTPUser)
	if err != nil {
		return 0, 0, err
	}
	groupItem, err = user.LookupGroup(constant.FTPUser)
	if err != nil {
		return 0, 0, err
	}
	if userItem.Gid != groupItem.Gid {
		return 0, 0, fmt.Errorf(
			"FTP user %s has GID %s, expected group GID %s",
			constant.FTPUser,
			userItem.Gid,
			groupItem.Gid,
		)
	}
	uid, err := strconv.ParseUint(userItem.Uid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	gid, err := strconv.ParseUint(groupItem.Gid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	if uid < standaloneFTPMinID || uid > standaloneFTPMaxID ||
		gid < standaloneFTPMinID || gid > standaloneFTPMaxID {
		return 0, 0, fmt.Errorf(
			"FTP identity %s must use UID and GID between %d and %d, got %d:%d",
			constant.FTPUser,
			standaloneFTPMinID,
			standaloneFTPMaxID,
			uid,
			gid,
		)
	}
	return uint(uid), uint(gid), nil
}

func findAvailableFtpIdentityID(checkUser, checkGroup bool) (string, error) {
	for id := standaloneFTPMinID; id <= standaloneFTPMaxID; id++ {
		idItem := strconv.Itoa(id)
		if checkUser {
			if _, err := user.LookupId(idItem); err == nil {
				continue
			} else {
				var unknownUser user.UnknownUserIdError
				if !errors.As(err, &unknownUser) {
					return "", err
				}
			}
		}
		if checkGroup {
			if _, err := user.LookupGroupId(idItem); err == nil {
				continue
			} else {
				var unknownGroup user.UnknownGroupIdError
				if !errors.As(err, &unknownGroup) {
					return "", err
				}
			}
		}
		return idItem, nil
	}
	return "", fmt.Errorf("no available FTP identity ID between %d and %d", standaloneFTPMinID, standaloneFTPMaxID)
}

func isUnknownUser(err error) bool {
	var unknownUser user.UnknownUserError
	return errors.As(err, &unknownUser)
}

func isUnknownGroup(err error) bool {
	var unknownGroup user.UnknownGroupError
	return errors.As(err, &unknownGroup)
}

func NewFtpClient() (*Ftp, error) {
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

func (f *Ftp) UserAdd(username, passwd, path string, uid, gid uint) error {
	if cmd.CheckIllegal(username, path) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := ValidateFtpRootPath(path); err != nil {
		return err
	}
	entry, err := generatePureFtpEntry(username, passwd, path, uid, gid)
	if err != nil {
		return fmt.Errorf("generate pure-ftpd entry failed, err: %v", err)
	}
	pwdFile, err := os.OpenFile("/etc/pure-ftpd/pureftpd.passwd", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = pwdFile.WriteString("\n" + entry + "\n")
	if err != nil {
		_ = pwdFile.Close()
		return err
	}
	if err := pwdFile.Close(); err != nil {
		return err
	}
	if err := f.Reload(); err != nil {
		return f.rollbackAddedUser(username, fmt.Errorf("reload FTP database after adding user failed: %w", err))
	}
	if err := chownFtpRoot(path, uid, gid); err != nil {
		return f.rollbackAddedUser(username, fmt.Errorf("change FTP root ownership failed: %w", err))
	}
	return nil
}

func (f *Ftp) rollbackAddedUser(username string, cause error) error {
	if rollbackErr := f.UserDel(username); rollbackErr != nil {
		return errors.Join(cause, fmt.Errorf("rollback FTP user %s failed: %w", username, rollbackErr))
	}
	return cause
}

func (f *Ftp) UserDel(username string) error {
	if cmd.CheckIllegal(username) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := cmd.NewCommandMgr().Run("pure-pw", "userdel", username); err != nil {
		return err
	}
	return f.Reload()
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

func (f *Ftp) SetPath(username, path string, uid, gid uint) error {
	if cmd.CheckIllegal(username, path) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := ValidateFtpRootPath(path); err != nil {
		return err
	}
	if err := cmd.NewCommandMgr().Run("pure-pw", "usermod", username, "-d", path); err != nil {
		return err
	}
	if err := chownFtpRoot(path, uid, gid); err != nil {
		return err
	}
	return nil
}

func chownFtpRoot(rootPath string, uid, gid uint) error {
	owner := fmt.Sprintf("%d:%d", uid, gid)
	return cmd.NewCommandMgr().Run("chown", "-R", owner, "--", rootPath)
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
	identities, err := loadPureFtpIdentities("/etc/pure-ftpd/pureftpd.passwd")
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
		identity, ok := identities[parts[0]]
		if !ok {
			return nil, fmt.Errorf("FTP identity for user %s was not found", parts[0])
		}
		lists = append(lists, FtpList{
			User:   parts[0],
			Path:   strings.ReplaceAll(parts[1], "/./", ""),
			Status: status,
			UID:    identity.UID,
			GID:    identity.GID,
		})
	}
	return lists, nil
}

type ftpIdentity struct {
	UID uint
	GID uint
}

func loadPureFtpIdentities(passwdPath string) (map[string]ftpIdentity, error) {
	pwdFile, err := os.Open(passwdPath)
	if err != nil {
		return nil, err
	}
	defer pwdFile.Close()

	identities := make(map[string]ftpIdentity)
	scanner := bufio.NewScanner(pwdFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) < 6 || parts[0] == "" {
			continue
		}
		uid, uidErr := strconv.ParseUint(parts[2], 10, 32)
		gid, gidErr := strconv.ParseUint(parts[3], 10, 32)
		if uidErr != nil || gidErr != nil {
			continue
		}
		identities[parts[0]] = ftpIdentity{UID: uint(uid), GID: uint(gid)}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return identities, nil
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

func generatePureFtpEntry(username, password, path string, uid, gid uint) (string, error) {
	if uid == 0 || gid == 0 {
		return "", errors.New("FTP UID and GID must be greater than zero")
	}
	passwdAfterSha512, err := helper.Generate([]byte(password))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s:%s:%d:%d::%s/./::::::::::::",
		username,
		passwdAfterSha512,
		uid,
		gid,
		path,
	), nil
}
