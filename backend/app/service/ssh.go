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
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/qqwry"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
	"github.com/pkg/errors"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHService struct{}

type ISSHService interface {
	GetSSHInfo() (*dto.SSHInfo, error)
	OperateSSH(operation string) error
	UpdateByFile(value string) error
	Update(req dto.SSHUpdate) error
	GenerateSSH(req dto.GenerateSSH) error
	AnalysisLog(req dto.SearchForAnalysis) (*dto.AnalysisRes, error)
	LoadSSHSecret(mode string) (string, error)
	LoadLog(req dto.SearchSSHLog) (*dto.SSHLog, error)

	LoadSSHConf() (string, error)
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

func (u *SSHService) GetSSHInfo() (*dto.SSHInfo, error) {
	data := dto.SSHInfo{
		AutoStart:              true,
		Status:                 constant.StatusEnable,
		Message:                "",
		Port:                   "22",
		ListenAddress:          "0.0.0.0",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	serviceName, err := loadServiceName()
	if err != nil {
		data.Status = constant.StatusDisable
		data.Message = err.Error()
	} else {
		active, err := systemctl.IsActive(serviceName)
		if !active {
			data.Status = constant.StatusDisable
			data.Message = err.Error()
		} else {
			data.Status = constant.StatusEnable
		}
	}

	out, err := systemctl.RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		data.AutoStart = false
	} else {
		if out == "alias\n" {
			data.AutoStart, _ = systemctl.IsEnable("ssh")
		} else {
			data.AutoStart = out == "enabled\n"
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
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	if operation == "enable" || operation == "disable" {
		serviceName += ".service"
	}
	stdout, err := cmd.Execf("%s systemctl %s %s", sudo, operation, serviceName)
	if err != nil {
		if strings.Contains(stdout, "alias name or linked unit file") {
			stdout, err := cmd.Execf("%s systemctl %s ssh", sudo, operation)
			if err != nil {
				return fmt.Errorf("%s ssh(alias name or linked unit file) failed, stdout: %s, err: %v", operation, stdout, err)
			}
		}
		return fmt.Errorf("%s %s failed, stdout: %s, err: %v", operation, serviceName, stdout, err)
	}
	return nil
}

func (u *SSHService) Update(req dto.SSHUpdate) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}

	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(sshConf), "\n")
	newFiles := updateSSHConf(lines, req.Key, req.NewValue)
	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	if req.Key == "Port" {
		stdout, _ := cmd.Execf("%s getenforce", sudo)
		if stdout == "Enforcing\n" {
			_, _ = cmd.Execf("%s semanage port -a -t ssh_port_t -p tcp %s", sudo, req.NewValue)
		}

		ruleItem := dto.PortRuleUpdate{
			OldRule: dto.PortRuleOperate{
				Operation: "remove",
				Port:      req.OldValue,
				Protocol:  "tcp",
				Strategy:  "accept",
			},
			NewRule: dto.PortRuleOperate{
				Operation: "add",
				Port:      req.NewValue,
				Protocol:  "tcp",
				Strategy:  "accept",
			},
		}
		if err := NewIFirewallService().UpdatePortRule(ruleItem); err != nil {
			global.LOG.Errorf("reset firewall rules %s -> %s failed, err: %v", req.OldValue, req.OldValue, err)
		}
	}

	_, _ = cmd.Execf("%s systemctl restart %s", sudo, serviceName)
	return nil
}

func (u *SSHService) UpdateByFile(value string) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(value); err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	_, _ = cmd.Execf("%s systemctl restart %s", sudo, serviceName)
	return nil
}

func (u *SSHService) GenerateSSH(req dto.GenerateSSH) error {
	if cmd.CheckIllegal(req.EncryptionMode, req.Password) {
		return buserr.New(constant.ErrCmdIllegal)
	}
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

type sshFileItem struct {
	Name string
	Year int
}

func (u *SSHService) LoadLog(req dto.SearchSSHLog) (*dto.SSHLog, error) {
	var fileList []sshFileItem
	var data dto.SSHLog
	baseDir := "/var/log"
	if err := filepath.Walk(baseDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasPrefix(info.Name(), "secure") || strings.HasPrefix(info.Name(), "auth")) {
			if !strings.HasSuffix(info.Name(), ".gz") {
				fileList = append(fileList, sshFileItem{Name: pathItem, Year: info.ModTime().Year()})
				return nil
			}
			itemFileName := strings.TrimSuffix(pathItem, ".gz")
			if _, err := os.Stat(itemFileName); err != nil && os.IsNotExist(err) {
				if err := handleGunzip(pathItem); err == nil {
					fileList = append(fileList, sshFileItem{Name: itemFileName, Year: info.ModTime().Year()})
				}
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

	showCountFrom := (req.Page - 1) * req.PageSize
	showCountTo := req.Page * req.PageSize
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	qqWry, err := qqwry.NewQQwry()
	if err != nil {
		global.LOG.Errorf("load qqwry datas failed: %s", err)
	}
	for _, file := range fileList {
		commandItem := ""
		if strings.HasPrefix(path.Base(file.Name), "secure") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -a 'Failed password for' | grep -v 'invalid' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE '(Failed password for|Accepted)' | grep -v 'invalid' %s", file.Name, command)
			}
		}
		if strings.HasPrefix(path.Base(file.Name), "auth.log") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -aE 'Failed password for|Connection closed by authenticating user' | grep -v 'invalid' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE \"(Failed password for|Connection closed by authenticating user|Accepted)\" | grep -v 'invalid' %s", file.Name, command)
			}
		}
		dataItem, successCount, failedCount := loadSSHData(commandItem, showCountFrom, showCountTo, file.Year, qqWry, nyc)
		data.FailedCount += failedCount
		data.TotalCount += successCount + failedCount
		showCountFrom = showCountFrom - (successCount + failedCount)
		showCountTo = showCountTo - (successCount + failedCount)
		data.Logs = append(data.Logs, dataItem...)
	}

	data.SuccessfulCount = data.TotalCount - data.FailedCount
	return &data, nil
}

func (u *SSHService) AnalysisLog(req dto.SearchForAnalysis) (*dto.AnalysisRes, error) {
	var fileList []string
	baseDir := "/var/log"
	if err := filepath.Walk(baseDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasPrefix(info.Name(), "secure") || strings.HasPrefix(info.Name(), "auth")) {
			if !strings.HasSuffix(info.Name(), ".gz") {
				fileList = append(fileList, pathItem)
				return nil
			}
			itemFileName := strings.TrimSuffix(pathItem, ".gz")
			if _, err := os.Stat(itemFileName); err != nil && os.IsNotExist(err) {
				if err := handleGunzip(pathItem); err == nil {
					fileList = append(fileList, itemFileName)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	command := ""
	sortMap := make(map[string]dto.SSHLogAnalysis)
	for _, file := range fileList {
		commandItem := ""
		if strings.HasPrefix(path.Base(file), "secure") {
			commandItem = fmt.Sprintf("cat %s | grep -aE '(Failed password for|Accepted)' | grep -v 'invalid' %s", file, command)
		}
		if strings.HasPrefix(path.Base(file), "auth.log") {
			commandItem = fmt.Sprintf("cat %s | grep -aE \"(Connection closed by authenticating user|Accepted)\" | grep -v 'invalid' %s", file, command)
		}
		loadSSHDataForAnalysis(sortMap, commandItem)
	}
	var sortSlice []dto.SSHLogAnalysis
	for key, value := range sortMap {
		sortSlice = append(sortSlice, dto.SSHLogAnalysis{Address: key, SuccessfulCount: value.SuccessfulCount, FailedCount: value.FailedCount, Status: "accept"})
	}
	if req.OrderBy == constant.StatusSuccess {
		sort.Slice(sortSlice, func(i, j int) bool {
			return sortSlice[i].SuccessfulCount > sortSlice[j].SuccessfulCount
		})
	} else {
		sort.Slice(sortSlice, func(i, j int) bool {
			return sortSlice[i].FailedCount > sortSlice[j].FailedCount
		})
	}
	qqWry, _ := qqwry.NewQQwry()
	rules, _ := listIpRules("drop")
	for i := 0; i < len(sortSlice); i++ {
		sortSlice[i].Area = qqWry.Find(sortSlice[i].Address).Area
		for _, rule := range rules {
			if sortSlice[i].Address == rule {
				sortSlice[i].Status = "drop"
				break
			}
		}
	}

	var backData dto.AnalysisRes
	for _, item := range sortSlice {
		backData.FailedCount += item.FailedCount
		backData.SuccessfulCount += item.SuccessfulCount
	}

	var data []dto.SSHLogAnalysis
	total, start, end := len(sortSlice), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		data = make([]dto.SSHLogAnalysis, 0)
	} else {
		if end >= total {
			end = total
		}
		data = sortSlice[start:end]
	}
	backData.Items = data
	backData.Total = int64(total)

	return &backData, nil
}

func (u *SSHService) LoadSSHConf() (string, error) {
	if _, err := os.Stat("/etc/ssh/sshd_config"); err != nil {
		return "", buserr.New("ErrHttpReqNotFound")
	}
	content, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func sortFileList(fileNames []sshFileItem) []sshFileItem {
	if len(fileNames) < 2 {
		return fileNames
	}
	if strings.HasPrefix(path.Base(fileNames[0].Name), "secure") {
		var itemFile []sshFileItem
		sort.Slice(fileNames, func(i, j int) bool {
			return fileNames[i].Name > fileNames[j].Name
		})
		itemFile = append(itemFile, fileNames[len(fileNames)-1])
		itemFile = append(itemFile, fileNames[:len(fileNames)-1]...)
		return itemFile
	}
	sort.Slice(fileNames, func(i, j int) bool {
		return fileNames[i].Name < fileNames[j].Name
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

func loadSSHData(command string, showCountFrom, showCountTo, currentYear int, qqWry *qqwry.QQwry, nyc *time.Location) ([]dto.SSHHistory, int, int) {
	var (
		datas        []dto.SSHHistory
		successCount int
		failedCount  int
	)
	stdout2, err := cmd.Exec(command)
	if err != nil {
		return datas, 0, 0
	}
	lines := strings.Split(string(stdout2), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		var itemData dto.SSHHistory
		switch {
		case strings.Contains(lines[i], "Failed password for"):
			itemData = loadFailedSecureDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = qqWry.Find(itemData.Address).Area
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Connection closed by authenticating user"):
			itemData = loadFailedAuthDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = qqWry.Find(itemData.Address).Area
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Accepted "):
			itemData = loadSuccessDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = qqWry.Find(itemData.Address).Area
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				successCount++
			}
		}
	}
	return datas, successCount, failedCount
}

func loadSSHDataForAnalysis(analysisMap map[string]dto.SSHLogAnalysis, commandItem string) {
	stdout, err := cmd.Exec(commandItem)
	if err != nil {
		return
	}
	lines := strings.Split(string(stdout), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		var itemData dto.SSHHistory
		switch {
		case strings.Contains(lines[i], "Failed password for"):
			itemData = loadFailedSecureDatas(lines[i])
		case strings.Contains(lines[i], "Connection closed by authenticating user"):
			itemData = loadFailedAuthDatas(lines[i])
		case strings.Contains(lines[i], "Accepted "):
			itemData = loadSuccessDatas(lines[i])
		}
		if len(itemData.Address) != 0 {
			if val, ok := analysisMap[itemData.Address]; ok {
				if itemData.Status == constant.StatusSuccess {
					val.SuccessfulCount++
				} else {
					val.FailedCount++
				}
				analysisMap[itemData.Address] = val
			} else {
				item := dto.SSHLogAnalysis{
					Address:         itemData.Address,
					SuccessfulCount: 0,
					FailedCount:     0,
				}
				if itemData.Status == constant.StatusSuccess {
					item.SuccessfulCount = 1
				} else {
					item.FailedCount = 1
				}
				analysisMap[itemData.Address] = item
			}
		}
	}
}

func loadSuccessDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", parts[0])
	if err != nil {
		if len(parts) < 14 {
			return data
		}
		data = dto.SSHHistory{
			DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
			AuthMode: parts[6],
			User:     parts[8],
			Address:  parts[10],
			Port:     parts[12],
			Status:   constant.StatusSuccess,
		}
	} else {
		if len(parts) < 12 {
			return data
		}
		data = dto.SSHHistory{
			DateStr:  t.Format("2006 Jan 2 15:04:05"),
			AuthMode: parts[4],
			User:     parts[6],
			Address:  parts[8],
			Port:     parts[10],
			Status:   constant.StatusSuccess,
		}
	}
	return data
}

func loadFailedAuthDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", parts[0])
	if err != nil {
		if len(parts) < 14 {
			return data
		}
		data = dto.SSHHistory{
			DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
			AuthMode: parts[8],
			User:     parts[10],
			Address:  parts[11],
			Port:     parts[13],
			Status:   constant.StatusFailed,
		}
	} else {
		if len(parts) < 12 {
			return data
		}
		data = dto.SSHHistory{
			DateStr:  t.Format("2006 Jan 2 15:04:05"),
			AuthMode: parts[6],
			User:     parts[7],
			Address:  parts[9],
			Port:     parts[11],
			Status:   constant.StatusFailed,
		}
	}
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}

func loadFailedSecureDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", parts[0])
	if err != nil {
		if len(parts) < 14 {
			return data
		}
		data = dto.SSHHistory{
			DateStr:  fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]),
			AuthMode: parts[6],
			User:     parts[8],
			Address:  parts[10],
			Port:     parts[12],
			Status:   constant.StatusFailed,
		}
	} else {
		if len(parts) < 12 {
			return data
		}
		index := 0
		if strings.Contains(line, " invalid user") {
			index = 2
		}
		data = dto.SSHHistory{
			DateStr:  t.Format("2006 Jan 2 15:04:05"),
			AuthMode: parts[4],
			User:     parts[index+6],
			Address:  parts[index+8],
			Port:     parts[index+10],
			Status:   constant.StatusFailed,
		}
	}
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}

func handleGunzip(path string) error {
	if _, err := cmd.Execf("gunzip %s", path); err != nil {
		return err
	}
	return nil
}

func loadServiceName() (string, error) {
	if exist, _ := systemctl.IsExist("sshd"); exist {
		return "sshd", nil
	} else if exist, _ := systemctl.IsExist("ssh"); exist {
		return "ssh", nil
	}
	return "", errors.New("The ssh or sshd service is unavailable")
}

func loadDate(currentYear int, DateStr string, nyc *time.Location) time.Time {
	itemDate, err := time.ParseInLocation("2006 Jan 2 15:04:05", fmt.Sprintf("%d %s", currentYear, DateStr), nyc)
	if err != nil {
		itemDate, _ = time.ParseInLocation("2006 Jan 2 15:04:05", DateStr, nyc)
	}
	return itemDate
}
