package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/ai"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
	"path"
	"strconv"
	"strings"
)

type McpServerService struct{}

type IMcpServerService interface {
	Page(req request.McpServerSearch) response.McpServersRes
	Create(create request.McpServerCreate) error
	Update(req request.McpServerUpdate) error
	Delete(id uint) error
	Operate(req request.McpServerOperate) error
	GetBindDomain() (response.McpBindDomainRes, error)
	BindDomain(req request.McpBindDomain) error
	UpdateBindDomain(req request.McpBindDomainUpdate) error
}

func NewIMcpServerService() IMcpServerService {
	return &McpServerService{}
}

func (m McpServerService) Page(req request.McpServerSearch) response.McpServersRes {
	var (
		res   response.McpServersRes
		items []response.McpServerDTO
	)

	total, data, _ := mcpServerRepo.Page(req.PageInfo.Page, req.PageInfo.PageSize)
	for _, item := range data {
		_ = syncMcpServerContainerStatus(&item)
		serverDTO := response.McpServerDTO{
			McpServer:    item,
			Environments: make([]request.Environment, 0),
			Volumes:      make([]request.Volume, 0),
		}
		project, err := docker.GetComposeProject(item.Name, path.Join(global.Dir.McpDir, item.Name), []byte(item.DockerCompose), []byte(item.Env), true)
		if err != nil {
			global.LOG.Errorf("get mcp compose project error: %s", err.Error())
			continue
		}
		for _, service := range project.Services {
			if service.Environment != nil {
				for key, value := range service.Environment {
					serverDTO.Environments = append(serverDTO.Environments, request.Environment{
						Key:   key,
						Value: *value,
					})
				}
			}
			if service.Volumes != nil {
				for _, volume := range service.Volumes {
					serverDTO.Volumes = append(serverDTO.Volumes, request.Volume{
						Source: volume.Source,
						Target: volume.Target,
					})
				}
			}
		}
		items = append(items, serverDTO)
	}
	res.Total = total
	res.Items = items
	return res
}

func (m McpServerService) Update(req request.McpServerUpdate) error {
	go pullImage(req.Type)
	mcpServer, err := mcpServerRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if mcpServer.Port != req.Port {
		if err := checkPortExist(req.Port); err != nil {
			return err
		}
	}
	if mcpServer.ContainerName != req.ContainerName {
		if err := checkContainerName(req.ContainerName); err != nil {
			return err
		}
	}
	req.Command = strings.TrimSuffix(req.Command, "\n")
	mcpServer.Name = req.Name
	mcpServer.ContainerName = req.ContainerName
	mcpServer.Port = req.Port
	mcpServer.Command = req.Command
	mcpServer.BaseURL = req.BaseURL
	mcpServer.HostIP = req.HostIP
	mcpServer.OutputTransport = req.OutputTransport
	mcpServer.Type = req.Type
	if req.OutputTransport == "sse" {
		mcpServer.SsePath = req.SsePath
	} else {
		mcpServer.StreamableHttpPath = req.StreamableHttpPath
	}
	if err := handleCreateParams(mcpServer, req.Environments, req.Volumes); err != nil {
		return err
	}
	env := handleEnv(mcpServer)
	mcpDir := path.Join(global.Dir.McpDir, mcpServer.Name)
	envPath := path.Join(mcpDir, ".env")
	if err := gotenv.Write(env, envPath); err != nil {
		return err
	}
	dockerComposePath := path.Join(mcpDir, "docker-compose.yml")
	if err := files.NewFileOp().SaveFile(dockerComposePath, mcpServer.DockerCompose, 0644); err != nil {
		return err
	}
	mcpServer.Status = constant.StatusStarting
	if err := mcpServerRepo.Save(mcpServer); err != nil {
		return err
	}
	go startMcp(mcpServer)
	return nil
}

func (m McpServerService) Create(create request.McpServerCreate) error {
	go pullImage(create.Type)
	servers, _ := mcpServerRepo.List()
	for _, server := range servers {
		if server.Port == create.Port {
			return buserr.New("ErrPortInUsed")
		}
		if server.ContainerName == create.ContainerName {
			return buserr.New("ErrContainerName")
		}
		if server.Name == create.Name {
			return buserr.New("ErrNameIsExist")
		}
		if server.SsePath == create.SsePath {
			return buserr.New("ErrSsePath")
		}
	}
	create.Command = strings.TrimSuffix(create.Command, "\n")
	if err := checkPortExist(create.Port); err != nil {
		return err
	}
	if err := checkContainerName(create.ContainerName); err != nil {
		return err
	}
	mcpDir := path.Join(global.Dir.McpDir, create.Name)
	mcpServer := &model.McpServer{
		Name:            create.Name,
		ContainerName:   create.ContainerName,
		Port:            create.Port,
		Command:         create.Command,
		Status:          constant.StatusStarting,
		BaseURL:         create.BaseURL,
		Dir:             mcpDir,
		HostIP:          create.HostIP,
		OutputTransport: create.OutputTransport,
		Type:            create.Type,
	}
	if create.OutputTransport == "sse" {
		mcpServer.SsePath = create.SsePath
	} else {
		mcpServer.StreamableHttpPath = create.StreamableHttpPath
	}

	if err := handleCreateParams(mcpServer, create.Environments, create.Volumes); err != nil {
		return err
	}

	env := handleEnv(mcpServer)
	filesOP := files.NewFileOp()
	if !filesOP.Stat(mcpDir) {
		_ = filesOP.CreateDir(mcpDir, 0644)
	}
	envPath := path.Join(mcpDir, ".env")
	if err := gotenv.Write(env, envPath); err != nil {
		return err
	}
	dockerComposePath := path.Join(mcpDir, "docker-compose.yml")
	if err := filesOP.SaveFile(dockerComposePath, mcpServer.DockerCompose, 0644); err != nil {
		return err
	}
	if err := mcpServerRepo.Create(mcpServer); err != nil {
		return err
	}
	addProxy(mcpServer)
	go startMcp(mcpServer)
	return nil
}

func (m McpServerService) Delete(id uint) error {
	mcpServer, err := mcpServerRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	composePath := path.Join(global.Dir.McpDir, mcpServer.Name, "docker-compose.yml")
	_, _ = compose.Down(composePath)
	_ = files.NewFileOp().DeleteDir(path.Join(global.Dir.McpDir, mcpServer.Name))

	websiteID := GetWebsiteID()
	if websiteID > 0 {
		websiteService := NewIWebsiteService()
		delProxyReq := request.WebsiteProxyDel{
			ID:   websiteID,
			Name: mcpServer.Name,
		}
		_ = websiteService.DeleteProxy(delProxyReq)
	}
	return mcpServerRepo.DeleteBy(repo.WithByID(id))
}

func (m McpServerService) Operate(req request.McpServerOperate) error {
	mcpServer, err := mcpServerRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	composePath := path.Join(mcpServer.Dir, "docker-compose.yml")
	var out string
	switch req.Operate {
	case "start":
		out, err = compose.Up(composePath)
		mcpServer.Status = constant.StatusRunning
	case "stop":
		out, err = compose.Down(composePath)
		mcpServer.Status = constant.StatusStopped
	case "restart":
		out, err = compose.Restart(composePath)
		mcpServer.Status = constant.StatusRunning
	}
	if err != nil {
		mcpServer.Status = constant.StatusError
		mcpServer.Message = out
	}
	return mcpServerRepo.Save(mcpServer)
}

func (m McpServerService) GetBindDomain() (response.McpBindDomainRes, error) {
	var res response.McpBindDomainRes
	websiteID := GetWebsiteID()
	if websiteID == 0 {
		return res, nil
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		return res, nil
	}
	res.WebsiteID = website.ID
	res.Domain = website.PrimaryDomain
	if website.WebsiteSSLID > 0 {
		res.SSLID = website.WebsiteSSLID
		ssl, _ := websiteSSLRepo.GetFirst(repo.WithByID(website.WebsiteSSLID))
		res.AcmeAccountID = ssl.AcmeAccountID
	}
	res.ConnUrl = fmt.Sprintf("%s://%s", strings.ToLower(website.Protocol), website.PrimaryDomain)
	res.AllowIPs = GetAllowIps(website)
	return res, nil

}

func (m McpServerService) BindDomain(req request.McpBindDomain) error {
	nginxInstall, _ := getAppInstallByKey(constant.AppOpenresty)
	if nginxInstall.ID == 0 {
		return buserr.New("ErrOpenrestyInstall")
	}
	var (
		ipList []string
		err    error
	)
	if len(req.IPList) > 0 {
		ipList, err = common.HandleIPList(req.IPList)
		if err != nil {
			return err
		}
	}
	if req.SSLID > 0 {
		ssl, err := websiteSSLRepo.GetFirst(repo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		if ssl.Pem == "" {
			return buserr.New("ErrSSL")
		}
	}
	group, _ := groupRepo.Get(groupRepo.WithByWebsiteDefault())

	domain, err := ParseDomain(req.Domain)
	if err != nil {
		return err
	}
	if domain.Port == 0 {
		domain.Port = nginxInstall.HttpPort
	}
	createWebsiteReq := request.WebsiteCreate{
		Domains: []request.WebsiteDomain{
			{
				Domain: domain.Domain,
				Port:   domain.Port,
			},
		},
		Alias:          strings.ToLower(req.Domain),
		Type:           constant.Static,
		WebsiteGroupID: group.ID,
	}
	if req.SSLID > 0 {
		createWebsiteReq.WebsiteSSLID = req.SSLID
		createWebsiteReq.EnableSSL = true
	}
	websiteService := NewIWebsiteService()
	if err := websiteService.CreateWebsite(createWebsiteReq); err != nil {
		return err
	}
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(strings.ToLower(req.Domain)))
	if err != nil {
		return err
	}
	_ = settingRepo.UpdateOrCreate("MCP_WEBSITE_ID", fmt.Sprintf("%d", website.ID))
	if len(ipList) > 0 {
		if err = ConfigAllowIPs(ipList, website); err != nil {
			return err
		}
	}
	if err = addMCPProxy(website.ID); err != nil {
		return err
	}
	return nil
}

func (m McpServerService) UpdateBindDomain(req request.McpBindDomainUpdate) error {
	nginxInstall, _ := getAppInstallByKey(constant.AppOpenresty)
	if nginxInstall.ID == 0 {
		return buserr.New("ErrOpenrestyInstall")
	}
	var (
		ipList []string
		err    error
	)
	if len(req.IPList) > 0 {
		ipList, err = common.HandleIPList(req.IPList)
		if err != nil {
			return err
		}
	}
	if req.SSLID > 0 {
		ssl, err := websiteSSLRepo.GetFirst(repo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		if ssl.Pem == "" {
			return buserr.New("ErrSSL")
		}
	}
	websiteService := NewIWebsiteService()
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if err = ConfigAllowIPs(ipList, website); err != nil {
		return err
	}
	if req.SSLID > 0 {
		sslReq := request.WebsiteHTTPSOp{
			WebsiteID:    website.ID,
			Enable:       true,
			Type:         "existed",
			WebsiteSSLID: req.SSLID,
			HttpConfig:   "HTTPSOnly",
		}
		if _, err = websiteService.OpWebsiteHTTPS(context.Background(), sslReq); err != nil {
			return err
		}
	}
	if website.WebsiteSSLID > 0 && req.SSLID == 0 {
		sslReq := request.WebsiteHTTPSOp{
			WebsiteID: website.ID,
			Enable:    false,
		}
		if _, err = websiteService.OpWebsiteHTTPS(context.Background(), sslReq); err != nil {
			return err
		}
	}
	go updateMcpConfig(website.ID)
	return nil
}

func updateMcpConfig(websiteID uint) {
	servers, _ := mcpServerRepo.List()
	if len(servers) == 0 {
		return
	}
	website, _ := websiteRepo.GetFirst(repo.WithByID(websiteID))
	websiteDomain := website.Domains[0]
	var baseUrl string
	if website.Protocol == constant.ProtocolHTTP {
		baseUrl = fmt.Sprintf("http://%s", websiteDomain.Domain)
	} else {
		baseUrl = fmt.Sprintf("https://%s", websiteDomain.Domain)
	}

	go func() {
		for _, server := range servers {
			if server.BaseURL != baseUrl {
				server.BaseURL = baseUrl
				server.HostIP = "127.0.0.1"
				_ = updateMcpServer(&server)
			}
		}
	}()
}

func addProxy(server *model.McpServer) {
	websiteID := GetWebsiteID()
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		global.LOG.Errorf("[mcp] add proxy failed, err: %v", err)
		return
	}
	fileOp := files.NewFileOp()
	includeDir := GetSitePath(website, SiteProxyDir)
	if !fileOp.Stat(includeDir) {
		if err = fileOp.CreateDir(includeDir, 0644); err != nil {
			return
		}
	}
	config, err := parser.NewStringParser(string(nginx_conf.SSE)).Parse()
	if err != nil {
		return
	}
	includePath := path.Join(includeDir, server.Name+".conf")
	config.FilePath = includePath
	directives := config.Directives
	location, ok := directives[0].(*components.Location)
	if !ok {
		return
	}
	var proxyPath string
	if server.OutputTransport == "sse" {
		proxyPath = server.SsePath
	} else {
		proxyPath = server.StreamableHttpPath
	}
	location.UpdateDirective("proxy_pass", []string{fmt.Sprintf("http://127.0.0.1:%d%s", server.Port, proxyPath)})
	location.ChangePath("^~", proxyPath)
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		global.LOG.Errorf("write config failed, err: %v", buserr.WithErr("ErrUpdateBuWebsite", err))
		return
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		global.LOG.Errorf("update nginx config failed, err: %v", err)
		return
	}
}

func addMCPProxy(websiteID uint) error {
	servers, _ := mcpServerRepo.List()
	if len(servers) == 0 {
		return nil
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	includeDir := GetSitePath(website, SiteProxyDir)
	if !fileOp.Stat(includeDir) {
		if err = fileOp.CreateDir(includeDir, 0644); err != nil {
			return err
		}
	}
	config, err := parser.NewStringParser(string(nginx_conf.SSE)).Parse()
	if err != nil {
		return err
	}
	websiteDomain := website.Domains[0]
	var baseUrl string
	if website.Protocol == constant.ProtocolHTTP {
		baseUrl = fmt.Sprintf("http://%s", websiteDomain.Domain)
	} else {
		baseUrl = fmt.Sprintf("https://%s", websiteDomain.Domain)
	}
	if websiteDomain.Port != 80 && websiteDomain.Port != 443 {
		baseUrl = fmt.Sprintf("%s:%d", baseUrl, websiteDomain.Port)
	}
	for _, server := range servers {
		includePath := path.Join(includeDir, server.Name+".conf")
		config.FilePath = includePath
		directives := config.Directives
		location, ok := directives[0].(*components.Location)
		if !ok {
			err = errors.New("error")
			return err
		}
		var proxyPath string
		if server.OutputTransport == "sse" {
			proxyPath = server.SsePath
		} else {
			proxyPath = server.StreamableHttpPath
		}
		location.UpdateDirective("proxy_pass", []string{fmt.Sprintf("http://127.0.0.1:%d%s", server.Port, proxyPath)})
		location.ChangePath("^~", proxyPath)
		if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
			return buserr.WithErr("ErrUpdateBuWebsite", err)
		}
		server.BaseURL = baseUrl
		server.HostIP = "127.0.0.1"
		go updateMcpServer(&server)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return err
	}
	return nil
}

func updateMcpServer(mcpServer *model.McpServer) error {
	env := handleEnv(mcpServer)
	if err := gotenv.Write(env, path.Join(mcpServer.Dir, ".env")); err != nil {
		return err
	}
	_ = mcpServerRepo.Save(mcpServer)
	composePath := path.Join(global.Dir.McpDir, mcpServer.Name, "docker-compose.yml")
	_, _ = compose.Down(composePath)
	if _, err := compose.Up(composePath); err != nil {
		mcpServer.Status = constant.StatusError
		mcpServer.Message = err.Error()
	}
	return mcpServerRepo.Save(mcpServer)
}

func handleEnv(mcpServer *model.McpServer) gotenv.Env {
	env := make(gotenv.Env)
	env["CONTAINER_NAME"] = mcpServer.ContainerName
	env["COMMAND"] = mcpServer.Command
	env["PANEL_APP_PORT_HTTP"] = strconv.Itoa(mcpServer.Port)
	env["BASE_URL"] = mcpServer.BaseURL
	env["SSE_PATH"] = mcpServer.SsePath
	env["HOST_IP"] = mcpServer.HostIP
	env["STREAMABLE_HTTP_PATH"] = mcpServer.StreamableHttpPath
	env["OUTPUT_TRANSPORT"] = mcpServer.OutputTransport
	envStr, _ := gotenv.Marshal(env)
	mcpServer.Env = envStr
	return env
}

func handleCreateParams(mcpServer *model.McpServer, environments []request.Environment, volumes []request.Volume) error {
	var composeContent []byte
	if mcpServer.ID == 0 {
		composeContent = ai.DefaultMcpCompose
	} else {
		composeContent = []byte(mcpServer.DockerCompose)
	}
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return err
	}
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		return buserr.New("ErrFileParse")
	}
	serviceName := ""
	serviceValue := make(map[string]interface{})

	if mcpServer.ID > 0 {
		serviceName = mcpServer.Name
		serviceValue = services[serviceName].(map[string]interface{})
	} else {
		for name, service := range services {
			serviceName = name
			serviceValue = service.(map[string]interface{})
			break
		}
		delete(services, serviceName)
	}
	delete(serviceValue, "environment")
	if len(environments) > 0 {
		envMap := make(map[string]string)
		for _, env := range environments {
			envMap[env.Key] = env.Value
		}
		serviceValue["environment"] = envMap
	}
	delete(serviceValue, "volumes")
	if len(volumes) > 0 {
		volumeList := make([]string, 0)
		for _, volume := range volumes {
			volumeList = append(volumeList, fmt.Sprintf("%s:%s", volume.Source, volume.Target))
		}
		serviceValue["volumes"] = volumeList
	}
	if mcpServer.Type == "npx" {
		serviceValue["image"] = "supercorp/supergateway:latest"
	} else {
		serviceValue["image"] = "supercorp/supergateway:uvx"
	}

	services[mcpServer.Name] = serviceValue
	composeByte, err := yaml.Marshal(composeMap)
	if err != nil {
		return err
	}
	mcpServer.DockerCompose = string(composeByte)
	return nil
}

func startMcp(mcpServer *model.McpServer) {
	composePath := path.Join(global.Dir.McpDir, mcpServer.Name, "docker-compose.yml")
	if mcpServer.Status != constant.StatusNormal {
		_, _ = compose.Down(composePath)
	}
	if out, err := compose.Up(composePath); err != nil {
		mcpServer.Status = constant.StatusError
		mcpServer.Message = out
	} else {
		mcpServer.Status = constant.StatusRunning
		mcpServer.Message = ""
	}
	_ = syncMcpServerContainerStatus(mcpServer)
}

func syncMcpServerContainerStatus(mcpServer *model.McpServer) error {
	containerNames := []string{mcpServer.ContainerName}
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	containers, err := cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		mcpServer.Status = constant.StatusStopped
		return mcpServerRepo.Save(mcpServer)
	}
	container := containers[0]
	switch container.State {
	case "exited":
		mcpServer.Status = constant.StatusError
	case "running":
		mcpServer.Status = constant.StatusRunning
	case "paused":
		mcpServer.Status = constant.StatusStopped
	case "restarting":
		mcpServer.Status = constant.StatusRestarting
	default:
		if mcpServer.Status != constant.StatusStarting {
			mcpServer.Status = constant.StatusStopped
		}
	}
	return mcpServerRepo.Save(mcpServer)
}

func GetWebsiteID() uint {
	websiteID, _ := settingRepo.Get(settingRepo.WithByKey("MCP_WEBSITE_ID"))
	if websiteID.Value == "" {
		return 0
	}
	websiteIDUint, _ := strconv.ParseUint(websiteID.Value, 10, 64)
	return uint(websiteIDUint)
}

func pullImage(imageType string) {
	if global.CONF.Base.IsOffLine {
		return
	}
	if imageType == "npx" {
		if err := docker.PullImage("supercorp/supergateway:latest"); err != nil {
			global.LOG.Errorf("docker pull mcp image error: %s", err.Error())
		}
	} else {
		if err := docker.PullImage("supercorp/supergateway:uvx"); err != nil {
			global.LOG.Errorf("docker pull mcp image error: %s", err.Error())
		}
	}
}
