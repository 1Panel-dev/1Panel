package service

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/qqwry"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHService struct{}

type ISSHService interface {
	GetSSHInfo() (*dto.SSHInfo, error)
	OperateSSH(operation string) error
	UpdateByFile(value string) error
	Update(key, value string) error
	GenerateSSH(req dto.GenerateSSH) error
	LoadSSHSecret(mode string) (string, error)
	LoadLog(req dto.SearchSSHLog) (*dto.SSHLog, error)
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

func (u *SSHService) GetSSHInfo() (*dto.SSHInfo, error) {
	data := dto.SSHInfo{
		Status:                 constant.StatusEnable,
		Message:                "",
		Port:                   "22",
		ListenAddress:          "0.0.0.0",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	sudo := cmd.SudoHandleCmd()
	stdout, err := cmd.Execf("%s systemctl status sshd", sudo)
	if err != nil {
		data.Message = stdout
		data.Status = constant.StatusDisable
	}
	stdLines := strings.Split(stdout, "\n")
	for _, stdline := range stdLines {
		if strings.Contains(stdline, "active (running)") {
			data.Status = constant.StatusEnable
			break
		}
	}
	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		data.Message = err.Error()
		data.Status = constant.StatusDisable
	}
	lines := strings.Split(string(sshConf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			data.Port = strings.ReplaceAll(line, "Port ", "")
		}
		if strings.HasPrefix(line, "ListenAddress ") {
			data.ListenAddress = strings.ReplaceAll(line, "ListenAddress ", "")
		}
		if strings.HasPrefix(line, "PasswordAuthentication ") {
			data.PasswordAuthentication = strings.ReplaceAll(line, "PasswordAuthentication ", "")
		}
		if strings.HasPrefix(line, "PubkeyAuthentication ") {
			data.PubkeyAuthentication = strings.ReplaceAll(line, "PubkeyAuthentication ", "")
		}
		if strings.HasPrefix(line, "PermitRootLogin ") {
			data.PermitRootLogin = strings.ReplaceAll(line, "PermitRootLogin ", "")
		}
		if strings.HasPrefix(line, "UseDNS ") {
			data.UseDNS = strings.ReplaceAll(line, "UseDNS ", "")
		}
	}
	return &data, nil
}

func (u *SSHService) OperateSSH(operation string) error {
	if operation == "start" || operation == "stop" || operation == "restart" {
		sudo := cmd.SudoHandleCmd()
		stdout, err := cmd.Execf("%s systemctl %s sshd", sudo, operation)
		if err != nil {
			return fmt.Errorf("%s sshd failed, stdout: %s, err: %v", operation, stdout, err)
		}
		return nil
	}
	return fmt.Errorf("not support such operation: %s", operation)
}

func (u *SSHService) Update(key, value string) error {
	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(sshConf), "\n")
	newFiles := updateSSHConf(lines, key, value)
	if err := settingRepo.Update(key, value); err != nil {
		return err
	}
	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	if key == "Port" {
		stdout, _ := cmd.Execf("%s getenforce", sudo)
		if stdout == "Enforcing\n" {
			_, _ = cmd.Execf("%s semanage port -a -t ssh_port_t -p tcp %s", sudo, value)
		}
	}
	_, _ = cmd.Execf("%s systemctl restart sshd", sudo)
	return nil
}

func (u *SSHService) UpdateByFile(value string) error {
	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(value); err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	_, _ = cmd.Execf("%s systemctl restart sshd", sudo)
	return nil
}

func (u *SSHService) GenerateSSH(req dto.GenerateSSH) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	secretFile := fmt.Sprintf("%s/.ssh/id_item_%s", currentUser.HomeDir, req.EncryptionMode)
	secretPubFile := fmt.Sprintf("%s/.ssh/id_item_%s.pub", currentUser.HomeDir, req.EncryptionMode)
	authFile := currentUser.HomeDir + "/.ssh/authorized_keys"

	command := fmt.Sprintf("ssh-keygen -t %s -f %s/.ssh/id_item_%s | echo y", req.EncryptionMode, currentUser.HomeDir, req.EncryptionMode)
	if len(req.Password) != 0 {
		command = fmt.Sprintf("ssh-keygen -t %s -P %s -f %s/.ssh/id_item_%s | echo y", req.EncryptionMode, req.Password, currentUser.HomeDir, req.EncryptionMode)
	}
	stdout, err := cmd.Exec(command)
	if err != nil {
		return fmt.Errorf("generate failed, err: %v, message: %s", err, stdout)
	}
	defer func() {
		_ = os.Remove(secretFile)
	}()
	defer func() {
		_ = os.Remove(secretPubFile)
	}()

	if _, err := os.Stat(authFile); err != nil {
		_, _ = os.Create(authFile)
	}
	stdout1, err := cmd.Execf("cat %s >> %s/.ssh/authorized_keys", secretPubFile, currentUser.HomeDir)
	if err != nil {
		return fmt.Errorf("generate failed, err: %v, message: %s", err, stdout1)
	}

	fileOp := files.NewFileOp()
	if err := fileOp.Rename(secretFile, fmt.Sprintf("%s/.ssh/id_%s", currentUser.HomeDir, req.EncryptionMode)); err != nil {
		return err
	}
	if err := fileOp.Rename(secretPubFile, fmt.Sprintf("%s/.ssh/id_%s.pub", currentUser.HomeDir, req.EncryptionMode)); err != nil {
		return err
	}

	return nil
}

func (u *SSHService) LoadSSHSecret(mode string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("load current user failed, err: %v", err)
	}

	homeDir := currentUser.HomeDir
	if _, err := os.Stat(fmt.Sprintf("%s/.ssh/id_%s", homeDir, mode)); err != nil {
		return "", nil
	}
	file, err := os.ReadFile(fmt.Sprintf("%s/.ssh/id_%s", homeDir, mode))
	return string(file), err
}

func (u *SSHService) LoadLog(req dto.SearchSSHLog) (*dto.SSHLog, error) {
	var fileList []string
	var data dto.SSHLog
	baseDir := "/var/log"
	if err := filepath.Walk(baseDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "secure") || strings.HasPrefix(info.Name(), "auth") {
			if strings.HasSuffix(info.Name(), ".gz") {
				if err := handleGunzip(pathItem); err == nil {
					fileList = append(fileList, strings.ReplaceAll(pathItem, ".gz", ""))
				}
			} else {
				fileList = append(fileList, pathItem)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	fileList = sortFileList(fileList)

	command := ""
	if len(req.Info) != 0 {
		command = fmt.Sprintf(" | grep '%s'", req.Info)
	}

	for i := 0; i < len(fileList); i++ {
		withAppend := len(data.Logs) < req.Page*req.PageSize
		if req.Status != constant.StatusSuccess {
			if strings.HasPrefix(path.Base(fileList[i]), "secure") {
				commandItem := fmt.Sprintf("cat %s | grep -a 'Failed password for' | grep -v 'invalid' %s", fileList[i], command)
				dataItem, itemTotal := loadFailedSecureDatas(commandItem, withAppend)
				data.FailedCount += itemTotal
				data.TotalCount += itemTotal
				data.Logs = append(data.Logs, dataItem...)
			}
			if strings.HasPrefix(path.Base(fileList[i]), "auth.log") {
				commandItem := fmt.Sprintf("cat %s | grep -a 'Connection closed by authenticating user' | grep -a 'preauth' %s", fileList[i], command)
				dataItem, itemTotal := loadFailedAuthDatas(commandItem, withAppend)
				data.FailedCount += itemTotal
				data.TotalCount += itemTotal
				data.Logs = append(data.Logs, dataItem...)
			}
		}
		if req.Status != constant.StatusFailed {
			commandItem := fmt.Sprintf("cat %s | grep -a Accepted %s", fileList[i], command)
			dataItem, itemTotal := loadSuccessDatas(commandItem, withAppend)
			data.TotalCount += itemTotal
			data.Logs = append(data.Logs, dataItem...)
		}
	}
	data.SuccessfulCount = data.TotalCount - data.FailedCount
	if len(data.Logs) < 1 {
		return nil, nil
	}

	var itemDatas []dto.SSHHistory
	total, start, end := len(data.Logs), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		itemDatas = make([]dto.SSHHistory, 0)
	} else {
		if end >= total {
			end = total
		}
		itemDatas = data.Logs[start:end]
	}
	data.Logs = itemDatas

	timeNow := time.Now()
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	qqWry, err := qqwry.NewQQwry()
	if err != nil {
		global.LOG.Errorf("load qqwry datas failed: %s", err)
	}
	var itemLogs []dto.SSHHistory
	for i := 0; i < len(data.Logs); i++ {
		data.Logs[i].Area = qqWry.Find(data.Logs[i].Address).Area
		data.Logs[i].Date, _ = time.ParseInLocation("2006 Jan 2 15:04:05", fmt.Sprintf("%d %s", timeNow.Year(), data.Logs[i].DateStr), nyc)
		itemLogs = append(itemLogs, data.Logs[i])
	}
	data.Logs = itemLogs

	return &data, nil
}

func sortFileList(fileNames []string) []string {
	if len(fileNames) < 2 {
		return fileNames
	}
	if strings.HasPrefix(path.Base(fileNames[0]), "secure") {
		var itemFile []string
		sort.Slice(fileNames, func(i, j int) bool {
			return fileNames[i] > fileNames[j]
		})
		itemFile = append(itemFile, fileNames[len(fileNames)-1])
		itemFile = append(itemFile, fileNames[:len(fileNames)-2]...)
		return itemFile
	}
	sort.Slice(fileNames, func(i, j int) bool {
		return fileNames[i] < fileNames[j]
	})
	return fileNames
}

func updateSSHConf(oldFiles []string, param string, value interface{}) []string {
	hasKey := false
	var newFiles []string
	for _, line := range oldFiles {
		if strings.HasPrefix(line, param+" ") {
			newFiles = append(newFiles, fmt.Sprintf("%s %v", param, value))
			hasKey = true
			continue
		}
		newFiles = append(newFiles, line)
	}
	if !hasKey {
		newFiles = []string{}
		for _, line := range oldFiles {
			if strings.HasPrefix(line, fmt.Sprintf("#%s ", param)) && !hasKey {
				newFiles = append(newFiles, fmt.Sprintf("%s %v", param, value))
				hasKey = true
				continue
			}
			newFiles = append(newFiles, line)
		}
	}
	if !hasKey {
		newFiles = []string{}
		newFiles = append(newFiles, oldFiles...)
		newFiles = append(newFiles, fmt.Sprintf("%s %v", param, value))
	}
	return newFiles
}

func loadSuccessDatas(command string, withAppend bool) ([]dto.SSHHistory, int) {
	var (
		datas    []dto.SSHHistory
		totalNum int
	)
	stdout2, err := cmd.Exec(command)
	if err == nil {
		lines := strings.Split(string(stdout2), "\n")
		if len(lines) == 0 {
			return datas, 0
		}
		for i := len(lines) - 1; i >= 0; i-- {
			parts := strings.Fields(lines[i])
			if len(parts) < 14 {
				continue
			}
			totalNum++
			if withAppend {
				historyItem := dto.SSHHistory{
					DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
					AuthMode: parts[6],
					User:     parts[8],
					Address:  parts[10],
					Port:     parts[12],
					Status:   constant.StatusSuccess,
				}
				datas = append(datas, historyItem)
			}
		}
	}
	return datas, totalNum
}

func loadFailedAuthDatas(command string, withAppend bool) ([]dto.SSHHistory, int) {
	var (
		datas    []dto.SSHHistory
		totalNum int
	)
	stdout2, err := cmd.Exec(command)
	if err == nil {
		lines := strings.Split(string(stdout2), "\n")
		if len(lines) == 0 {
			return datas, 0
		}
		for i := len(lines) - 1; i >= 0; i-- {
			parts := strings.Fields(lines[i])
			if len(parts) < 14 {
				continue
			}
			totalNum++
			if withAppend {
				historyItem := dto.SSHHistory{
					DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
					AuthMode: parts[8],
					User:     parts[10],
					Address:  parts[11],
					Port:     parts[13],
					Status:   constant.StatusFailed,
				}
				if strings.Contains(lines[i], ": ") {
					historyItem.Message = strings.Split(lines[i], ": ")[1]
				}
				datas = append(datas, historyItem)
			}
		}
	}
	return datas, totalNum
}

func loadFailedSecureDatas(command string, withAppend bool) ([]dto.SSHHistory, int) {
	var (
		datas    []dto.SSHHistory
		totalNum int
	)
	stdout2, err := cmd.Exec(command)
	if err == nil {
		lines := strings.Split(string(stdout2), "\n")
		if len(lines) == 0 {
			return datas, 0
		}
		for i := len(lines) - 1; i >= 0; i-- {
			parts := strings.Fields(lines[i])
			if len(parts) < 14 {
				continue
			}
			totalNum++
			if withAppend {
				historyItem := dto.SSHHistory{
					DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
					AuthMode: parts[6],
					User:     parts[8],
					Address:  parts[10],
					Port:     parts[12],
					Status:   constant.StatusFailed,
				}
				if strings.Contains(lines[i], ": ") {
					historyItem.Message = strings.Split(lines[i], ": ")[1]
				}
				datas = append(datas, historyItem)
			}
		}
	}
	return datas, totalNum
}

func handleGunzip(path string) error {
	if _, err := cmd.Execf("gunzip %s", path); err != nil {
		return err
	}
	return nil
}
