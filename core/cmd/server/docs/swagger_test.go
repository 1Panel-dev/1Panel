package docs

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestGenerateXlog(t *testing.T) {
	workDir := "/usr/songliu/1Panel"
	fset := token.NewFileSet()

	apiDirs := []string{workDir + "/agent/app/api/v2", workDir + "/core/app/api/v2", workDir + "/agent/xpack/app/api/v2", workDir + "/core/xpack/app/api/v2"}

	xlogMap := make(map[string]operationJson)
	for _, dir := range apiDirs {
		entries, _ := os.ReadDir(dir)
		for _, info := range entries {
			if info.IsDir() {
				continue
			}
			fileItem, err := parser.ParseFile(fset, path.Join(dir, info.Name()), nil, parser.ParseComments)
			if err != nil {
				continue
			}
			for _, decl := range fileItem.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Doc != nil {
						routerContent := ""
						logContent := ""
						for _, comment := range d.Doc.List {
							if strings.HasPrefix(comment.Text, "// @Router") {
								routerContent = replaceStr(comment.Text, "// @Router", "[post]", "[get]", " ")
							}
							if strings.HasPrefix(comment.Text, "// @x-panel-log") {
								logContent = replaceStr(comment.Text, "// @x-panel-log", " ")
							}
						}
						if len(routerContent) != 0 && len(logContent) != 0 {
							var item operationJson
							if err := json.Unmarshal([]byte(logContent), &item); err != nil {
								panic(fmt.Sprintf("json unamrshal failed, err: %v", err))
							}
							xlogMap[routerContent] = item
						}
					}
				}
			}
		}
	}

	newJson, err := json.MarshalIndent(xlogMap, "", "\t")
	if err != nil {
		panic(fmt.Sprintf("json marshal for new file failed, err: %v", err))
	}
	if err := os.WriteFile("x-log.json", newJson, 0640); err != nil {
		panic(fmt.Sprintf("write new swagger.json failed, err: %v", err))
	}
}

func TestGenerateSwaggerDoc(t *testing.T) {
	workDir := "/usr/songliu/1Panel"
	swagBin := "/root/go/bin/swag"

	cmd1 := exec.Command(swagBin, "init", "-o", workDir+"/core/cmd/server/docs/docs_agent", "-d", workDir+"/agent", "-g", "../agent/cmd/server/main.go")
	cmd1.Dir = workDir
	std1, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Printf("generate swagger doc of agent failed, std1: %v, err: %v", string(std1), err)
		return
	}
	cmd2 := exec.Command(swagBin, "init", "-o", workDir+"/core/cmd/server/docs/docs_core", "-d", workDir+"/core", "-g", "./cmd/server/main.go")
	cmd2.Dir = workDir
	std2, err := cmd2.CombinedOutput()
	if err != nil {
		fmt.Printf("generate swagger doc of core failed, std1: %v, err: %v", string(std2), err)
		return
	}

	agentJson := workDir + "/core/cmd/server/docs/docs_agent/swagger.json"
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

	coreJson := workDir + "/core/cmd/server/docs/docs_core/swagger.json"
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
	for key, val := range coreSwagger.Definitions {
		if _, ok := newSwagger.Definitions[key]; ok {
			fmt.Printf("duplicate definitions were found: %s \n", key)
		}
		newSwagger.Definitions[key] = val
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
	if err := os.WriteFile(workDir+"/core/cmd/server/docs/docs.go", []byte(docTemplate), 0640); err != nil {
		fmt.Printf("write new docs.go failed, err: %v", err)
		return
	}

	_ = os.RemoveAll(workDir + "/core/cmd/server/docs/docs_agent")
	_ = os.RemoveAll(workDir + "/core/cmd/server/docs/docs_core")
}

type Swagger struct {
	Swagger     string                 `json:"swagger"`
	Info        interface{}            `json:"info"`
	Host        string                 `json:"host"`
	BasePath    string                 `json:"basePath"`
	Paths       map[string]interface{} `json:"paths"`
	Definitions map[string]interface{} `json:"definitions"`
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
	Description:      "Top-Rated Web-based Linux Server Management Tool",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}`
}

func replaceStr(val string, rep ...string) string {
	for _, item := range rep {
		val = strings.ReplaceAll(val, item, "")
	}
	return val
}

type operationJson struct {
	BodyKeys        []string       `json:"bodyKeys"`
	ParamKeys       []string       `json:"paramKeys"`
	BeforeFunctions []functionInfo `json:"beforeFunctions"`
	FormatZH        string         `json:"formatZH"`
	FormatEN        string         `json:"formatEN"`
}
type functionInfo struct {
	InputColumn  string `json:"input_column"`
	InputValue   string `json:"input_value"`
	IsList       bool   `json:"isList"`
	DB           string `json:"db"`
	OutputColumn string `json:"output_column"`
	OutputValue  string `json:"output_value"`
}
