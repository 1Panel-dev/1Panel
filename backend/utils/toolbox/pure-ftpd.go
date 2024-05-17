package toolbox

import (
	"errors"
	"os/user"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
)

type Ftp struct {
	DefaultUser string
}

type FtpClient interface {
	Status() (bool, error)
	LoadList() ([]FtpList, error)
	UserAdd(username, path, passwd string) error
	UserDel(username string) error
	SetPasswd(username, passwd string) error
	Reload() error
}

func NewFtpClient() (*Ftp, error) {
	userItem, err := user.LookupId("1000")
	if err == nil {
		return &Ftp{DefaultUser: userItem.Username}, err
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
		return &Ftp{DefaultUser: "1panel"}, nil
	}
	if err.Error() != user.UnknownGroupIdError("1000").Error() {
		return nil, err
	}
	stdout, err := cmd.Exec("groupadd -g 1000 1panel")
	if err != nil {
		return nil, errors.New(string(stdout))
	}
	stdout2, err := cmd.Execf("useradd -u 1000 -g %s %s", groupItem.Name, userItem.Username)
	if err != nil {
		return nil, errors.New(stdout2)
	}
	return &Ftp{DefaultUser: "1panel"}, nil
}

func (f *Ftp) Status() (bool, error) {
	return systemctl.IsActive("pure-ftpd.service")
}

func (f *Ftp) UserAdd(username, passwd, path string) error {
	std, err := cmd.Execf("pure-pw useradd %s -u %s -d %s <<EOF \n%s\n%s\nEOF", username, f.DefaultUser, path, passwd, passwd)
	if err != nil {
		return errors.New(std)
	}
	_ = f.Reload()
	std2, err := cmd.Execf("chown %s %s", f.DefaultUser, path)
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
	std, err := cmd.Execf("pure-pw passwd %s <<EOF \n%s\n%s\nEOF", username, passwd, passwd)
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (f *Ftp) SetPath(username, path string) error {
	std, err := cmd.Execf("pure-pw usermod %s -d %s", username, path)
	if err != nil {
		return errors.New(std)
	}
	std2, err := cmd.Execf("chown %s %s", f.DefaultUser, path)
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
		lists = append(lists, FtpList{User: parts[0], Path: strings.ReplaceAll(parts[1], "/./", "")})
	}
	return lists, nil
}

type FtpList struct {
	User string
	Path string
}

func (f *Ftp) Reload() error {
	std, err := cmd.Exec("pure-pw mkdb")
	if err != nil {
		return errors.New(std)
	}
	return nil
}
