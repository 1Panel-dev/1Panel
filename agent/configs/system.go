package configs

type System struct {
	MasterRequestAddr  string `mapstructure:"master_request_addr"`
	MasterRequestToken string `mapstructure:"master_request_token"`
	DbFile             string `mapstructure:"db_agent_file"`
	DbPath             string `mapstructure:"db_path"`
	LogPath            string `mapstructure:"log_path"`
	DataDir            string `mapstructure:"data_dir"`
	TmpDir             string `mapstructure:"tmp_dir"`
	Cache              string `mapstructure:"cache"`
	Backup             string `mapstructure:"backup"`
	EncryptKey         string `mapstructure:"encrypt_key"`
	BaseDir            string `mapstructure:"base_dir"`
	Mode               string `mapstructure:"mode"`
	RepoUrl            string `mapstructure:"repo_url"`
	Version            string `mapstructure:"version"`
	IsDemo             bool   `mapstructure:"is_demo"`
	AppRepo            string `mapstructure:"app_repo"`
}
