package cmd

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/server"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {}

var RootCmd = &cobra.Command{
	Use:   "1panel",
	Short: "1Panel ，一款现代化的 Linux 面板",
	RunE: func(cmd *cobra.Command, args []string) error {
		server.Start()
		return nil
	},
}

type setting struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Key       string    `json:"key" gorm:"type:varchar(256);not null;"`
	Value     string    `json:"value" gorm:"type:varchar(256)"`
	About     string    `json:"about" gorm:"type:longText"`
}

func getSettingByKey(db *gorm.DB, key string) string {
	var setting setting
	_ = db.Where("key = ?", key).First(&setting).Error
	return setting.Value
}
