package toolbox

import (
	"bufio"
	"fmt"
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
	"golang.org/x/crypto/bcrypt"
)

type Ftp struct {
	DefaultUser  string
	DefaultGroup string
	serviceConf  *systemctl.ServiceConfig
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

var pureftpdService = &systemctl.ServiceConfig{
	ID:          "pure-ftpd",
	DisplayName: "Pure-FTPD Service",
	ServiceName: map[string]string{
		"systemd":  "pure-ftpd.service",
		"openrc":   "pure-ftpd",
		"sysvinit": "pure-ftpd",
	},
	UseSocket:   false,
	Description: "Pure-FTPD service management",
}

// NewFtpClient 创建 FTP 客户端实例
// 初始化时自动检查/创建 UID=1000 的默认用户和组
// 返回值:
//   - *Ftp: 初始化完成的 FTP 客户端对象
//   - error: 初始化过程中出现的错误
func NewFtpClient() (*Ftp, error) {
	ftp := &Ftp{serviceConf: pureftpdService}

	userItem, err := user.LookupId("1000")
	if err == nil {
		groupItem, err := user.LookupGroupId(userItem.Gid)
		if err != nil {
			global.LOG.Errorf("Lookup group failed: %v", err)
			return nil, err
		}
		ftp.DefaultUser = userItem.Username
		ftp.DefaultGroup = groupItem.Name
		return ftp, nil
	}

	if err.Error() != user.UnknownUserIdError(1000).Error() {
		global.LOG.Errorf("User lookup error: %v", err)
		return nil, err
	}

	if groupItem, err := user.LookupGroupId("1000"); err == nil {
		if _, err := cmd.Execf("useradd -u 1000 -g %s %s", groupItem.Name, "1panel"); err != nil {
			global.LOG.Errorf("Create user failed: %v", err)
			return nil, fmt.Errorf("create user failed: %v", err)
		}
		ftp.DefaultUser = "1panel"
		ftp.DefaultGroup = groupItem.Name
		return ftp, nil
	}

	if _, err := cmd.Exec("groupadd -g 1000 1panel"); err != nil {
		global.LOG.Errorf("Create group failed: %v", err)
		return nil, fmt.Errorf("create group failed: %v", err)
	}
	if _, err := cmd.Exec("useradd -u 1000 -g 1panel 1panel"); err != nil {
		global.LOG.Errorf("Create user failed: %v", err)
		return nil, fmt.Errorf("create user failed: %v", err)
	}
	ftp.DefaultUser = "1panel"
	ftp.DefaultGroup = "1panel"
	return ftp, nil
}

func (f *Ftp) Status() (bool, bool) {
	active, err := systemctl.IsActive(f.serviceConf)
	if err != nil {
		global.LOG.Warnf("Check service active status failed: %v", err)
	}
	exist, err := systemctl.IsExist(f.serviceConf)
	if err != nil {
		global.LOG.Warnf("Check service existence failed: %v", err)
	}
	return active, exist
}

func (f *Ftp) Operate(operate string) error {
	if err := systemctl.Operate(operate, f.serviceConf); err != nil {
		global.LOG.Errorf("%s service failed: %v", operate, err)
		return fmt.Errorf("%s service failed: %v", operate, err)
	}
	return nil
}

func (f *Ftp) UserAdd(username, passwd, path string) error {
	entry, err := generatePureFtpEntrySimple(username, passwd, path)
	if err != nil {
		global.LOG.Errorf("Generate user entry failed: %v", err)
		return err
	}

	pwdFile, err := os.OpenFile("/etc/pure-ftpd/pureftpd.passwd", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		global.LOG.Errorf("Open passwd file failed: %v", err)
		return err
	}
	defer pwdFile.Close()

	if _, err = pwdFile.WriteString("\n" + entry + "\n"); err != nil {
		global.LOG.Errorf("Write to passwd file failed: %v", err)
		return err
	}
	if err := f.Reload(); err != nil {
		global.LOG.Errorf("Reload service failed: %v", err)
		return err
	}

	if _, err := cmd.Execf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path); err != nil {
		global.LOG.Errorf("Chown command failed: %v", err)
		return fmt.Errorf("chown failed: %v", err)
	}
	return nil
}

func (f *Ftp) UserDel(username string) error {
	if _, err := cmd.Execf("pure-pw userdel %s", username); err != nil {
		return fmt.Errorf("userdel failed: %v", err)
	}
	return f.Reload()
}

func (f *Ftp) SetPasswd(username, passwd string) error {
	hashedPassword, err := hashPassword(passwd)
	if err != nil {
		return err
	}

	pwdFile, err := os.Open("/etc/pure-ftpd/pureftpd.passwd")
	if err != nil {
		return err
	}
	defer pwdFile.Close()

	var entries []string
	scanner := bufio.NewScanner(pwdFile)
	for scanner.Scan() {
		line := scanner.Text()
		if parts := strings.Split(line, ":"); len(parts) > 1 && parts[0] == username {
			parts[1] = string(hashedPassword)
			line = strings.Join(parts, ":")
		}
		entries = append(entries, line)
	}

	if err := os.WriteFile("/etc/pure-ftpd/pureftpd.passwd", []byte(strings.Join(entries, "\n")), 0644); err != nil {
		return err
	}
	return f.Reload()
}

func (f *Ftp) SetPath(username, path string) error {
	if _, err := cmd.Execf("pure-pw usermod %s -d %s", username, path); err != nil {
		return fmt.Errorf("usermod failed: %v", err)
	}
	if _, err := cmd.Execf("chown -R %s:%s %s", f.DefaultUser, f.DefaultGroup, path); err != nil {
		return fmt.Errorf("chown failed: %v", err)
	}
	return nil
}

func (f *Ftp) SetStatus(username, status string) error {
	statusFlag := "''"
	if status == constant.StatusDisable {
		statusFlag = "1"
	}
	if _, err := cmd.Execf("pure-pw usermod %s -r %s", username, statusFlag); err != nil {
		return fmt.Errorf("status update failed: %v", err)
	}
	return nil
}

func (f *Ftp) LoadList() ([]FtpList, error) {
	stdout, err := cmd.Exec("pure-pw list")
	if err != nil {
		global.LOG.Errorf("List users failed: %v", err)
		return nil, fmt.Errorf("list failed: %v", err)
	}

	var lists []FtpList
	for _, line := range strings.Split(stdout, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		stdout, err := cmd.Execf("pure-pw show %s | grep 'Allowed client IPs :'", parts[0])
		if err != nil {
			global.LOG.Warnf("Check user %s status failed: %v", parts[0], err)
			continue
		}

		status := constant.StatusEnable
		if cleaned := strings.TrimSpace(strings.ReplaceAll(stdout, "Allowed client IPs :", "")); cleaned != "" {
			status = constant.StatusDisable
		}

		lists = append(lists, FtpList{
			User:   parts[0],
			Path:   strings.ReplaceAll(parts[1], "/./", ""),
			Status: status,
		})
	}
	return lists, nil
}

func (f *Ftp) Reload() error {
	if _, err := cmd.Exec("pure-pw mkdb"); err != nil {
		global.LOG.Errorf("Reload database failed: %v", err)
		return fmt.Errorf("reload failed: %v", err)
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
