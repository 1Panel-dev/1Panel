package toolbox

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
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
	stdout2, err := cmd.Exec("useradd -u 1000 -g 1panel 1panel")
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
	entry, err := generatePureFtpEntrySimple(username, passwd, path)
	if err != nil {
		return err
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
	hashedPassword, err := hashPassword(passwd)
	if err != nil {
		return err
	}
	// read now
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

	// write new
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
	std, err := cmd.Execf("pure-pw usermod %s -d %s", username, path)
	if err != nil {
		return errors.New(std)
	}
	std2, err := cmd.Execf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path)
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
		std2, err := cmd.Execf("pure-pw  show %s | grep 'Allowed client IPs :'", parts[0])
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

func hashPassword(password string) ([]byte, error) {
	// Hash the password using bcrypt with a cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func generatePureFtpEntrySimple(username, password, path string) (string, error) {
	return generatePureFtpEntry(username, password, 1000, 1000, "", path+"/./",
		"", "", "", "", "",
		"", "", "", "", "", "", "")
}

func generatePureFtpEntry(username, password string, uid, gid int, gecos, homedir,
	uploadBandwidth, downloadBandwidth, uploadRatio, downloadRatio, maxConnections, filesQuota, sizeQuota,
	authorizedLocalIPs, refusedLocalIPs, authorizedClientIPs, refusedClientIPs, timeRestrictions string) (string, error) {

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	// Format the entry
	entry := fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		username,
		hashedPassword,
		uid,
		gid,
		gecos,
		homedir,
		uploadBandwidth,
		downloadBandwidth,
		uploadRatio,
		downloadRatio,
		maxConnections,
		filesQuota,
		sizeQuota,
		authorizedLocalIPs,
		refusedLocalIPs,
		authorizedClientIPs,
		refusedClientIPs,
		timeRestrictions,
	)

	return entry, nil
}
