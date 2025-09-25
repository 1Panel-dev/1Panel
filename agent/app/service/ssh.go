package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	csvexport "github.com/1Panel-dev/1Panel/agent/utils/csv_export"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/geo"
	"github.com/gin-gonic/gin"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/systemctl"
	"github.com/pkg/errors"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHService struct{}

type ISSHService interface {
	GetSSHInfo() (*dto.SSHInfo, error)
	OperateSSH(operation string) error
	Update(req dto.SSHUpdate) error
	LoadSSHFile(name string) (string, error)
	UpdateByFile(req dto.SettingUpdate) error

	LoadLog(ctx *gin.Context, req dto.SearchSSHLog) (int64, []dto.SSHHistory, error)
	ExportLog(ctx *gin.Context, req dto.SearchSSHLog) (string, error)

	SyncRootCert() error
	CreateRootCert(req dto.CreateRootCert) error
	SearchRootCerts(req dto.SearchWithPage) (int64, interface{}, error)
	DeleteRootCerts(req dto.ForceDelete) error
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

func (u *SSHService) GetSSHInfo() (*dto.SSHInfo, error) {
	data := dto.SSHInfo{
		AutoStart:              true,
		IsExist:                true,
		IsActive:               true,
		Message:                "",
		Port:                   "22",
		ListenAddress:          "",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	serviceName, err := loadServiceName()
	if err != nil {
		data.IsExist = false
		data.Message = err.Error()
	} else {
		active, err := systemctl.IsActive(serviceName)
		data.IsActive = active
		if !active && err != nil {
			data.Message = err.Error()
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
		data.IsActive = false
	}
	lines := strings.Split(string(sshConf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			data.Port = strings.ReplaceAll(line, "Port ", "")
		}
		if strings.HasPrefix(line, "ListenAddress ") {
			itemAddr := strings.ReplaceAll(line, "ListenAddress ", "")
			if len(data.ListenAddress) != 0 {
				data.ListenAddress += ("," + itemAddr)
			} else {
				data.ListenAddress = itemAddr
			}
		}
		if strings.HasPrefix(line, "PasswordAuthentication ") {
			data.PasswordAuthentication = strings.ReplaceAll(line, "PasswordAuthentication ", "")
		}
		if strings.HasPrefix(line, "PubkeyAuthentication ") {
			data.PubkeyAuthentication = strings.ReplaceAll(line, "PubkeyAuthentication ", "")
		}
		if strings.HasPrefix(line, "PermitRootLogin ") {
			data.PermitRootLogin = strings.ReplaceAll(strings.ReplaceAll(line, "PermitRootLogin ", ""), "prohibit-password", "without-password")
		}
		if strings.HasPrefix(line, "UseDNS ") {
			data.UseDNS = strings.ReplaceAll(line, "UseDNS ", "")
		}
	}

	currentUser, err := user.Current()
	if err != nil || len(currentUser.Name) == 0 {
		data.CurrentUser = "root"
	} else {
		data.CurrentUser = currentUser.Name
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
	if operation == "stop" {
		isSocketActive, _ := systemctl.IsActive(serviceName + ".socket")
		if isSocketActive {
			std, err := cmd.RunDefaultWithStdoutBashCf("%s systemctl stop %s", sudo, serviceName+".socket")
			if err != nil {
				global.LOG.Errorf("handle systemctl stop %s.socket failed, err: %v", serviceName, std)
			}
		}
	}

	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s systemctl %s %s", sudo, operation, serviceName)
	if err != nil {
		if strings.Contains(stdout, "alias name or linked unit file") {
			stdout, err := cmd.RunDefaultWithStdoutBashCf("%s systemctl %s ssh", sudo, operation)
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
	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	if req.Key == "Port" {
		stdout, _ := cmd.RunDefaultWithStdoutBashCf("%s getenforce", sudo)
		if stdout == "Enforcing\n" {
			_, _ = cmd.RunDefaultWithStdoutBashCf("%s semanage port -a -t ssh_port_t -p tcp %s", sudo, req.NewValue)
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
			global.LOG.Errorf("reset firewall rules %s -> %s failed, err: %v", req.OldValue, req.NewValue, err)
		}
		newPort, _ := strconv.Atoi(req.NewValue)
		if err := updateLocalConn(uint(newPort)); err != nil {
			global.LOG.Errorf("update local conn for terminal failed, err: %v", err)
		}

		if err := updateSSHSocketFile(req.NewValue); err != nil {
			global.LOG.Errorf("update port for ssh.socket failed, err: %v", err)
		}
	}

	_, _ = cmd.RunDefaultWithStdoutBashCf("%s systemctl restart %s", sudo, serviceName)
	return nil
}

func (u *SSHService) SyncRootCert() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	sshDir := fmt.Sprintf("%s/.ssh", currentUser.HomeDir)

	fileList, err := os.ReadDir(sshDir)
	if err != nil {
		return err
	}
	var rootCerts []model.RootCert
	fileMap := make(map[string]bool)
	for _, item := range fileList {
		if !item.IsDir() {
			fileMap[item.Name()] = true
		}
	}
	for item := range fileMap {
		if !strings.HasSuffix(item, ".pub") {
			continue
		}
		if !fileMap[strings.TrimSuffix(item, ".pub")] {
			continue
		}
		cert := model.RootCert{Name: strings.TrimSuffix(item, ".pub"), PublicKeyPath: path.Join(sshDir, item), PrivateKeyPath: path.Join(sshDir, strings.TrimSuffix(item, ".pub"))}
		pubItem, err := os.ReadFile(path.Join(sshDir, item))
		if err != nil {
			global.LOG.Errorf("read pubic key of %s for sync failed, err: %v", item, err)
			continue
		}
		cert.EncryptionMode = loadEncryptioMode(string(pubItem))
		rootCerts = append(rootCerts, cert)
	}
	return hostRepo.SyncCert(rootCerts)
}

func (u *SSHService) CreateRootCert(req dto.CreateRootCert) error {
	if cmd.CheckIllegal(req.EncryptionMode, req.PassPhrase) {
		return buserr.New("ErrCmdIllegal")
	}
	certItem, _ := hostRepo.GetCert(repo.WithByName(req.Name))
	if certItem.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	var cert model.RootCert
	if err := copier.Copy(&cert, req); err != nil {
		return err
	}
	privatePath := fmt.Sprintf("%s/.ssh/%s", currentUser.HomeDir, req.Name)
	publicPath := fmt.Sprintf("%s/.ssh/%s.pub", currentUser.HomeDir, req.Name)
	authFilePath := currentUser.HomeDir + "/.ssh/authorized_keys"

	if req.Mode == "input" || req.Mode == "import" {
		if err := os.WriteFile(privatePath, []byte(req.PrivateKey), constant.FilePerm); err != nil {
			return err
		}
		if err := os.WriteFile(publicPath, []byte(req.PublicKey), constant.FilePerm); err != nil {
			return err
		}
	} else {
		command := fmt.Sprintf("ssh-keygen -t %s -f %s/.ssh/%s -N ''", req.EncryptionMode, currentUser.HomeDir, req.Name)
		if len(req.PassPhrase) != 0 {
			command = fmt.Sprintf("ssh-keygen -t %s -P %s -f %s/.ssh/%s | echo y", req.EncryptionMode, req.PassPhrase, currentUser.HomeDir, req.Name)
		}
		stdout, err := cmd.RunDefaultWithStdoutBashC(command)
		if err != nil {
			return fmt.Errorf("generate failed, err: %v, message: %s", err, stdout)
		}
	}

	stdout, err := cmd.RunDefaultWithStdoutBashCf("cat %s >> %s", publicPath, authFilePath)
	if err != nil {
		return fmt.Errorf("generate failed, err: %v, message: %s", err, stdout)
	}

	cert.PrivateKeyPath = privatePath
	cert.PublicKeyPath = publicPath
	if len(cert.PassPhrase) != 0 {
		cert.PassPhrase, _ = encrypt.StringEncrypt(cert.PassPhrase)
	}
	return hostRepo.CreateCert(&cert)
}

func (u *SSHService) SearchRootCerts(req dto.SearchWithPage) (int64, interface{}, error) {
	total, records, err := hostRepo.PageCert(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	var datas []dto.RootCert
	for i := 0; i < len(records); i++ {
		publicItem, err := os.ReadFile(records[i].PublicKeyPath)
		var publicBase64 string
		if err == nil && len(publicItem) != 0 {
			publicBase64 = base64.StdEncoding.EncodeToString(publicItem)
		}
		privateItem, _ := os.ReadFile(records[i].PrivateKeyPath)
		var privateBase64 string
		if err == nil && len(publicItem) != 0 {
			privateBase64 = base64.StdEncoding.EncodeToString(privateItem)
		}
		passPhrase, _ := encrypt.StringDecryptWithBase64(records[i].PassPhrase)
		datas = append(datas, dto.RootCert{
			ID:             records[i].ID,
			CreatedAt:      records[i].CreatedAt,
			Name:           records[i].Name,
			EncryptionMode: records[i].EncryptionMode,
			PassPhrase:     passPhrase,
			PublicKey:      publicBase64,
			PrivateKey:     privateBase64,
			Description:    records[i].Description,
		})
	}
	return total, datas, err
}

func (u *SSHService) DeleteRootCerts(req dto.ForceDelete) error {
	currentUser, err := user.Current()
	if err != nil && !req.ForceDelete {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	authFilePath := currentUser.HomeDir + "/.ssh/authorized_keys"
	authItem, err := os.ReadFile(authFilePath)
	if err != nil && !req.ForceDelete {
		return err
	}
	for _, id := range req.IDs {
		cert, _ := hostRepo.GetCert(repo.WithByID(id))
		if cert.ID == 0 {
			if !req.ForceDelete {
				return buserr.New("ErrRecordNotFound")
			} else {
				continue
			}
		}
		publicItem, err := os.ReadFile(cert.PublicKeyPath)
		if err != nil && !req.ForceDelete {
			return err
		}
		newFile := bytes.ReplaceAll(authItem, publicItem, nil)
		if err := os.WriteFile(authFilePath, newFile, constant.FilePerm); err != nil && !req.ForceDelete {
			return fmt.Errorf("refresh authorized_keys failed, err: %v", err)
		}
		_ = os.Remove(cert.PublicKeyPath)
		_ = os.Remove(cert.PrivateKeyPath)
		if err := hostRepo.DeleteCert(repo.WithByID(id)); err != nil && !req.ForceDelete {
			return err
		}
	}

	return nil
}

type sshFileItem struct {
	Name string
	Year int
}

func (u *SSHService) LoadLog(ctx *gin.Context, req dto.SearchSSHLog) (int64, []dto.SSHHistory, error) {
	var fileList []sshFileItem
	var data []dto.SSHHistory
	baseDir := "/var/log"
	fileItems, err := os.ReadDir(baseDir)
	if err != nil {
		return 0, data, err
	}
	for _, item := range fileItems {
		if item.IsDir() || (!strings.HasPrefix(item.Name(), "secure") && !strings.HasPrefix(item.Name(), "auth")) {
			continue
		}
		info, _ := item.Info()
		itemPath := path.Join(baseDir, info.Name())
		if !strings.HasSuffix(item.Name(), ".gz") {
			fileList = append(fileList, sshFileItem{Name: itemPath, Year: info.ModTime().Year()})
			continue
		}
		itemFileName := strings.TrimSuffix(itemPath, ".gz")
		if _, err := os.Stat(itemFileName); err != nil && os.IsNotExist(err) {
			if err := handleGunzip(itemPath); err == nil {
				fileList = append(fileList, sshFileItem{Name: itemFileName, Year: info.ModTime().Year()})
			}
		}
	}
	fileList = sortFileList(fileList)

	command := ""
	if len(req.Info) != 0 {
		command = fmt.Sprintf(" | grep '%s'", req.Info)
	}

	showCountFrom := (req.Page - 1) * req.PageSize
	showCountTo := req.Page * req.PageSize
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	itemFailed, itemTotal := 0, 0
	for _, file := range fileList {
		commandItem := ""
		if strings.HasPrefix(path.Base(file.Name), "secure") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -a 'Failed password for' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE '(Failed password for|Accepted)' %s", file.Name, command)
			}
		}
		if strings.HasPrefix(path.Base(file.Name), "auth.log") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -aE 'Failed password for|Connection closed by authenticating user' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE \"(Failed password for|Connection closed by authenticating user|Accepted)\" %s", file.Name, command)
			}
		}
		dataItem, successCount, failedCount := loadSSHData(ctx, commandItem, showCountFrom, showCountTo, file.Year, nyc)
		itemFailed += failedCount
		itemTotal += successCount + failedCount
		showCountFrom = showCountFrom - (successCount + failedCount)
		if showCountTo != -1 {
			showCountTo = showCountTo - (successCount + failedCount)
		}
		data = append(data, dataItem...)
	}

	total := itemTotal
	if req.Status == constant.StatusFailed {
		total = itemFailed
	}
	if req.Status == constant.StatusSuccess {
		total = itemTotal - itemFailed
	}
	return int64(total), data, nil
}

func (u *SSHService) ExportLog(ctx *gin.Context, req dto.SearchSSHLog) (string, error) {
	_, logs, err := u.LoadLog(ctx, req)
	if err != nil {
		return "", err
	}
	if len(logs) == 0 {
		return "", buserr.New("ErrRecordNotFound")
	}
	tmpFileName := path.Join(global.Dir.TmpDir, "export/ssh-log", fmt.Sprintf("1panel-ssh-log-%s.csv", time.Now().Format(constant.DateTimeSlimLayout)))
	if _, err := os.Stat(path.Dir(tmpFileName)); err != nil {
		_ = os.MkdirAll(path.Dir(tmpFileName), constant.DirPerm)
	}
	if err := csvexport.ExportSSHLogs(tmpFileName, logs); err != nil {
		return "", err
	}
	return tmpFileName, nil
}

func (u *SSHService) LoadSSHFile(name string) (string, error) {
	var fileName string
	switch name {
	case "authKeys":
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("load current user failed, err: %v", err)
		}
		fileName = currentUser.HomeDir + "/.ssh/authorized_keys"
	case "sshdConf":
		fileName = "/etc/ssh/sshd_config"
	default:
		return "", buserr.WithName("ErrNotSupportType", name)
	}
	if _, err := os.Stat(fileName); err != nil {
		return "", buserr.WithErr("ErrHttpReqNotFound", err)
	}
	content, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (u *SSHService) UpdateByFile(req dto.SettingUpdate) error {
	var fileName string
	switch req.Key {
	case "authKeys":
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("load current user failed, err: %v", err)
		}
		fileName = currentUser.HomeDir + "/.ssh/authorized_keys"
	case "sshdConf":
		fileName = "/etc/ssh/sshd_config"
	default:
		return buserr.WithName("ErrNotSupportType", req.Key)
	}
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(req.Value); err != nil {
		return err
	}
	if req.Key == "authKeys" {
		return nil
	}
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}
	sudo := cmd.SudoHandleCmd()
	_, _ = cmd.RunDefaultWithStdoutBashCf("%s systemctl restart %s", sudo, serviceName)
	return nil
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

func updateSSHConf(oldFiles []string, param string, value string) []string {
	var valueItems []string
	if param != "ListenAddress" {
		valueItems = append(valueItems, value)
	} else {
		if value != "" {
			valueItems = strings.Split(value, ",")
		}
	}
	var newFiles []string
	for _, line := range oldFiles {
		lineItem := strings.TrimSpace(line)
		if (strings.HasPrefix(lineItem, param) || strings.HasPrefix(lineItem, fmt.Sprintf("#%s", param))) && len(valueItems) != 0 {
			newFiles = append(newFiles, fmt.Sprintf("%s %s", param, valueItems[0]))
			valueItems = valueItems[1:]
			continue
		}
		if strings.HasPrefix(lineItem, param) && len(valueItems) == 0 {
			newFiles = append(newFiles, fmt.Sprintf("#%s", line))
			continue
		}
		newFiles = append(newFiles, line)
	}
	if len(valueItems) != 0 {
		for _, item := range valueItems {
			newFiles = append(newFiles, fmt.Sprintf("%s %s", param, item))
		}
	}
	return newFiles
}

func loadSSHData(ctx *gin.Context, command string, showCountFrom, showCountTo, currentYear int, nyc *time.Location) ([]dto.SSHHistory, int, int) {
	var (
		datas        []dto.SSHHistory
		successCount int
		failedCount  int
	)
	getLoc, err := geo.NewGeo()
	if err != nil {
		return datas, 0, 0
	}
	stdout, err := cmd.RunDefaultWithStdoutBashC(command)
	if err != nil {
		return datas, 0, 0
	}
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		var itemData dto.SSHHistory
		switch {
		case strings.Contains(lines[i], "Failed password for"):
			itemData = loadFailedSecureDatas(lines[i])
			if checkIsStandard(itemData) {
				if successCount+failedCount >= showCountFrom && (showCountTo == -1 || successCount+failedCount < showCountTo) {
					itemData.Area, _ = geo.GetIPLocation(getLoc, itemData.Address, common.GetLang(ctx))
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Connection closed by authenticating user"):
			itemData = loadFailedAuthDatas(lines[i])
			if checkIsStandard(itemData) {
				if successCount+failedCount >= showCountFrom && (showCountTo == -1 || successCount+failedCount < showCountTo) {
					itemData.Area, _ = geo.GetIPLocation(getLoc, itemData.Address, common.GetLang(ctx))
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Accepted "):
			itemData = loadSuccessDatas(lines[i])
			if checkIsStandard(itemData) {
				if successCount+failedCount >= showCountFrom && (showCountTo == -1 || successCount+failedCount < showCountTo) {
					itemData.Area, _ = geo.GetIPLocation(getLoc, itemData.Address, common.GetLang(ctx))
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				successCount++
			}
		}
	}
	return datas, successCount, failedCount
}

func loadSuccessDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	data.AuthMode = parts[4+index]
	data.User = parts[6+index]
	data.Address = parts[8+index]
	data.Port = parts[10+index]
	data.Status = constant.StatusSuccess
	return data
}
func loadFailedAuthDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	switch index {
	case 1:
		data.User = parts[9]
	case 2:
		data.User = parts[10]
	default:
		data.User = parts[7]
	}
	data.AuthMode = parts[6+index]
	data.Address = parts[9+index]
	data.Port = parts[11+index]
	data.Status = constant.StatusFailed
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}
func loadFailedSecureDatas(line string) dto.SSHHistory {
	var data dto.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	if strings.Contains(line, " invalid ") {
		data.AuthMode = parts[4+index]
		index += 2
	} else {
		data.AuthMode = parts[4+index]
	}
	data.User = parts[6+index]
	data.Address = parts[8+index]
	data.Port = parts[10+index]
	data.Status = constant.StatusFailed
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}

func checkIsStandard(item dto.SSHHistory) bool {
	if len(item.Address) == 0 {
		return false
	}
	portItem, _ := strconv.Atoi(item.Port)
	return portItem != 0
}

func handleGunzip(path string) error {
	if _, err := cmd.RunDefaultWithStdoutBashCf("gunzip %s", path); err != nil {
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

func analyzeDateStr(parts []string) (int, string) {
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err == nil {
		if len(parts) < 12 {
			return 0, ""
		}
		return 0, t.Format("2006 Jan 2 15:04:05")
	}
	t, err = time.Parse(constant.DateTimeLayout, fmt.Sprintf("%s %s", parts[0], parts[1]))
	if err == nil {
		if len(parts) < 14 {
			return 0, ""
		}
		return 1, t.Format("2006 Jan 2 15:04:05")
	}

	if len(parts) < 14 {
		return 0, ""
	}
	return 2, fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
}

func loadEncryptioMode(content string) string {
	if strings.HasPrefix(content, "ssh-rsa") {
		return "rsa"
	}
	if strings.HasPrefix(content, "ssh-ed25519") {
		return "ed25519"
	}
	if strings.HasPrefix(content, "ssh-ecdsa") {
		return "ecdsa"
	}
	if strings.HasPrefix(content, "ssh-dsa") {
		return "dsa"
	}
	return ""
}

func updateLocalConn(newPort uint) error {
	conn, _ := settingRepo.GetValueByKey("LocalSSHConn")
	if len(conn) == 0 {
		return nil
	}
	connItem, err := encrypt.StringDecrypt(conn)
	if err != nil {
		return err
	}
	var data dto.SSHConnData
	if err := json.Unmarshal([]byte(connItem), &data); err != nil {
		return err
	}
	data.Port = newPort
	connNew, err := json.Marshal(data)
	if err != nil {
		return err
	}
	connNewItem, err := encrypt.StringEncrypt(string(connNew))
	if err != nil {
		return err
	}
	return settingRepo.Update("LocalSSHConn", connNewItem)
}

func updateSSHSocketFile(newPort string) error {
	active, _ := systemctl.IsActive("ssh.socket")
	if !active {
		return nil
	}
	filepath := "/usr/lib/systemd/system/ssh.socket"
	file, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(file), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "ListenStream=") {
			parts := strings.Split(lines[i], ":")
			if len(parts) > 1 {
				lines[i] = strings.ReplaceAll(lines[i], parts[len(parts)-1], newPort)
				continue
			}
			parts = strings.Split(lines[i], "=")
			if len(parts) > 1 {
				lines[i] = strings.ReplaceAll(lines[i], parts[len(parts)-1], newPort)
			}
		}
	}
	fileItem, err := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return err
	}
	defer fileItem.Close()
	if _, err = fileItem.WriteString(strings.Join(lines, "\n")); err != nil {
		return err
	}
	_ = cmd.RunDefaultBashC("systemctl daemon-reload && systemctl restart ssh.socket")
	return nil
}
