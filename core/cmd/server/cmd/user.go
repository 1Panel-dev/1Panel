package cmd

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	RootCmd.AddCommand(userListCmd)
}

var userListCmd = &cobra.Command{
	Use:   "user-list",
	Short: i18n.GetMsgByKeyForCmd("UserList"),
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl user-list"}))
			return nil
		}
		if isEnterprise() {
			db, err := loadDBConn("enterprise.db")
			if err != nil {
				return err
			}
			return listEnterpriseUsers(db)
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			return err
		}
		return listCommunityUsers(db)
	},
}

func listEnterpriseUsers(db *gorm.DB) error {
	var userModels []User
	if err := db.Order("created_at desc").Find(&userModels).Error; err != nil {
		return err
	}

	rows := make([]enterpriseUserRow, 0, len(userModels))
	for _, user := range userModels {
		superAdmin := "no"
		if user.IsSuperAdmin {
			superAdmin = "yes"
		}
		rows = append(rows, enterpriseUserRow{
			Name:       user.Name,
			MFAStatus:  user.MFAStatus,
			SuperAdmin: superAdmin,
			CreatedAt:  user.CreatedAt.Format(time.RFC3339),
		})
	}

	if len(rows) == 0 {
		fmt.Println(i18n.GetMsgByKeyForCmd("UserEmptyList"))
		return nil
	}

	return printUserRows(rows)
}

func listCommunityUsers(db *gorm.DB) error {
	nameSetting, err := loadSettingRecord(db, "UserName")
	if err != nil {
		return err
	}
	mfaSetting, err := loadSettingRecord(db, "MFAStatus")
	if err != nil {
		return err
	}

	rows := []enterpriseUserRow{
		{
			Name:       nameSetting.Value,
			MFAStatus:  mfaSetting.Value,
			SuperAdmin: "yes",
			CreatedAt:  createdAtString(nameSetting.CreatedAt),
		},
	}

	return printUserRows(rows)
}

func printUserRows(rows []enterpriseUserRow) error {
	headers := []string{
		i18n.GetMsgByKeyForCmd("UserTableName"),
		i18n.GetMsgByKeyForCmd("UserTableMFAStatus"),
		i18n.GetMsgByKeyForCmd("UserTableSuperAdmin"),
		i18n.GetMsgByKeyForCmd("UserTableCreatedAt"),
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = displayWidth(header)
	}
	for _, row := range rows {
		values := []string{row.Name, row.MFAStatus, row.SuperAdmin, row.CreatedAt}
		for i, value := range values {
			if w := displayWidth(value); w > widths[i] {
				widths[i] = w
			}
		}
	}
	fmt.Println(joinColumns(headers, widths))
	for _, row := range rows {
		fmt.Println(joinColumns([]string{row.Name, row.MFAStatus, row.SuperAdmin, row.CreatedAt}, widths))
	}
	return nil
}

func loadSettingRecord(db *gorm.DB, key string) (setting, error) {
	var record setting
	if err := db.Where("key = ?", key).First(&record).Error; err != nil {
		return setting{}, err
	}
	return record, nil
}

func createdAtString(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC3339)
}

type enterpriseUserRow struct {
	Name       string
	MFAStatus  string
	SuperAdmin string
	CreatedAt  string
}

type User struct {
	Name         string
	MFAStatus    string
	IsSuperAdmin bool
	CreatedAt    time.Time
}

func joinColumns(values []string, widths []int) string {
	out := ""
	for i, value := range values {
		out += padDisplayWidth(value, widths[i])
		if i != len(values)-1 {
			out += "  "
		}
	}
	return out
}

func padDisplayWidth(value string, width int) string {
	pad := width - displayWidth(value)
	if pad <= 0 {
		return value
	}
	return value + spaces(pad)
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeDisplayWidth(r)
	}
	return width
}

func runeDisplayWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if r < 32 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if isWideRune(r) {
		return 2
	}
	return 1
}

func isWideRune(r rune) bool {
	if r < 0x1100 {
		return false
	}
	if r <= 0x115f || r == 0x2329 || r == 0x232a {
		return true
	}
	if r >= 0x2e80 && r <= 0xa4cf && r != 0x303f {
		return true
	}
	if r >= 0xac00 && r <= 0xd7a3 {
		return true
	}
	if r >= 0xf900 && r <= 0xfaff {
		return true
	}
	if r >= 0xfe10 && r <= 0xfe19 {
		return true
	}
	if r >= 0xfe30 && r <= 0xfe6f {
		return true
	}
	if r >= 0xff00 && r <= 0xff60 {
		return true
	}
	if r >= 0xffe0 && r <= 0xffe6 {
		return true
	}
	return r >= 0x20000 && r <= 0x3fffd
}
