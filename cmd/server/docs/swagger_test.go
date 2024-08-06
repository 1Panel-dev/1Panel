package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGenerateSwaggerDoc(t *testing.T) {
	workDir := "/Users/slooop/Documents/mycode/1Panel"
	swagBin := "/Users/slooop/.gvm/pkgsets/go1.22.4/global/bin/swag"

	cmd1 := exec.Command(swagBin, "init", "-o", workDir+"/cmd/server/docs/docs_agent", "-d", workDir+"/agent", "-g", "./cmd/server/main.go")
	cmd1.Dir = workDir
	std1, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Printf("generate swagger doc of agent failed, std1: %v, err: %v", string(std1), err)
		return
	}
	cmd2 := exec.Command(swagBin, "init", "-o", workDir+"/cmd/server/docs/docs_core", "-d", workDir+"/core", "-g", "../cmd/server/main.go")
	cmd2.Dir = workDir
	std2, err := cmd2.CombinedOutput()
	if err != nil {
		fmt.Printf("generate swagger doc of core failed, std1: %v, err: %v", string(std2), err)
		return
	}

	agentJson := workDir + "/cmd/server/docs/docs_agent/swagger.json"
	agentFile, err := os.ReadFile(agentJson)
	if err != nil {
		fmt.Printf("read file docs_agent failed, err: %v", err)
		return
	}
	var agentSwagger Swagger
	if err := json.Unmarshal(agentFile, &agentSwagger); err != nil {
		fmt.Printf("agent json unmarshal failed, err: %v", err)
		return
	}

	coreJson := workDir + "/cmd/server/docs/docs_core/swagger.json"
	coreFile, err := os.ReadFile(coreJson)
	if err != nil {
		fmt.Printf("read file docs_core failed, err: %v", err)
		return
	}
	var coreSwagger Swagger
	if err := json.Unmarshal(coreFile, &coreSwagger); err != nil {
		fmt.Printf("core json unmarshal failed, err: %v", err)
		return
	}

	newSwagger := Swagger{
		Swagger:     agentSwagger.Swagger,
		Info:        agentSwagger.Info,
		Host:        agentSwagger.Host,
		BasePath:    agentSwagger.BasePath,
		Paths:       agentSwagger.Paths,
		Definitions: agentSwagger.Definitions,
	}

	for key, val := range coreSwagger.Paths {
		if _, ok := newSwagger.Paths[key]; ok {
			fmt.Printf("duplicate interfaces were found: %s \n", key)
		}
		newSwagger.Paths[key] = val
	}

	newXLog := make(map[string]interface{})
	for key, val := range newSwagger.Paths {
		methodMap, isMethodMap := val.(map[string]interface{})
		if !isMethodMap {
			continue
		}
		dataMap, hasPost := methodMap["post"]
		if !hasPost {
			continue
		}
		data, isDataMap := dataMap.(map[string]interface{})
		if !isDataMap {
			continue
		}
		xLog, hasXLog := data["x-panel-log"]
		if !hasXLog {
			continue
		}
		newXLog[key] = xLog
	}
	newJson, err := json.MarshalIndent(newSwagger, "", "\t")
	if err != nil {
		fmt.Printf("json marshal for new file failed, err: %v", err)
		return
	}
	if err := os.WriteFile("swagger.json", newJson, 0640); err != nil {
		fmt.Printf("write new swagger.json failed, err: %v", err)
		return
	}
	docTemplate := strings.ReplaceAll(loadDefaultDocs(), "const docTemplate = \"aa\"", fmt.Sprintf("const docTemplate = `%s`", string(newJson)))
	if err := os.WriteFile(workDir+"/cmd/server/docs/docs.go", []byte(docTemplate), 0640); err != nil {
		fmt.Printf("write new docs.go failed, err: %v", err)
		return
	}

	newXLogFile, err := json.MarshalIndent(newXLog, "", "\t")
	if err != nil {
		fmt.Printf("json marshal for new x-log file failed, err: %v", err)
		return
	}
	if err := os.WriteFile("x-log.json", newXLogFile, 0640); err != nil {
		fmt.Printf("write new x-log.json failed, err: %v", err)
		return
	}

	_ = os.RemoveAll(workDir + "/cmd/server/docs/docs_agent")
	_ = os.RemoveAll(workDir + "/cmd/server/docs/docs_core")
}

type Swagger struct {
	Swagger     string                 `json:"swagger"`
	Info        interface{}            `json:"info"`
	Host        string                 `json:"host"`
	BasePath    string                 `json:"basePath"`
	Paths       map[string]interface{} `json:"paths"`
	Definitions interface{}            `json:"definitions"`
}

func loadDefaultDocs() string {
	return `package docs

import "github.com/swaggo/swag"

const docTemplate = "aa"

var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "localhost",
	BasePath:         "/api/v2",
	Schemes:          []string{},
	Title:            "1Panel",
	Description:      "开源Linux面板",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}`
}
