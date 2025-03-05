package service

import (
	"encoding/base64"
	"fmt"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type HostService struct{}

type IHostService interface {
	TestLocalConn(id uint) bool
	TestByInfo(req dto.HostConnTest) bool
	GetHostByID(id uint) (*dto.HostInfo, error)
	SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error)
	SearchWithPage(search dto.SearchPageWithGroup) (int64, interface{}, error)
	Create(hostDto dto.HostOperate) (*dto.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) (*dto.HostInfo, error)
	Delete(id []uint) error

	EncryptHost(itemVal string) (string, error)
}

func NewIHostService() IHostService {
	return &HostService{}
}

func (u *HostService) TestByInfo(req dto.HostConnTest) bool {
	if req.AuthMode == "password" && len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return false
		}
		req.Password = string(password)
	}
	if req.AuthMode == "key" && len(req.PrivateKey) != 0 {
		privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
		if err != nil {
			return false
		}
		req.PrivateKey = string(privateKey)
	}
	if len(req.Password) == 0 && len(req.PrivateKey) == 0 {
		host, err := hostRepo.Get(hostRepo.WithByAddr(req.Addr))
		if err != nil {
			return false
		}
		req.Password = host.Password
		req.AuthMode = host.AuthMode
		req.PrivateKey = host.PrivateKey
		req.PassPhrase = host.PassPhrase
	}

	var connInfo ssh.ConnInfo
	_ = copier.Copy(&connInfo, &req)
	connInfo.PrivateKey = []byte(req.PrivateKey)
	if len(req.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(req.PassPhrase)
	}
	client, err := ssh.NewClient(connInfo)
	if err != nil {
		return false
	}
	defer client.Close()
	return true
}

func (u *HostService) TestLocalConn(id uint) bool {
	var (
		host model.Host
		err  error
	)
	if id == 0 {
		host, err = hostRepo.Get(hostRepo.WithByAddr("127.0.0.1"))
		if err != nil {
			return false
		}
	} else {
		host, err = hostRepo.Get(repo.WithByID(id))
		if err != nil {
			return false
		}
	}
	var connInfo ssh.ConnInfo
	if err := copier.Copy(&connInfo, &host); err != nil {
		return false
	}
	if len(host.Password) != 0 {
		host.Password, err = encrypt.StringDecrypt(host.Password)
		if err != nil {
			return false
		}
		connInfo.Password = host.Password
	}
	if len(host.PrivateKey) != 0 {
		host.PrivateKey, err = encrypt.StringDecrypt(host.PrivateKey)
		if err != nil {
			return false
		}
		connInfo.PrivateKey = []byte(host.PrivateKey)
	}
	if len(host.PassPhrase) != 0 {
		host.PassPhrase, err = encrypt.StringDecrypt(host.PassPhrase)
		if err != nil {
			return false
		}
		connInfo.PassPhrase = []byte(host.PassPhrase)
	}
	client, err := ssh.NewClient(connInfo)
	if err != nil {
		return false
	}
	defer client.Close()

	return true
}

func (u *HostService) SearchWithPage(req dto.SearchPageWithGroup) (int64, interface{}, error) {
	var options []global.DBOption
	if len(req.Info) != 0 {
		options = append(options, hostRepo.WithByInfo(req.Info))
	}
	if req.GroupID != 0 {
		options = append(options, repo.WithByGroupID(req.GroupID))
	}
	total, hosts, err := hostRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	var dtoHosts []dto.HostInfo
	for _, host := range hosts {
		var item dto.HostInfo
		if err := copier.Copy(&item, &host); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		group, _ := groupRepo.Get(repo.WithByID(host.GroupID))
		item.GroupBelong = group.Name
		if !item.RememberPassword {
			item.Password = ""
			item.PrivateKey = ""
			item.PassPhrase = ""
		} else {
			if len(host.Password) != 0 {
				item.Password, err = encrypt.StringDecrypt(host.Password)
				if err != nil {
					return 0, nil, err
				}
			}
			if len(host.PrivateKey) != 0 {
				item.PrivateKey, err = encrypt.StringDecrypt(host.PrivateKey)
				if err != nil {
					return 0, nil, err
				}
			}
			if len(host.PassPhrase) != 0 {
				item.PassPhrase, err = encrypt.StringDecrypt(host.PassPhrase)
				if err != nil {
					return 0, nil, err
				}
			}
		}
		dtoHosts = append(dtoHosts, item)
	}
	return total, dtoHosts, err
}

func (u *HostService) SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error) {
	hosts, err := hostRepo.GetList(hostRepo.WithByInfo(search.Info))
	if err != nil {
		return nil, err
	}
	groups, err := groupRepo.GetList(repo.WithByType("host"))
	if err != nil {
		return nil, err
	}
	var datas []dto.HostTree
	for _, group := range groups {
		var data dto.HostTree
		data.ID = group.ID + 10000
		data.Label = group.Name
		for _, host := range hosts {
			label := fmt.Sprintf("%s@%s:%d", host.User, host.Addr, host.Port)
			if len(host.Name) != 0 {
				label = fmt.Sprintf("%s - %s@%s:%d", host.Name, host.User, host.Addr, host.Port)
			}
			if host.GroupID == group.ID {
				data.Children = append(data.Children, dto.TreeChild{ID: host.ID, Label: label})
			}
		}
		if len(data.Children) != 0 {
			datas = append(datas, data)
		}
	}
	return datas, err
}

func (u *HostService) GetHostByID(id uint) (*dto.HostInfo, error) {
	var item dto.HostInfo
	var host model.Host
	if id == 0 {
		host, _ = hostRepo.Get(repo.WithByName("local"))
	} else {
		host, _ = hostRepo.Get(repo.WithByID(id))
	}
	if host.ID == 0 {
		return nil, buserr.New("ErrRecordNotFound")
	}
	if err := copier.Copy(&item, &host); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if !item.RememberPassword {
		item.Password = ""
		item.PrivateKey = ""
		item.PassPhrase = ""
		return &item, nil
	}
	var err error
	if len(host.Password) != 0 {
		item.Password, err = encrypt.StringDecrypt(host.Password)
		if err != nil {
			return nil, err
		}
	}
	if len(host.PrivateKey) != 0 {
		item.PrivateKey, err = encrypt.StringDecrypt(host.PrivateKey)
		if err != nil {
			return nil, err
		}
	}
	if len(host.PassPhrase) != 0 {
		item.PassPhrase, err = encrypt.StringDecrypt(host.PassPhrase)
		if err != nil {
			return nil, err
		}
	}
	return &item, err
}

func (u *HostService) Create(req dto.HostOperate) (*dto.HostInfo, error) {
	if req.Name == "local" {
		return nil, buserr.New("ErrRecordExist")
	}
	hostItem, _ := hostRepo.Get(hostRepo.WithByAddr(req.Addr), hostRepo.WithByUser(req.User), hostRepo.WithByPort(req.Port))
	if hostItem.ID != 0 {
		return nil, buserr.New("ErrRecordExist")
	}

	var err error
	if len(req.Password) != 0 && req.AuthMode == "password" {
		req.Password, err = u.EncryptHost(req.Password)
		if err != nil {
			return nil, err
		}
		req.PrivateKey = ""
		req.PassPhrase = ""
	}
	if len(req.PrivateKey) != 0 && req.AuthMode == "key" {
		req.PrivateKey, err = u.EncryptHost(req.PrivateKey)
		if err != nil {
			return nil, err
		}
		if len(req.PassPhrase) != 0 {
			req.PassPhrase, err = encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				return nil, err
			}
		}
		req.Password = ""
	}
	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if req.GroupID == 0 {
		group, err := groupRepo.Get(repo.WithByType("host"), groupRepo.WithByDefault(true))
		if err != nil {
			return nil, errors.New("get default group failed")
		}
		host.GroupID = group.ID
		req.GroupID = group.ID
	}

	if err := hostRepo.Create(&host); err != nil {
		return nil, err
	}
	var hostinfo dto.HostInfo
	if err := copier.Copy(&hostinfo, &host); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	return &hostinfo, nil
}

func (u *HostService) Delete(ids []uint) error {
	hosts, _ := hostRepo.GetList(repo.WithByIDs(ids))
	for _, host := range hosts {
		if host.ID == 0 {
			return buserr.New("ErrRecordNotFound")
		}
		if host.Addr == "127.0.0.1" {
			return errors.New("the local connection information cannot be deleted!")
		}
	}
	return hostRepo.Delete(repo.WithByIDs(ids))
}

func (u *HostService) Update(id uint, upMap map[string]interface{}) (*dto.HostInfo, error) {
	if err := hostRepo.Update(id, upMap); err != nil {
		return nil, err
	}
	hostItem, _ := hostRepo.Get(repo.WithByID(id))
	var hostinfo dto.HostInfo
	if err := copier.Copy(&hostinfo, &hostItem); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	return &hostinfo, nil
}

func (u *HostService) EncryptHost(itemVal string) (string, error) {
	privateKey, err := base64.StdEncoding.DecodeString(itemVal)
	if err != nil {
		return "", err
	}
	keyItem, err := encrypt.StringEncrypt(string(privateKey))
	return keyItem, err
}

func GetHostInfo(id uint) (*model.Host, error) {
	host, err := hostRepo.Get(repo.WithByID(id))
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	if len(host.Password) != 0 {
		host.Password, err = encrypt.StringDecrypt(host.Password)
		if err != nil {
			return nil, err
		}
	}
	if len(host.PrivateKey) != 0 {
		host.PrivateKey, err = encrypt.StringDecrypt(host.PrivateKey)
		if err != nil {
			return nil, err
		}
	}

	if len(host.PassPhrase) != 0 {
		host.PassPhrase, err = encrypt.StringDecrypt(host.PassPhrase)
		if err != nil {
			return nil, err
		}
	}
	return &host, err
}
