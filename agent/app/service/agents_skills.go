package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

const clawhubGlobalRegistry = "https://clawhub.com"
const clawhubChinaRegistry = "https://mirror-cn.clawhub.com"
const localSkillHubSource = "local-hub"
const localSkillHubPublishedStatus = "published"
const hermesManagedSkillsDir = "/opt/data/skills"

type openclawSkillsList struct {
	Skills []openclawSkillListItem `json:"skills"`
}

type openclawSkillListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Bundled     bool   `json:"bundled"`
	Disabled    bool   `json:"disabled"`
}

type openclawSkillInfo struct {
	SkillKey string `json:"skillKey"`
}

type skillhubSearchPayload struct {
	Skills  []dto.AgentSkillSearchItem `json:"skills"`
	Results []dto.AgentSkillSearchItem `json:"results"`
}

type localSkillHubItem struct {
	ID              uint
	Name            string
	Status          string
	ApplicableAgent string
	PackagePath     string
}

var clawhubSearchLinePattern = regexp.MustCompile(`^(\S+)\s+(.+?)\s+\(([\d.]+)\)$`)
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var safeLocalSkillDirNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func (a AgentService) ListSkills(req dto.AgentIDReq) ([]dto.AgentSkillItem, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return listHermesSkills(install.ContainerName)
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	output, err := cmd.RunDockerExecWithStdout(5*time.Minute, install.ContainerName, "openclaw", "skills", "list", "--json")
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	return parseOpenclawSkillsList(output)
}

func (a AgentService) SearchSkills(req dto.AgentSkillSearchReq) ([]dto.AgentSkillSearchItem, error) {
	if req.Source == localSkillHubSource {
		return nil, fmt.Errorf("local skills hub is provided by enterprise edition")
	}
	if global.CONF.Base.IsOffline {
		return nil, fmt.Errorf("offline environment cannot access remote Skills Hub")
	}
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return searchHermesSkills(install.ContainerName, req.Source, req.Keyword)
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	output, err := loadOpenclawSkillSearchOutput(install.ContainerName, req.Source, req.Keyword)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	switch req.Source {
	case "skillhub":
		return parseSkillhubSearchResult(output)
	default:
		return parseClawhubSearchResult(output, req.Source), nil
	}
}

func (a AgentService) UpdateSkill(req dto.AgentSkillUpdateReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	skillKey, err := getOpenclawSkillKey(install.ContainerName, req.Name)
	if err != nil {
		return err
	}
	setOpenclawSkillEnabled(conf, skillKey, req.Enabled)
	return writeOpenclawConfigRaw(agent.ConfigPath, conf)
}

func (a AgentService) InstallSkill(req dto.AgentSkillInstallReq) error {
	if global.CONF.Base.IsOffline && req.Source != localSkillHubSource {
		return fmt.Errorf("offline environment cannot access remote Skills Hub")
	}
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	if req.Source == localSkillHubSource {
		if !global.IsMaster {
			return fmt.Errorf("local skills hub is only available on the master node")
		}
		hostPath, skillName, err := resolveLocalSkillPackage(agent.AgentType, req.Slug)
		if err != nil {
			return err
		}
		installTask, err := task.NewTaskWithOps(skillName, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
		if err != nil {
			return err
		}
		installTask.AddSubTask("Install local skill", func(t *task.Task) error {
			mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
			return installLocalSkillPackage(mgr, install.ContainerName, agent.AgentType, agent.ConfigPath, hostPath, skillName)
		}, nil)
		go func() {
			if err := installTask.Execute(); err != nil {
				global.LOG.Errorf("install local skill failed: %v", err)
			}
		}()
		return nil
	}
	installTask, err := task.NewTaskWithOps(req.Slug, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		installTask.AddSubTask("Install Hermes skill", func(t *task.Task) error {
			mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
			return mgr.Run("docker", buildHermesDockerExecArgs(install.ContainerName, "skills", "install", req.Slug, "--yes")...)
		}, nil)
		go func() {
			if err := installTask.Execute(); err != nil {
				global.LOG.Errorf("install hermes skill failed: %v", err)
			}
		}()
		return nil
	}
	if agent.AgentType != constant.AppOpenclaw {
		return fmt.Errorf("%s does not support", agent.AgentType)
	}
	installTask.AddSubTask("Install OpenClaw skill", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
		return installOpenclawSkill(mgr, install.ContainerName, req.Source, req.Slug)
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw skill failed: %v", err)
		}
	}()
	return nil
}

func installLocalSkillPackage(mgr *cmd.CommandHelper, containerName, agentType, configPath, hostPath, skillName string) error {
	targetDir, err := resolveAgentLocalSkillDir(agentType)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp("", "1panel-agent-skill-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	extractRoot := filepath.Join(tempDir, "extract")
	if err := os.MkdirAll(extractRoot, 0755); err != nil {
		return err
	}
	if err := extractLocalSkillPackage(hostPath, extractRoot); err != nil {
		return err
	}
	copyRoot, installedSkillName, err := normalizeLocalSkillInstallRoot(extractRoot, filepath.Join(tempDir, "install"), skillName)
	if err != nil {
		return err
	}
	if err := mgr.Run("docker", "exec", containerName, "mkdir", "-p", targetDir); err != nil {
		return err
	}
	if err := mgr.Run("docker", "cp", copyRoot+"/.", containerName+":"+targetDir); err != nil {
		return err
	}
	if agentType == constant.AppOpenclaw {
		return registerOpenclawLocalSkill(containerName, configPath, installedSkillName)
	}
	return nil
}

func resolveAgentLocalSkillDir(agentType string) (string, error) {
	switch agentType {
	case constant.AppOpenclaw:
		return openclawManagedSkillsDir, nil
	case constant.AppHermesAgent:
		return hermesManagedSkillsDir, nil
	default:
		return "", fmt.Errorf("%s does not support", agentType)
	}
}

func resolveLocalSkillPackage(agentType, skillID string) (string, string, error) {
	skillID = strings.TrimSpace(skillID)
	id, err := strconv.ParseUint(skillID, 10, 64)
	if err != nil || id == 0 {
		return "", "", fmt.Errorf("invalid local skill id")
	}
	dbPath := filepath.Join(global.CONF.Base.InstallDir, "1panel", "db", "enterprise.db")
	if _, err := os.Stat(dbPath); err != nil {
		return "", "", fmt.Errorf("local skills hub is unavailable")
	}
	db, err := common.GetDBWithPath(dbPath)
	if err != nil {
		return "", "", err
	}
	defer common.CloseDB(db)
	var item localSkillHubItem
	if err := db.Table("skill_hub_items").
		Select("id,name,status,applicable_agent,package_path").
		Where("id = ? AND status = ?", uint(id), localSkillHubPublishedStatus).
		First(&item).Error; err != nil {
		return "", "", fmt.Errorf("local skill is not published or does not exist")
	}
	if !matchLocalSkillAgent(item.ApplicableAgent, agentType) {
		return "", "", fmt.Errorf("local skill does not support %s", agentType)
	}
	packagePath, err := validateLocalSkillPackagePath(item.PackagePath)
	if err != nil {
		return "", "", err
	}
	return packagePath, sanitizeLocalSkillDirName(item.Name, packagePath, skillID), nil
}

func matchLocalSkillAgent(applicableAgent, agentType string) bool {
	applicableAgent = strings.TrimSpace(applicableAgent)
	if applicableAgent == "" {
		return true
	}
	for _, item := range strings.Split(applicableAgent, ",") {
		normalized := normalizeLocalSkillAgent(strings.TrimSpace(item))
		if normalized == "common" || normalized == normalizeLocalSkillAgent(agentType) {
			return true
		}
	}
	return false
}

func normalizeLocalSkillAgent(agentType string) string {
	switch strings.ToLower(strings.TrimSpace(agentType)) {
	case constant.AppHermesAgent:
		return "hermes"
	default:
		return strings.ToLower(strings.TrimSpace(agentType))
	}
}

func sanitizeLocalSkillDirName(name, packagePath, fallback string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(packagePath), localSkillPackageExtension(packagePath))
	}
	name = strings.Trim(safeLocalSkillDirNamePattern.ReplaceAllString(name, "-"), "-")
	if name == "" || name == "." || name == ".." {
		name = "skill-" + strings.TrimSpace(fallback)
	}
	if name == "skill-" {
		return "skill"
	}
	return name
}

func validateLocalSkillPackagePath(packagePath string) (string, error) {
	packagePath = strings.TrimSpace(packagePath)
	if packagePath == "" {
		return "", fmt.Errorf("skill package path is required")
	}
	if _, err := localSkillPackageFormat(packagePath); err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		return "", err
	}
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", err
	}
	root := filepath.Join(global.CONF.Base.InstallDir, "1panel", "uploads", "skills-hub")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		// upload directory may not exist yet on agent host; treat as invalid
		return "", fmt.Errorf("invalid local skill package path")
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		return "", fmt.Errorf("invalid local skill package path")
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("skill package must be a file")
	}
	return resolvedPath, nil
}

func normalizeLocalSkillInstallRoot(extractRoot, installRoot, skillName string) (string, string, error) {
	entries, err := os.ReadDir(extractRoot)
	if err != nil {
		return "", "", err
	}
	if len(entries) == 0 {
		return "", "", fmt.Errorf("skill package is empty")
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return extractRoot, entries[0].Name(), nil
	}
	skillRoot := filepath.Join(installRoot, skillName)
	if err := os.MkdirAll(skillRoot, 0755); err != nil {
		return "", "", err
	}
	if err := copyLocalDirContents(extractRoot, skillRoot); err != nil {
		return "", "", err
	}
	return installRoot, skillName, nil
}

func registerOpenclawLocalSkill(containerName, configPath, skillName string) error {
	conf, err := readOpenclawConfig(configPath)
	if err != nil {
		return err
	}
	skillKey, err := getOpenclawSkillKey(containerName, skillName)
	if err != nil {
		return err
	}
	setOpenclawSkillEnabled(conf, skillKey, true)
	return writeOpenclawConfigRaw(configPath, conf)
}

func extractLocalSkillPackage(packagePath, targetDir string) error {
	format, err := localSkillPackageFormat(packagePath)
	if err != nil {
		return err
	}
	switch format {
	case "zip":
		return unzipLocalSkillPackage(packagePath, targetDir)
	case "tar", "targz":
		return untarLocalSkillPackage(packagePath, targetDir, format == "targz")
	case "7z":
		return un7zLocalSkillPackage(packagePath, targetDir)
	default:
		return fmt.Errorf("unsupported skill package format")
	}
}

func unzipLocalSkillPackage(packagePath, targetDir string) error {
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	root, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	for _, file := range reader.File {
		name, err := safeLocalSkillZipEntryName(file.Name)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(root, name)
		rel, err := filepath.Rel(root, targetPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip entry: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsupported zip entry: %s", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, copyErr := io.Copy(dst, rc)
		closeErr := dst.Close()
		_ = rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func untarLocalSkillPackage(packagePath, targetDir string, gzipped bool) error {
	src, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer src.Close()
	var reader io.Reader = src
	var gzipReader *gzip.Reader
	if gzipped {
		gzipReader, err = gzip.NewReader(src)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	root, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name, err := safeLocalSkillZipEntryName(header.Name)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(root, name)
		rel, err := filepath.Rel(root, targetPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar entry: %s", header.Name)
		}
		info := header.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return err
			}
			continue
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(dst, tarReader)
		closeErr := dst.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func un7zLocalSkillPackage(packagePath, targetDir string) error {
	if _, err := listLocal7zEntries(packagePath); err != nil {
		return err
	}
	root, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	bin, err := findLocal7zBinary()
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp("", "1panel-agent-skill-7z-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).Run(bin, "x", "-y", "-o"+tempDir, packagePath); err != nil {
		return err
	}
	if err := validateExtractedLocalSkillPackage(tempDir); err != nil {
		return err
	}
	return copyLocalDirContents(tempDir, root)
}

func safeLocalSkillZipEntryName(name string) (string, error) {
	clean := path.Clean(strings.ReplaceAll(name, "\\", "/"))
	clean = strings.TrimLeft(clean, "/")
	if clean == "." || clean == "" || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("invalid zip entry: %s", name)
	}
	return clean, nil
}

func copyLocalDirContents(sourceRoot, targetRoot string) error {
	return filepath.Walk(sourceRoot, func(sourcePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if sourcePath == sourceRoot {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return fmt.Errorf("unsupported skill package entry: %s", sourcePath)
		}
		rel, err := filepath.Rel(sourceRoot, sourcePath)
		if err != nil {
			return err
		}
		name, err := safeLocalSkillZipEntryName(rel)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetRoot, name)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		return copyLocalSkillFile(sourcePath, targetPath)
	})
}

func copyLocalSkillFile(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return err
	}
	return dst.Close()
}

func validateExtractedLocalSkillPackage(root string) error {
	return filepath.Walk(root, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if currentPath == root {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return fmt.Errorf("unsupported skill package entry: %s", currentPath)
		}
		return nil
	})
}

func localSkillPackageExtension(name string) string {
	lower := strings.ToLower(filepath.Base(strings.TrimSpace(name)))
	if lower == "" || lower == "." {
		return ""
	}
	if strings.HasSuffix(lower, ".tar.gz") {
		return ".tar.gz"
	}
	return strings.ToLower(filepath.Ext(lower))
}

func localSkillPackageFormat(name string) (string, error) {
	switch localSkillPackageExtension(name) {
	case ".zip":
		return "zip", nil
	case ".7z":
		return "7z", nil
	case ".tar":
		return "tar", nil
	case ".tar.gz":
		return "targz", nil
	default:
		return "", fmt.Errorf("only .zip, .7z, .tar, and .tar.gz skill packages are supported")
	}
}

func findLocal7zBinary() (string, error) {
	for _, name := range []string{"7z", "7za", "7zr"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("7z command is required to process .7z skill packages")
}

func listLocal7zEntries(packagePath string) ([]string, error) {
	bin, err := findLocal7zBinary()
	if err != nil {
		return nil, err
	}
	output, err := exec.Command(bin, "l", "-slt", packagePath).CombinedOutput()
	if err != nil {
		if message := strings.TrimSpace(string(output)); message != "" {
			return nil, fmt.Errorf("7z failed: %w: %s", err, message)
		}
		return nil, fmt.Errorf("7z failed: %w", err)
	}
	entries := make([]string, 0)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Path = ") {
			continue
		}
		entry := strings.TrimSpace(strings.TrimPrefix(line, "Path = "))
		if entry == "" || entry == packagePath || entry == filepath.Base(packagePath) {
			continue
		}
		name, err := safeLocalSkillZipEntryName(entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, name)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("skill package is empty")
	}
	return entries, nil
}

func (a AgentService) UninstallSkill(req dto.AgentSkillUninstallReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType != constant.AppHermesAgent {
		return fmt.Errorf("%s does not support", agent.AgentType)
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(5*time.Minute)).Run(
		"docker",
		buildHermesDockerExecCommandArgs(
			install.ContainerName,
			"sh",
			"-lc",
			fmt.Sprintf(`printf 'y\n' | %s skills uninstall "$1"`, hermesExecutablePath),
			"sh",
			req.Name,
		)...,
	)
}

func parseOpenclawSkillsList(output string) ([]dto.AgentSkillItem, error) {
	payloadBytes, err := extractEmbeddedJSON(output)
	if err != nil {
		return nil, err
	}
	if len(payloadBytes) == 0 {
		return nil, nil
	}
	var payload openclawSkillsList
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}
	items := make([]dto.AgentSkillItem, 0, len(payload.Skills))
	for _, item := range payload.Skills {
		items = append(items, dto.AgentSkillItem{
			Name:        item.Name,
			Description: item.Description,
			Source:      item.Source,
			Bundled:     item.Bundled,
			Disabled:    item.Disabled,
		})
	}
	return items, nil
}

func loadOpenclawSkillSearchOutput(containerName, source, keyword string) (string, error) {
	switch source {
	case "skillhub":
		return cmd.RunDockerExecWithStdout(2*time.Minute, containerName, "skillhub", "search", keyword, "--json")
	default:
		return cmd.NewCommandMgr(cmd.WithTimeout(2*time.Minute)).
			RunWithStdout(
				"docker",
				"exec",
				"-e",
				"CLAWHUB_REGISTRY="+resolveClawhubRegistry(source),
				containerName,
				"clawhub",
				"search",
				keyword,
			)
	}
}

func parseSkillhubSearchResult(output string) ([]dto.AgentSkillSearchItem, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}
	var list []dto.AgentSkillSearchItem
	if err := json.Unmarshal([]byte(trimmed), &list); err == nil {
		for i := range list {
			list[i].Source = "skillhub"
		}
		return list, nil
	}
	var payload skillhubSearchPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, err
	}
	items := payload.Skills
	if len(items) == 0 {
		items = payload.Results
	}
	for i := range items {
		items[i].Source = "skillhub"
	}
	return items, nil
}

func parseClawhubSearchResult(output, source string) []dto.AgentSkillSearchItem {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	items := make([]dto.AgentSkillSearchItem, 0, len(lines))
	for _, line := range lines {
		matches := clawhubSearchLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) != 4 {
			continue
		}
		items = append(items, dto.AgentSkillSearchItem{
			Slug:   matches[1],
			Name:   matches[2],
			Score:  matches[3],
			Source: source,
		})
	}
	return items
}

func installOpenclawSkill(mgr *cmd.CommandHelper, containerName, source, slug string) error {
	if err := mgr.Run("docker", "exec", containerName, "mkdir", "-p", openclawManagedSkillsDir); err != nil {
		return err
	}
	switch source {
	case "clawhub-global", "clawhub-cn":
		return mgr.Run(
			"docker",
			"exec",
			"-e",
			"CLAWHUB_REGISTRY="+resolveClawhubRegistry(source),
			containerName,
			"clawhub",
			"--workdir",
			"/home/node/.openclaw",
			"--dir",
			"skills",
			"install",
			slug,
		)
	default:
		return mgr.Run("docker", "exec", containerName, "skillhub", "--dir", openclawManagedSkillsDir, "install", slug)
	}
}

func resolveClawhubRegistry(source string) string {
	switch source {
	case "clawhub-cn":
		return clawhubChinaRegistry
	default:
		return clawhubGlobalRegistry
	}
}

func getOpenclawSkillKey(containerName, name string) (string, error) {
	output, err := cmd.RunDockerExecWithStdout(2*time.Minute, containerName, "openclaw", "skills", "info", name, "--json")
	if err != nil {
		return "", err
	}
	return parseOpenclawSkillKey(name, output)
}

func parseOpenclawSkillKey(name, output string) (string, error) {
	payloadBytes, err := extractEmbeddedJSON(output)
	if err != nil {
		return "", err
	}
	var payload openclawSkillInfo
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", err
	}
	if payload.SkillKey == "" {
		return name, nil
	}
	return payload.SkillKey, nil
}

func extractEmbeddedJSON(output string) ([]byte, error) {
	trimmed := strings.TrimSpace(ansiEscapePattern.ReplaceAllString(output, ""))
	if trimmed == "" {
		return nil, nil
	}
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] != '{' && trimmed[i] != '[' {
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(trimmed[i:]))
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err == nil {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("json payload not found")
}

func setOpenclawSkillEnabled(conf map[string]interface{}, skillKey string, enabled bool) {
	skills := ensureChildMap(conf, "skills")
	entries := ensureChildMap(skills, "entries")
	entry := ensureChildMap(entries, skillKey)
	entry["enabled"] = enabled
}
