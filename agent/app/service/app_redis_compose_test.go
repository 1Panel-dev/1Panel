package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const redisServiceForm = `{"formFields":[{"envKey":"PANEL_REDIS_HOST","type":"service","required":true}]}`

const affineComposeYAML = `services:
  affine:
    image: ghcr.io/toeverything/affine:0.27.4
    container_name: ${CONTAINER_NAME}
    restart: unless-stopped
    ports:
      - ${PANEL_APP_PORT_HTTP}:3010
    depends_on:
      affine_migration:
        condition: service_completed_successfully
    labels:
      createdBy: "Apps"
  affine_migration:
    image: ghcr.io/toeverything/affine:0.27.4
    container_name: ${CONTAINER_NAME}_migration_job
    command: ['sh', '-c', 'node ./scripts/self-host-predeploy.js']
    labels:
      createdBy: "Apps"
      skipStatusCheck: "true"
`

const affineComposeReversedYAML = `services:
  affine_migration:
    image: ghcr.io/toeverything/affine:0.27.4
    container_name: ${CONTAINER_NAME}_migration_job
    command: ['sh', '-c', 'node ./scripts/self-host-predeploy.js']
    labels:
      createdBy: "Apps"
      skipStatusCheck: "true"
  affine:
    image: ghcr.io/toeverything/affine:0.27.4
    container_name: ${CONTAINER_NAME}
    restart: unless-stopped
    ports:
      - ${PANEL_APP_PORT_HTTP}:3010
    depends_on:
      affine_migration:
        condition: service_completed_successfully
    labels:
      createdBy: "Apps"
`

func setupDatabaseTestEnv(t *testing.T) {
	t.Helper()
	db := iconSyncDB(t)
	if err := db.AutoMigrate(&model.Database{}); err != nil {
		t.Fatal(err)
	}
	oldKey := global.CONF.Base.EncryptKey
	global.CONF.Base.EncryptKey = "1234567890abcdef"
	t.Cleanup(func() {
		global.CONF.Base.EncryptKey = oldKey
	})
}

func TestGetServicesRemoteRedisConfig(t *testing.T) {
	setupDatabaseTestEnv(t)

	if err := databaseRepo.Create(context.Background(), &model.Database{
		Name:     "affine_redis",
		Type:     constant.AppRedis,
		From:     constant.AppResourceRemote,
		Address:  "192.168.31.165",
		Port:     6380,
		Password: "redis-secret-password",
	}); err != nil {
		t.Fatal(err)
	}
	if err := databaseRepo.Create(context.Background(), &model.Database{
		Name:     "affine_pg",
		Type:     constant.AppPostgresql,
		From:     constant.AppResourceRemote,
		Address:  "192.168.31.165",
		Port:     5432,
		Username: "postgres",
		Password: "pg-secret-password",
	}); err != nil {
		t.Fatal(err)
	}

	svc := &AppInstallService{}
	redisServices, err := svc.GetServices(constant.AppRedis)
	if err != nil {
		t.Fatalf("GetServices(redis) returned error: %v", err)
	}
	if len(redisServices) != 1 {
		t.Fatalf("expected 1 redis service, got %d", len(redisServices))
	}
	redisConfig, ok := redisServices[0].Config.(map[string]string)
	if !ok {
		t.Fatalf("unexpected config type %T", redisServices[0].Config)
	}
	if got := redisConfig["PANEL_REDIS_ROOT_PASSWORD"]; got != "redis-secret-password" {
		t.Fatalf("PANEL_REDIS_ROOT_PASSWORD = %q, want %q", got, "redis-secret-password")
	}
	if _, exists := redisConfig["PANEL_DB_ROOT_PASSWORD"]; exists {
		t.Fatalf("remote redis config should not set PANEL_DB_ROOT_PASSWORD: %+v", redisServices[0].Config)
	}

	pgServices, err := svc.GetServices(constant.AppPostgresql)
	if err != nil {
		t.Fatalf("GetServices(postgresql) returned error: %v", err)
	}
	if len(pgServices) != 1 {
		t.Fatalf("expected 1 postgresql service, got %d", len(pgServices))
	}
	pgConfig, ok := pgServices[0].Config.(map[string]string)
	if !ok {
		t.Fatalf("unexpected config type %T", pgServices[0].Config)
	}
	if got := pgConfig["PANEL_DB_ROOT_PASSWORD"]; got != "pg-secret-password" {
		t.Fatalf("PANEL_DB_ROOT_PASSWORD = %q, want %q", got, "pg-secret-password")
	}
	if _, exists := pgConfig["PANEL_REDIS_ROOT_PASSWORD"]; exists {
		t.Fatalf("remote postgresql config should not set PANEL_REDIS_ROOT_PASSWORD: %+v", pgServices[0].Config)
	}
}

func TestResolveRedisParamsAndStatusCheck(t *testing.T) {
	setupDatabaseTestEnv(t)

	if err := databaseRepo.Create(context.Background(), &model.Database{
		Name:     "affine_redis",
		Type:     constant.AppRedis,
		From:     constant.AppResourceRemote,
		Address:  "192.168.31.165",
		Port:     6380,
		Password: "redis-secret-password",
	}); err != nil {
		t.Fatal(err)
	}
	if err := databaseRepo.Create(context.Background(), &model.Database{
		Name:     "other_pg",
		Type:     constant.AppPostgresql,
		From:     constant.AppResourceRemote,
		Address:  "10.0.0.10",
		Port:     5432,
		Password: "pg-password",
	}); err != nil {
		t.Fatal(err)
	}

	params := map[string]interface{}{
		"PANEL_REDIS_HOST":          "affine_redis",
		"PANEL_REDIS_ROOT_PASSWORD": "",
	}
	if err := resolveRedisParams(params, redisServiceForm, false); err != nil {
		t.Fatalf("remote redis status check failed: %v", err)
	}
	if params["PANEL_REDIS_HOST"] != "192.168.31.165" {
		t.Fatalf("PANEL_REDIS_HOST = %v, want 192.168.31.165", params["PANEL_REDIS_HOST"])
	}
	if params["PANEL_REDIS_PORT"] != uint(6380) {
		t.Fatalf("PANEL_REDIS_PORT = %v, want 6380", params["PANEL_REDIS_PORT"])
	}
	if params["PANEL_REDIS_HOST_NAME"] != "affine_redis" {
		t.Fatalf("PANEL_REDIS_HOST_NAME = %v, want affine_redis", params["PANEL_REDIS_HOST_NAME"])
	}
	if params["PANEL_REDIS_ROOT_PASSWORD"] != "redis-secret-password" {
		t.Fatalf("PANEL_REDIS_ROOT_PASSWORD = %v, want redis-secret-password", params["PANEL_REDIS_ROOT_PASSWORD"])
	}

	// Non-redis database name must not be resolved as Redis.
	nonRedisParams := map[string]interface{}{
		"PANEL_REDIS_HOST":          "other_pg",
		"PANEL_REDIS_ROOT_PASSWORD": "",
	}
	if err := resolveRedisParams(nonRedisParams, redisServiceForm, false); err == nil {
		t.Fatal("accepted PostgreSQL as Redis")
	}
	if nonRedisParams["PANEL_REDIS_HOST"] != "other_pg" {
		t.Fatalf("non-redis database was wrongly resolved: %v", nonRedisParams["PANEL_REDIS_HOST"])
	}

	// Local stopped Redis must fail running check.
	stoppedInstall := model.AppInstall{
		Name:        "local-redis",
		ServiceName: "local-redis",
		Status:      constant.StatusStopped,
	}
	if err := global.DB.Create(&stoppedInstall).Error; err != nil {
		t.Fatal(err)
	}
	if err := databaseRepo.Create(context.Background(), &model.Database{
		AppInstallID: stoppedInstall.ID,
		Name:         stoppedInstall.Name,
		Type:         constant.AppRedis,
		From:         constant.AppResourceLocal,
		Address:      stoppedInstall.ServiceName,
		Port:         6379,
	}); err != nil {
		t.Fatal(err)
	}
	localParams := map[string]interface{}{"PANEL_REDIS_HOST": "local-redis"}
	if err := resolveRedisParams(localParams, redisServiceForm, false); err == nil {
		t.Fatal("expected error for stopped local redis, got nil")
	}
}

func TestSelectPrimaryComposeServiceAndRestartPolicy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		compose string
	}{
		{name: "main_first", compose: affineComposeYAML},
		{name: "migration_first", compose: affineComposeReversedYAML},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				composeMap := make(map[string]interface{})
				if err := yaml.Unmarshal([]byte(tc.compose), &composeMap); err != nil {
					t.Fatal(err)
				}
				servicesMap := composeMap["services"].(map[string]interface{})
				primary := selectPrimaryComposeService(servicesMap, tc.compose, "affine")
				if primary != "affine" {
					t.Fatalf("iteration %d: selectPrimaryComposeService = %q, want %q", i, primary, "affine")
				}

				params := map[string]interface{}{}
				cfg := request.AppContainerConfig{
					Advanced:      true,
					RestartPolicy: "always",
				}
				if err := AddDockerComposeCommonParam(composeMap, primary, cfg, params, "affine"); err != nil {
					t.Fatalf("AddDockerComposeCommonParam error: %v", err)
				}
				affineSvc := servicesMap["affine"].(map[string]interface{})
				migrationSvc := servicesMap["affine_migration"].(map[string]interface{})
				if affineSvc["restart"] != "always" {
					t.Fatalf("affine restart = %v, want always", affineSvc["restart"])
				}
				if _, hasRestart := migrationSvc["restart"]; hasRestart {
					t.Fatalf("affine_migration should not receive restart policy, got %v", migrationSvc["restart"])
				}
			}
		})
	}

	// Defense-in-depth: even when targeting a one-shot migration service directly, do not inject restart.
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(affineComposeYAML), &composeMap); err != nil {
		t.Fatal(err)
	}
	if err := AddDockerComposeCommonParam(composeMap, "affine_migration", request.AppContainerConfig{
		RestartPolicy: "always",
	}, map[string]interface{}{}, "affine"); err != nil {
		t.Fatalf("AddDockerComposeCommonParam error: %v", err)
	}
	migrationSvc := composeMap["services"].(map[string]interface{})["affine_migration"].(map[string]interface{})
	if _, hasRestart := migrationSvc["restart"]; hasRestart {
		t.Fatalf("one-shot migration service received restart policy: %v", migrationSvc["restart"])
	}

	// getRestartPolicy should read the primary service deterministically.
	for i := 0; i < 20; i++ {
		if got := getRestartPolicy(affineComposeYAML, "affine", "affine"); got != "unless-stopped" {
			t.Fatalf("getRestartPolicy(affine) = %q, want unless-stopped", got)
		}
		if got := getRestartPolicy(affineComposeReversedYAML, "", "affine"); got != "unless-stopped" {
			t.Fatalf("getRestartPolicy(reversed, empty) = %q, want unless-stopped", got)
		}
	}
}

func TestRedisServiceSelectionAndLegacyUpgrade(t *testing.T) {
	setupDatabaseTestEnv(t)
	if err := databaseRepo.Create(context.Background(), &model.Database{
		Name: "shared-cache", Type: constant.AppRedis, From: constant.AppResourceRemote,
		Address: "redis.example", Port: 6380, Password: "registered-password",
	}); err != nil {
		t.Fatal(err)
	}

	selected := map[string]interface{}{"PANEL_REDIS_HOST": "deleted"}
	if err := resolveRedisParams(selected, redisServiceForm, false); err == nil {
		t.Fatal("accepted deleted Redis service")
	}
	textForm := strings.Replace(redisServiceForm, "service", "text", 1)
	literal := map[string]interface{}{"PANEL_REDIS_HOST": "shared-cache", "PANEL_REDIS_PORT": "6391"}
	if err := resolveRedisParams(literal, textForm, false); err != nil || literal["PANEL_REDIS_HOST"] != "shared-cache" || literal["PANEL_REDIS_PORT"] != "6391" {
		t.Fatalf("free-text host or port changed: %v, %v", literal, err)
	}
	for _, tc := range []struct{ name, host, port, password, form, wantHost, wantPort, wantPassword, wantAlias string }{
		{"old service", "shared-cache", "", "", redisServiceForm, "redis.example", "6380", "registered-password", "shared-cache"},
		{"explicit port and password", "shared-cache", "6385", "custom", redisServiceForm, "redis.example", "6385", "custom", "shared-cache"},
		{"external host", "external.example", "6390", "custom", redisServiceForm, "external.example", "6390", "custom", ""},
		{"text field matching database name", "shared-cache", "6390", "custom", textForm, "shared-cache", "6390", "custom", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			install := model.AppInstall{Env: `{"PANEL_REDIS_HOST":"shared-cache","PANEL_REDIS_ROOT_PASSWORD":""}`}
			live := []byte("PANEL_REDIS_HOST=" + tc.host + "\nPANEL_REDIS_PORT=" + tc.port + "\nPANEL_REDIS_ROOT_PASSWORD=" + tc.password + "\n")
			for i := 0; i < 2; i++ {
				content, err := renderUpgradeEnv(&install, live, tc.form)
				if err != nil {
					t.Fatal(err)
				}
				actual, err := godotenv.UnmarshalBytes(content)
				if err != nil {
					t.Fatal(err)
				}
				var saved map[string]interface{}
				if err := json.Unmarshal([]byte(install.Env), &saved); err != nil {
					t.Fatal(err)
				}
				if actual["PANEL_REDIS_HOST"] != tc.wantHost || actual["PANEL_REDIS_PORT"] != tc.wantPort || actual["PANEL_REDIS_HOST_NAME"] != tc.wantAlias || actual["PANEL_REDIS_ROOT_PASSWORD"] != tc.wantPassword || saved["PANEL_REDIS_HOST"] != tc.wantHost {
					t.Fatal("unexpected Redis connection in upgraded environment")
				}
				live = content
			}
		})
	}
}

func TestLegacyMigrationServiceOnEditAndUpgrade(t *testing.T) {
	re.Init()
	const job = "affine_migration"
	old := strings.Replace(affineComposeYAML, "  affine_migration:\n    image", "  affine_migration:\n    restart: always\n    image", 1)
	current := strings.Replace(affineComposeReversedYAML, "  affine_migration:\n    image", "  affine_migration:\n    restart: \"no\"\n    image", 1)
	original := model.AppInstall{App: model.App{Key: "affine"}, ServiceName: job, DockerCompose: old, Env: "{}", Name: "affine", ContainerName: "affine"}
	if got := getRestartPolicy(old, job, "affine"); got != "unless-stopped" {
		t.Fatalf("shown restart policy = %q", got)
	}
	for _, op := range []string{"edit", "upgrade", "custom upgrade"} {
		t.Run(op, func(t *testing.T) {
			install := original
			var err error
			switch op {
			case "edit":
				err = repairLegacyComposePrimary(&install)
			case "upgrade":
				install.DockerCompose, err = renderUpgradeCompose(&install, model.AppDetail{DockerCompose: current}, "")
			case "custom upgrade":
				install.DockerCompose, err = renderUpgradeCompose(&install, model.AppDetail{}, old)
			}
			if err != nil {
				t.Fatal(err)
			}
			if install.ServiceName != "affine" {
				t.Fatalf("saved service = %q", install.ServiceName)
			}
			project, err := docker.GetComposeProject("affine", t.TempDir(), []byte(install.DockerCompose), []byte("CONTAINER_NAME=affine\nPANEL_APP_PORT_HTTP=3010\nCPUS=1\nMEMORY_LIMIT=1G\n"), false)
			if err != nil {
				t.Fatal(err)
			}
			if got := project.Services["affine"].Restart; got != "unless-stopped" {
				t.Fatalf("main restart policy = %q", got)
			}
			if got := project.Services[job].Restart; got != "" && got != "no" {
				t.Fatalf("migration restart policy = %q", got)
			}
			if op == "upgrade" && project.Services[job].Restart != "no" {
				t.Fatal("upgrade overwrote the template's migration policy")
			}
		})
	}
}
