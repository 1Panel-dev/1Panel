package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type WebsiteTemplateService struct {
}

type IWebsiteTemplateService interface {
	PageTemplate(req request.WebsiteTemplateSearch) (int64, []response.WebsiteTemplateDTO, error)
	CreateTemplate(req request.WebsiteTemplateCreate) error
	UpdateTemplate(req request.WebsiteTemplateUpdate) error
	DeleteTemplate(id uint) error
	GetTemplate(id uint) (*response.WebsiteTemplateDTO, error)
	SaveUploadZip(fileName string, content []byte) (string, []string, error)
	PageOutput(req request.WebsiteTemplateOutputSearch) (int64, []response.WebsiteTemplateOutputDTO, error)
	CreateOutput(req request.WebsiteTemplateOutputCreate) error
	DeleteOutput(id uint) error
	GetOutput(id uint) (*response.WebsiteTemplateOutputDTO, error)
	Preview(req request.WebsitePreviewReq) (*response.WebsitePreviewDTO, error)
}

func NewIWebsiteTemplateService() IWebsiteTemplateService {
	return &WebsiteTemplateService{}
}

func templateBaseDir() string {
	return path.Join(global.Dir.DataDir, "templates")
}

func templateFileDir() string {
	return path.Join(templateBaseDir(), "files")
}

func templateOutputDir(id uint) string {
	return path.Join(templateBaseDir(), "outputs", fmt.Sprintf("%d", id))
}

func (w WebsiteTemplateService) PageTemplate(req request.WebsiteTemplateSearch) (int64, []response.WebsiteTemplateDTO, error) {
	var opts []repo.DBOption
	if req.Name != "" {
		opts = append(opts, websiteTemplateRepo.WithName(req.Name))
	}
	if req.Type != "" {
		opts = append(opts, websiteTemplateRepo.WithType(req.Type))
	}
	opts = append(opts, repo.WithOrderDesc("created_at"))
	total, templates, err := websiteTemplateRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	var dtos []response.WebsiteTemplateDTO
	for _, t := range templates {
		dtos = append(dtos, response.WebsiteTemplateDTO{WebsiteTemplate: t})
	}
	return total, dtos, nil
}

func (w WebsiteTemplateService) CreateTemplate(req request.WebsiteTemplateCreate) error {
	if exist, _ := websiteTemplateRepo.GetFirst(repo.WithByName(req.Name)); exist != nil {
		return buserr.New("ErrNameIsExist")
	}
	template := &model.WebsiteTemplate{
		Name:      req.Name,
		Type:      req.Type,
		Content:   req.Content,
		FilePath:  req.FilePath,
		Variables: req.Variables,
		Remark:    req.Remark,
	}
	return websiteTemplateRepo.Create(template)
}

func (w WebsiteTemplateService) UpdateTemplate(req request.WebsiteTemplateUpdate) error {
	template, err := websiteTemplateRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if exist, _ := websiteTemplateRepo.GetFirst(repo.WithByName(req.Name), repo.WithByNOTID(req.ID)); exist != nil {
		return buserr.New("ErrNameIsExist")
	}
	template.Name = req.Name
	template.Type = req.Type
	template.Content = req.Content
	if req.FilePath != template.FilePath && template.FilePath != "" && strings.HasPrefix(template.FilePath, templateBaseDir()) {
		_ = os.Remove(template.FilePath)
	}
	template.FilePath = req.FilePath
	template.Variables = req.Variables
	template.Remark = req.Remark
	return websiteTemplateRepo.Save(template)
}

func (w WebsiteTemplateService) DeleteTemplate(id uint) error {
	template, err := websiteTemplateRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	outputs, _ := websiteTemplateOutputRepo.List(websiteTemplateOutputRepo.WithByTemplateID(id))
	for _, output := range outputs {
		if output.OutputPath != "" && strings.HasPrefix(output.OutputPath, templateBaseDir()) {
			_ = os.RemoveAll(output.OutputPath)
		}
	}
	if template.FilePath != "" && strings.HasPrefix(template.FilePath, templateBaseDir()) {
		_ = os.Remove(template.FilePath)
	}
	if err := websiteTemplateOutputRepo.DeleteBy(websiteTemplateOutputRepo.WithByTemplateID(id)); err != nil {
		return err
	}
	return websiteTemplateRepo.DeleteBy(repo.WithByID(id))
}

func (w WebsiteTemplateService) GetTemplate(id uint) (*response.WebsiteTemplateDTO, error) {
	template, err := websiteTemplateRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	return &response.WebsiteTemplateDTO{WebsiteTemplate: *template}, nil
}

func (w WebsiteTemplateService) SaveUploadZip(fileName string, content []byte) (string, []string, error) {
	if !strings.HasSuffix(strings.ToLower(fileName), ".zip") {
		return "", nil, buserr.WithName("ErrNotSupportType", fileName)
	}
	dir := templateFileDir()
	if err := os.MkdirAll(dir, constant.DirPerm); err != nil {
		return "", nil, err
	}
	filePath := path.Join(dir, fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(fileName)))
	if err := os.WriteFile(filePath, content, constant.FilePerm); err != nil {
		return "", nil, err
	}
	return filePath, scanZipVariables(filePath), nil
}

func scanZipVariables(zipPath string) []string {
	variables := []string{}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return variables
	}
	defer reader.Close()
	seen := make(map[string]struct{})
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !isTextFile(file.Name) {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		for _, match := range templateVarRegex.FindAllStringSubmatch(string(content), -1) {
			if _, ok := seen[match[1]]; !ok {
				seen[match[1]] = struct{}{}
				variables = append(variables, match[1])
			}
		}
	}
	return variables
}

func (w WebsiteTemplateService) PageOutput(req request.WebsiteTemplateOutputSearch) (int64, []response.WebsiteTemplateOutputDTO, error) {
	var opts []repo.DBOption
	if req.TemplateID > 0 {
		opts = append(opts, websiteTemplateOutputRepo.WithByTemplateID(req.TemplateID))
	}
	opts = append(opts, repo.WithOrderDesc("created_at"))
	total, outputs, err := websiteTemplateOutputRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	var dtos []response.WebsiteTemplateOutputDTO
	for _, output := range outputs {
		item := response.WebsiteTemplateOutputDTO{WebsiteTemplateOutput: output}
		if template, err := websiteTemplateRepo.GetFirst(repo.WithByID(output.TemplateID)); err == nil {
			item.TemplateName = template.Name
		}
		dtos = append(dtos, item)
	}
	return total, dtos, nil
}

func (w WebsiteTemplateService) CreateOutput(req request.WebsiteTemplateOutputCreate) error {
	template, err := websiteTemplateRepo.GetFirst(repo.WithByID(req.TemplateID))
	if err != nil {
		return err
	}
	valuesJSON, err := json.Marshal(req.VariableValues)
	if err != nil {
		return err
	}
	output := &model.WebsiteTemplateOutput{
		Name:           req.Name,
		TemplateID:     template.ID,
		TemplateType:   template.Type,
		VariableValues: string(valuesJSON),
	}
	if err := websiteTemplateOutputRepo.Create(output); err != nil {
		return err
	}
	outputDir := templateOutputDir(output.ID)
	if err := renderToDir(template, req.VariableValues, outputDir); err != nil {
		_ = websiteTemplateOutputRepo.DeleteBy(repo.WithByID(output.ID))
		return err
	}
	output.OutputPath = outputDir
	return websiteTemplateOutputRepo.Save(output)
}

func (w WebsiteTemplateService) DeleteOutput(id uint) error {
	output, err := websiteTemplateOutputRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	if output.OutputPath != "" && strings.HasPrefix(output.OutputPath, templateBaseDir()) {
		_ = os.RemoveAll(output.OutputPath)
	}
	return websiteTemplateOutputRepo.DeleteBy(repo.WithByID(id))
}

func (w WebsiteTemplateService) GetOutput(id uint) (*response.WebsiteTemplateOutputDTO, error) {
	output, err := websiteTemplateOutputRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	item := &response.WebsiteTemplateOutputDTO{WebsiteTemplateOutput: *output}
	if template, err := websiteTemplateRepo.GetFirst(repo.WithByID(output.TemplateID)); err == nil {
		item.TemplateName = template.Name
	}
	return item, nil
}

func (w WebsiteTemplateService) Preview(req request.WebsitePreviewReq) (*response.WebsitePreviewDTO, error) {
	template, err := websiteTemplateRepo.GetFirst(repo.WithByID(req.TemplateID))
	if err != nil {
		return nil, err
	}
	var html string
	if template.Type == "single" {
		html = renderContent(template.Content, req.VariableValues)
	} else {
		html, err = renderMainHTMLFromZip(template.FilePath, req.VariableValues)
		if err != nil {
			return nil, err
		}
	}
	return &response.WebsitePreviewDTO{HTML: html}, nil
}

var templateVarRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

func renderContent(content string, values map[string]string) string {
	rendered := templateVarRegex.ReplaceAllStringFunc(content, func(match string) string {
		key := templateVarRegex.FindStringSubmatch(match)[1]
		if val, ok := values[key]; ok {
			return val
		}
		return ""
	})
	return rendered
}

func renderToDir(template *model.WebsiteTemplate, values map[string]string, outputDir string) error {
	if err := os.MkdirAll(outputDir, constant.DirPerm); err != nil {
		return err
	}
	if template.Type == "single" {
		html := renderContent(template.Content, values)
		return os.WriteFile(path.Join(outputDir, "index.html"), []byte(html), constant.FilePerm)
	}
	return unzipAndRender(template.FilePath, outputDir, values)
}

func unzipAndRender(zipPath, outputDir string, values map[string]string) error {
	if zipPath == "" {
		return buserr.New("ErrFileNotFound")
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if err := extractAndRenderFile(file, outputDir, values); err != nil {
			return err
		}
	}
	return nil
}

func extractAndRenderFile(file *zip.File, outputDir string, values map[string]string) error {
	targetPath := filepath.Join(outputDir, file.Name)
	if !strings.HasPrefix(targetPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path in zip: %s", file.Name)
	}
	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, constant.DirPerm)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), constant.DirPerm); err != nil {
		return err
	}
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	content, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	if isTextFile(file.Name) {
		content = []byte(renderContent(string(content), values))
	}
	return os.WriteFile(targetPath, content, constant.FilePerm)
}

func isTextFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".html", ".htm", ".css", ".js", ".json", ".xml", ".txt", ".svg", ".md":
		return true
	}
	return false
}

func renderMainHTMLFromZip(zipPath string, values map[string]string) (string, error) {
	if zipPath == "" {
		return "", buserr.New("ErrFileNotFound")
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	var mainFile *zip.File
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name := strings.ToLower(filepath.Base(file.Name))
		if name == "index.html" || name == "index.htm" {
			if mainFile == nil || len(file.Name) < len(mainFile.Name) {
				mainFile = file
			}
		}
	}
	if mainFile == nil {
		return "", buserr.New("ErrFileNotFound")
	}
	rc, err := mainFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	content, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return renderContent(string(content), values), nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, constant.DirPerm)
		}
		return copyFile(p, targetPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	if err := os.MkdirAll(filepath.Dir(dst), constant.DirPerm); err != nil {
		return err
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

