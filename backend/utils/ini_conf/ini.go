package ini_conf

import "gopkg.in/ini.v1"

func GetIniValue(filePath, Group, Key string) (string, error) {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return "", err
	}
	service, err := cfg.GetSection(Group)
	if err != nil {
		return "", err
	}
	startKey, err := service.GetKey(Key)
	if err != nil {
		return "", err
	}
	return startKey.Value(), nil
}

func SetIniValue(filePath, Group, Key, value string) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return err
	}
	service, err := cfg.GetSection(Group)
	if err != nil {
		return err
	}
	targetKey := service.Key(Key)
	if err != nil {
		return err
	}
	targetKey.SetValue(value)
	return cfg.SaveTo(filePath)
}
