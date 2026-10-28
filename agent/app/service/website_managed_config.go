package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
)

type websiteManagedConfigPolicy struct {
	protectedDirectories func() []string
	backupIncludes       func(model.Website) []string
}

var websiteManagedConfigs websiteManagedConfigPolicy

func configureWebsiteManagedConfig(policy websiteManagedConfigPolicy) {
	if !policy.configured() || websiteManagedConfigs.configured() {
		panic("invalid website managed configuration policy")
	}
	websiteManagedConfigs = policy
}

func (p websiteManagedConfigPolicy) configured() bool {
	return p.protectedDirectories != nil && p.backupIncludes != nil
}

func (p websiteManagedConfigPolicy) pathAllowed(target string) bool {
	if !p.configured() {
		return false
	}
	resolved, err := websiteProxyResolvePath(target)
	if err != nil {
		return false
	}
	for _, directory := range p.protectedDirectories() {
		if directory == "" {
			return false
		}
		root, err := websiteProxyResolvePath(directory)
		if err != nil {
			return false
		}
		if relative, err := filepath.Rel(root, resolved); err != nil || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return false
		}
	}
	return true
}

func exportWebsiteNginxConfig(site model.Website, source, destination string) error {
	return websiteManagedConfigs.exportNginxConfig(site, source, destination)
}

func (p websiteManagedConfigPolicy) exportNginxConfig(site model.Website, source, destination string) error {
	if !p.configured() {
		return errors.New("website managed configuration policy is unavailable")
	}
	includes := p.backupIncludes(site)
	if len(includes) == 0 {
		return files.NewFileOp().CopyFile(source, destination)
	}
	excluded := make(map[string]struct{}, len(includes))
	for _, include := range includes {
		if include == "" {
			return errors.New("invalid managed website backup include")
		}
		excluded[include] = struct{}{}
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	config, err := parser.NewStringParser(string(content)).Parse()
	if err != nil {
		return errors.New("invalid managed website backup configuration")
	}
	changed := false
	for _, server := range config.FindServers() {
		for _, directive := range server.FindDirectives("include") {
			parameters := directive.GetParameters()
			if len(parameters) != 1 {
				continue
			}
			if _, found := excluded[strings.Trim(parameters[0], "\"'")]; found {
				server.RemoveDirective("include", parameters)
				changed = true
			}
		}
	}
	if changed {
		content = []byte(nginx.DumpConfig(config, nginx.IndentedStyle))
	}
	return os.WriteFile(filepath.Join(destination, filepath.Base(source)), content, 0600)
}
