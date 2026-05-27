package service

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/controller"
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
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/pkg/errors"
)

const sshPath = "/etc/ssh/sshd_config"
const defaultSSHPort = "22"
const sshManagedMarker = "# config by 1panel"

type SSHService struct{}

type ISSHService interface {
	GetSSHInfo() (*dto.SSHInfo, error)
	OperateSSH(operation string) error
	Update(req dto.SSHUpdate) error
	LoadSSHFile(name string) (string, error)
	UpdateByFile(req dto.SSHConfUpdate) error

	LoadLog(ctx *gin.Context, req dto.SearchSSHLog) (int64, []dto.SSHHistory, error)
	ExportLog(ctx *gin.Context, req dto.SearchSSHLog) (string, error)

	SyncRootCert() error
	CreateRootCert(req dto.RootCertOperate) error
	EditRootCert(req dto.RootCertOperate) error
	SearchRootCerts(req dto.SearchWithPage) (int64, interface{}, error)
	DeleteRootCerts(req dto.ForceDelete) error
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

type sshDirective struct {
	Key     string
	Value   string
	File    string
	Line    int
	InMatch bool
}

type sshConfigFile struct {
	Path     string `json:"path"`
	Priority int    `json:"priority"`
}

type sshManagedBlock struct {
	Start int
	End   int
}

func (b sshManagedBlock) contains(line int) bool {
	return line >= b.Start && line < b.End
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
	loadSSHServiceStatus(&data)
	loadSSHConfigInfo(&data)
	data.CurrentUser = loadCurrentUserName()
	return &data, nil
}

func (u *SSHService) OperateSSH(operation string) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}
	if operation == "enable" || operation == "disable" {
		serviceName += ".service"
	}
	if operation == "stop" {
		isSocketActive, _ := controller.CheckActive(serviceName + ".socket")
		if isSocketActive {
			if err := controller.HandleStop(serviceName + ".socket"); err != nil {
				global.LOG.Errorf("handle stop %s.socket failed, err: %v", serviceName, err)
			}
		}
	}

	if err := controller.Handle(operation, serviceName); err != nil {
		return fmt.Errorf("%s %s failed, err: %v", operation, serviceName, err)
	}
	return nil
}

func (u *SSHService) Update(req dto.SSHUpdate) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}

	directives, _, err := parseSSHConfigTree(sshPath)
	if err != nil {
		return err
	}
	oldPortValue := strings.Join(loadSSHPortValues(directives), ",")
	if err := updateSSHDirectiveValue(req.Key, req.NewValue, directives); err != nil {
		return err
	}
	if req.Key == "Port" {
		handleSSHPortUpdate(oldPortValue, req.NewValue)
	}

	_ = controller.HandleRestart(serviceName)
	return nil
}

func loadSSHServiceStatus(data *dto.SSHInfo) {
	serviceName, err := loadServiceName()
	if err != nil {
		data.IsExist = false
		data.Message = err.Error()
		return
	}

	active, err := controller.CheckActive(serviceName)
	data.IsActive = active
	if !active && err != nil {
		data.Message = err.Error()
	}

	enable, err := controller.CheckEnable(serviceName)
	if err != nil {
		data.AutoStart = false
		return
	}
	data.AutoStart = enable
}

func loadSSHConfigInfo(data *dto.SSHInfo) {
	directives, _, err := parseSSHConfigTree(sshPath)
	if err != nil {
		data.Message = err.Error()
		data.IsActive = false
		return
	}

	if port := strings.Join(loadSSHPortValues(directives), ","); port != "" {
		data.Port = port
	}
	data.ListenAddress = strings.Join(loadSSHDirectiveValues(directives, "ListenAddress"), ",")
	if value, ok := loadFirstSSHDirectiveValue(directives, "PasswordAuthentication"); ok {
		data.PasswordAuthentication = value
	}
	if value, ok := loadFirstSSHDirectiveValue(directives, "PubkeyAuthentication"); ok {
		data.PubkeyAuthentication = value
	}
	if value, ok := loadFirstSSHDirectiveValue(directives, "PermitRootLogin"); ok {
		data.PermitRootLogin = strings.ReplaceAll(value, "prohibit-password", "without-password")
	}
	if value, ok := loadFirstSSHDirectiveValue(directives, "UseDNS"); ok {
		data.UseDNS = value
	}
}

func loadCurrentUserName() string {
	currentUser, err := user.Current()
	if err != nil || currentUser.Name == "" {
		return "root"
	}
	return currentUser.Name
}

func handleSSHPortUpdate(oldValue, newValue string) {
	sudo := cmd.SudoHandleCmd()
	newPorts := splitSSHPorts(newValue)
	oldPorts := splitSSHPorts(oldValue)

	stdout, _ := runWithOptionalSudo(sudo, "getenforce")
	if stdout == "Enforcing\n" {
		for _, port := range diffSSHPorts(oldPorts, newPorts) {
			if _, err := runWithOptionalSudo(sudo, "semanage", "port", "-d", "-t", "ssh_port_t", "-p", "tcp", port); err != nil {
				global.LOG.Warnf("remove selinux ssh port %s failed, err: %v", port, err)
			}
		}
		for _, port := range newPorts {
			_, _ = runWithOptionalSudo(sudo, "semanage", "port", "-a", "-t", "ssh_port_t", "-p", "tcp", port)
		}
	}

	removedPorts, err := parseSSHPortsToInts(diffSSHPorts(oldPorts, newPorts))
	if err != nil {
		global.LOG.Errorf("parse removed ssh ports failed, err: %v", err)
	} else {
		addedPorts, err := parseSSHPortsToInts(diffSSHPorts(newPorts, oldPorts))
		if err != nil {
			global.LOG.Errorf("parse added ssh ports failed, err: %v", err)
		} else if err := OperateFirewallPort(removedPorts, addedPorts); err != nil {
			global.LOG.Errorf("reset firewall rules %s -> %s failed, err: %v", oldValue, newValue, err)
		}
	}

	primaryPort, err := loadPrimarySSHPort(newValue)
	if err != nil {
		global.LOG.Errorf("load primary ssh port from %s failed, err: %v", newValue, err)
		return
	}
	if err := updateLocalConn(uint(primaryPort)); err != nil {
		global.LOG.Errorf("update local conn for terminal failed, err: %v", err)
	}
	if err := updateSSHSocketFile(strconv.Itoa(primaryPort)); err != nil {
		global.LOG.Errorf("update port for ssh.socket failed, err: %v", err)
	}
}

func splitSSHPorts(value string) []string {
	var ports []string
	for _, item := range strings.Split(value, ",") {
		port := strings.TrimSpace(item)
		if port != "" {
			ports = append(ports, port)
		}
	}
	return ports
}

func diffSSHPorts(left, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}
	var diff []string
	for _, item := range left {
		if _, ok := rightSet[item]; ok {
			continue
		}
		diff = append(diff, item)
	}
	return diff
}

func loadPrimarySSHPort(value string) (int, error) {
	ports := splitSSHPorts(value)
	if len(ports) == 0 {
		return 0, fmt.Errorf("ssh port is empty")
	}
	return strconv.Atoi(ports[0])
}

func parseSSHPortsToInts(ports []string) ([]int, error) {
	var values []int
	for _, port := range ports {
		value, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func runWithOptionalSudo(sudo, name string, args ...string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Second))
	if sudo != "" {
		commandArgs := append([]string{name}, args...)
		return cmdMgr.RunWithStdout("sudo", commandArgs...)
	}
	return cmdMgr.RunWithStdout(name, args...)
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

func (u *SSHService) CreateRootCert(req dto.RootCertOperate) error {
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

	if info, err := os.Stat(authFilePath); err == nil && info.Size() > 0 {
		f, err := os.Open(authFilePath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		if _, err := f.Seek(-1, 2); err == nil {
			buf := make([]byte, 1)
			if _, err := f.Read(buf); err == nil && buf[0] != '\n' {
				appendFile, err := os.OpenFile(authFilePath, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				if _, err := appendFile.Write([]byte("\n")); err != nil {
					_ = appendFile.Close()
					return err
				}
				_ = appendFile.Close()
			}
		}
	}

	if req.Mode == "input" || req.Mode == "import" {
		if err := os.WriteFile(privatePath, []byte(req.PrivateKey), constant.FilePerm); err != nil {
			return err
		}
		if err := os.WriteFile(publicPath, []byte(req.PublicKey), constant.FilePerm); err != nil {
			return err
		}
	} else {
		tmpPrivatePath := privatePath + ".tmp"
		tmpPublicPath := privatePath + ".tmp.pub"
		cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Minute))
		args := []string{
			"-t", req.EncryptionMode,
			"-f", tmpPrivatePath,
			"-N", req.PassPhrase,
			"-q",
		}

		if err := cmdMgr.Run("ssh-keygen", args...); err != nil {
			_ = os.Remove(tmpPrivatePath)
			_ = os.Remove(tmpPublicPath)
			return fmt.Errorf("generate failed, %v", err)
		}
		if err := os.Rename(tmpPrivatePath, privatePath); err != nil {
			return fmt.Errorf("replace private key failed, %v", err)
		}
		if err := os.Rename(tmpPublicPath, publicPath); err != nil {
			return fmt.Errorf("replace public key failed, %v", err)
		}
	}

	publicKeyBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return fmt.Errorf("read public key failed, %v", err)
	}

	cleanKey := strings.TrimRight(string(publicKeyBytes), "\n") + "\n"
	authFile, err := os.OpenFile(authFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("open authorized_keys failed, %v", err)
	}
	defer func() { _ = authFile.Close() }()

	if _, err := authFile.Write([]byte(cleanKey)); err != nil {
		return fmt.Errorf("append authorized_keys failed, %v", err)
	}

	cert.PrivateKeyPath = privatePath
	cert.PublicKeyPath = publicPath
	if len(cert.PassPhrase) != 0 {
		cert.PassPhrase, _ = encrypt.StringEncrypt(cert.PassPhrase)
	}
	return hostRepo.SaveCert(&cert)
}

func (u *SSHService) EditRootCert(req dto.RootCertOperate) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	certItem, _ := hostRepo.GetCert(repo.WithByID(req.ID))
	if certItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	oldPublicItem, err := os.ReadFile(certItem.PublicKeyPath)
	if err != nil {
		return err
	}

	var cert model.RootCert
	if err := copier.Copy(&cert, req); err != nil {
		return err
	}
	cert.PrivateKeyPath = fmt.Sprintf("%s/.ssh/%s", currentUser.HomeDir, req.Name)
	cert.PublicKeyPath = fmt.Sprintf("%s/.ssh/%s.pub", currentUser.HomeDir, req.Name)
	if err := os.WriteFile(cert.PrivateKeyPath, []byte(req.PrivateKey), constant.FilePerm); err != nil {
		return err
	}
	if err := os.WriteFile(cert.PublicKeyPath, []byte(req.PublicKey), constant.FilePerm); err != nil {
		return err
	}

	authFilePath := currentUser.HomeDir + "/.ssh/authorized_keys"
	authItem, err := os.ReadFile(authFilePath)
	if err != nil {
		return err
	}
	oldPublic := strings.ReplaceAll(string(oldPublicItem), "\n", "")
	newPublic := strings.ReplaceAll(string(req.PublicKey), "\n", "")
	lines := strings.Split(string(authItem), "\n")
	var newFiles []string
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) != 0 && lines[i] != oldPublic && lines[i] != newPublic {
			newFiles = append(newFiles, lines[i])
		}
	}
	newFiles = append(newFiles, newPublic)
	if err := os.WriteFile(authFilePath, []byte(strings.Join(newFiles, "\n")), constant.FilePerm); err != nil {
		return fmt.Errorf("refresh authorized_keys failed, err: %v", err)
	}
	if len(cert.PassPhrase) != 0 {
		cert.PassPhrase, _ = encrypt.StringEncrypt(cert.PassPhrase)
	}
	return hostRepo.SaveCert(&cert)
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
		if item.IsDir() || (!strings.HasPrefix(item.Name(), "secure") && !strings.HasPrefix(item.Name(), "auth.log")) {
			continue
		}
		info, _ := item.Info()
		itemPath := path.Join(baseDir, info.Name())
		if strings.HasSuffix(itemPath, ".gz") {
			if _, err := os.Stat(strings.TrimSuffix(itemPath, ".gz")); err == nil {
				continue
			}
		}
		fileList = append(fileList, sshFileItem{Name: itemPath, Year: info.ModTime().Year()})
	}
	fileList = sortFileList(fileList)

	filter := ""
	if len(req.Info) != 0 {
		if cmd.CheckIllegal(req.Info) {
			return 0, data, buserr.New("ErrCmdIllegal")
		}
		filter = req.Info
	}

	showCountFrom := (req.Page - 1) * req.PageSize
	showCountTo := req.Page * req.PageSize
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	itemFailed, itemTotal := 0, 0
	for _, file := range fileList {
		dataItem, successCount, failedCount := loadSSHData(
			ctx,
			file.Name,
			req.Status,
			filter,
			req.StartTime,
			req.EndTime,
			showCountFrom,
			showCountTo,
			file.Year,
			nyc,
		)
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
	case "sshdConfOptions":
		_, fileList, err := parseSSHConfigTree(sshPath)
		if err != nil {
			return "", err
		}
		content, err := json.Marshal(fileList)
		if err != nil {
			return "", err
		}
		return string(content), nil
	default:
		if strings.HasPrefix(name, "sshdConfPath:") {
			fileName = strings.TrimPrefix(name, "sshdConfPath:")
			if !isSSHConfigPathAllowed(fileName) {
				return "", buserr.WithName("ErrNotSupportType", name)
			}
		} else {
			return "", buserr.WithName("ErrNotSupportType", name)
		}
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

func (u *SSHService) UpdateByFile(req dto.SSHConfUpdate) error {
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
	case "sshdConfPath":
		fileName = req.Path
		if !isSSHConfigPathAllowed(fileName) {
			return buserr.WithName("ErrNotSupportType", req.Key)
		}
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
	_ = controller.HandleRestart(serviceName)
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

func parseSSHConfigTree(root string) ([]sshDirective, []sshConfigFile, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, err
	}
	var (
		directives []sshDirective
		fileList   []sshConfigFile
		order      int
		priority   int
	)
	visited := make(map[string]bool)
	if err := parseSSHConfigFile(rootPath, false, visited, &directives, &fileList, &order, &priority); err != nil {
		return nil, nil, err
	}
	return directives, fileList, nil
}

func parseSSHConfigFile(
	fileName string,
	inMatch bool,
	visited map[string]bool,
	directives *[]sshDirective,
	fileList *[]sshConfigFile,
	order *int,
	priority *int,
) error {
	absolutePath, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	if visited[absolutePath] {
		return nil
	}
	visited[absolutePath] = true

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	currentMatch := inMatch
	fileRegistered := false
	for index, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := parseSSHConfigLine(line)
		if !ok {
			continue
		}
		if strings.EqualFold(key, "Match") {
			currentMatch = true
			continue
		}
		if strings.EqualFold(key, "Include") {
			for _, includePath := range expandSSHIncludePaths(absolutePath, value) {
				if err := parseSSHConfigFile(includePath, currentMatch, visited, directives, fileList, order, priority); err != nil {
					global.LOG.Warnf("parse ssh include %s failed, err: %v", includePath, err)
				}
			}
			continue
		}
		if !fileRegistered {
			registerSSHConfigFile(absolutePath, fileList, priority)
			fileRegistered = true
		}
		*order++
		*directives = append(*directives, sshDirective{
			Key:     key,
			Value:   normalizeSSHDirectiveValue(key, value),
			File:    absolutePath,
			Line:    index,
			InMatch: currentMatch,
		})
	}
	if !fileRegistered {
		registerSSHConfigFile(absolutePath, fileList, priority)
	}
	return nil
}

func registerSSHConfigFile(fileName string, fileList *[]sshConfigFile, priority *int) {
	for _, item := range *fileList {
		if item.Path == fileName {
			return
		}
	}
	*priority++
	*fileList = append(*fileList, sshConfigFile{Path: fileName, Priority: *priority})
}

func parseSSHConfigLine(line string) (string, string, bool) {
	line = trimSSHInlineComment(line)
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}
	return fields[0], strings.Join(fields[1:], " "), true
}

func trimSSHInlineComment(line string) string {
	inQuote := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inQuote = !inQuote
		case '#':
			if !inQuote && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
				return strings.TrimSpace(line[:i])
			}
		}
	}
	return strings.TrimSpace(line)
}

func expandSSHIncludePaths(baseFile, value string) []string {
	var includeFiles []string
	for _, item := range strings.Fields(strings.ReplaceAll(value, "\"", "")) {
		pattern := item
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(filepath.Dir(baseFile), pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		sort.Strings(matches)
		includeFiles = append(includeFiles, matches...)
	}
	return includeFiles
}

func normalizeSSHDirectiveValue(key, value string) string {
	if strings.EqualFold(key, "PermitRootLogin") && value == "prohibit-password" {
		return "without-password"
	}
	return value
}

func loadFirstSSHDirectiveValue(directives []sshDirective, key string) (string, bool) {
	for _, item := range directives {
		if item.InMatch || !strings.EqualFold(item.Key, key) {
			continue
		}
		return item.Value, true
	}
	return "", false
}

func loadSSHDirectiveValues(directives []sshDirective, key string) []string {
	var values []string
	for _, item := range directives {
		if item.InMatch || !strings.EqualFold(item.Key, key) {
			continue
		}
		values = append(values, item.Value)
	}
	return values
}

func loadSSHPortValues(directives []sshDirective) []string {
	values := loadSSHDirectiveValues(directives, "Port")
	if len(values) == 0 {
		return []string{defaultSSHPort}
	}
	return values
}

func updateSSHDirectiveValue(key, value string, directives []sshDirective) error {
	if key == "Port" || key == "ListenAddress" {
		return rewriteSSHMultiValueDirective(key, value, directives)
	}
	return updateSSHSingleDirectiveValue(key, value, directives)
}

func updateSSHSingleDirectiveValue(key, value string, directives []sshDirective) error {
	block, err := loadSSHManagedBlockFromFile(sshPath)
	if err != nil {
		return err
	}
	for _, item := range directives {
		if item.InMatch || !strings.EqualFold(item.Key, key) {
			continue
		}
		if item.File == sshPath && block.contains(item.Line) {
			continue
		}
		if err := commentSSHConfigLines(item.File, []int{item.Line}); err != nil {
			return err
		}
	}
	return rewriteSSHManagedDirectives(sshPath, key, []string{fmt.Sprintf("%s %s", key, value)})
}

func rewriteSSHMultiValueDirective(key, value string, directives []sshDirective) error {
	block, err := loadSSHManagedBlockFromFile(sshPath)
	if err != nil {
		return err
	}
	targetFiles := make(map[string][]int)
	for _, item := range directives {
		if item.InMatch || !strings.EqualFold(item.Key, key) {
			continue
		}
		if item.File == sshPath && block.contains(item.Line) {
			continue
		}
		targetFiles[item.File] = append(targetFiles[item.File], item.Line)
	}
	for fileName, lines := range targetFiles {
		if err := commentSSHConfigLines(fileName, lines); err != nil {
			return err
		}
	}

	return rewriteSSHManagedDirectives(sshPath, key, buildSSHDirectiveLines(key, value))
}

func buildSSHDirectiveLines(key, value string) []string {
	var directives []string
	for _, item := range strings.Split(value, ",") {
		if item != "" {
			directives = append(directives, fmt.Sprintf("%s %s", key, item))
		}
	}
	return directives
}

func rewriteSSHManagedDirectives(fileName, key string, newDirectives []string) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	lines, insertAt := ensureSSHManagedMarker(lines)

	block, ok := loadSSHManagedBlock(lines)
	if ok {
		var filtered []string
		for index := block.Start; index < block.End; index++ {
			line := strings.TrimSpace(lines[index])
			if line == "" || strings.HasPrefix(line, "#") {
				filtered = append(filtered, lines[index])
				continue
			}
			itemKey, _, ok := parseSSHConfigLine(line)
			if ok && strings.EqualFold(itemKey, key) {
				continue
			}
			filtered = append(filtered, lines[index])
		}
		lines = append(append(lines[:block.Start], filtered...), lines[block.End:]...)
		insertAt = block.Start
	}

	if len(newDirectives) != 0 {
		lines = insertSSHDirectivesAt(lines, insertAt, newDirectives)
	}
	return os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), constant.FilePerm)
}

func commentSSHConfigLines(fileName string, targetLines []int) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	targetSet := make(map[int]struct{}, len(targetLines))
	for _, line := range targetLines {
		targetSet[line] = struct{}{}
	}
	for index := range lines {
		if _, ok := targetSet[index]; !ok {
			continue
		}
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines[index] = "#" + lines[index]
	}
	return os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), constant.FilePerm)
}

func loadSSHInsertIndex(lines []string) int {
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "match ") {
			return index
		}
	}
	return len(lines)
}

func ensureSSHManagedMarker(lines []string) ([]string, int) {
	for index, line := range lines {
		if strings.TrimSpace(line) == sshManagedMarker {
			return lines, index + 1
		}
	}
	insertAt := loadSSHInsertIndex(lines)
	lines = insertSSHDirectivesAt(lines, insertAt, []string{sshManagedMarker})
	return lines, insertAt + 1
}

func loadSSHManagedBlock(lines []string) (sshManagedBlock, bool) {
	for index, line := range lines {
		if strings.TrimSpace(line) != sshManagedMarker {
			continue
		}
		end := len(lines)
		for next := index + 1; next < len(lines); next++ {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(lines[next])), "match ") {
				end = next
				break
			}
		}
		return sshManagedBlock{Start: index + 1, End: end}, true
	}
	return sshManagedBlock{}, false
}

func loadSSHManagedBlockFromFile(fileName string) (sshManagedBlock, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return sshManagedBlock{}, err
	}
	block, ok := loadSSHManagedBlock(strings.Split(string(content), "\n"))
	if !ok {
		return sshManagedBlock{}, nil
	}
	return block, nil
}

func insertSSHDirectivesAt(lines []string, insertAt int, directives []string) []string {
	newLines := make([]string, 0, len(lines)+len(directives))
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, directives...)
	newLines = append(newLines, lines[insertAt:]...)
	return newLines
}

func isSSHConfigPathAllowed(fileName string) bool {
	if fileName == "" {
		return false
	}
	absolutePath, err := filepath.Abs(fileName)
	if err != nil {
		return false
	}
	if absolutePath == sshPath {
		return true
	}
	_, fileList, err := parseSSHConfigTree(sshPath)
	if err != nil {
		return false
	}
	for _, item := range fileList {
		if item.Path == absolutePath {
			return true
		}
	}
	return false
}

type sshLogHeader struct {
	DateStr string
	Process string
	PID     string
	Message string
}

type sshParsedLog struct {
	History    dto.SSHHistory
	SessionKey string
	Raw        string
	Search     string
	Index      int
}

func loadSSHData(
	ctx *gin.Context,
	filePath, status, filter string,
	startTime, endTime time.Time,
	showCountFrom, showCountTo, currentYear int,
	nyc *time.Location,
) ([]dto.SSHHistory, int, int) {
	var (
		datas        []dto.SSHHistory
		successCount int
		failedCount  int
	)
	getLoc, err := geo.NewGeo()
	if err != nil {
		return datas, 0, 0
	}
	lines, err := loadSSHLogLines(filePath)
	if err != nil {
		return datas, 0, 0
	}
	items := collectSSHLogItems(lines, filter, status)
	for i := len(items) - 1; i >= 0; i-- {
		itemData := items[i].History
		if !matchSSHLogStatus(status, itemData.Status) || !checkIsStandard(itemData) {
			continue
		}
		itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
		if !isSSHLogWithinTimeRange(itemData.Date, startTime, endTime) {
			continue
		}
		if successCount+failedCount >= showCountFrom && (showCountTo == -1 || successCount+failedCount < showCountTo) {
			itemData.Area, _ = geo.GetIPLocation(getLoc, itemData.Address, common.GetLang(ctx))
			datas = append(datas, itemData)
		}
		if itemData.Status == constant.StatusSuccess {
			successCount++
		} else {
			failedCount++
		}
	}
	return datas, successCount, failedCount
}

func isSSHLogWithinTimeRange(itemTime, startTime, endTime time.Time) bool {
	if startTime.IsZero() || endTime.IsZero() {
		return true
	}
	return itemTime.After(startTime) && itemTime.Before(endTime)
}

func collectSSHLogItems(lines []string, filter, status string) []sshParsedLog {
	var items []sshParsedLog
	auxiliaryIndex := make(map[string]int)
	sessionHasAuthEvent := make(map[string]bool)
	for lineIndex, line := range lines {
		if !shouldParseSSHLogLine(line, status, filter) {
			continue
		}
		item, ok := parseSSHLogLine(line)
		if !ok || !checkIsStandard(item.History) {
			continue
		}
		item.Index = lineIndex
		item.Search = line
		if isSSHAuthEvent(item) && item.SessionKey != "" {
			sessionHasAuthEvent[item.SessionKey] = true
			items = append(items, item)
			continue
		}
		if index, ok := auxiliaryIndex[item.SessionKey]; ok {
			items[index].Search += "\n" + line
			continue
		}
		auxiliaryIndex[item.SessionKey] = len(items)
		items = append(items, item)
	}
	items = filterRedundantSSHSessionItems(items, sessionHasAuthEvent)
	if filter != "" {
		filteredItems := make([]sshParsedLog, 0, len(items))
		for _, item := range items {
			if strings.Contains(item.Search, filter) {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})
	return items
}

func filterRedundantSSHSessionItems(items []sshParsedLog, sessionHasAuthEvent map[string]bool) []sshParsedLog {
	filteredItems := make([]sshParsedLog, 0, len(items))
	for _, item := range items {
		if !isSSHAuthEvent(item) && item.SessionKey != "" && sessionHasAuthEvent[item.SessionKey] {
			continue
		}
		filteredItems = append(filteredItems, item)
	}
	return filteredItems
}

func isSSHAuthEvent(item sshParsedLog) bool {
	return item.History.Status == constant.StatusSuccess || item.History.AuthMode != ""
}

func shouldParseSSHLogLine(line, status, filter string) bool {
	if !strings.Contains(line, "sshd") {
		return false
	}
	if filter != "" {
		return containsKnownSSHLogMessage(line)
	}
	switch status {
	case constant.StatusSuccess:
		return strings.Contains(line, "Accepted ")
	case constant.StatusFailed:
		// Accepted lines are still needed here to suppress redundant session failure records from the same SSH session.
		return containsFailedSSHLogMessage(line) || strings.Contains(line, "Accepted ")
	default:
		return containsKnownSSHLogMessage(line)
	}
}

func containsKnownSSHLogMessage(line string) bool {
	return strings.Contains(line, "Accepted ") || containsFailedSSHLogMessage(line)
}

func containsFailedSSHLogMessage(line string) bool {
	return strings.Contains(line, "Failed ") ||
		strings.Contains(line, "Invalid user ") ||
		strings.Contains(line, "Connection closed by ") ||
		strings.Contains(line, "Disconnected from ") ||
		strings.Contains(line, "Received disconnect from ") ||
		strings.Contains(line, "maximum authentication attempts exceeded") ||
		strings.Contains(line, " not allowed")
}

func loadSSHLogLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func matchSSHLogStatus(requestStatus, itemStatus string) bool {
	switch requestStatus {
	case constant.StatusSuccess:
		return itemStatus == constant.StatusSuccess
	case constant.StatusFailed:
		return itemStatus == constant.StatusFailed
	default:
		return true
	}
}

func parseSSHLogLine(line string) (sshParsedLog, bool) {
	header, ok := parseSSHLogHeader(line)
	if !ok {
		return sshParsedLog{}, false
	}
	data, ok := parseSSHLogMessage(header.Message)
	if !ok {
		return sshParsedLog{}, false
	}
	data.DateStr = header.DateStr
	data.Message = header.Message
	sessionKey := loadSSHLogSessionKey(header, data, line)
	return sshParsedLog{History: data, SessionKey: sessionKey, Raw: line}, true
}

func parseSSHLogHeader(line string) (sshLogHeader, bool) {
	for _, pattern := range []string{re.SSHRFC3339LinePattern, re.SSHDateTimeLinePattern, re.SSHSyslogLinePattern} {
		matches := re.GetRegex(pattern).FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}
		dateStr := normalizeSSHLogDate(matches[1])
		if dateStr == "" {
			return sshLogHeader{}, false
		}
		return sshLogHeader{DateStr: dateStr, Process: matches[2], PID: matches[3], Message: matches[4]}, true
	}
	return sshLogHeader{}, false
}

func loadSSHLogSessionKey(header sshLogHeader, data dto.SSHHistory, line string) string {
	if header.Process == "sshd-session" && data.Address != "" && data.Port != "" {
		return data.Address + ":" + data.Port
	}
	if header.PID != "" {
		return "pid:" + header.PID
	}
	if data.Address != "" && data.Port != "" {
		return data.Address + ":" + data.Port
	}
	return line
}

func normalizeSSHLogDate(dateStr string) string {
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t.Format("2006 Jan 2 15:04:05")
	}
	if t, err := time.Parse(constant.DateTimeLayout, dateStr); err == nil {
		return t.Format("2006 Jan 2 15:04:05")
	}
	if _, err := time.Parse("Jan 2 15:04:05", dateStr); err == nil {
		return dateStr
	}
	return ""
}

func parseSSHLogMessage(message string) (dto.SSHHistory, bool) {
	if matches := re.GetRegex(re.SSHAcceptedPattern).FindStringSubmatch(message); len(matches) == 5 {
		return dto.SSHHistory{
			AuthMode: matches[1],
			User:     matches[2],
			Address:  matches[3],
			Port:     matches[4],
			Status:   constant.StatusSuccess,
		}, true
	}
	if matches := re.GetRegex(re.SSHFailedPattern).FindStringSubmatch(message); len(matches) == 6 {
		return dto.SSHHistory{
			AuthMode: matches[1],
			User:     matches[3],
			Address:  matches[4],
			Port:     matches[5],
			Status:   constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHInvalidUserPattern).FindStringSubmatch(message); len(matches) == 4 {
		return dto.SSHHistory{
			User:    matches[1],
			Address: matches[2],
			Port:    matches[3],
			Status:  constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHClosedPattern).FindStringSubmatch(message); len(matches) == 4 {
		return dto.SSHHistory{
			User:    matches[1],
			Address: matches[2],
			Port:    matches[3],
			Status:  constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHDisconnectedPattern).FindStringSubmatch(message); len(matches) == 4 {
		return dto.SSHHistory{
			User:    matches[1],
			Address: matches[2],
			Port:    matches[3],
			Status:  constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHDisconnectPattern).FindStringSubmatch(message); len(matches) == 3 {
		return dto.SSHHistory{
			Address: matches[1],
			Port:    matches[2],
			Status:  constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHMaxAuthPattern).FindStringSubmatch(message); len(matches) == 4 {
		return dto.SSHHistory{
			User:    matches[1],
			Address: matches[2],
			Port:    matches[3],
			Status:  constant.StatusFailed,
		}, true
	}
	if matches := re.GetRegex(re.SSHNotAllowedPattern).FindStringSubmatch(message); len(matches) == 3 {
		return dto.SSHHistory{
			User:    matches[1],
			Address: matches[2],
			Status:  constant.StatusFailed,
		}, true
	}
	return dto.SSHHistory{}, false
}

func checkIsStandard(item dto.SSHHistory) bool {
	if len(item.Address) == 0 || net.ParseIP(item.Address) == nil {
		return false
	}
	if item.Port == "" {
		return true
	}
	portItem, _ := strconv.Atoi(item.Port)
	return portItem > 0 && portItem < 65536
}

func handleGunzip(path string) error {
	cmdMgr := cmd.NewCommandMgr()
	if err := cmdMgr.Run("gunzip", path); err != nil {
		return err
	}
	return nil
}

func loadServiceName() (string, error) {
	if exist, _ := controller.CheckExist("sshd"); exist {
		return "sshd", nil
	} else if exist, _ := controller.CheckExist("ssh"); exist {
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
	active, _ := controller.CheckActive("ssh.socket")
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
	_ = controller.Reload()
	_ = controller.HandleRestart("ssh.socket")
	return nil
}

func loadSSHPort() string {
	port := "22"
	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		return port
	}
	lines := strings.Split(string(sshConf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			portStr := strings.ReplaceAll(line, "Port ", "")
			portItem, _ := strconv.Atoi(portStr)
			if portItem > 0 && portItem < 65535 {
				return portStr
			}
		}
	}
	return port
}
