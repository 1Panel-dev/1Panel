package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
)

func parseWebsiteSelection(value string) (bool, []uint, []uint, error) {
	if value == "all" {
		return true, nil, nil, nil
	}
	var ids, groups []uint
	seen := map[string]bool{}
	for _, token := range strings.Split(value, ",") {
		group := strings.HasPrefix(token, "group:")
		raw := token
		if group {
			raw = strings.TrimPrefix(token, "group:")
		}
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 || strconv.FormatUint(id, 10) != raw {
			return false, nil, nil, errors.New(i18n.GetMsgWithMap("ErrCronjobWebsiteSelection", map[string]interface{}{"value": token}))
		}
		if seen[token] {
			continue
		}
		seen[token] = true
		if group {
			groups = append(groups, uint(id))
		} else {
			ids = append(ids, uint(id))
		}
	}
	return false, ids, groups, nil
}

func validateWebsiteSelection(value string) (string, error) {
	all, ids, groups, err := parseWebsiteSelection(value)
	if err != nil {
		return "", err
	}
	if all {
		return "all", nil
	}
	var tokens []string
	for _, id := range groups {
		if _, err := groupRepo.Get(repo.WithByID(id), repo.WithByType("website")); err != nil {
			return "", fmt.Errorf("%s: %w", i18n.GetMsgWithMap("ErrCronjobWebsiteGroup", map[string]interface{}{"id": id}), err)
		}
		tokens = append(tokens, fmt.Sprintf("group:%d", id))
	}
	for _, id := range ids {
		tokens = append(tokens, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(tokens, ","), nil
}

func loadWebsForJob(job model.Cronjob) ([]model.Website, error) {
	all, ids, groups, err := parseWebsiteSelection(job.Website)
	if err != nil {
		return nil, err
	}
	if all {
		return websiteRepo.List()
	}
	// Verify group references so deleted groups are reported in the task log.
	for _, id := range groups {
		if _, err := groupRepo.Get(repo.WithByID(id), repo.WithByType("website")); err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.GetMsgWithMap("ErrCronjobWebsiteGroup", map[string]interface{}{"id": id}), err)
		}
	}
	var websites []model.Website
	err = global.DB.Where("id IN ? OR website_group_id IN ?", ids, groups).Order("id ASC").Find(&websites).Error
	return websites, err
}

func exportWebsiteSelection(job model.Cronjob, item *dto.CronjobTrans) error {
	all, ids, groups, err := parseWebsiteSelection(job.Website)
	if err != nil {
		return err
	}
	if all {
		return nil
	}
	if len(groups) > 0 {
		item.WebsiteSelectionVersion = 2
	}
	for _, id := range groups {
		group, err := groupRepo.Get(repo.WithByID(id), repo.WithByType("website"))
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMsgWithMap("ErrCronjobWebsiteGroup", map[string]interface{}{"id": id}), err)
		}
		websites, err := websiteRepo.GetBy(websiteRepo.WithGroupID(id))
		if err != nil {
			return err
		}
		snapshot := dto.CronjobWebsiteGroup{Name: group.Name, Websites: []string{}}
		for _, web := range websites {
			snapshot.Websites = append(snapshot.Websites, web.Alias)
		}
		item.WebsiteGroups = append(item.WebsiteGroups, snapshot)
	}
	if len(ids) > 0 {
		websites, err := websiteRepo.GetBy(repo.WithByIDs(ids))
		if err != nil {
			return err
		}
		for _, web := range websites {
			item.Websites = append(item.Websites, web.Alias)
		}
	}
	// An empty legacy list means all websites. Never export a stale fixed selection as all.
	if len(groups) == 0 && len(item.Websites) == 0 {
		return errors.New(i18n.GetMsgWithMap("ErrCronjobWebsiteExportEmpty", map[string]interface{}{"name": job.Name}))
	}
	return nil
}

func importWebsiteSelection(item dto.CronjobTrans, groups []model.Group, websites []model.Website) (string, []string, error) {
	if item.WebsiteSelectionVersion != 2 {
		return "", nil, errors.New(i18n.GetMsgWithMap("ErrCronjobWebsiteVersion", map[string]interface{}{"version": item.WebsiteSelectionVersion}))
	}
	if item.AllWebsites {
		if len(item.WebsiteGroups) > 0 || len(item.Websites) > 0 {
			return "", nil, errors.New(i18n.GetMsgByKey("ErrCronjobWebsiteAllConflict"))
		}
		return "all", nil, nil
	}
	groupIDs := map[string]uint{}
	ambiguous := map[string]bool{}
	for _, group := range groups {
		if _, ok := groupIDs[group.Name]; ok {
			ambiguous[group.Name] = true
		}
		groupIDs[group.Name] = group.ID
	}
	aliases := map[string]uint{}
	for _, web := range websites {
		aliases[web.Alias] = web.ID
	}
	var tokens, warnings []string
	seen := map[string]bool{}
	add := func(token string) {
		if !seen[token] {
			tokens = append(tokens, token)
			seen[token] = true
		}
	}
	for _, group := range item.WebsiteGroups {
		if group.Name == "" {
			return "", nil, errors.New(i18n.GetMsgByKey("ErrCronjobWebsiteGroupName"))
		}
		if ambiguous[group.Name] {
			return "", nil, fmt.Errorf("ambiguous website group name: %s", group.Name)
		}
		if id, ok := groupIDs[group.Name]; ok {
			add(fmt.Sprintf("group:%d", id))
			continue
		}
		matched := 0
		var missing []string
		for _, alias := range group.Websites {
			if id, ok := aliases[alias]; ok {
				add(strconv.FormatUint(uint64(id), 10))
				matched++
			} else {
				missing = append(missing, alias)
			}
		}
		if matched == 0 {
			return "", nil, errors.New(i18n.GetMsgWithMap("ErrCronjobWebsiteGroupMissing", map[string]interface{}{"group": group.Name}))
		}
		warnings = append(warnings, i18n.GetMsgWithMap("CronjobWebsiteGroupFallback", map[string]interface{}{"name": item.Name, "group": group.Name}))
		if len(missing) > 0 {
			warnings = append(warnings, i18n.GetMsgWithMap("CronjobWebsiteMissing", map[string]interface{}{"name": item.Name, "website": strings.Join(missing, ", ")}))
		}
	}
	for _, alias := range item.Websites {
		if id, ok := aliases[alias]; ok {
			add(strconv.FormatUint(uint64(id), 10))
		} else {
			warnings = append(warnings, i18n.GetMsgWithMap("CronjobWebsiteMissing", map[string]interface{}{"name": item.Name, "website": alias}))
		}
	}
	if len(tokens) == 0 {
		return "", nil, errors.New(i18n.GetMsgByKey("ErrCronjobWebsiteImportEmpty"))
	}
	return strings.Join(tokens, ","), warnings, nil
}
