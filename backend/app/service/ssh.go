package service

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHService struct{}

type ISSHService interface {
	GetSSHInfo() (*dto.SSHInfo, error)
	Update(key, value string) error
	GenerateSSH(req dto.GenerateSSH) error
	LoadSSHSecret(mode string) (string, error)
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

func (u *SSHService) GetSSHInfo() (*dto.SSHInfo, error) {
	data := dto.SSHInfo{
		Port:                   "22",
		ListenAddress:          "0.0.0.0",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		return &data, err
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
	return &data, err
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
