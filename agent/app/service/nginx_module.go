package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	dockerUtils "github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/mattn/go-shellwords"
	"github.com/subosito/gotenv"
)

const (
	nginxModuleBuildDynamic = "dynamic"
	nginxModuleBuildStatic  = "static"

	nginxModuleProviderLocal = "local"

	nginxModuleStatusPending = "pending"
	nginxModuleStatusReady   = "ready"
	nginxModuleStatusFailed  = "failed"

	nginxModuleLoadEnabled  = "enabled"
	nginxModuleLoadDisabled = "disabled"

	nginxModuleOperateCreate = "create"
	nginxModuleOperateUpdate = "update"
	nginxModuleOperateDelete = "delete"

	nginxModuleBuildDir           = "build"
	nginxModuleModulesDir         = "modules"
	nginxModuleConfDir            = "conf"
	nginxModuleTmpDir             = "tmp"
	nginxModuleEnabledConfDir     = "modules-enabled"
	nginxModuleStagingDir         = ".staging"
	nginxModuleLibDir             = "lib"
	nginxModuleBuilderFile        = "Dockerfile.modules"
	nginxModuleStoreFile          = "module.json"
	nginxModuleCatalogFile        = "module.catalog.json"
	nginxModuleCatalogPendingFile = "module.catalog.pending.json"
	nginxModulePreScriptFile      = "module-pre.sh"
	nginxModuleConfigArgsFile     = "module-config.args"
	nginxModuleStaticPreScript    = "pre.sh"
	nginxModuleManifestFile       = "manifest.json"

	nginxModuleContainerRoot = "/usr/local/openresty/nginx/modules/1panel"
	nginxModuleConfigPrefix  = "1panel-module-"

	nginxModuleStaticBuildHint = "; switch the module to static build mode to compile it with the full image"
)

var errNginxModuleBuilderMissing = errors.New("dynamic module builder not found")

type nginxModuleBuildSpec struct {
	Install   model.AppInstall
	Module    dto.NginxModule
	Target    dto.NginxModuleTarget
	BuildPath string
	Force     bool
	Mirror    string
	Task      *task.Task
}

// nginxModuleArtifactProvider keeps local builds replaceable by a future
// precompiled-artifact resolver without changing module state or API contracts.
type nginxModuleArtifactProvider interface {
	Name() string
	Resolve(spec nginxModuleBuildSpec) (dto.NginxModuleBuild, error)
}

type localNginxModuleProvider struct{}

type nginxModuleState struct {
	Name      string                 `json:"name"`
	Custom    bool                   `json:"custom,omitempty"`
	Script    string                 `json:"script,omitempty"`
	Packages  []string               `json:"packages,omitempty"`
	Params    string                 `json:"params,omitempty"`
	Enable    bool                   `json:"enable"`
	BuildMode string                 `json:"buildMode,omitempty"`
	Provider  string                 `json:"provider,omitempty"`
	LoadOrder int                    `json:"loadOrder,omitempty"`
	Builds    []dto.NginxModuleBuild `json:"builds,omitempty"`
	LastError string                 `json:"lastError,omitempty"`
}

func (localNginxModuleProvider) Name() string {
	return nginxModuleProviderLocal
}

func resolveNginxModuleTarget(install model.AppInstall) (dto.NginxModuleTarget, string, error) {
	target, warning, err := resolveNginxRuntimeTarget(install)
	if err != nil {
		return target, warning, err
	}
	builderPath := path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleBuilderFile)
	builderContent, err := os.ReadFile(builderPath)
	if err != nil {
		return target, warning, fmt.Errorf("%w: %v", errNginxModuleBuilderMissing, err)
	}
	builderSum := sha256.Sum256(builderContent)
	target.BuilderDigest = hex.EncodeToString(builderSum[:])
	setNginxModuleTargetKey(&target)
	return target, warning, nil
}

func resolveNginxRuntimeTarget(install model.AppInstall) (dto.NginxModuleTarget, string, error) {
	target := dto.NginxModuleTarget{
		OpenRestyVersion: install.Version,
		Architecture:     runtime.GOARCH,
	}
	envContent, _ := os.ReadFile(install.GetEnvPath())
	images, imageErr := dockerUtils.GetImagesFromDockerCompose(envContent, []byte(install.DockerCompose))
	if imageErr != nil || len(images) == 0 {
		images, imageErr = compose.GetComposeImages(install.GetComposePath())
	}
	var warning string
	if imageErr == nil && len(images) > 0 {
		target.Image = images[0]
		inspectMgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Minute))
		inspectOut, inspectErr := inspectMgr.RunWithStdout("docker", "image", "inspect", "--format={{.Id}}\t{{.Architecture}}", target.Image)
		if inspectErr != nil {
			warning = fmt.Sprintf("inspect OpenResty image %s failed, module rebuilds will fall back to host architecture %s: %v", target.Image, target.Architecture, inspectErr)
		} else if fields := strings.Fields(inspectOut); len(fields) >= 2 {
			target.ImageDigest = fields[0]
			target.Architecture = fields[1]
		}
	}
	setNginxModuleTargetKey(&target)
	return target, warning, nil
}

func setNginxModuleTargetKey(target *dto.NginxModuleTarget) {
	keyInput := strings.Join([]string{target.OpenRestyVersion, target.Architecture, target.ImageDigest, target.BuilderDigest}, "\x00")
	keySum := sha256.Sum256([]byte(keyInput))
	target.Key = fmt.Sprintf("%s-%s-%s", sanitizeModulePathPart(target.OpenRestyVersion), target.Architecture, hex.EncodeToString(keySum[:6]))
}

// nginxModuleDynamicSupported reports whether the installed version ships both
// the dynamic module builder and the module catalog.
func nginxModuleDynamicSupported(install model.AppInstall) bool {
	fileOp := files.NewFileOp()
	buildPath := path.Join(install.GetPath(), nginxModuleBuildDir)
	return fileOp.Stat(path.Join(buildPath, nginxModuleBuilderFile)) &&
		fileOp.Stat(path.Join(buildPath, nginxModuleCatalogFile))
}

// nginxModuleStaticSupported reports whether the install can recompile its own
// OpenResty image, which is what a static module build needs. Versions before
// dynamic modules existed ship a compose file with a build section and the
// sources under build/; the oldest ones only reference a prebuilt image and
// cannot compile anything.
func nginxModuleStaticSupported(install model.AppInstall) bool {
	if !files.NewFileOp().Stat(path.Join(install.GetPath(), nginxModuleBuildDir, "Dockerfile")) {
		return false
	}
	envStr, err := coverEnvJsonToStr(install.Env)
	if err != nil {
		return false
	}
	project, err := dockerUtils.GetComposeProject(install.Name, install.GetPath(),
		[]byte(install.DockerCompose), []byte(envStr), true)
	if err != nil {
		return false
	}
	for _, service := range project.AllServices() {
		if service.Build != nil {
			return true
		}
	}
	return false
}

// defaultNginxModuleBuildMode picks the mode an install can actually perform.
//
// Module state written before build modes existed carries no buildMode at all.
// Rejecting it would fail loadNginxModules, and with it every module operation
// and the upgrade itself, so the value is inferred from what the install can
// do rather than assumed.
func defaultNginxModuleBuildMode(install model.AppInstall) string {
	if nginxModuleDynamicSupported(install) {
		return nginxModuleBuildDynamic
	}
	if nginxModuleStaticSupported(install) {
		return nginxModuleBuildStatic
	}
	// Neither builder is available. Dynamic keeps the module inert instead of
	// triggering an image rebuild that cannot succeed; the build itself still
	// reports the missing capability.
	return nginxModuleBuildDynamic
}

func syncNginxModuleBuilder(detailBuildDir, installBuildDir string) error {
	sourcePath := path.Join(detailBuildDir, nginxModuleBuilderFile)
	targetPath := path.Join(installBuildDir, nginxModuleBuilderFile)
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			if removeErr := os.Remove(targetPath); removeErr != nil && !os.IsNotExist(removeErr) {
				return removeErr
			}
			return nil
		}
		return err
	}
	return files.NewFileOp().CopyFile(sourcePath, installBuildDir)
}

// resolveNginxModuleBuildMirror picks the apt mirror for module builds: the
// explicit request value wins, otherwise the install's saved
// CONTAINER_PACKAGE_URL.
func resolveNginxModuleBuildMirror(install model.AppInstall, mirror string) string {
	if mirror != "" {
		return mirror
	}
	envs, err := gotenv.Read(install.GetEnvPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(envs["CONTAINER_PACKAGE_URL"])
}

func buildDynamicNginxModules(install model.AppInstall, modules []dto.NginxModule, selected []string, force bool, mirror string, catalogPath string, parentTask *task.Task) ([]dto.NginxModule, error) {
	// Skip target resolution entirely when nothing needs a dynamic build, so
	// installs without dynamic modules do not require Dockerfile.modules.
	if !hasDynamicNginxModuleBuildTask(modules, selected) {
		return modules, nil
	}
	target, targetWarning, err := resolveNginxModuleTarget(install)
	if err != nil {
		return modules, err
	}
	if targetWarning != "" {
		parentTask.Logf("WARNING: %s", targetWarning)
	}
	originalModules := cloneNginxModules(modules)
	selectedNames := make(map[string]struct{}, len(selected))
	for _, name := range selected {
		selectedNames[name] = struct{}{}
	}
	buildPath := path.Join(install.GetPath(), nginxModuleBuildDir)
	buildMirror := resolveNginxModuleBuildMirror(install, mirror)
	for i := range modules {
		module := &modules[i]
		normalizeNginxModule(module)
		if !nginxModuleNeedsDynamicBuild(*module, selectedNames) {
			continue
		}
		provider, providerErr := getNginxModuleProvider(module.Provider)
		if providerErr != nil {
			return modules, providerErr
		}
		previousBuild := findCurrentNginxModuleBuild(*module, target)
		build, buildErr := provider.Resolve(nginxModuleBuildSpec{
			Install: install, Module: *module, Target: target, BuildPath: buildPath, Force: force, Mirror: buildMirror, Task: parentTask,
		})
		if buildErr != nil {
			build.Provider = provider.Name()
			build.Status = nginxModuleStatusFailed
			build.Target = target
			build.Error = buildErr.Error()
			build.BuiltAt = time.Now()
			failedModules := recordNginxModuleBuildFailure(originalModules, module.Name, build, previousBuild)
			removeNginxModuleOutputsNotReferenced(install, modules, failedModules)
			_ = saveNginxModulesWithCatalog(install, failedModules, catalogPath)
			return failedModules, fmt.Errorf("build dynamic module %s: %w%s", module.Name, buildErr, nginxModuleStaticBuildErrorHint(*module))
		}
		err = validateNginxModuleArtifacts(install, build.Artifacts)
		if err != nil {
			for outputDir := range nginxModuleOutputDirectories(install, []dto.NginxModule{{Builds: []dto.NginxModuleBuild{build}}}) {
				_ = os.RemoveAll(outputDir)
			}
			build.Status = nginxModuleStatusFailed
			build.Error = err.Error()
			build.Artifacts = nil
			build.BuiltAt = time.Now()
			failedModules := recordNginxModuleBuildFailure(originalModules, module.Name, build, previousBuild)
			removeNginxModuleOutputsNotReferenced(install, modules, failedModules)
			_ = saveNginxModulesWithCatalog(install, failedModules, catalogPath)
			return failedModules, fmt.Errorf("validate dynamic module %s: %w%s", module.Name, err, nginxModuleStaticBuildErrorHint(*module))
		}
		module.LastError = ""
		upsertNginxModuleBuild(module, build)
	}
	return modules, nil
}

func (localNginxModuleProvider) Resolve(spec nginxModuleBuildSpec) (dto.NginxModuleBuild, error) {
	params, err := normalizeDynamicModuleParams(spec.Module.Params)
	if err != nil {
		return dto.NginxModuleBuild{}, err
	}
	configureArgs, err := parseDynamicModuleParams(params)
	if err != nil {
		return dto.NginxModuleBuild{}, err
	}
	if err = validateNginxModulePackages(spec.Module.Packages); err != nil {
		return dto.NginxModuleBuild{}, err
	}
	buildHash, err := nginxModuleBuildHash(spec.Module, spec.Target, params)
	if err != nil {
		return dto.NginxModuleBuild{}, err
	}
	result := dto.NginxModuleBuild{
		Provider:  nginxModuleProviderLocal,
		BuildMode: nginxModuleBuildDynamic,
		Status:    nginxModuleStatusPending,
		Hash:      buildHash,
		Target:    spec.Target,
	}
	modulesRoot := path.Join(spec.Install.GetPath(), nginxModuleModulesDir)
	if !spec.Force {
		if current := findCurrentNginxModuleBuild(spec.Module, spec.Target); current != nil && current.Status == nginxModuleStatusReady {
			if validateNginxModuleArtifacts(spec.Install, current.Artifacts) == nil {
				return *current, nil
			}
		}
	}
	outputRevision := buildHash
	if spec.Force {
		outputRevision = fmt.Sprintf("%s-r%d", buildHash, time.Now().UnixNano())
	}
	modulePathName := nginxModulePathName(spec.Module.Name)
	finalPath := path.Join(modulesRoot, spec.Target.Key, modulePathName, outputRevision)
	buildComplete := false
	defer func() {
		if !buildComplete {
			_ = os.RemoveAll(finalPath)
		}
	}()

	preScriptPath := path.Join(spec.BuildPath, nginxModuleTmpDir, nginxModulePreScriptFile)
	if err = os.WriteFile(preScriptPath, []byte("#!/bin/bash\nset -e\n"+spec.Module.Script+"\n"), constant.FilePerm); err != nil {
		return result, err
	}
	defer os.Remove(preScriptPath)
	configureArgsPath := path.Join(spec.BuildPath, nginxModuleTmpDir, nginxModuleConfigArgsFile)
	if err = os.WriteFile(configureArgsPath, []byte(strings.Join(configureArgs, "\n")+"\n"), constant.FilePerm); err != nil {
		return result, err
	}
	defer os.Remove(configureArgsPath)

	shortHash := buildHash
	if len(shortHash) > 12 {
		shortHash = shortHash[:12]
	}
	taskSuffix := sanitizeModulePathPart(spec.Task.TaskID)
	if len(taskSuffix) > 8 {
		taskSuffix = taskSuffix[:8]
	}
	tempImage := fmt.Sprintf("1panel/openresty-module-builder:%s-%s-%s", modulePathName, shortHash, taskSuffix)
	tempContainer := fmt.Sprintf("1panel-module-%s-%s-%s", modulePathName, shortHash, taskSuffix)
	commandMgr := cmd.NewCommandMgr(cmd.WithTask(*spec.Task), cmd.WithTimeout(120*time.Minute))
	buildArgs := []string{
		"build", "--target", "module-output",
		"-f", path.Join(spec.BuildPath, nginxModuleBuilderFile),
		"-t", tempImage,
		"--build-arg", "PANEL_OPENRESTY_VERSION=" + spec.Install.Version,
		"--build-arg", "RESTY_ADD_PACKAGE_BUILDDEPS=" + strings.Join(spec.Module.Packages, " "),
	}
	if spec.Mirror != "" {
		buildArgs = append(buildArgs, "--build-arg", "CONTAINER_PACKAGE_URL="+spec.Mirror)
	}
	buildArgs = append(buildArgs, spec.BuildPath)
	if err = commandMgr.Run("docker", buildArgs...); err != nil {
		return result, err
	}
	cleanupMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Minute))
	defer func() {
		_ = cleanupMgr.Run("docker", "rm", "-f", tempContainer)
		_ = cleanupMgr.Run("docker", "image", "rm", "-f", tempImage)
	}()
	if err = commandMgr.Run("docker", "create", "--name", tempContainer, tempImage, "/bin/true"); err != nil {
		return result, err
	}

	stagingPath := path.Join(spec.Install.GetPath(), nginxModuleModulesDir, nginxModuleStagingDir, modulePathName+"-"+shortHash)
	_ = os.RemoveAll(stagingPath)
	if err = os.MkdirAll(stagingPath, constant.DirPerm); err != nil {
		return result, err
	}
	defer os.RemoveAll(stagingPath)
	if err = commandMgr.Run("docker", "cp", tempContainer+":/out/.", stagingPath); err != nil {
		return result, err
	}
	if artifacts, artifactErr := collectNginxModuleArtifacts(stagingPath, stagingPath); artifactErr != nil || len(artifacts) == 0 {
		if artifactErr != nil {
			return result, artifactErr
		}
		return result, errors.New("dynamic module build produced no loadable .so files")
	}
	if err = os.MkdirAll(path.Dir(finalPath), constant.DirPerm); err != nil {
		return result, err
	}
	_ = os.RemoveAll(finalPath)
	if err = os.Rename(stagingPath, finalPath); err != nil {
		return result, err
	}
	artifacts, err := collectNginxModuleArtifacts(finalPath, modulesRoot)
	if err != nil {
		return result, err
	}
	result.Status = nginxModuleStatusReady
	result.Artifacts = artifacts
	result.BuiltAt = time.Now()
	manifest, _ := json.MarshalIndent(result, "", "  ")
	_ = os.WriteFile(path.Join(finalPath, nginxModuleManifestFile), manifest, constant.FilePerm)
	buildComplete = true
	return result, nil
}

func commitNginxModuleBuilds(install model.AppInstall, previousModules, modules []dto.NginxModule, reload bool, catalogPath string) error {
	if err := saveNginxModulesWithCatalog(install, modules, catalogPath); err != nil {
		removeNginxModuleOutputsNotReferenced(install, modules, previousModules)
		return err
	}
	if err := reconcileDynamicNginxModuleConfig(install, modules, reload); err != nil {
		rollbackModules := recordNginxModuleActivationFailure(previousModules, modules, err)
		_ = saveNginxModulesWithCatalog(install, rollbackModules, catalogPath)
		removeNginxModuleOutputsNotReferenced(install, modules, previousModules)
		return err
	}
	removeNginxModuleOutputsNotReferenced(install, previousModules, modules)
	return nil
}

func validateNginxModuleArtifacts(install model.AppInstall, artifacts []dto.NginxModuleArtifact) error {
	if len(artifacts) == 0 {
		return errors.New("module build produced no loadable artifacts")
	}
	modulesRoot := filepath.Clean(path.Join(install.GetPath(), nginxModuleModulesDir))
	seen := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		if _, ok := seen[artifact.Path]; ok {
			return fmt.Errorf("duplicate module artifact path %q", artifact.Path)
		}
		seen[artifact.Path] = struct{}{}
		if artifact.Name != path.Base(artifact.Path) {
			return fmt.Errorf("module artifact name %q does not match path %q", artifact.Name, artifact.Path)
		}
		if !re.IsValidNginxModuleChecksum(artifact.Checksum) {
			return fmt.Errorf("invalid checksum for module artifact %q", artifact.Path)
		}
		if !re.IsValidNginxModuleArtifact(artifact.Path) {
			return fmt.Errorf("invalid module artifact path %q", artifact.Path)
		}
		artifactFullPath := filepath.Clean(filepath.Join(modulesRoot, filepath.FromSlash(artifact.Path)))
		if !strings.HasPrefix(artifactFullPath, modulesRoot+string(os.PathSeparator)) {
			return fmt.Errorf("module artifact path %q escapes the module directory", artifact.Path)
		}
		info, err := os.Lstat(artifactFullPath)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("module artifact %q is not a regular file", artifact.Path)
		}
		file, err := os.Open(artifactFullPath)
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		actualChecksum := hex.EncodeToString(hash.Sum(nil))
		if !strings.EqualFold(actualChecksum, artifact.Checksum) {
			return fmt.Errorf("checksum mismatch for module artifact %q", artifact.Path)
		}
	}
	return nil
}

func collectNginxModuleArtifacts(root, relativeRoot string) ([]dto.NginxModuleArtifact, error) {
	var artifacts []dto.NginxModuleArtifact
	err := filepath.Walk(root, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ".so" || filepath.Base(filepath.Dir(filePath)) == nginxModuleLibDir {
			return nil
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		_ = file.Close()
		if copyErr != nil {
			return copyErr
		}
		relativePath, err := filepath.Rel(relativeRoot, filePath)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, dto.NginxModuleArtifact{
			Name: info.Name(), Path: filepath.ToSlash(relativePath), Checksum: hex.EncodeToString(hash.Sum(nil)),
		})
		return nil
	})
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Name < artifacts[j].Name })
	return artifacts, err
}

func validateNginxModuleLoadConfig(install model.AppInstall, target dto.NginxModuleTarget, validationName, loadDirectives string) error {
	if target.Image == "" {
		return errors.New("target OpenResty image is not resolved")
	}
	var config strings.Builder
	config.WriteString(loadDirectives)
	config.WriteString("events {}\nhttp {}\n")
	configSum := sha256.Sum256([]byte(config.String()))
	testConfig := path.Join(install.GetPath(), nginxModuleModulesDir, ".test-"+nginxModulePathName(validationName)+"-"+hex.EncodeToString(configSum[:6])+".conf")
	if err := os.WriteFile(testConfig, []byte(config.String()), constant.FilePerm); err != nil {
		return err
	}
	defer os.Remove(testConfig)
	modulesRoot := path.Join(install.GetPath(), nginxModuleModulesDir)
	args := []string{
		"run", "--rm", "--network", "none",
		"-v", modulesRoot + ":" + nginxModuleContainerRoot + ":ro",
		"-v", testConfig + ":/tmp/1panel-module-test.conf:ro",
		"--entrypoint", "/usr/local/openresty/nginx/sbin/nginx",
		target.Image, "-t", "-c", "/tmp/1panel-module-test.conf",
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(5*time.Minute)).Run("docker", args...)
}

type nginxModuleConfigSnapshot map[string][]byte

type openrestyUpgradeSnapshot struct {
	installPath string
	backupPath  string
	existing    map[string]bool
}

var openrestyUpgradeSnapshotPaths = []string{
	nginxModuleBuildDir,
	nginxModuleModulesDir,
	"scripts",
	path.Join(nginxModuleConfDir, nginxModuleEnabledConfDir),
	path.Join(nginxModuleConfDir, "nginx.conf"),
	"docker-compose.yml",
	".env",
}

func createOpenrestyUpgradeSnapshot(installPath string) (*openrestyUpgradeSnapshot, error) {
	backupPath, err := os.MkdirTemp("", "1panel-openresty-upgrade-*")
	if err != nil {
		return nil, err
	}
	snapshot := &openrestyUpgradeSnapshot{
		installPath: installPath,
		backupPath:  backupPath,
		existing:    make(map[string]bool, len(openrestyUpgradeSnapshotPaths)),
	}
	for _, relativePath := range openrestyUpgradeSnapshotPaths {
		sourcePath := path.Join(installPath, relativePath)
		if _, err = os.Stat(sourcePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			snapshot.Cleanup()
			return nil, err
		}
		snapshot.existing[relativePath] = true
		if err = copyOpenrestyUpgradeSnapshotEntry(sourcePath, path.Join(backupPath, relativePath)); err != nil {
			snapshot.Cleanup()
			return nil, err
		}
	}
	return snapshot, nil
}

func copyOpenrestyUpgradeSnapshotEntry(sourcePath, targetPath string) error {
	if err := os.MkdirAll(path.Dir(targetPath), constant.DirPerm); err != nil {
		return err
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	if info.IsDir() {
		return fileOp.CopyDir(sourcePath, path.Dir(targetPath))
	}
	return fileOp.CopyFile(sourcePath, path.Dir(targetPath))
}

func (s *openrestyUpgradeSnapshot) Restore() error {
	for _, relativePath := range openrestyUpgradeSnapshotPaths {
		targetPath := path.Join(s.installPath, relativePath)
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
		if !s.existing[relativePath] {
			continue
		}
		if err := copyOpenrestyUpgradeSnapshotEntry(path.Join(s.backupPath, relativePath), targetPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *openrestyUpgradeSnapshot) Cleanup() {
	if s != nil && s.backupPath != "" {
		_ = os.RemoveAll(s.backupPath)
	}
}

func reconcileDynamicNginxModuleConfig(install model.AppInstall, modules []dto.NginxModule, reload bool) error {
	var target dto.NginxModuleTarget
	if hasDynamicNginxModuleBuildTask(modules, nil) {
		var targetWarning string
		var err error
		target, targetWarning, err = resolveNginxModuleTarget(install)
		if err != nil {
			return err
		}
		if targetWarning != "" {
			global.LOG.Warn(targetWarning)
		}
	}
	configDir := path.Join(install.GetPath(), nginxModuleConfDir, nginxModuleEnabledConfDir)
	if err := os.MkdirAll(configDir, constant.DirPerm); err != nil {
		return err
	}
	snapshot, err := snapshotManagedNginxModuleConfigs(configDir)
	if err != nil {
		return err
	}
	desired := make(map[string][]byte)
	var combinedLoadDirectives strings.Builder
	sortedModules := append([]dto.NginxModule(nil), modules...)
	sort.SliceStable(sortedModules, func(i, j int) bool {
		if sortedModules[i].LoadOrder == sortedModules[j].LoadOrder {
			return sortedModules[i].Name < sortedModules[j].Name
		}
		return sortedModules[i].LoadOrder < sortedModules[j].LoadOrder
	})
	for _, module := range sortedModules {
		normalizeNginxModule(&module)
		if !module.Enable || module.BuildMode == nginxModuleBuildStatic {
			continue
		}
		build := findCurrentNginxModuleBuild(module, target)
		if build == nil || build.Status != nginxModuleStatusReady {
			build = findLatestNginxModuleBuild(module, target)
		}
		if build == nil || build.Status != nginxModuleStatusReady {
			continue
		}
		// Artifacts were already validated after the build; re-validating here
		// is a deliberate guard against on-disk tampering before writing
		// load_module directives.
		if err = validateNginxModuleArtifacts(install, build.Artifacts); err != nil {
			return fmt.Errorf("validate artifacts for dynamic module %s: %w", module.Name, err)
		}
		var content strings.Builder
		content.WriteString("# Managed by 1Panel. Manual changes will be overwritten.\n")
		for _, artifact := range build.Artifacts {
			content.WriteString("load_module ")
			content.WriteString(path.Join(nginxModuleContainerRoot, artifact.Path))
			content.WriteString(";\n")
			combinedLoadDirectives.WriteString("load_module ")
			combinedLoadDirectives.WriteString(path.Join(nginxModuleContainerRoot, artifact.Path))
			combinedLoadDirectives.WriteString(";\n")
		}
		fileName := fmt.Sprintf("%s%04d-%s.conf", nginxModuleConfigPrefix, module.LoadOrder, nginxModulePathName(module.Name))
		desired[fileName] = []byte(content.String())
	}
	if combinedLoadDirectives.Len() > 0 {
		if err = validateNginxModuleLoadConfig(install, target, "combined", combinedLoadDirectives.String()); err != nil {
			return fmt.Errorf("validate combined dynamic module configuration: %w", err)
		}
	}

	// Runtime directives live in http.d because load_module is main-context
	// while directives such as "brotli on" are http-context. Both sets are
	// written before nginx -t runs, so nginx only ever observes the final,
	// consistent state; on failure both are rolled back together.
	httpConfigDir := nginxHTTPConfigDir(install)
	httpSupported := nginxHTTPConfigSupported(install)
	var httpSnapshot nginxModuleConfigSnapshot
	if httpSupported {
		if httpSnapshot, err = snapshotManagedNginxHTTPConfigs(httpConfigDir); err != nil {
			return err
		}
	}
	restore := func() {
		_ = applyManagedNginxModuleConfigs(configDir, snapshot)
		if httpSupported {
			_ = applyManagedNginxHTTPConfigs(httpConfigDir, httpSnapshot)
		}
	}

	if err = applyManagedNginxModuleConfigs(configDir, desired); err != nil {
		restore()
		return err
	}
	if httpSupported {
		desiredHTTP := desiredNginxModuleRuntimeConfigs(install, modules, target)
		if err = applyManagedNginxHTTPConfigs(httpConfigDir, desiredHTTP); err != nil {
			restore()
			return err
		}
	}
	if !reload {
		return nil
	}
	status, statusErr := checkContainerStatus(install.ContainerName)
	if statusErr != nil || status != "running" {
		return nil
	}
	if err = opNginx(install.ContainerName, constant.NginxCheck); err != nil {
		restore()
		return err
	}
	if err = opNginx(install.ContainerName, constant.NginxReload); err != nil {
		restore()
		return err
	}
	return nil
}

func snapshotManagedNginxModuleConfigs(configDir string) (nginxModuleConfigSnapshot, error) {
	snapshot := make(nginxModuleConfigSnapshot)
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), nginxModuleConfigPrefix) {
			continue
		}
		content, readErr := os.ReadFile(path.Join(configDir, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		snapshot[entry.Name()] = content
	}
	return snapshot, nil
}

func applyManagedNginxModuleConfigs(configDir string, desired map[string][]byte) error {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}
	for fileName, content := range desired {
		tmpPath := path.Join(configDir, "."+fileName+".tmp")
		if err = os.WriteFile(tmpPath, content, constant.FilePerm); err != nil {
			return err
		}
		if err = os.Rename(tmpPath, path.Join(configDir, fileName)); err != nil {
			return err
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), nginxModuleConfigPrefix) {
			continue
		}
		if _, ok := desired[entry.Name()]; !ok {
			if err = os.Remove(path.Join(configDir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

// hasEnabledStaticNginxModules reports whether a full image rebuild is needed.
//
// Module state is the only input on purpose. RESTY_CONFIG_OPTIONS_MORE in .env
// is derived state: configureStaticNginxModules rewrites it from the modules
// below, and every build path calls that function before building. Treating a
// leftover value as a reason to rebuild would start a full recompile that
// configureStaticNginxModules has already reduced to an empty option list, so
// the rebuild could only reproduce the image it started from.
func hasEnabledStaticNginxModules(modules []dto.NginxModule) bool {
	for _, module := range modules {
		normalizeNginxModule(&module)
		if module.Enable && module.BuildMode == nginxModuleBuildStatic {
			return true
		}
	}
	return false
}

func configureStaticNginxModules(install model.AppInstall, modules []dto.NginxModule, mirror string) error {
	buildPath := path.Join(install.GetPath(), nginxModuleBuildDir)
	var params, packages []string
	preScriptPath := path.Join(buildPath, nginxModuleTmpDir, nginxModuleStaticPreScript)
	preScript, err := os.OpenFile(preScriptPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return err
	}
	for _, module := range modules {
		normalizeNginxModule(&module)
		if !module.Enable || module.BuildMode != nginxModuleBuildStatic {
			continue
		}
		if _, err = preScript.WriteString(module.Script + "\n"); err != nil {
			_ = preScript.Close()
			return err
		}
		params = append(params, module.Params)
		packages = append(packages, module.Packages...)
	}
	if err = preScript.Close(); err != nil {
		return err
	}
	envs, err := gotenv.Read(install.GetEnvPath())
	if err != nil {
		return err
	}
	if mirror != "" {
		envs["CONTAINER_PACKAGE_URL"] = mirror
	}
	envs["RESTY_CONFIG_OPTIONS_MORE"] = strings.Join(compactStrings(params), " ")
	envs["RESTY_ADD_PACKAGE_BUILDDEPS"] = strings.Join(compactStrings(packages), " ")
	return gotenv.Write(envs, install.GetEnvPath())
}

func executeStaticNginxModuleBuild(install model.AppInstall, modules []dto.NginxModule, mirror string, force bool, parentTask *task.Task) error {
	if err := configureStaticNginxModules(install, modules, mirror); err != nil {
		return err
	}
	commandMgr := cmd.NewCommandMgr(cmd.WithTask(*parentTask), cmd.WithTimeout(120*time.Minute))
	if err := commandMgr.Run("docker", "compose", "-f", install.GetComposePath(), "build"); err != nil {
		return err
	}
	previousModules := cloneNginxModules(modules)
	// A rebuilt runtime changes the target ABI, so every enabled dynamic module
	// must be rebuilt even when the user selected only one module.
	modules, err := buildDynamicNginxModules(install, modules, nil, force, mirror, "", parentTask)
	if err != nil {
		return err
	}
	if err = commitNginxModuleBuilds(install, previousModules, modules, false, ""); err != nil {
		return err
	}
	if _, err = compose.DownAndUp(install.GetComposePath()); err != nil {
		return err
	}
	return commitStaticNginxModuleBuilds(install, "", parentTask)
}

func executeNginxModuleBuild(install model.AppInstall, reqModules []string, force bool, mirror string, parentTask *task.Task, reload bool) error {
	modules, err := loadNginxModules(install)
	if err != nil {
		return err
	}
	// Only the module list decides this. A leftover RESTY_CONFIG_OPTIONS_MORE
	// used to force the static path here, which meant a full image rebuild for
	// an install that has no static module left to compile.
	staticBuild := hasEnabledStaticNginxModules(modules)
	if !staticBuild && hasDynamicNginxModuleBuildTask(modules, reqModules) && !nginxModuleDynamicSupported(install) {
		// The install predates dynamic modules. Compile the same modules into
		// the image instead of refusing the build, which is how these versions
		// have always produced modules.
		if !nginxModuleStaticSupported(install) {
			return buserr.New("ErrModuleBuildUnsupported")
		}
		if modules, err = convertNginxModulesToStatic(modules, reqModules); err != nil {
			return err
		}
		staticBuild = true
	}
	if staticBuild {
		return executeStaticNginxModuleBuild(install, modules, mirror, force, parentTask)
	}
	previousModules := cloneNginxModules(modules)
	modules, err = buildDynamicNginxModules(install, modules, reqModules, force, mirror, "", parentTask)
	if err != nil {
		return err
	}
	return commitNginxModuleBuilds(install, previousModules, modules, reload, "")
}

func removeNginxModuleArtifacts(install model.AppInstall, module dto.NginxModule) error {
	for _, build := range module.Builds {
		moduleDir := path.Join(install.GetPath(), nginxModuleModulesDir, build.Target.Key, nginxModulePathName(module.Name))
		if err := files.NewFileOp().DeleteDir(moduleDir); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadNginxModules(install model.AppInstall) ([]dto.NginxModule, error) {
	return loadNginxModulesWithCatalog(install, "")
}

func loadNginxModulesWithCatalog(install model.AppInstall, catalogPath string) ([]dto.NginxModule, error) {
	modulePath := path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleStoreFile)
	states, err := readNginxModuleStateFile(modulePath)
	if err != nil {
		return nil, err
	}
	if catalogPath == "" {
		catalogPath = path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleCatalogFile)
	}
	modules, err := readNginxModuleFile(catalogPath)
	if err != nil {
		return nil, err
	}
	moduleIndexes := make(map[string]int, len(modules))
	for i := range modules {
		modules[i].Custom = false
		modules[i].Enable = false
		modules[i].Builds = nil
		modules[i].LastError = ""
		if _, exists := moduleIndexes[modules[i].Name]; exists {
			return nil, fmt.Errorf("duplicate OpenResty module catalog name %q", modules[i].Name)
		}
		moduleIndexes[modules[i].Name] = i
	}
	stateNames := make(map[string]struct{}, len(states))
	for _, state := range states {
		if _, exists := stateNames[state.Name]; exists {
			return nil, fmt.Errorf("duplicate OpenResty module state name %q", state.Name)
		}
		stateNames[state.Name] = struct{}{}
		if index, ok := moduleIndexes[state.Name]; ok {
			if state.Custom {
				return nil, fmt.Errorf("custom OpenResty module %s conflicts with the module catalog", state.Name)
			}
			modules[index].Enable = state.Enable
			modules[index].Builds = state.Builds
			modules[index].LastError = state.LastError
			continue
		}
		if !state.Custom {
			return nil, fmt.Errorf("OpenResty module state %s is missing from the module catalog", state.Name)
		}
		modules = append(modules, dto.NginxModule{
			Name: state.Name, Custom: true, Script: state.Script, Packages: state.Packages, Params: state.Params,
			Enable: state.Enable, BuildMode: state.BuildMode, Provider: state.Provider, LoadOrder: state.LoadOrder,
			Builds: state.Builds, LastError: state.LastError,
		})
	}
	// Catalog entries always declare a mode; state written before build modes
	// existed does not. Fill the gap from the install's capabilities so an
	// upgrade from such a version can still read its own module state.
	fallbackMode := ""
	for i := range modules {
		if modules[i].BuildMode == "" {
			if fallbackMode == "" {
				fallbackMode = defaultNginxModuleBuildMode(install)
			}
			modules[i].BuildMode = fallbackMode
		}
		if err = validateNginxModuleBuildMode(modules[i]); err != nil {
			return nil, err
		}
		normalizeNginxModule(&modules[i])
	}
	return modules, nil
}

func readNginxModuleStateFile(filePath string) ([]nginxModuleState, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []nginxModuleState{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return []nginxModuleState{}, nil
	}
	var states []nginxModuleState
	if err = json.Unmarshal(content, &states); err != nil {
		return nil, fmt.Errorf("parse OpenResty module state %s: %w", filePath, err)
	}
	return states, nil
}

func readNginxModuleFile(filePath string) ([]dto.NginxModule, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.NginxModule{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return []dto.NginxModule{}, nil
	}
	var modules []dto.NginxModule
	if err = json.Unmarshal(content, &modules); err != nil {
		return nil, fmt.Errorf("parse OpenResty module configuration %s: %w", filePath, err)
	}
	return modules, nil
}

func saveNginxModules(install model.AppInstall, modules []dto.NginxModule) error {
	return saveNginxModulesWithCatalog(install, modules, "")
}

func saveNginxModulesWithCatalog(install model.AppInstall, modules []dto.NginxModule, catalogPath string) error {
	if catalogPath == "" {
		catalogPath = path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleCatalogFile)
	}
	catalog, err := readNginxModuleFile(catalogPath)
	if err != nil {
		return err
	}
	catalogNames := make(map[string]struct{}, len(catalog))
	for _, module := range catalog {
		if _, exists := catalogNames[module.Name]; exists {
			return fmt.Errorf("duplicate OpenResty module catalog name %q", module.Name)
		}
		catalogNames[module.Name] = struct{}{}
	}
	states := make([]nginxModuleState, 0, len(modules))
	for i := range modules {
		module := &modules[i]
		if err = validateNginxModuleBuildMode(*module); err != nil {
			return err
		}
		normalizeNginxModule(module)
		if _, builtin := catalogNames[module.Name]; builtin {
			if module.Custom {
				return fmt.Errorf("custom OpenResty module %s conflicts with the module catalog", module.Name)
			}
			if !module.Enable && len(module.Builds) == 0 && module.LastError == "" {
				continue
			}
			states = append(states, nginxModuleState{
				Name: module.Name, Enable: module.Enable, Builds: module.Builds, LastError: module.LastError,
			})
			continue
		}
		if !module.Custom {
			return fmt.Errorf("OpenResty module %s is missing from the module catalog", module.Name)
		}
		states = append(states, nginxModuleState{
			Name: module.Name, Custom: true, Script: module.Script, Packages: module.Packages, Params: module.Params,
			Enable: module.Enable, BuildMode: module.BuildMode, Provider: module.Provider, LoadOrder: module.LoadOrder,
			Builds: module.Builds, LastError: module.LastError,
		})
	}
	content, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	modulePath := path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleStoreFile)
	if err = os.MkdirAll(path.Dir(modulePath), constant.DirPerm); err != nil {
		return err
	}
	tmpPath := modulePath + ".tmp"
	if err := os.WriteFile(tmpPath, content, constant.FilePerm); err != nil {
		return err
	}
	return os.Rename(tmpPath, modulePath)
}

func activateNginxModuleCatalog(pendingPath, activePath string) error {
	return os.Rename(pendingPath, activePath)
}

func activateNginxModuleCatalogAndCommit(pendingPath, activePath string, commit func() error) error {
	previous, err := os.ReadFile(activePath)
	if err != nil {
		return err
	}
	if err = activateNginxModuleCatalog(pendingPath, activePath); err != nil {
		return err
	}
	if err = commit(); err == nil {
		return nil
	}
	if restoreErr := writeNginxModuleCatalog(activePath, previous); restoreErr != nil {
		return fmt.Errorf("%w; restore previous OpenResty module catalog: %v", err, restoreErr)
	}
	return err
}

func stageNginxModuleCatalog(sourcePath, pendingPath string) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return writeNginxModuleCatalog(pendingPath, content)
}

func writeNginxModuleCatalog(targetPath string, content []byte) error {
	tmpPath := targetPath + ".tmp"
	defer func() { _ = os.Remove(tmpPath) }()
	if err := os.WriteFile(tmpPath, content, constant.FilePerm); err != nil {
		return err
	}
	return os.Rename(tmpPath, targetPath)
}

func normalizeNginxModule(module *dto.NginxModule) {
	if module.Provider == "" {
		module.Provider = nginxModuleProviderLocal
	}
	module.Packages = compactStrings(module.Packages)
}

func validateNginxModuleBuildMode(module dto.NginxModule) error {
	if module.BuildMode != nginxModuleBuildDynamic && module.BuildMode != nginxModuleBuildStatic {
		return fmt.Errorf("OpenResty module %s has invalid build mode %q", module.Name, module.BuildMode)
	}
	return nil
}

func compactStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func nginxModuleStaticBuildErrorHint(module dto.NginxModule) string {
	if module.Custom {
		return nginxModuleStaticBuildHint
	}
	return ""
}

// nginxModuleNeedsDynamicBuild mirrors the dynamic-build filter of
// buildDynamicNginxModules. It normalizes a copy so prescan callers never
// mutate the stored entities.
func nginxModuleNeedsDynamicBuild(module dto.NginxModule, selectedNames map[string]struct{}) bool {
	normalizeNginxModule(&module)
	if module.BuildMode != nginxModuleBuildDynamic {
		return false
	}
	if len(selectedNames) > 0 {
		_, ok := selectedNames[module.Name]
		return ok
	}
	return module.Enable
}

func hasDynamicNginxModuleBuildTask(modules []dto.NginxModule, selected []string) bool {
	selectedNames := make(map[string]struct{}, len(selected))
	for _, name := range selected {
		selectedNames[name] = struct{}{}
	}
	for _, module := range modules {
		if nginxModuleNeedsDynamicBuild(module, selectedNames) {
			return true
		}
	}
	return false
}

func findLatestNginxModuleBuild(module dto.NginxModule, target dto.NginxModuleTarget) *dto.NginxModuleBuild {
	var latest *dto.NginxModuleBuild
	for i := range module.Builds {
		if module.Builds[i].BuildMode != module.BuildMode ||
			module.Builds[i].Target.Key != target.Key ||
			module.Builds[i].Status != nginxModuleStatusReady {
			continue
		}
		if latest == nil || module.Builds[i].BuiltAt.After(latest.BuiltAt) {
			latest = &module.Builds[i]
		}
	}
	return latest
}

func findCurrentNginxModuleBuild(module dto.NginxModule, target dto.NginxModuleTarget) *dto.NginxModuleBuild {
	params, err := nginxModuleBuildParams(module)
	if err != nil {
		return nil
	}
	buildHash, err := nginxModuleBuildHash(module, target, params)
	if err != nil {
		return nil
	}
	for i := range module.Builds {
		if module.Builds[i].BuildMode == module.BuildMode &&
			module.Builds[i].Target.Key == target.Key &&
			module.Builds[i].Hash == buildHash {
			return &module.Builds[i]
		}
	}
	return nil
}

func nginxModuleBuildParams(module dto.NginxModule) (string, error) {
	if module.BuildMode == nginxModuleBuildStatic {
		params := strings.TrimSpace(module.Params)
		if params == "" {
			return "", errors.New("static module parameters are empty")
		}
		return params, nil
	}
	return normalizeDynamicModuleParams(module.Params)
}

func recordStaticNginxModuleBuilds(modules []dto.NginxModule, target dto.NginxModuleTarget) ([]dto.NginxModule, error) {
	for i := range modules {
		module := &modules[i]
		normalizeNginxModule(module)
		if !module.Enable || module.BuildMode != nginxModuleBuildStatic {
			continue
		}
		params, err := nginxModuleBuildParams(*module)
		if err != nil {
			return modules, err
		}
		buildHash, err := nginxModuleBuildHash(*module, target, params)
		if err != nil {
			return modules, err
		}
		upsertNginxModuleBuild(module, dto.NginxModuleBuild{
			Provider:  nginxModuleProviderLocal,
			BuildMode: nginxModuleBuildStatic,
			Status:    nginxModuleStatusReady,
			Hash:      buildHash,
			Target:    target,
			BuiltAt:   time.Now(),
		})
		module.LastError = ""
	}
	return modules, nil
}

func commitStaticNginxModuleBuilds(install model.AppInstall, catalogPath string, parentTask *task.Task) error {
	modules, err := loadNginxModulesWithCatalog(install, catalogPath)
	if err != nil || !hasEnabledStaticNginxModules(modules) {
		return err
	}
	status, err := checkContainerStatus(install.ContainerName)
	if err != nil {
		return err
	}
	if status != "running" {
		return fmt.Errorf("OpenResty container %s is not running after static module build", install.ContainerName)
	}
	if err = opNginx(install.ContainerName, constant.NginxCheck); err != nil {
		return err
	}
	target, targetWarning, err := resolveNginxRuntimeTarget(install)
	if err != nil {
		return err
	}
	if targetWarning != "" && parentTask != nil {
		parentTask.Logf("WARNING: %s", targetWarning)
	}
	modules, err = recordStaticNginxModuleBuilds(modules, target)
	if err != nil {
		return err
	}
	return saveNginxModulesWithCatalog(install, modules, catalogPath)
}

func upsertNginxModuleBuild(module *dto.NginxModule, build dto.NginxModuleBuild) {
	for i := range module.Builds {
		if module.Builds[i].Target.Key == build.Target.Key && module.Builds[i].Hash == build.Hash {
			module.Builds[i] = build
			return
		}
	}
	module.Builds = append(module.Builds, build)
}

func cloneNginxModules(modules []dto.NginxModule) []dto.NginxModule {
	cloned := make([]dto.NginxModule, len(modules))
	copy(cloned, modules)
	for i := range cloned {
		cloned[i].Packages = append([]string(nil), modules[i].Packages...)
		cloned[i].Builds = make([]dto.NginxModuleBuild, len(modules[i].Builds))
		copy(cloned[i].Builds, modules[i].Builds)
		for j := range cloned[i].Builds {
			cloned[i].Builds[j].Artifacts = append([]dto.NginxModuleArtifact(nil), modules[i].Builds[j].Artifacts...)
		}
	}
	return cloned
}

func recordNginxModuleBuildFailure(originalModules []dto.NginxModule, moduleName string, build dto.NginxModuleBuild, previousBuild *dto.NginxModuleBuild) []dto.NginxModule {
	failedModules := cloneNginxModules(originalModules)
	for i := range failedModules {
		if failedModules[i].Name != moduleName {
			continue
		}
		failedModules[i].LastError = build.Error
		if previousBuild == nil || previousBuild.Status != nginxModuleStatusReady {
			upsertNginxModuleBuild(&failedModules[i], build)
		}
		break
	}
	return failedModules
}

func recordNginxModuleActivationFailure(previousModules, candidateModules []dto.NginxModule, activationErr error) []dto.NginxModule {
	rollbackModules := cloneNginxModules(previousModules)
	candidates := make(map[string]dto.NginxModule, len(candidateModules))
	for _, module := range candidateModules {
		candidates[module.Name] = module
	}
	for i := range rollbackModules {
		candidate, ok := candidates[rollbackModules[i].Name]
		if !ok || !candidate.Enable || candidate.BuildMode == nginxModuleBuildStatic {
			continue
		}
		rollbackModules[i].LastError = activationErr.Error()
	}
	return rollbackModules
}

func normalizeDynamicModuleParams(params string) (string, error) {
	params = strings.TrimSpace(params)
	params = strings.ReplaceAll(params, "--add-module=", "--add-dynamic-module=")
	if !strings.Contains(params, "--add-dynamic-module=") && !strings.Contains(params, "=dynamic") {
		return "", errors.New("module does not declare a dynamic configure option")
	}
	if _, err := parseDynamicModuleParams(params); err != nil {
		return "", err
	}
	return params, nil
}

// normalizeStaticModuleParams is the inverse of normalizeDynamicModuleParams:
// a module compiled into the image uses --add-module and the plain form of
// nginx's built-in switches.
func normalizeStaticModuleParams(params string) string {
	params = strings.TrimSpace(params)
	params = strings.ReplaceAll(params, "--add-dynamic-module=", "--add-module=")
	return strings.ReplaceAll(params, "=dynamic", "")
}

// convertNginxModulesToStatic retargets the modules a build was asked to
// produce so they are compiled into the image. It is used when the install
// has no dynamic builder, where this is the only way to produce a module.
//
// The change is confined to the returned slice: it drives one build and is
// never persisted, so the catalog's declared mode stays authoritative and the
// modules go back to dynamic once the install gains a builder.
func convertNginxModulesToStatic(modules []dto.NginxModule, selected []string) ([]dto.NginxModule, error) {
	selectedNames := make(map[string]struct{}, len(selected))
	for _, name := range selected {
		selectedNames[name] = struct{}{}
	}
	converted := cloneNginxModules(modules)
	for i := range converted {
		if !nginxModuleNeedsDynamicBuild(converted[i], selectedNames) {
			continue
		}
		params := normalizeStaticModuleParams(converted[i].Params)
		if params == "" {
			return nil, fmt.Errorf("OpenResty module %s has no configure options to compile into the image", converted[i].Name)
		}
		converted[i].BuildMode = nginxModuleBuildStatic
		converted[i].Params = params
		converted[i].Enable = true
	}
	return converted, nil
}

func parseDynamicModuleParams(params string) ([]string, error) {
	// shellwords silently stops at unquoted shell metacharacters instead of
	// reporting them, so reject them on the raw input before parsing.
	if strings.ContainsAny(params, "\x00\r\n;&|<>") {
		return nil, errors.New("dynamic module parameters contain unsupported characters")
	}
	parser := shellwords.NewParser()
	parser.ParseBacktick = false
	parser.ParseEnv = false
	args, err := parser.Parse(params)
	if err != nil {
		return nil, fmt.Errorf("parse dynamic module parameters: %w", err)
	}
	if len(args) == 0 {
		return nil, errors.New("dynamic module parameters are empty")
	}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			return nil, fmt.Errorf("unsupported configure argument %q", arg)
		}
		if strings.ContainsAny(arg, "\x00\r\n;&|<>") {
			return nil, fmt.Errorf("configure argument %q contains unsupported characters", arg)
		}
	}
	return args, nil
}

func validateNginxModulePackages(packages []string) error {
	for _, item := range packages {
		if !re.IsValidNginxModulePackage(item) {
			return fmt.Errorf("invalid build package %q", item)
		}
	}
	return nil
}

func nginxModuleBuildHash(module dto.NginxModule, target dto.NginxModuleTarget, params string) (string, error) {
	payload := struct {
		Name      string
		Script    string
		Packages  []string
		Params    string
		BuildMode string
		TargetKey string
		Provider  string
	}{module.Name, module.Script, module.Packages, params, module.BuildMode, target.Key, module.Provider}
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func getNginxModuleProvider(name string) (nginxModuleArtifactProvider, error) {
	switch name {
	case "", nginxModuleProviderLocal:
		return localNginxModuleProvider{}, nil
	case "prebuilt":
		return nil, errors.New("prebuilt module provider is not configured; use the local provider")
	default:
		return nil, fmt.Errorf("unknown module artifact provider %q", name)
	}
}

func nginxModuleOutputDirectories(install model.AppInstall, modules []dto.NginxModule) map[string]struct{} {
	result := make(map[string]struct{})
	modulesRoot := filepath.Clean(path.Join(install.GetPath(), nginxModuleModulesDir))
	for _, module := range modules {
		for _, build := range module.Builds {
			for _, artifact := range build.Artifacts {
				if !re.IsValidNginxModuleArtifact(artifact.Path) {
					continue
				}
				artifactFullPath := filepath.Clean(filepath.Join(modulesRoot, filepath.FromSlash(artifact.Path)))
				if !strings.HasPrefix(artifactFullPath, modulesRoot+string(os.PathSeparator)) {
					continue
				}
				outputDir := filepath.Dir(artifactFullPath)
				if outputDir != modulesRoot {
					result[outputDir] = struct{}{}
				}
			}
		}
	}
	return result
}

func removeNginxModuleOutputsNotReferenced(install model.AppInstall, fromModules, referencedModules []dto.NginxModule) {
	referenced := nginxModuleOutputDirectories(install, referencedModules)
	for outputDir := range nginxModuleOutputDirectories(install, fromModules) {
		if _, ok := referenced[outputDir]; !ok {
			_ = os.RemoveAll(outputDir)
		}
	}
}

func sanitizeModulePathPart(value string) string {
	var result strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '.' || char == '-' || char == '_' {
			result.WriteRune(char)
		} else {
			result.WriteByte('-')
		}
	}
	return strings.Trim(result.String(), "-")
}

func nginxModulePathName(value string) string {
	base := sanitizeModulePathPart(value)
	if base == "" {
		base = "module"
	}
	if len(base) > 48 {
		base = base[:48]
	}
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%s-%s", base, hex.EncodeToString(sum[:4]))
}
