package common

import (
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/utils/cmd"
)

func LoadTimeZoneByCmd() string {
	loc := time.Now().Location().String()
	if _, err := time.LoadLocation(loc); err != nil {
		loc = "Asia/Shanghai"
	}
	std, err := cmd.RunDefaultWithStdoutBashC("timedatectl | grep 'Time zone'")
	if err != nil {
		return loc
	}
	fields := strings.Fields(string(std))
	if len(fields) != 5 {
		return loc
	}
	if _, err := time.LoadLocation(fields[2]); err != nil {
		return loc
	}
	return fields[2]
}

func LoadExpiredLocation() *time.Location {
	var (
		expiredLoc     *time.Location
		expiredLocOnce sync.Once
	)
	expiredLocOnce.Do(func() {
		loc, err := time.LoadLocation(LoadTimeZoneByCmd())
		if err != nil {
			expiredLoc = time.Local
			return
		}
		expiredLoc = loc
	})
	if expiredLoc == nil {
		return time.Local
	}
	return expiredLoc
}
