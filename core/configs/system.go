package configs

type System struct {
	Port           string `mapstructure:"port"`
	Ipv6           string `mapstructure:"ipv6"`
	BindAddress    string `mapstructure:"bindAddress"`
	SSL            string `mapstructure:"ssl"`
	DbCoreFile     string `mapstructure:"db_core_file"`
	EncryptKey     string `mapstructure:"encrypt_key"`
	BaseDir        string `mapstructure:"base_dir"`
	BackupDir      string `mapstructure:"backup_dir"`
	Mode           string `mapstructure:"mode"`
	RepoUrl        string `mapstructure:"repo_url"`
	ResourceUrl    string `mapstructure:"resource_url"`
	Version        string `mapstructure:"version"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Entrance       string `mapstructure:"entrance"`
	Language       string `mapstructure:"language"`
	IsDemo         bool   `mapstructure:"is_demo"`
	IsIntl         bool   `mapstructure:"is_intl"`
	ChangeUserInfo string `mapstructure:"change_user_info"`

	ApiInterfaceStatus string `mapstructure:"api_interface_status"`
	ApiKey             string `mapstructure:"api_key"`
	IpWhiteList        string `mapstructure:"ip_white_list"`
	ApiKeyValidityTime string `mapstructure:"api_key_validity_time"`
}
