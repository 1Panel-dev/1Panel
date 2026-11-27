package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// The sidebar is built from the saved menu, so a new page needs an entry there.
var AddDnsRecordMenu = &gormigrate.Migration{
	ID: "20261127-add-dns-record-menu",
	Migrate: func(tx *gorm.DB) error {
		return helper.UpsertChildMenuByLabel(tx, "Website-Menu", dto.ShowMenu{
			ID:       "35",
			Disabled: false,
			Title:    "menu.dnsRecord",
			IsShow:   true,
			Label:    "SSL-DNS-RECORD",
			Path:     "/websites/dns-record",
			Sort:     220,
		}, "SSL")
	},
}
