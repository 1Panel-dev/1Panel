package systemctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"go4.org/syncutil/singleflight"
	"golang.org/x/sync/errgroup"
)

var (
	aliasFile      string
	serviceAliases sync.Map
	saveTimer      *time.Timer
	saveMutex      sync.Mutex
	afterSaveTime  = 20 * time.Second
)
var (
	ErrServiceNotFound  = errors.New("service not found")
	ErrDiscoveryTimeout = errors.New("service discovery timeout for: %w")
	ErrServiceDiscovery = errors.New("service discovery failed for: %w")
	ErrNoValidService   = errors.New("no valid service found for: %w")
)

func loadPredefinedAliases() map[string][]string {
	return map[string][]string{
		"clam":       {"clamav-daemon.service", "clamd@scan.service", "clamd"},
		"freshclam":  {"clamav-freshclam.service", "freshclam.service"},
		"fail2ban":   {"fail2ban.service", "fail2ban"},
		"supervisor": {"supervisord.service", "supervisor.service", "supervisord", "supervisor"},
		"ssh":        {"sshd.service", "ssh.service", "sshd", "ssh"},
		"1panel":     {"1panel.service", "1paneld"},
		"docker":     {"docker.service", "dockerd"},
	}
}

func InitializeServiceDiscovery() {
	svcName := loadAliasesFromConfig()
	if len(svcName) > 0 {
		RegisterServiceAliases(svcName)
	}
}

func RegisterServiceAliases(aliases map[string][]string) {
	for key, values := range aliases {
		existing, loaded := serviceAliases.LoadOrStore(key, values)
		if loaded {
			merged := append(existing.([]string), values...)
			serviceAliases.Store(key, merged)
		}
	}
}

func loadAliasesFromConfig() map[string][]string {
	data, err := os.ReadFile(aliasFile)
	if err != nil {
		return nil
	}
	var rawAliases map[string][]string
	json.Unmarshal(data, &rawAliases)

	validAliases := make(map[string][]string)
	for key, aliases := range rawAliases {
		valid := []string{}
		for _, alias := range aliases {
			confirmed, _ := confirmServiceExists(alias)
			if confirmed {
				valid = append(valid, alias)
			}
		}
		if len(valid) > 0 {
			validAliases[key] = valid
		}
	}
	return validAliases
}

func cleanupKeywordAliases(keyword string) {
	serviceAliases.Range(func(k, v interface{}) bool {
		if k.(string) != keyword {
			return true
		}
		aliases := v.([]string)
		valid := make([]string, 0)
		for _, alias := range aliases {
			confirmed, _ := confirmServiceExists(alias)
			if confirmed {
				valid = append(valid, alias)
			}
		}
		if len(valid) == 0 {
			serviceAliases.Delete(k)
			serviceExistenceCache.Delete(k)
		} else {
			serviceAliases.Store(k, valid)
		}
		return true
	})
	go scheduleSave()
}

func smartServiceName(keyword string) (string, error) {
	mgr := GetGlobalManager()
	processedName := handleServiceNaming(mgr, keyword)

	confirmed, _ := confirmServiceExists(processedName)
	if confirmed {
		updateAliases(keyword, processedName)
		return processedName, nil
	}

	candidates := append([]string{processedName}, getAliases(keyword)...)
	if name, err := validateCandidatesConcurrently(candidates); err == nil {
		updateAliases(keyword, name)
		return name, nil
	}

	discoveredName, err := discoverAndSelectService(keyword)
	if err != nil {
		cleanupKeywordAliases(keyword)
		return "", ErrServiceNotFound
	}
	updateAliases(keyword, discoveredName)
	return discoveredName, nil
}
func handleServiceNaming(mgr ServiceManager, keyword string) string {
	keyword = strings.ToLower(keyword)
	// 处理 .service.socket 后缀
	if strings.HasSuffix(keyword, ".service.socket") {
		keyword = strings.TrimSuffix(keyword, ".service.socket") + ".socket"
	}
	if mgr.Name() != "systemd" {
		keyword = strings.TrimSuffix(keyword, ".service")
		return keyword
	}
	// 自动补全 .service 后缀
	if !strings.HasSuffix(keyword, ".service") &&
		!strings.HasSuffix(keyword, ".socket") {
		keyword += ".service"
	}
	return keyword
}
func validateCandidatesConcurrently(candidates []string) (string, error) {
	var (
		g     errgroup.Group
		found = make(chan string, 1) // 缓冲确保首个结果不阻塞
	)

	// 启动并发检查
	for _, candidate := range candidates {
		cand := candidate // 避免闭包循环引用
		g.Go(func() error {
			confirmed, _ := confirmServiceExists(cand)
			if confirmed {
				select {
				case found <- cand: // 发送首个成功结果
				default: // 如果已有结果，忽略后续
				}
				return nil
			}
			return ErrServiceNotFound
		})
	}

	// 处理结果
	resultErr := make(chan error, 1)
	go func() {
		defer close(found)
		resultErr <- g.Wait()
	}()

	select {
	case name := <-found:
		return name, nil
	case <-time.After(1000 * time.Millisecond):
		return "", fmt.Errorf(ErrDiscoveryTimeout.Error(), candidates[0])
	case err := <-resultErr:
		if err != nil {
			return "", ErrServiceNotFound
		}
		return "", ErrServiceNotFound
	}
}
func discoverAndSelectService(keyword string) (string, error) {
	discovered, err := discoverServices(keyword)
	if err != nil {
		return "", ErrServiceNotFound
	}

	if len(discovered) == 0 {
		return "", ErrServiceNotFound
	}
	selected, err := selectBestMatch(keyword, discovered)
	if err != nil {
		return "", ErrServiceNotFound
	}

	confirmed, err := confirmServiceExists(selected)
	if err != nil {
		return "", fmt.Errorf("service existence check failed: %w", err)
	}
	if confirmed {
		return selected, nil
	}
	return "", ErrServiceNotFound
}
func selectBestMatch(keyword string, candidates []string) (string, error) {
	if len(candidates) == 0 {
		return "", ErrServiceNotFound
	}
	lowerKeyword := strings.ToLower(keyword)
	var exactMatch string
	var firstContainMatch string
	// 第一轮遍历：严格匹配完全一致的名称（不区分大小写）
	for _, name := range candidates {
		if strings.EqualFold(name, keyword) {
			exactMatch = name
			break // 完全匹配直接终止循环
		}
	}
	if exactMatch != "" {
		return exactMatch, nil
	}
	// 第二轮遍历：寻找首个包含关键字的名称（不区分大小写）
	for _, name := range candidates {
		if strings.Contains(strings.ToLower(name), lowerKeyword) {
			firstContainMatch = name
			global.LOG.Debugf("[%s] [keyword: %s] Found first contain match: %s", getManagerName(), keyword, firstContainMatch)
			break
		}
	}
	if firstContainMatch != "" {
		return firstContainMatch, nil
	}
	// 无任何匹配项时返回明确错误
	return "", fmt.Errorf("%w: %q (no exact or partial match)", ErrNoValidService, keyword)
}

type cacheItem struct {
	services []string
	expires  time.Time
	exists   bool
}

var (
	discoveryCache sync.Map
	discoveryGroup singleflight.Group
)

func discoverServices(keyword string) ([]string, error) {
	result, err := discoveryGroup.Do(keyword, func() (interface{}, error) {
		if cached, ok := discoveryCache.Load(keyword); ok {
			item := cached.(cacheItem)
			if time.Now().Before(item.expires) {
				return item.services, nil
			}
			discoveryCache.Delete(keyword)
		}
		manager := GetGlobalManager()
		results, err := manager.FindServices(keyword)

		if err != nil {
			global.LOG.Errorf("Find services failed for %s: %v", keyword, err)
			return nil, fmt.Errorf("%w: %q (%v)", ErrServiceDiscovery, keyword, err)
		} else {
			discoveryCache.Store(keyword, cacheItem{
				services: results,
				expires:  time.Now().Add(5 * time.Minute),
			})
		}
		return results, err
	})
	if err != nil {
		return nil, err
	}
	return result.([]string), nil
}

func updateAliases(keyword, alias string) {
	if keyword == alias {
		return
	}
	existing, _ := serviceAliases.LoadOrStore(keyword, []string{})
	aliases := existing.([]string)
	if contains(aliases, alias) {
		return
	}
	serviceAliases.Store(keyword, append(aliases, alias))
	go scheduleSave()
}

func scheduleSave() {
	saveMutex.Lock()
	defer saveMutex.Unlock()

	if saveTimer != nil {
		saveTimer.Stop()
	}

	dataSnapshot := make(map[string][]string)
	serviceAliases.Range(func(k, v interface{}) bool {
		dataSnapshot[k.(string)] = append([]string{}, v.([]string)...)
		return true
	})
	aliasFile = filepath.Join(constant.ResourceDir, "svcaliases.json")
	saveTimer = time.AfterFunc(afterSaveTime, func() {
		tmpFile := aliasFile + ".tmp"
		if err := saveAliasesToFile(dataSnapshot, tmpFile); err == nil {
			os.Rename(tmpFile, aliasFile)
		}
	})
}
func saveAliasesToFile(data map[string][]string, path string) error {
	fileData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}

	if err := os.WriteFile(path, fileData, 0644); err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}
	return nil
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var serviceExistenceCache sync.Map

func confirmServiceExists(serviceName string) (bool, error) {
	if val, ok := serviceExistenceCache.Load(serviceName); ok {
		if item, ok := val.(cacheItem); ok && time.Now().Before(item.expires) {
			return item.exists, nil
		}
		serviceExistenceCache.Delete(serviceName)
	}
	handler := NewServiceHandler(defaultServiceConfig(serviceName))
	isExist, err := handler.IsExists()
	if err != nil {
		return false, fmt.Errorf("check service existence failed: %w", err)
	}
	serviceExistenceCache.Store(serviceName, cacheItem{
		exists:  isExist.IsExists,
		expires: time.Now().Add(30 * time.Second),
	})
	return isExist.IsExists, nil
}

func getAliases(keyword string) []string {
	predefined := loadPredefinedAliases()[keyword]
	runtimeAliases, _ := serviceAliases.LoadOrStore(keyword, []string{})
	merged := make(map[string]struct{})
	for _, alias := range predefined {
		merged[alias] = struct{}{}
	}
	for _, alias := range runtimeAliases.([]string) {
		merged[alias] = struct{}{}
	}
	result := make([]string, 0, len(merged))
	for k := range merged {
		result = append(result, k)
	}
	return result
}

type ConfigOption struct {
	TailLines string
}

func ViewConfig(path string, opt ConfigOption) (string, error) {
	var cmd []string
	if opt.TailLines != "" && opt.TailLines != "0" {
		cmd = []string{"tail", "-n", opt.TailLines, path}
	} else {
		cmd = []string{"cat", path}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := executeCommand(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		// global.LOG.Errorf("View config command failed: %v", err)
		return "", fmt.Errorf("view config failed: %w", err)
	}
	return string(output), nil
}
