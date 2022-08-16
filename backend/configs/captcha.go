package configs

type Captcha struct {
	Enable     bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
	Source     string `mapstructure:"source" json:"source" yaml:"source"`
	Length     int    `mapstructure:"length" json:"length" yaml:"length"`
	NoiseCount int    `mapstructure:"noise-count" json:"noise-count" yaml:"noise-count"`
	ImgWidth   int    `mapstructure:"img-width" json:"img-width" yaml:"img-width"`
	ImgHeight  int    `mapstructure:"img-height" json:"img-height" yaml:"img-height"`
}
