package configs

import (
	"fmt"
	"os"
)

type Sqlite struct {
	Path   string `mapstructure:"path"`
	DbFile string `mapstructure:"db_file"`
}

func (s *Sqlite) Dsn() string {
	if _, err := os.Stat(s.Path); err != nil {
		if err := os.MkdirAll(s.Path, os.ModePerm); err != nil {
			panic(fmt.Errorf("init db dir falied, err: %v", err))
		}
	}
	if _, err := os.Stat(s.Path + "/" + s.DbFile); err != nil {
		if _, err := os.Create(s.Path + "/" + s.DbFile); err != nil {
			panic(fmt.Errorf("init db file falied, err: %v", err))
		}
	}
	return s.Path + "/" + s.DbFile
}
