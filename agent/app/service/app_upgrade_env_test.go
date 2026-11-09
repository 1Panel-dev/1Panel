package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func TestRenderUpgradeEnvKeepsValuesExactly(t *testing.T) {
	re.Init()
	compose := []byte("services:\n  demo:\n    image: demo:1.0\n    environment:\n      PASSWORD: ${PANEL_DB_ROOT_PASSWORD}\n")
	for _, password := range []string{"0123", "+123456", "-01"} {
		install := &model.AppInstall{
			App: model.App{Key: "demo"},
			Env: `{"PANEL_DB_ROOT_PASSWORD":"` + password + `"}`,
		}
		content, err := renderUpgradeEnv(install, []byte("PANEL_DB_ROOT_PASSWORD=\""+password+"\"\n"))
		if err != nil {
			t.Fatal(err)
		}
		project, err := docker.GetComposeProject("demo", t.TempDir(), compose, content, false)
		if err != nil {
			t.Fatal(err)
		}
		got := ""
		if value := project.Services["demo"].Environment["PASSWORD"]; value != nil {
			got = *value
		}
		if got != password {
			t.Errorf("password %q rendered as %q, which the compose project reads as %q", password, content, got)
		}
	}
}
