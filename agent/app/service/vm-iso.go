package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/jinzhu/copier"
)

func (v *VMService) SearchIso(req dto.SearchPageWithType) (int64, interface{}, error) {
	options := []repo.DBOption{repo.WithOrderDesc("created_at"), repo.WithOrderDesc("id")}
	if req.Type != "" {
		options = append(options, repo.WithByType(req.Type))
	}
	if req.Info != "" {
		options = append(options, repo.WithByLikeName(req.Info))
	}
	total, isos, err := vmRepo.PageIso(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	dtoIsos := make([]dto.VMIsoInfo, 0, len(isos))
	for _, iso := range isos {
		var item dto.VMIsoInfo
		if err := copier.Copy(&item, &iso); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		item.IsExist, item.Size = loadVMResourceFileInfo(item.Path)
		dtoIsos = append(dtoIsos, item)
	}
	return total, dtoIsos, err
}

func (v *VMService) ListIso() ([]dto.VMIsoInfo, error) {
	isos, err := vmRepo.ListIsos()
	if err != nil {
		return nil, err
	}
	dtoIsos := make([]dto.VMIsoInfo, 0, len(isos))
	for _, iso := range isos {
		if iso.Status == "uploading" || iso.Status == "downloading" || iso.Status == "failed" {
			continue
		}
		var item dto.VMIsoInfo
		if err := copier.Copy(&item, &iso); err != nil {
			return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		item.IsExist, item.Size = loadVMResourceFileInfo(item.Path)
		dtoIsos = append(dtoIsos, item)
	}
	return dtoIsos, nil
}

func (v *VMService) CreateIso(req dto.VMIsoCreate) error {
	isoItem, _ := vmRepo.GetIso(repo.WithByName(req.Name))
	if isoItem.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	exists, _ := loadVMResourceFileInfo(req.Path)
	if !exists {
		return buserr.WithDetail("ErrResourceNotExists", "Resource file does not exist", nil)
	}
	return vmRepo.CreateIso(&model.VMIso{
		Type: req.Type, Name: req.Name, Path: req.Path, Status: "ready", Description: req.Description,
	})
}

func (v *VMService) StartIsoImport(req dto.VMIsoImportStart) (uint, error) {
	isoItem, _ := vmRepo.GetIso(repo.WithByName(req.Name))
	if isoItem.ID != 0 {
		return 0, buserr.New("ErrRecordExist")
	}
	item := model.VMIso{Type: req.Type, Name: req.Name, Path: req.Path, Status: req.Status, TempFileID: req.TempFileID, TempFileType: req.TempFileType, Description: req.Description}
	if err := vmRepo.CreateIso(&item); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (v *VMService) UpdateIsoImport(req dto.VMIsoImportUpdate) error {
	isoItem, err := vmRepo.GetIso(repo.WithByID(req.ID))
	if err != nil || isoItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if isoItem.Status != "uploading" && isoItem.Status != "downloading" {
		return buserr.WithDetail("ErrVMInvalidInput", "ISO import is not active", nil)
	}
	if (isoItem.Status == "uploading" && req.Status == "downloading") ||
		(isoItem.Status == "downloading" && req.Status == "uploading") {
		return buserr.WithDetail("ErrVMInvalidInput", "ISO import type cannot be changed", nil)
	}
	vars := map[string]interface{}{"status": req.Status, "error": req.Error}
	if req.TempFileID != "" {
		vars["temp_file_id"] = req.TempFileID
	}
	if req.TempFileType != "" {
		vars["temp_file_type"] = req.TempFileType
	}
	if req.Status == "ready" {
		exists, _ := loadVMResourceFileInfo(isoItem.Path)
		if !exists {
			return buserr.WithDetail("ErrResourceNotExists", "Resource file does not exist", nil)
		}
		vars["error"] = ""
	}
	if req.Status == "ready" {
		vars["temp_file_id"] = ""
		vars["temp_file_type"] = ""
	}
	return vmRepo.UpdateIso(req.ID, vars)
}

func (v *VMService) UpdateIso(req dto.VMIsoUpdate) error {
	isoItem, err := vmRepo.GetIso(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if isoItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	exists, _ := loadVMResourceFileInfo(req.Path)
	if !exists {
		return buserr.WithDetail("ErrResourceNotExists", "Resource file does not exist", nil)
	}
	return vmRepo.UpdateIso(req.ID, map[string]interface{}{
		"name": req.Name, "type": req.Type, "path": req.Path, "description": req.Description,
	})
}

func (v *VMService) DeleteIso(req dto.VMIsoDelete) error {
	isoItem, err := vmRepo.GetIso(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if isoItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if req.DeleteFile && isoItem.Path != "" {
		exists, _ := loadVMResourceFileInfo(isoItem.Path)
		if exists {
			if err := os.Remove(isoItem.Path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	if isoItem.TempFileType == "upload" && validTempFileID(isoItem.TempFileID) {
		tmpPath := filepath.Join(global.CONF.Base.InstallDir, "1panel", "tmp", "upload", isoItem.TempFileID)
		if err := os.RemoveAll(tmpPath); err != nil {
			return err
		}
	}
	return vmRepo.DeleteIso(repo.WithByID(req.ID))
}

func loadVMResourceFileInfo(filePath string) (bool, int64) {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false, 0
	}
	return true, info.Size()
}

func validTempFileID(value string) bool {
	return value != "" && filepath.Base(value) == value && !strings.ContainsAny(value, `/\\`)
}
