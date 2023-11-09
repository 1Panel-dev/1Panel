package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/toolbox"
)

const defaultFail2banPath = "/etc/fail2ban/jail.local"

type Fail2banService struct{}

type IFail2banService interface {
	LoadBaseInfo() (dto.Fail2banBaseInfo, error)
	SearchWithPage(search dto.Fail2banSearch) (int64, interface{}, error)
	Operate(operation string) error
	OperateSSHD(req dto.Fail2banSet) error
	UpdateConf(req dto.Fail2banUpdate) error
}

func NewIFail2banService() IFail2banService {
	return &Fail2banService{}
}

func (u *Fail2banService) LoadBaseInfo() (dto.Fail2banBaseInfo, error) {
	var baseInfo dto.Fail2banBaseInfo
	client, err := toolbox.NewFail2ban()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.IsActive, baseInfo.IsEnable = client.Status()
	baseInfo.Version = client.Version()
	conf, err := os.ReadFile(defaultFail2banPath)
	if err != nil {
		return baseInfo, fmt.Errorf("read fail2ban conf of %s failed, err: %v", defaultFail2banPath, err)
	}
	lines := strings.Split(string(conf), "\n")

	isStart := false
	for _, line := range lines {
		if strings.HasPrefix(line, "[sshd]") {
			isStart = true
		}
		if !isStart {
			continue
		}
		loadFailValue(line, &baseInfo)
		if strings.HasPrefix(line, "[") {
			break
		}
	}

	return baseInfo, nil
}

func (u *Fail2banService) SearchWithPage(req dto.Fail2banSearch) (int64, interface{}, error) {
	var (
		list      []string
		backDatas []string
		err       error
	)
	client, err := toolbox.NewFail2ban()
	if err != nil {
		return 0, nil, err
	}
	if req.Status == "banned" {
		list, err = client.ListBanned()

	} else {
		list, err = client.ListIgnore()
	}
	if err != nil {
		return 0, nil, err
	}

	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]string, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = list[start:end]
	}
	return int64(total), backDatas, nil
}

func (u *Fail2banService) Operate(operation string) error {
	client, err := toolbox.NewFail2ban()
	if err != nil {
		return err
	}
	return client.Operate(operation)
}

func (u *Fail2banService) UpdateConf(req dto.Fail2banUpdate) error {
	conf, err := os.ReadFile(defaultFail2banPath)
	if err != nil {
		return fmt.Errorf("read fail2ban conf of %s failed, err: %v", defaultFail2banPath, err)
	}
	lines := strings.Split(string(conf), "\n")

	isStart, isEnd, hasKey := false, false, false
	newFile := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "[sshd]") {
			isStart = true
		}
		if !isStart || isEnd {
			newFile += fmt.Sprintf("%s\n", line)
			continue
		}

		if strings.HasPrefix(line, strings.ToLower(req.Key)) {
			hasKey = true
			newFile += fmt.Sprintf("%s = %s\n", req.Key, req.Value)
			continue
		}
		if strings.HasPrefix(line, "[") {
			isEnd = true
			if !hasKey {
				newFile += fmt.Sprintf("%s = %s\n\n", req.Key, req.Value)
			}
			newFile += fmt.Sprintf("%s\n", line)
		}
	}
	return nil
}

func (u *Fail2banService) OperateSSHD(req dto.Fail2banSet) error {
	if strings.HasSuffix(req.Operate, "ignore") {
		if err := u.UpdateConf(dto.Fail2banUpdate{Key: "ignoreip", Value: strings.Join(req.IPs, ",")}); err != nil {
			return err
		}
	}
	client, err := toolbox.NewFail2ban()
	if err != nil {
		return err
	}
	for _, item := range req.IPs {
		if err := client.OperateSSHD(req.Operate, item); err != nil {
			return err
		}
	}
	return nil
}

func loadFailValue(line string, baseInfo *dto.Fail2banBaseInfo) {
	if strings.HasPrefix(line, "port") {
		itemValue := strings.ReplaceAll(line, "port", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.Port, _ = strconv.Atoi(strings.TrimSpace(itemValue))
	}
	if strings.HasPrefix(line, "maxretry") {
		itemValue := strings.ReplaceAll(line, "maxretry", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.MaxRetry, _ = strconv.Atoi(strings.TrimSpace(itemValue))
	}
	if strings.HasPrefix(line, "findtime") {
		itemValue := strings.ReplaceAll(line, "findtime", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.FindTime = strings.TrimSpace(itemValue)
	}
	if strings.HasPrefix(line, "bantime") {
		itemValue := strings.ReplaceAll(line, "bantime", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.BanTime = strings.TrimSpace(itemValue)
	}
	if strings.HasPrefix(line, "banaction") {
		itemValue := strings.ReplaceAll(line, "banaction", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.BanAction = strings.TrimSpace(itemValue)
	}
	if strings.HasPrefix(line, "logpath") {
		itemValue := strings.ReplaceAll(line, "logpath", "")
		itemValue = strings.ReplaceAll(itemValue, "=", "")
		baseInfo.BanTime = strings.TrimSpace(itemValue)
	}
}
