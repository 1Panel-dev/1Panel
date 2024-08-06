package docs

import "github.com/swaggo/swag"

const docTemplate = `{
	"swagger": "2.0",
	"info": {
		"contact": {},
		"description": "开源Linux面板",
		"license": {
			"name": "Apache 2.0",
			"url": "http://www.apache.org/licenses/LICENSE-2.0.html"
		},
		"termsOfService": "http://swagger.io/terms/",
		"title": "1Panel",
		"version": "1.0"
	},
	"host": "localhost",
	"basePath": "/api/v1",
	"paths": {
		"/apps/:key": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 key 获取应用信息",
				"parameters": [
					{
						"description": "app key",
						"in": "path",
						"name": "key",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.AppDTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app by key",
				"tags": [
					"App"
				]
			}
		},
		"/apps/checkupdate": {
			"get": {
				"description": "获取应用更新版本",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get app list update",
				"tags": [
					"App"
				]
			}
		},
		"/apps/detail/:appId/:version/:type": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 appid 获取应用详情",
				"parameters": [
					{
						"description": "app id",
						"in": "path",
						"name": "appId",
						"required": true,
						"type": "integer"
					},
					{
						"description": "app 版本",
						"in": "path",
						"name": "version",
						"required": true,
						"type": "string"
					},
					{
						"description": "app 类型",
						"in": "path",
						"name": "version",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.AppDetailDTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app detail by appid",
				"tags": [
					"App"
				]
			}
		},
		"/apps/details/:id": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 id 获取应用详情",
				"parameters": [
					{
						"description": "id",
						"in": "path",
						"name": "appId",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.AppDetailDTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get app detail by id",
				"tags": [
					"App"
				]
			}
		},
		"/apps/ignored": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取忽略的应用版本",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.IgnoredApp"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get Ignore App",
				"tags": [
					"App"
				]
			}
		},
		"/apps/install": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "安装应用",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstallCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/model.AppInstall"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Install app",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "app_installs",
							"input_column": "name",
							"input_value": "name",
							"isList": false,
							"output_column": "app_id",
							"output_value": "appId"
						},
						{
							"db": "apps",
							"info": "appId",
							"isList": false,
							"output_column": "key",
							"output_value": "appKey"
						}
					],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Install app [appKey]-[name]",
					"formatZH": "安装应用 [appKey]-[name]",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "检查应用安装情况",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstalledInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.AppInstalledCheck"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check app installed",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "通过 key 获取应用默认配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search default config by key",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/conninfo/:key": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取应用连接信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app password by key",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/delete/check/:appInstallId": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "删除前检查",
				"parameters": [
					{
						"description": "App install id",
						"in": "path",
						"name": "appInstallId",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.AppResource"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check before delete",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/ignore": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "忽略应用升级版本",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstalledIgnoreUpgrade"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "ignore App Update",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"installId"
					],
					"formatEN": "Application param update [installId]",
					"formatZH": "忽略应用 [installId] 版本升级",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/list": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取已安装应用列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.AppInstallInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List app installed",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/loadport": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取应用端口",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "integer"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app port by key",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/op": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作已安装应用",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstalledOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate installed app",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "app_installs",
							"input_column": "id",
							"input_value": "installId",
							"isList": false,
							"output_column": "app_id",
							"output_value": "appId"
						},
						{
							"db": "app_installs",
							"input_column": "id",
							"input_value": "installId",
							"isList": false,
							"output_column": "name",
							"output_value": "appName"
						},
						{
							"db": "apps",
							"input_column": "id",
							"input_value": "appId",
							"isList": false,
							"output_column": "key",
							"output_value": "appKey"
						}
					],
					"bodyKeys": [
						"installId",
						"operate"
					],
					"formatEN": "[operate] App [appKey][appName]",
					"formatZH": "[operate] 应用 [appKey][appName]",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/params/:appInstallId": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 install id 获取应用参数",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "appInstallId",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.AppParam"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search params by appInstallId",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/params/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改应用参数",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstalledUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change app params",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"installId"
					],
					"formatEN": "Application param update [installId]",
					"formatZH": "应用参数修改 [installId]",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/port/change": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改应用端口",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.PortUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change app port",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"name",
						"port"
					],
					"formatEN": "Application port update [key]-[name] =\u003e [port]",
					"formatZH": "应用端口修改 [key]-[name] =\u003e [port]",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "分页获取已安装应用列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppInstalledSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page app installed",
				"tags": [
					"App"
				]
			}
		},
		"/apps/installed/sync": {
			"post": {
				"description": "同步已安装应用列表",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Sync app installed",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Sync the list of installed apps",
					"formatZH": "同步已安装应用列表",
					"paramKeys": []
				}
			}
		},
		"/apps/installed/update/versions": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "通过 install id 获取应用更新版本",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "appInstallId",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.AppVersion"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app update version by install id",
				"tags": [
					"App"
				]
			}
		},
		"/apps/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取应用列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.AppSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List apps",
				"tags": [
					"App"
				]
			}
		},
		"/apps/services/:key": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 key 获取应用 service",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "key",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.AppService"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search app service by key",
				"tags": [
					"App"
				]
			}
		},
		"/apps/sync": {
			"post": {
				"description": "同步应用列表",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Sync app list",
				"tags": [
					"App"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "App store synchronization",
					"formatZH": "应用商店同步",
					"paramKeys": []
				}
			}
		},
		"/containers": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建容器",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"image"
					],
					"formatEN": "create container [name][image]",
					"formatZH": "创建容器 [name][image]",
					"paramKeys": []
				}
			}
		},
		"/containers/clean/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清理容器日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean container log",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "clean container [name] logs",
					"formatZH": "清理容器 [name] 日志",
					"paramKeys": []
				}
			}
		},
		"/containers/commit": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器提交生成新镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerCommit"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"summary": "Commit Container",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/compose": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建容器编排",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create compose",
				"tags": [
					"Container Compose"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create compose [name]",
					"formatZH": "创建 compose [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/compose/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器编排操作",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeOperation"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate compose",
				"tags": [
					"Container Compose"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"operation"
					],
					"formatEN": "compose [operation] [name]",
					"formatZH": "compose [operation] [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/compose/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取编排列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page composes",
				"tags": [
					"Container Compose"
				]
			}
		},
		"/containers/compose/search/log": {
			"get": {
				"description": "docker-compose 日志",
				"parameters": [
					{
						"description": "compose 文件地址",
						"in": "query",
						"name": "compose",
						"type": "string"
					},
					{
						"description": "时间筛选",
						"in": "query",
						"name": "since",
						"type": "string"
					},
					{
						"description": "是否追踪",
						"in": "query",
						"name": "follow",
						"type": "string"
					},
					{
						"description": "显示行号",
						"in": "query",
						"name": "tail",
						"type": "string"
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Container Compose logs",
				"tags": [
					"Container Compose"
				]
			}
		},
		"/containers/compose/test": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "测试 compose 是否可用",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Test compose",
				"tags": [
					"Container Compose"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "check compose [name]",
					"formatZH": "检测 compose [name] 格式",
					"paramKeys": []
				}
			}
		},
		"/containers/compose/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新容器编排",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update compose",
				"tags": [
					"Container Compose"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "update compose information [name]",
					"formatZH": "更新 compose [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/daemonjson": {
			"get": {
				"description": "获取 docker 配置信息",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DaemonJsonConf"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load docker daemon.json",
				"tags": [
					"Container Docker"
				]
			}
		},
		"/containers/daemonjson/file": {
			"get": {
				"description": "获取 docker 配置信息(表单)",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load docker daemon.json",
				"tags": [
					"Container Docker"
				]
			}
		},
		"/containers/daemonjson/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 docker 配置信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update docker daemon.json",
				"tags": [
					"Container Docker"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "Updated configuration [key]",
					"formatZH": "更新配置 [key]",
					"paramKeys": []
				}
			}
		},
		"/containers/daemonjson/update/byfile": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "上传替换 docker 配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DaemonJsonUpdateByFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update docker daemon.json by upload file",
				"tags": [
					"Container Docker"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Updated configuration file",
					"formatZH": "更新配置文件",
					"paramKeys": []
				}
			}
		},
		"/containers/docker/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Docker 操作",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DockerOperation"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate docker",
				"tags": [
					"Container Docker"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] docker service",
					"formatZH": "docker 服务 [operation]",
					"paramKeys": []
				}
			}
		},
		"/containers/docker/status": {
			"get": {
				"description": "获取 docker 服务状态",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load docker status",
				"tags": [
					"Container Docker"
				]
			}
		},
		"/containers/download/log": {
			"post": {
				"description": "下载容器日志",
				"responses": {}
			}
		},
		"/containers/image": {
			"get": {
				"description": "获取镜像名称列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.Options"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "load images options",
				"tags": [
					"Container Image"
				]
			}
		},
		"/containers/image/all": {
			"get": {
				"description": "获取所有镜像列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.ImageInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List all images",
				"tags": [
					"Container Image"
				]
			}
		},
		"/containers/image/build": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "构建镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageBuild"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Build image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "build image [name]",
					"formatZH": "构建镜像 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/image/load": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "导入镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageLoad"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "load image from [path]",
					"formatZH": "从 [path] 加载镜像",
					"paramKeys": []
				}
			}
		},
		"/containers/image/pull": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "拉取镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImagePull"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Pull image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "image_repos",
							"input_column": "id",
							"input_value": "repoID",
							"isList": false,
							"output_column": "name",
							"output_value": "reponame"
						}
					],
					"bodyKeys": [
						"repoID",
						"imageName"
					],
					"formatEN": "image pull [reponame][imageName]",
					"formatZH": "镜像拉取 [reponame][imageName]",
					"paramKeys": []
				}
			}
		},
		"/containers/image/push": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "推送镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImagePush"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Push image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "image_repos",
							"input_column": "id",
							"input_value": "repoID",
							"isList": false,
							"output_column": "name",
							"output_value": "reponame"
						}
					],
					"bodyKeys": [
						"repoID",
						"tagName",
						"name"
					],
					"formatEN": "push [tagName] to [reponame][name]",
					"formatZH": "[tagName] 推送到 [reponame][name]",
					"paramKeys": []
				}
			}
		},
		"/containers/image/remove": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"names"
					],
					"formatEN": "remove image [names]",
					"formatZH": "移除镜像 [names]",
					"paramKeys": []
				}
			}
		},
		"/containers/image/save": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "导出镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageSave"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Save image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"tagName",
						"path",
						"name"
					],
					"formatEN": "save [tagName] as [path]/[name]",
					"formatZH": "保留 [tagName] 为 [path]/[name]",
					"paramKeys": []
				}
			}
		},
		"/containers/image/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取镜像列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page images",
				"tags": [
					"Container Image"
				]
			}
		},
		"/containers/image/tag": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Tag 镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageTag"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Tag image",
				"tags": [
					"Container Image"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "image_repos",
							"input_column": "id",
							"input_value": "repoID",
							"isList": false,
							"output_column": "name",
							"output_value": "reponame"
						}
					],
					"bodyKeys": [
						"repoID",
						"targetName"
					],
					"formatEN": "tag image [reponame][targetName]",
					"formatZH": "tag 镜像 [reponame][targetName]",
					"paramKeys": []
				}
			}
		},
		"/containers/info": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器表单信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.ContainerOperate"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load container info",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/inspect": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器详情",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.InspectReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Container inspect",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/ipv6option/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 docker ipv6 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.LogOption"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update docker daemon.json ipv6 option",
				"tags": [
					"Container Docker"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Updated the ipv6 option",
					"formatZH": "更新 ipv6 配置",
					"paramKeys": []
				}
			}
		},
		"/containers/limit": {
			"get": {
				"description": "获取容器限制",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.ResourceLimit"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load container limits"
			}
		},
		"/containers/list": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器名称",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List containers",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/list/stats": {
			"get": {
				"description": "获取容器列表资源占用",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.ContainerListStats"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load container stats"
			}
		},
		"/containers/load/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器操作日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load container log",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/logoption/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 docker 日志配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.LogOption"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update docker daemon.json log option",
				"tags": [
					"Container Docker"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Updated the log option",
					"formatZH": "更新日志配置",
					"paramKeys": []
				}
			}
		},
		"/containers/network": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器网络列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.Options"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List networks",
				"tags": [
					"Container Network"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建容器网络",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.NetworkCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create network",
				"tags": [
					"Container Network"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create container network [name]",
					"formatZH": "创建容器网络 name",
					"paramKeys": []
				}
			}
		},
		"/containers/network/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除容器网络",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete network",
				"tags": [
					"Container Network"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"names"
					],
					"formatEN": "delete container network [names]",
					"formatZH": "删除容器网络 [names]",
					"paramKeys": []
				}
			}
		},
		"/containers/network/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器网络列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page networks",
				"tags": [
					"Container Network"
				]
			}
		},
		"/containers/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器操作",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerOperation"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate Container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"names",
						"operation"
					],
					"formatEN": "container [operation] [names]",
					"formatZH": "容器 [names] 执行 [operation]",
					"paramKeys": []
				}
			}
		},
		"/containers/prune": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器清理",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerPrune"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.ContainerPruneReport"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"pruneType"
					],
					"formatEN": "clean container [pruneType]",
					"formatZH": "清理容器 [pruneType]",
					"paramKeys": []
				}
			}
		},
		"/containers/rename": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "容器重命名",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerRename"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Rename Container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"newName"
					],
					"formatEN": "rename container [name] =\u003e [newName]",
					"formatZH": "容器重命名 [name] =\u003e [newName]",
					"paramKeys": []
				}
			}
		},
		"/containers/repo": {
			"get": {
				"description": "获取镜像仓库列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.ImageRepoOption"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List image repos",
				"tags": [
					"Container Image-repo"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建镜像仓库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageRepoDelete"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create image repo",
				"tags": [
					"Container Image-repo"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create image repo [name]",
					"formatZH": "创建镜像仓库 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/repo/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除镜像仓库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageRepoDelete"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete image repo",
				"tags": [
					"Container Image-repo"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "image_repos",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete image repo [names]",
					"formatZH": "删除镜像仓库 [names]",
					"paramKeys": []
				}
			}
		},
		"/containers/repo/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取镜像仓库列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page image repos",
				"tags": [
					"Container Image-repo"
				]
			}
		},
		"/containers/repo/status": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 docker 仓库状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load repo status",
				"tags": [
					"Container Image-repo"
				]
			}
		},
		"/containers/repo/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新镜像仓库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ImageRepoUpdate"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update image repo",
				"tags": [
					"Container Image-repo"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "image_repos",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "update image repo information [name]",
					"formatZH": "更新镜像仓库 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageContainer"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page containers",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/search/log": {
			"post": {
				"description": "容器日志",
				"parameters": [
					{
						"description": "容器名称",
						"in": "query",
						"name": "container",
						"type": "string"
					},
					{
						"description": "时间筛选",
						"in": "query",
						"name": "since",
						"type": "string"
					},
					{
						"description": "是否追踪",
						"in": "query",
						"name": "follow",
						"type": "string"
					},
					{
						"description": "显示行号",
						"in": "query",
						"name": "tail",
						"type": "string"
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Container logs",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/stats/:id": {
			"get": {
				"description": "容器监控信息",
				"parameters": [
					{
						"description": "容器id",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.ContainerStats"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Container stats",
				"tags": [
					"Container"
				]
			}
		},
		"/containers/template": {
			"get": {
				"description": "获取容器编排模版列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.ComposeTemplateInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List compose templates",
				"tags": [
					"Container Compose-template"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建容器编排模版",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeTemplateCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create compose template",
				"tags": [
					"Container Compose-template"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create compose template [name]",
					"formatZH": "创建 compose 模版 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/template/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除容器编排模版",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete compose template",
				"tags": [
					"Container Compose-template"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "compose_templates",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete compose template [names]",
					"formatZH": "删除 compose 模版 [names]",
					"paramKeys": []
				}
			}
		},
		"/containers/template/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器编排模版列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page compose templates",
				"tags": [
					"Container Compose-template"
				]
			}
		},
		"/containers/template/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新容器编排模版",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ComposeTemplateUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update compose template",
				"tags": [
					"Container Compose-template"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "compose_templates",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "update compose template information [name]",
					"formatZH": "更新 compose 模版 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新容器",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"image"
					],
					"formatEN": "update container [name][image]",
					"formatZH": "更新容器 [name][image]",
					"paramKeys": []
				}
			}
		},
		"/containers/upgrade": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新容器镜像",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ContainerUpgrade"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Upgrade container",
				"tags": [
					"Container"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"image"
					],
					"formatEN": "upgrade container image [name][image]",
					"formatZH": "更新容器镜像 [name][image]",
					"paramKeys": []
				}
			}
		},
		"/containers/volume": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器存储卷列表",
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.Options"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List volumes",
				"tags": [
					"Container Volume"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建容器存储卷",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.VolumeCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create volume",
				"tags": [
					"Container Volume"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create container volume [name]",
					"formatZH": "创建容器存储卷 [name]",
					"paramKeys": []
				}
			}
		},
		"/containers/volume/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除容器存储卷",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete volume",
				"tags": [
					"Container Volume"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"names"
					],
					"formatEN": "delete container volume [names]",
					"formatZH": "删除容器存储卷 [names]",
					"paramKeys": []
				}
			}
		},
		"/containers/volume/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取容器存储卷分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page volumes",
				"tags": [
					"Container Volume"
				]
			}
		},
		"/core/auth/captcha": {
			"get": {
				"description": "加载验证码",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.CaptchaResponse"
						}
					}
				},
				"summary": "Load captcha",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/demo": {
			"get": {
				"description": "判断是否为demo环境",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"summary": "Check System isDemo",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/issafety": {
			"get": {
				"description": "获取系统安全登录状态",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"summary": "Load safety status",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/language": {
			"get": {
				"description": "获取系统语言设置",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"summary": "Load System Language",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/login": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "用户登录",
				"parameters": [
					{
						"description": "安全入口 base64 加密串",
						"in": "header",
						"name": "EntranceCode",
						"required": true,
						"type": "string"
					},
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Login"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.UserLoginInfo"
						}
					}
				},
				"summary": "User login",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/logout": {
			"post": {
				"description": "用户登出",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "User logout",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/auth/mfalogin": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "用户 mfa 登录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MFALogin"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"headers": {
							"EntranceCode": {
								"description": "安全入口",
								"type": "string"
							}
						},
						"schema": {
							"$ref": "#/definitions/dto.UserLoginInfo"
						}
					}
				},
				"summary": "User login with mfa",
				"tags": [
					"Auth"
				]
			}
		},
		"/core/logs/clean": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清空操作日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CleanLog"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean operation logs",
				"tags": [
					"Logs"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"logType"
					],
					"formatEN": "Clean the [logType] log information",
					"formatZH": "清空 [logType] 日志信息",
					"paramKeys": []
				}
			}
		},
		"/core/logs/login": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统登录日志列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchLgLogWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page login logs",
				"tags": [
					"Logs"
				]
			}
		},
		"/core/logs/operation": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统操作日志列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchOpLogWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page operation logs",
				"tags": [
					"Logs"
				]
			}
		},
		"/core/settings/basedir": {
			"get": {
				"description": "获取安装根目录",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load local backup dir",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/bind/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统监听信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BindInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system bind info",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"ipv6",
						"bindAddress"
					],
					"formatEN": "update system bind info =\u003e ipv6: [ipv6], 监听 IP: [bindAddress]",
					"formatZH": "修改系统监听信息 =\u003e ipv6: [ipv6], 监听 IP: [bindAddress]",
					"paramKeys": []
				}
			}
		},
		"/core/settings/expired/handle": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "重置过期系统登录密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PasswordUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Reset system password expired",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "reset an expired Password",
					"formatZH": "重置过期密码",
					"paramKeys": []
				}
			}
		},
		"/core/settings/interface": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统地址信息",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system address",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/menu/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "隐藏高级功能菜单",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system setting",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Hide advanced feature menu.",
					"formatZH": "隐藏高级功能菜单",
					"paramKeys": []
				}
			}
		},
		"/core/settings/mfa": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mfa 信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MfaCredential"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/mfa.Otp"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load mfa info",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/mfa/bind": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Mfa 绑定",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MfaCredential"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Bind mfa",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "bind mfa",
					"formatZH": "mfa 绑定",
					"paramKeys": []
				}
			}
		},
		"/core/settings/password/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统登录密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PasswordUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system password",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "update system password",
					"formatZH": "修改系统密码",
					"paramKeys": []
				}
			}
		},
		"/core/settings/port/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统端口",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PortUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system port",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"serverPort"
					],
					"formatEN": "update system port =\u003e [serverPort]",
					"formatZH": "修改系统端口 =\u003e [serverPort]",
					"paramKeys": []
				}
			}
		},
		"/core/settings/proxy/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "服务器代理配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ProxyUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update proxy setting",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"proxyUrl",
						"proxyPort"
					],
					"formatEN": "set proxy [proxyPort]:[proxyPort].",
					"formatZH": "服务器代理配置 [proxyPort]:[proxyPort]",
					"paramKeys": []
				}
			}
		},
		"/core/settings/search": {
			"post": {
				"description": "加载系统配置信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.SettingInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system setting info",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/search/available": {
			"get": {
				"description": "获取系统可用状态",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system available status",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/ssl/download": {
			"post": {
				"description": "下载证书",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download system cert",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/ssl/info": {
			"get": {
				"description": "获取证书信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.SettingInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system cert info",
				"tags": [
					"System Setting"
				]
			}
		},
		"/core/settings/ssl/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改系统 ssl 登录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SSLUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system ssl",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"ssl"
					],
					"formatEN": "update system ssl =\u003e [ssl]",
					"formatZH": "修改系统 ssl =\u003e [ssl]",
					"paramKeys": []
				}
			}
		},
		"/core/settings/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system setting",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update system setting [key] =\u003e [value]",
					"formatZH": "修改系统配置 [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/core/settings/upgrade": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取版本 release notes",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Upgrade"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load release notes by version",
				"tags": [
					"System Setting"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "系统更新",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Upgrade"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Upgrade",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"version"
					],
					"formatEN": "upgrade system =\u003e [version]",
					"formatZH": "更新系统 =\u003e [version]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建计划任务",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create cronjob",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type",
						"name"
					],
					"formatEN": "create cronjob [type][name]",
					"formatZH": "创建计划任务 [type][name]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除计划任务",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobBatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete cronjob",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "cronjobs",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete cronjob [names]",
					"formatZH": "删除计划任务 [names]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/download": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "下载计划任务记录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobDownload"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download cronjob records",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "job_records",
							"input_column": "id",
							"input_value": "recordID",
							"isList": false,
							"output_column": "file",
							"output_value": "file"
						}
					],
					"bodyKeys": [
						"recordID"
					],
					"formatEN": "download the cronjob record [file]",
					"formatZH": "下载计划任务记录 [file]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/handle": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "手动执行计划任务",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Handle cronjob once",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "cronjobs",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "manually execute the cronjob [name]",
					"formatZH": "手动执行计划任务 [name]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/records/clean": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清空计划任务记录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobClean"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean job records",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "cronjobs",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "clean cronjob [name] records",
					"formatZH": "清空计划任务记录 [name]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/records/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取计划任务记录日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load Cronjob record log",
				"tags": [
					"Cronjob"
				]
			}
		},
		"/cronjobs/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取计划任务分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageCronjob"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page cronjobs",
				"tags": [
					"Cronjob"
				]
			}
		},
		"/cronjobs/search/records": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取计划任务记录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchRecord"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page job records",
				"tags": [
					"Cronjob"
				]
			}
		},
		"/cronjobs/status": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新计划任务状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobUpdateStatus"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update cronjob status",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "cronjobs",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id",
						"status"
					],
					"formatEN": "change the status of cronjob [name] to [status].",
					"formatZH": "修改计划任务 [name] 状态为 [status]",
					"paramKeys": []
				}
			}
		},
		"/cronjobs/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新计划任务",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CronjobUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update cronjob",
				"tags": [
					"Cronjob"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "cronjobs",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "update cronjob [name]",
					"formatZH": "更新计划任务 [name]",
					"paramKeys": []
				}
			}
		},
		"/dashboard/base/:ioOption/:netOption": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取首页基础数据",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "ioOption",
						"required": true,
						"type": "string"
					},
					{
						"description": "request",
						"in": "path",
						"name": "netOption",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DashboardBase"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load dashboard base info",
				"tags": [
					"Dashboard"
				]
			}
		},
		"/dashboard/base/os": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取服务器基础数据",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.OsInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load os info",
				"tags": [
					"Dashboard"
				]
			}
		},
		"/dashboard/current/:ioOption/:netOption": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取首页实时数据",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "ioOption",
						"required": true,
						"type": "string"
					},
					{
						"description": "request",
						"in": "path",
						"name": "netOption",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DashboardCurrent"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load dashboard current info",
				"tags": [
					"Dashboard"
				]
			}
		},
		"/dashboard/system/restart/:operation": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "重启服务器/面板",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "operation",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "System restart",
				"tags": [
					"Dashboard"
				]
			}
		},
		"/databases": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建 mysql 数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlDBCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create mysql database",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create mysql database [name]",
					"formatZH": "创建 mysql 数据库 [name]",
					"paramKeys": []
				}
			}
		},
		"/databases/bind": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "绑定 mysql 数据库用户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BindUser"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Bind user of mysql database",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"database",
						"username"
					],
					"formatEN": "bind mysql database [database] [username]",
					"formatZH": "绑定 mysql 数据库名 [database] [username]",
					"paramKeys": []
				}
			}
		},
		"/databases/change/access": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 mysql 访问权限",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeDBInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change mysql access",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_mysqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update database [name] access",
					"formatZH": "更新数据库 [name] 访问权限",
					"paramKeys": []
				}
			}
		},
		"/databases/change/password": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 mysql 密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeDBInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change mysql password",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_mysqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update database [name] password",
					"formatZH": "更新数据库 [name] 密码",
					"paramKeys": []
				}
			}
		},
		"/databases/common/info": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取数据库基础信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DBBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load base info",
				"tags": [
					"Database Common"
				]
			}
		},
		"/databases/common/load/file": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取数据库配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load Database conf",
				"tags": [
					"Database Common"
				]
			}
		},
		"/databases/common/update/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "上传替换配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DBConfUpdateByFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update conf by upload file",
				"tags": [
					"Database Common"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type",
						"database"
					],
					"formatEN": "update the [type] [database] database configuration information",
					"formatZH": "更新 [type] 数据库 [database] 配置信息",
					"paramKeys": []
				}
			}
		},
		"/databases/db": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建远程数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DatabaseCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create database",
				"tags": [
					"Database"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"type"
					],
					"formatEN": "create database [name][type]",
					"formatZH": "创建远程数据库 [name][type]",
					"paramKeys": []
				}
			}
		},
		"/databases/db/:name": {
			"get": {
				"description": "获取远程数据库",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DatabaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get databases",
				"tags": [
					"Database"
				]
			}
		},
		"/databases/db/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "检测远程数据库连接性",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DatabaseCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check database",
				"tags": [
					"Database"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"type"
					],
					"formatEN": "check if database [name][type] is connectable",
					"formatZH": "检测远程数据库 [name][type] 连接性",
					"paramKeys": []
				}
			}
		},
		"/databases/db/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除远程数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DatabaseDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete database",
				"tags": [
					"Database"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "databases",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete database [names]",
					"formatZH": "删除远程数据库 [names]",
					"paramKeys": []
				}
			}
		},
		"/databases/db/item/:type": {
			"get": {
				"description": "获取数据库列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.DatabaseItem"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List databases",
				"tags": [
					"Database"
				]
			}
		},
		"/databases/db/list/:type": {
			"get": {
				"description": "获取远程数据库列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.DatabaseOption"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List databases",
				"tags": [
					"Database"
				]
			}
		},
		"/databases/db/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取远程数据库列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DatabaseSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page databases",
				"tags": [
					"Database"
				]
			}
		},
		"/databases/db/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新远程数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DatabaseUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update database",
				"tags": [
					"Database"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "update database [name]",
					"formatZH": "更新远程数据库 [name]",
					"paramKeys": []
				}
			}
		},
		"/databases/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除 mysql 数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlDBDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete mysql database",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_mysqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "delete mysql database [name]",
					"formatZH": "删除 mysql 数据库 [name]",
					"paramKeys": []
				}
			}
		},
		"/databases/del/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Mysql 数据库删除前检查",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlDBDeleteCheck"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check before delete mysql database",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/description/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 mysql 数据库库描述信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateDescription"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update mysql database description",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_mysqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id",
						"description"
					],
					"formatEN": "The description of the mysql database [name] is modified =\u003e [description]",
					"formatZH": "mysql 数据库 [name] 描述信息修改 [description]",
					"paramKeys": []
				}
			}
		},
		"/databases/load": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "从服务器获取",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlLoadDB"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load mysql database from remote",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/options": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mysql 数据库列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.MysqlOption"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List mysql database names",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/pg": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建 postgresql 数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlDBCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create postgresql database",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "create postgresql database [name]",
					"formatZH": "创建 postgresql 数据库 [name]",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/:database/load": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "从服务器获取",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlLoadDB"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load postgresql database from remote",
				"tags": [
					"Database Postgresql"
				]
			}
		},
		"/databases/pg/bind": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "绑定 postgresql 数据库用户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlBindUser"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Bind postgresql user",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"username"
					],
					"formatEN": "bind postgresql database [name] user [username]",
					"formatZH": "绑定 postgresql 数据库 [name] 用户 [username]",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除 postgresql 数据库",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlDBDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete postgresql database",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_postgresqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "delete postgresql database [name]",
					"formatZH": "删除 postgresql 数据库 [name]",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/del/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Postgresql 数据库删除前检查",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlDBDeleteCheck"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check before delete postgresql database",
				"tags": [
					"Database Postgresql"
				]
			}
		},
		"/databases/pg/description": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 postgresql 数据库库描述信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateDescription"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update postgresql database description",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_postgresqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id",
						"description"
					],
					"formatEN": "The description of the postgresql database [name] is modified =\u003e [description]",
					"formatZH": "postgresql 数据库 [name] 描述信息修改 [description]",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/password": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 postgresql 密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeDBInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change postgresql password",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "database_postgresqls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update database [name] password",
					"formatZH": "更新数据库 [name] 密码",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/privileges": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 postgresql 用户权限",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeDBInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change postgresql privileges",
				"tags": [
					"Database Postgresql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"database",
						"username"
					],
					"formatEN": "Update [user] privileges of database [database]",
					"formatZH": "更新数据库 [database] 用户 [username] 权限",
					"paramKeys": []
				}
			}
		},
		"/databases/pg/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 postgresql 数据库列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PostgresqlDBSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page postgresql databases",
				"tags": [
					"Database Postgresql"
				]
			}
		},
		"/databases/redis/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 redis 配置信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.RedisConf"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load redis conf",
				"tags": [
					"Database Redis"
				]
			}
		},
		"/databases/redis/conf/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 redis 配置信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RedisConfUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update redis conf",
				"tags": [
					"Database Redis"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "update the redis database configuration information",
					"formatZH": "更新 redis 数据库配置信息",
					"paramKeys": []
				}
			}
		},
		"/databases/redis/install/cli": {
			"post": {
				"description": "安装 redis cli",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Install redis-cli",
				"tags": [
					"Database Redis"
				]
			}
		},
		"/databases/redis/password": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 redis 密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeRedisPass"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change redis password",
				"tags": [
					"Database Redis"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "change the password of the redis database",
					"formatZH": "修改 redis 数据库密码",
					"paramKeys": []
				}
			}
		},
		"/databases/redis/persistence/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 redis 持久化配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.RedisPersistence"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load redis persistence conf",
				"tags": [
					"Database Redis"
				]
			}
		},
		"/databases/redis/persistence/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 redis 持久化配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RedisConfPersistenceUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update redis persistence conf",
				"tags": [
					"Database Redis"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "redis database persistence configuration update",
					"formatZH": "redis 数据库持久化配置更新",
					"paramKeys": []
				}
			}
		},
		"/databases/redis/status": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 redis 状态信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.RedisStatus"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load redis status info",
				"tags": [
					"Database Redis"
				]
			}
		},
		"/databases/remote": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mysql 远程访问权限",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "boolean"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load mysql remote access",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mysql 数据库列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlDBSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page mysql databases",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/status": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mysql 状态信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.MysqlStatus"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load mysql status info",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/variables": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 mysql 性能参数信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithNameAndType"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.MysqlVariables"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load mysql variables info",
				"tags": [
					"Database Mysql"
				]
			}
		},
		"/databases/variables/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "mysql 性能调优",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MysqlVariablesUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update mysql variables",
				"tags": [
					"Database Mysql"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "adjust mysql database performance parameters",
					"formatZH": "调整 mysql 数据库性能参数",
					"paramKeys": []
				}
			}
		},
		"/db/remote/del/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Mysql 远程数据库删除前检查",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check before delete remote database",
				"tags": [
					"Database"
				]
			}
		},
		"/files": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建文件/文件夹",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Create dir or file [path]",
					"formatZH": "创建文件/文件夹 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/batch/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "批量删除文件/文件夹",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileBatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Batch delete file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"paths"
					],
					"formatEN": "Batch delete dir or file [paths]",
					"formatZH": "批量删除文件/文件夹 [paths]",
					"paramKeys": []
				}
			}
		},
		"/files/batch/role": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "批量修改文件权限和用户/组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileRoleReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Batch change file mode and owner",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"paths",
						"mode",
						"user",
						"group"
					],
					"formatEN": "Batch change file mode and owner [paths] =\u003e [mode]/[user]/[group]",
					"formatZH": "批量修改文件权限和用户/组 [paths] =\u003e [mode]/[user]/[group]",
					"paramKeys": []
				}
			}
		},
		"/files/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "检测文件是否存在",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FilePathCheck"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check file exist",
				"tags": [
					"File"
				]
			}
		},
		"/files/chunkdownload": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "分片下载下载文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileDownload"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Chunk Download file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Download file [name]",
					"formatZH": "下载文件 [name]",
					"paramKeys": []
				}
			}
		},
		"/files/chunkupload": {
			"post": {
				"description": "分片上传文件",
				"parameters": [
					{
						"description": "request",
						"in": "formData",
						"name": "file",
						"required": true,
						"type": "file"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "ChunkUpload file",
				"tags": [
					"File"
				]
			}
		},
		"/files/compress": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "压缩文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileCompress"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Compress file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Compress file [name]",
					"formatZH": "压缩文件 [name]",
					"paramKeys": []
				}
			}
		},
		"/files/content": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取文件内容",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileContentReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.FileInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load file content",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Load file content [path]",
					"formatZH": "获取文件内容 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/decompress": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "解压文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileDeCompress"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Decompress file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Decompress file [path]",
					"formatZH": "解压 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除文件/文件夹",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Delete dir or file [path]",
					"formatZH": "删除文件/文件夹 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/download": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "下载文件",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download file",
				"tags": [
					"File"
				]
			}
		},
		"/files/favorite": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建收藏",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FavoriteCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create favorite",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "收藏文件/文件夹 [path]",
					"formatZH": "收藏文件/文件夹 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/favorite/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除收藏",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FavoriteDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete favorite",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "favorites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "path",
							"output_value": "path"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "delete avorite [path]",
					"formatZH": "删除收藏 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/favorite/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取收藏列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List favorites",
				"tags": [
					"File"
				]
			}
		},
		"/files/mode": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改文件权限",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change file mode",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path",
						"mode"
					],
					"formatEN": "Change mode [paths] =\u003e [mode]",
					"formatZH": "修改权限 [paths] =\u003e [mode]",
					"paramKeys": []
				}
			}
		},
		"/files/move": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "移动文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileMove"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Move file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"oldPaths",
						"newPath"
					],
					"formatEN": "Move [oldPaths] =\u003e [newPath]",
					"formatZH": "移动文件 [oldPaths] =\u003e [newPath]",
					"paramKeys": []
				}
			}
		},
		"/files/owner": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改文件用户/组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileRoleUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change file owner",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path",
						"user",
						"group"
					],
					"formatEN": "Change owner [paths] =\u003e [user]/[group]",
					"formatZH": "修改用户/组 [paths] =\u003e [user]/[group]",
					"paramKeys": []
				}
			}
		},
		"/files/read": {
			"post": {
				"description": "按行读取日志文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileReadByLineReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Read file by Line",
				"tags": [
					"File"
				]
			}
		},
		"/files/recycle/clear": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清空回收站文件",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clear RecycleBin files",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "清空回收站",
					"formatZH": "清空回收站",
					"paramKeys": []
				}
			}
		},
		"/files/recycle/reduce": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "还原回收站文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RecycleBinReduce"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Reduce RecycleBin files",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Reduce RecycleBin file [name]",
					"formatZH": "还原回收站文件 [name]",
					"paramKeys": []
				}
			}
		},
		"/files/recycle/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取回收站文件列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List RecycleBin files",
				"tags": [
					"File"
				]
			}
		},
		"/files/recycle/status": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取回收站状态",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get RecycleBin status",
				"tags": [
					"File"
				]
			}
		},
		"/files/rename": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改文件名称",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileRename"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change file name",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"oldName",
						"newName"
					],
					"formatEN": "Rename [oldName] =\u003e [newName]",
					"formatZH": "重命名 [oldName] =\u003e [newName]",
					"paramKeys": []
				}
			}
		},
		"/files/save": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新文件内容",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileEdit"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update file content",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Update file content [path]",
					"formatZH": "更新文件内容 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取文件列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileOption"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.FileInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List files",
				"tags": [
					"File"
				]
			}
		},
		"/files/size": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取文件夹大小",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.DirSizeReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load file size",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Load file size [path]",
					"formatZH": "获取文件夹大小 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/tree": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "加载文件树",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileOption"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.FileTree"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load files tree",
				"tags": [
					"File"
				]
			}
		},
		"/files/upload": {
			"post": {
				"description": "上传文件",
				"parameters": [
					{
						"description": "request",
						"in": "formData",
						"name": "file",
						"required": true,
						"type": "file"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Upload file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"path"
					],
					"formatEN": "Upload file [path]",
					"formatZH": "上传文件 [path]",
					"paramKeys": []
				}
			}
		},
		"/files/upload/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "分页获取上传文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.SearchUploadWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.FileInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page file",
				"tags": [
					"File"
				]
			}
		},
		"/files/wget": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "下载远端文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.FileWget"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Wget file",
				"tags": [
					"File"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"url",
						"path",
						"name"
					],
					"formatEN": "Download url =\u003e [path]/[name]",
					"formatZH": "下载 url =\u003e [path]/[name]",
					"paramKeys": []
				}
			}
		},
		"/groups": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建系统组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.GroupCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"System Group"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"type"
					],
					"formatEN": "create group [name][type]",
					"formatZH": "创建组 [name][type]",
					"paramKeys": []
				}
			}
		},
		"/groups/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除系统组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete group",
				"tags": [
					"System Group"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "groups",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						},
						{
							"db": "groups",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "type",
							"output_value": "type"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "delete group [type][name]",
					"formatZH": "删除组 [type][name]",
					"paramKeys": []
				}
			}
		},
		"/groups/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "查询系统组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.GroupSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.GroupInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List groups",
				"tags": [
					"System Group"
				]
			}
		},
		"/groups/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.GroupUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update group",
				"tags": [
					"System Group"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"type"
					],
					"formatEN": "update group [name][type]",
					"formatZH": "更新组 [name][type]",
					"paramKeys": []
				}
			}
		},
		"/host/conffile/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "上传文件更新 SSH 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SSHConf"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update host SSH setting by file",
				"tags": [
					"SSH"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "update SSH conf",
					"formatZH": "修改 SSH 配置文件",
					"paramKeys": []
				}
			}
		},
		"/host/ssh/conf": {
			"get": {
				"description": "获取 SSH 配置文件",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load host SSH conf",
				"tags": [
					"SSH"
				]
			}
		},
		"/host/ssh/generate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "生成 SSH 密钥",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.GenerateSSH"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Generate host SSH secret",
				"tags": [
					"SSH"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "generate SSH secret",
					"formatZH": "生成 SSH 密钥 ",
					"paramKeys": []
				}
			}
		},
		"/host/ssh/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 SSH 登录日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchSSHLog"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.SSHLog"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load host SSH logs",
				"tags": [
					"SSH"
				]
			}
		},
		"/host/ssh/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 SSH 服务状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Operate"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate SSH",
				"tags": [
					"SSH"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] SSH",
					"formatZH": "[operation] SSH ",
					"paramKeys": []
				}
			}
		},
		"/host/ssh/search": {
			"post": {
				"description": "加载 SSH 配置信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.SSHInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load host SSH setting info",
				"tags": [
					"SSH"
				]
			}
		},
		"/host/ssh/secret": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 SSH 密钥",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.GenerateLoad"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load host SSH secret",
				"tags": [
					"SSH"
				]
			}
		},
		"/host/ssh/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 SSH 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SSHUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update host SSH setting",
				"tags": [
					"SSH"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update SSH setting [key] =\u003e [value]",
					"formatZH": "修改 SSH 配置 [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/host/tool": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取主机工具状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.HostToolReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get tool",
				"tags": [
					"Host tool"
				]
			}
		},
		"/host/tool/config": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作主机工具配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.HostToolConfig"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get tool config",
				"tags": [
					"Host tool"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operate"
					],
					"formatEN": "[operate] tool config",
					"formatZH": "[operate] 主机工具配置文件 ",
					"paramKeys": []
				}
			}
		},
		"/host/tool/create": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建主机工具配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.HostToolCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create Host tool Config",
				"tags": [
					"Host tool"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type"
					],
					"formatEN": "create [type] config",
					"formatZH": "创建 [type] 配置",
					"paramKeys": []
				}
			}
		},
		"/host/tool/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取主机工具日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.HostToolLogReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get tool",
				"tags": [
					"Host tool"
				]
			}
		},
		"/host/tool/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作主机工具",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.HostToolReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate tool",
				"tags": [
					"Host tool"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operate",
						"type"
					],
					"formatEN": "[operate] [type]",
					"formatZH": "[operate] [type] ",
					"paramKeys": []
				}
			}
		},
		"/host/tool/supervisor/process": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 Supervisor 进程配置",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get Supervisor process config",
				"tags": [
					"Host tool"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作守护进程",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.SupervisorProcessConfig"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create Supervisor process",
				"tags": [
					"Host tool"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operate"
					],
					"formatEN": "[operate] process",
					"formatZH": "[operate] 守护进程 ",
					"paramKeys": []
				}
			}
		},
		"/host/tool/supervisor/process/file": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作 Supervisor 进程文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.SupervisorProcessFileReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get Supervisor process config",
				"tags": [
					"Host tool"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operate"
					],
					"formatEN": "[operate] Supervisor Process Config file",
					"formatZH": "[operate] Supervisor 进程文件 ",
					"paramKeys": []
				}
			}
		},
		"/hosts": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建主机",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.HostOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create host",
				"tags": [
					"Host"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"addr"
					],
					"formatEN": "create host [name][addr]",
					"formatZH": "创建主机 [name][addr]",
					"paramKeys": []
				}
			}
		},
		"/hosts/command": {
			"get": {
				"description": "获取快速命令列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.CommandInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List commands",
				"tags": [
					"Command"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建快速命令",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CommandOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create command",
				"tags": [
					"Command"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"command"
					],
					"formatEN": "create quick command [name][command]",
					"formatZH": "创建快捷命令 [name][command]",
					"paramKeys": []
				}
			}
		},
		"/hosts/command/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除快速命令",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete command",
				"tags": [
					"Command"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "commands",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete quick command [names]",
					"formatZH": "删除快捷命令 [names]",
					"paramKeys": []
				}
			}
		},
		"/hosts/command/redis": {
			"get": {
				"description": "获取 redis 快速命令列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "Array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List redis commands",
				"tags": [
					"Redis Command"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "保存 Redis 快速命令",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RedisCommand"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Save redis command",
				"tags": [
					"Redis Command"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"command"
					],
					"formatEN": "save quick command for redis [name][command]",
					"formatZH": "保存 redis 快捷命令 [name][command]",
					"paramKeys": []
				}
			}
		},
		"/hosts/command/redis/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除 redis 快速命令",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete redis command",
				"tags": [
					"Redis Command"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "redis_commands",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete quick command of redis [names]",
					"formatZH": "删除 redis 快捷命令 [names]",
					"paramKeys": []
				}
			}
		},
		"/hosts/command/redis/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 redis 快速命令列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page redis commands",
				"tags": [
					"Redis Command"
				]
			}
		},
		"/hosts/command/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取快速命令列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page commands",
				"tags": [
					"Command"
				]
			}
		},
		"/hosts/command/tree": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取快速命令树",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "Array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Tree commands",
				"tags": [
					"Command"
				]
			}
		},
		"/hosts/command/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新快速命令",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CommandOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update command",
				"tags": [
					"Command"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "update quick command [name]",
					"formatZH": "更新快捷命令 [name]",
					"paramKeys": []
				}
			}
		},
		"/hosts/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除主机",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete host",
				"tags": [
					"Host"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "hosts",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "addr",
							"output_value": "addrs"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete host [addrs]",
					"formatZH": "删除主机 [addrs]",
					"paramKeys": []
				}
			}
		},
		"/hosts/firewall/base": {
			"get": {
				"description": "获取防火墙基础信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.FirewallBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load firewall base info",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/firewall/batch": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "批量删除防火墙规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchRuleOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/firewall/forward": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新防火墙端口转发规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ForwardRuleOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"source_port"
					],
					"formatEN": "update port forward rules [source_port]",
					"formatZH": "更新端口转发规则 [source_port]",
					"paramKeys": []
				}
			}
		},
		"/hosts/firewall/ip": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建防火墙 IP 规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.AddrRuleOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"strategy",
						"address"
					],
					"formatEN": "create address rules [strategy][address]",
					"formatZH": "添加 ip 规则 [strategy] [address]",
					"paramKeys": []
				}
			}
		},
		"/hosts/firewall/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改防火墙状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.FirewallOperation"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page firewall status",
				"tags": [
					"Firewall"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] firewall",
					"formatZH": "[operation] 防火墙",
					"paramKeys": []
				}
			}
		},
		"/hosts/firewall/port": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建防火墙端口规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PortRuleOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"port",
						"strategy"
					],
					"formatEN": "create port rules [strategy][port]",
					"formatZH": "添加端口规则 [strategy] [port]",
					"paramKeys": []
				}
			}
		},
		"/hosts/firewall/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取防火墙规则列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RuleSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page firewall rules",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/firewall/update/addr": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 ip 防火墙规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.AddrRuleUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/firewall/update/description": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新防火墙描述",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateFirewallDescription"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update rule description",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/firewall/update/port": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新端口防火墙规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PortRuleUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create group",
				"tags": [
					"Firewall"
				]
			}
		},
		"/hosts/monitor/clean": {
			"post": {
				"description": "清空监控数据",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean monitor datas",
				"tags": [
					"Monitor"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "clean monitor datas",
					"formatZH": "清空监控数据",
					"paramKeys": []
				}
			}
		},
		"/hosts/monitor/search": {
			"post": {
				"description": "获取监控数据",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.MonitorSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load monitor datas",
				"tags": [
					"Monitor"
				]
			}
		},
		"/hosts/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取主机列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchHostWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.HostTree"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page host",
				"tags": [
					"Host"
				]
			}
		},
		"/hosts/test/byid/:id": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "测试主机连接",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "boolean"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Test host conn by host id",
				"tags": [
					"Host"
				]
			}
		},
		"/hosts/test/byinfo": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "测试主机连接",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.HostConnTest"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Test host conn by info",
				"tags": [
					"Host"
				]
			}
		},
		"/hosts/tree": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "加载主机树",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchForTree"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.HostTree"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load host tree",
				"tags": [
					"Host"
				]
			}
		},
		"/hosts/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新主机",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.HostOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update host",
				"tags": [
					"Host"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"addr"
					],
					"formatEN": "update host [name][addr]",
					"formatZH": "更新主机信息 [name][addr]",
					"paramKeys": []
				}
			}
		},
		"/hosts/update/group": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "切换分组",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangeHostGroup"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update host group",
				"tags": [
					"Host"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "hosts",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "addr",
							"output_value": "addr"
						}
					],
					"bodyKeys": [
						"id",
						"group"
					],
					"formatEN": "change host [addr] group =\u003e [group]",
					"formatZH": "切换主机[addr]分组 =\u003e [group]",
					"paramKeys": []
				}
			}
		},
		"/logs/system": {
			"post": {
				"description": "获取系统日志",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system logs",
				"tags": [
					"Logs"
				]
			}
		},
		"/logs/system/files": {
			"get": {
				"description": "获取系统日志文件列表",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system log files",
				"tags": [
					"Logs"
				]
			}
		},
		"/logs/tasks/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取任务日志列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchTaskLogReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page task logs",
				"tags": [
					"TaskLog"
				]
			}
		},
		"/openresty": {
			"get": {
				"description": "获取 OpenResty 配置信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.FileInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load OpenResty conf",
				"tags": [
					"OpenResty"
				]
			}
		},
		"/openresty/clear": {
			"post": {
				"description": "清理 OpenResty 代理缓存",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clear OpenResty proxy cache",
				"tags": [
					"OpenResty"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Clear nginx proxy cache",
					"formatZH": "清理 Openresty 代理缓存",
					"paramKeys": []
				}
			}
		},
		"/openresty/file": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "上传更新 OpenResty 配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxConfigFileUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update OpenResty conf by upload file",
				"tags": [
					"OpenResty"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Update nginx conf",
					"formatZH": "更新 nginx 配置",
					"paramKeys": []
				}
			}
		},
		"/openresty/scope": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取部分 OpenResty 配置信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxScopeReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.NginxParam"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load partial OpenResty conf",
				"tags": [
					"OpenResty"
				]
			}
		},
		"/openresty/status": {
			"get": {
				"description": "获取 OpenResty 状态信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.NginxStatus"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load OpenResty status info",
				"tags": [
					"OpenResty"
				]
			}
		},
		"/openresty/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 OpenResty 配置信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxConfigUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update OpenResty conf",
				"tags": [
					"OpenResty"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteId",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteId"
					],
					"formatEN": "Update nginx conf [domain]",
					"formatZH": "更新 nginx 配置 [domain]",
					"paramKeys": []
				}
			}
		},
		"/process/stop": {
			"post": {
				"description": "停止进程",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.ProcessReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Stop Process",
				"tags": [
					"Process"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"PID"
					],
					"formatEN": "结束进程 [PID]",
					"formatZH": "结束进程 [PID]",
					"paramKeys": []
				}
			}
		},
		"/runtimes": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建运行环境",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RuntimeCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create runtime",
				"tags": [
					"Runtime"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Create runtime [name]",
					"formatZH": "创建运行环境 [name]",
					"paramKeys": []
				}
			}
		},
		"/runtimes/:id": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取运行环境",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "string"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get runtime",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除运行环境",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RuntimeDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete runtime",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete website [name]",
					"formatZH": "删除网站 [name]",
					"paramKeys": []
				}
			}
		},
		"/runtimes/node/modules": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 Node 项目的 modules",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NodeModuleReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get Node modules",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/node/modules/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作 Node 项目 modules",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NodeModuleReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate Node modules",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/node/package": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 Node 项目的 scripts",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NodePackageReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get Node package scripts",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作运行环境",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RuntimeOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate runtime",
				"tags": [
					"Runtime"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Operate runtime [name]",
					"formatZH": "操作运行环境 [name]",
					"paramKeys": []
				}
			}
		},
		"/runtimes/php/extensions": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Create Extensions",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.PHPExtensionsCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create Extensions",
				"tags": [
					"PHP Extensions"
				]
			}
		},
		"/runtimes/php/extensions/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Delete Extensions",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.PHPExtensionsDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete Extensions",
				"tags": [
					"PHP Extensions"
				]
			}
		},
		"/runtimes/php/extensions/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Page Extensions",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.PHPExtensionsSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.PHPExtensionsDTO"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page Extensions",
				"tags": [
					"PHP Extensions"
				]
			}
		},
		"/runtimes/php/extensions/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "Update Extensions",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.PHPExtensionsUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update Extensions",
				"tags": [
					"PHP Extensions"
				]
			}
		},
		"/runtimes/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取运行环境列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RuntimeSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List runtimes",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/sync": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "同步运行环境状态",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Sync runtime status",
				"tags": [
					"Runtime"
				]
			}
		},
		"/runtimes/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新运行环境",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.RuntimeUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update runtime",
				"tags": [
					"Runtime"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Update runtime [name]",
					"formatZH": "更新运行环境 [name]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建备份账号",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BackupOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create backup account",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type"
					],
					"formatEN": "create backup account [type]",
					"formatZH": "创建备份账号 [type]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/backup": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "备份系统数据",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CommonBackup"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Backup system data",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type",
						"name",
						"detailName"
					],
					"formatEN": "backup [type] data [name][detailName]",
					"formatZH": "备份 [type] 数据 [name][detailName]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除备份账号",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete backup account",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "backup_accounts",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "type",
							"output_value": "types"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "delete backup account [types]",
					"formatZH": "删除备份账号 [types]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/onedrive": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 OneDrive 信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.OneDriveInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load OneDrive info",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/record/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除备份记录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete backup record",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "backup_records",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "file_name",
							"output_value": "files"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete backup records [files]",
					"formatZH": "删除备份记录 [files]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/record/download": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "下载备份记录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.DownloadRecord"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download backup record",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"source",
						"fileName"
					],
					"formatEN": "download backup records [source][fileName]",
					"formatZH": "下载备份记录 [source][fileName]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/record/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取备份记录列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RecordSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page backup records",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/record/search/bycronjob": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "通过计划任务获取备份记录列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.RecordSearchByCronjob"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page backup records by cronjob",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/recover": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "恢复系统数据",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CommonRecover"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Recover system data",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type",
						"name",
						"detailName",
						"file"
					],
					"formatEN": "recover [type] data [name][detailName] from [file]",
					"formatZH": "从 [file] 恢复 [type] 数据 [name][detailName]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/recover/byupload": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "从上传恢复系统数据",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.CommonRecover"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Recover system data by upload",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type",
						"name",
						"detailName",
						"file"
					],
					"formatEN": "recover [type] data [name][detailName] from [file]",
					"formatZH": "从 [file] 恢复 [type] 数据 [name][detailName]",
					"paramKeys": []
				}
			}
		},
		"/settings/backup/refresh/onedrive": {
			"post": {
				"description": "刷新 OneDrive token",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Refresh OneDrive token",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/search": {
			"get": {
				"description": "获取备份账号列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.BackupInfo"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List backup accounts",
				"tags": [
					"Backup Account"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 bucket 列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ForBuckets"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List buckets",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/search/files": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取备份账号内文件列表",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BackupSearchFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List files from backup accounts",
				"tags": [
					"Backup Account"
				]
			}
		},
		"/settings/backup/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新备份账号信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BackupOperate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update backup account",
				"tags": [
					"Backup Account"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type"
					],
					"formatEN": "update backup account [types]",
					"formatZH": "更新备份账号 [types]",
					"paramKeys": []
				}
			}
		},
		"/settings/basedir": {
			"get": {
				"description": "获取安装根目录",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "string"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load local backup dir",
				"tags": [
					"System Setting"
				]
			}
		},
		"/settings/search": {
			"post": {
				"description": "加载系统配置信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.SettingInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system setting info",
				"tags": [
					"System Setting"
				]
			}
		},
		"/settings/search/available": {
			"get": {
				"description": "获取系统可用状态",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load system available status",
				"tags": [
					"System Setting"
				]
			}
		},
		"/settings/snapshot": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建系统快照",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SnapshotCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create system snapshot",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"from",
						"description"
					],
					"formatEN": "Create system backup [description] to [from]",
					"formatZH": "创建系统快照 [description] 到 [from]",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除系统快照",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SnapshotBatchDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete system backup",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "snapshots",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "Delete system backup [name]",
					"formatZH": "删除系统快照 [name]",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/description/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新快照描述信息",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateDescription"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update snapshot description",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "snapshots",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id",
						"description"
					],
					"formatEN": "The description of the snapshot [name] is modified =\u003e [description]",
					"formatZH": "快照 [name] 描述信息修改 [description]",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/import": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "导入已有快照",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SnapshotImport"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Import system snapshot",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"from",
						"names"
					],
					"formatEN": "Sync system snapshots [names] from [from]",
					"formatZH": "从 [from] 同步系统快照 [names]",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/recover": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "从系统快照恢复",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SnapshotRecover"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Recover system backup",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "snapshots",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Recover from system backup [name]",
					"formatZH": "从系统快照 [name] 恢复",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/rollback": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "从系统快照回滚",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SnapshotRecover"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Rollback system backup",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "snapshots",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Rollback from system backup [name]",
					"formatZH": "从系统快照 [name] 回滚",
					"paramKeys": []
				}
			}
		},
		"/settings/snapshot/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统快照列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page system snapshot",
				"tags": [
					"System Setting"
				]
			}
		},
		"/settings/snapshot/status": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取快照状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load Snapshot status",
				"tags": [
					"System Setting"
				]
			}
		},
		"/settings/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新系统配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update system setting",
				"tags": [
					"System Setting"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update system setting [key] =\u003e [value]",
					"formatZH": "修改系统配置 [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建扫描规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create clam",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"path"
					],
					"formatEN": "create clam [name][path]",
					"formatZH": "创建扫描规则 [name][path]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/base": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 Clam 基础信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.ClamBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load clam base info",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除扫描规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete clam",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "clams",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "name",
							"output_value": "names"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete clam [names]",
					"formatZH": "删除扫描规则 [names]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/file/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取扫描文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamFileReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load clam file",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/file/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新病毒扫描配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateByNameAndFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update clam file",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/handle": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "执行病毒扫描",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Handle clam scan",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "clams",
							"input_column": "id",
							"input_value": "id",
							"isList": true,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "handle clam scan [name]",
					"formatZH": "执行病毒扫描 [name]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 Clam 状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Operate"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate Clam",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] FTP",
					"formatZH": "[operation] Clam",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/record/clean": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清空扫描报告",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperateByID"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean clam record",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "clams",
							"input_column": "id",
							"input_value": "id",
							"isList": true,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "clean clam record [name]",
					"formatZH": "清空扫描报告 [name]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/record/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取扫描结果详情",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamLogReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load clam record detail",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/record/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取扫描结果列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamLogSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page clam record",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取扫描规则列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchClamWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page clam",
				"tags": [
					"Clam"
				]
			}
		},
		"/toolbox/clam/status/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改扫描规则状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamUpdateStatus"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update clam status",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "clams",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id",
						"status"
					],
					"formatEN": "change the status of clam [name] to [status].",
					"formatZH": "修改扫描规则 [name] 状态为 [status]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clam/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改扫描规则",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ClamUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update clam",
				"tags": [
					"Clam"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name",
						"path"
					],
					"formatEN": "update clam [name][path]",
					"formatZH": "修改扫描规则 [name][path]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/clean": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "清理系统垃圾文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"items": {
								"$ref": "#/definitions/dto.Clean"
							},
							"type": "array"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Clean system",
				"tags": [
					"Device"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "Clean system junk files",
					"formatZH": "清理系统垃圾文件",
					"paramKeys": []
				}
			}
		},
		"/toolbox/device/base": {
			"post": {
				"description": "获取设备基础信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.DeviceBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load device base info",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/device/check/dns": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "检查系统 DNS 配置可用性",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check device DNS conf",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/device/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.OperationWithName"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "load conf",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/device/update/byconf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "通过文件修改配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateByNameAndFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update device conf by file",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/device/update/conf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改系统参数",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SettingUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update device",
				"tags": [
					"Device"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update device conf [key] =\u003e [value]",
					"formatZH": "修改主机参数 [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/device/update/host": {
			"post": {
				"description": "修改系统 hosts",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update device hosts",
				"tags": [
					"Device"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update device host [key] =\u003e [value]",
					"formatZH": "修改主机 Host [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/device/update/passwd": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改系统密码",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.ChangePasswd"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update device passwd",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/device/update/swap": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改系统 Swap",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SwapHelper"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update device swap",
				"tags": [
					"Device"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operate",
						"path"
					],
					"formatEN": "[operate] device swap [path]",
					"formatZH": "[operate] 主机 swap [path]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/device/zone/options": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取系统可用时区选项",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "Array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "list time zone options",
				"tags": [
					"Device"
				]
			}
		},
		"/toolbox/fail2ban/base": {
			"get": {
				"description": "获取 Fail2ban 基础信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.Fail2BanBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load fail2ban base info",
				"tags": [
					"Fail2ban"
				]
			}
		},
		"/toolbox/fail2ban/load/conf": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 fail2ban 配置文件",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load fail2ban conf",
				"tags": [
					"Fail2ban"
				]
			}
		},
		"/toolbox/fail2ban/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 Fail2ban 状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Operate"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate fail2ban",
				"tags": [
					"Fail2ban"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] Fail2ban",
					"formatZH": "[operation] Fail2ban",
					"paramKeys": []
				}
			}
		},
		"/toolbox/fail2ban/operate/sshd": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "配置 sshd",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Operate"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate sshd of fail2ban",
				"tags": [
					"Fail2ban"
				]
			}
		},
		"/toolbox/fail2ban/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 Fail2ban ip",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Fail2BanSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"type": "Array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page fail2ban ip list",
				"tags": [
					"Fail2ban"
				]
			}
		},
		"/toolbox/fail2ban/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 Fail2ban 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Fail2BanUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update fail2ban conf",
				"tags": [
					"Fail2ban"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"key",
						"value"
					],
					"formatEN": "update fail2ban conf [key] =\u003e [value]",
					"formatZH": "修改 Fail2ban 配置 [key] =\u003e [value]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/fail2ban/update/byconf": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "通过文件修改 fail2ban 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.UpdateByFile"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update fail2ban conf by file",
				"tags": [
					"Fail2ban"
				]
			}
		},
		"/toolbox/ftp": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建 FTP 账户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.FtpCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create FTP user",
				"tags": [
					"FTP"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"user",
						"path"
					],
					"formatEN": "create FTP [user][path]",
					"formatZH": "创建 FTP 账户 [user][path]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/ftp/base": {
			"get": {
				"description": "获取 FTP 基础信息",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.FtpBaseInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load FTP base info",
				"tags": [
					"FTP"
				]
			}
		},
		"/toolbox/ftp/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除 FTP 账户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete FTP user",
				"tags": [
					"FTP"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "ftps",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "user",
							"output_value": "users"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "delete FTP users [users]",
					"formatZH": "删除 FTP 账户 [users]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/ftp/log/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 FTP 操作日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.FtpLogSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load FTP operation log",
				"tags": [
					"FTP"
				]
			}
		},
		"/toolbox/ftp/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 FTP 状态",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.Operate"
						}
					}
				],
				"responses": {},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate FTP",
				"tags": [
					"FTP"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"operation"
					],
					"formatEN": "[operation] FTP",
					"formatZH": "[operation] FTP",
					"paramKeys": []
				}
			}
		},
		"/toolbox/ftp/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 FTP 账户列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.SearchWithPage"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page FTP user",
				"tags": [
					"FTP"
				]
			}
		},
		"/toolbox/ftp/sync": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "同步 FTP 账户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.BatchDeleteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Sync FTP user",
				"tags": [
					"FTP"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "sync FTP users",
					"formatZH": "同步 FTP 账户",
					"paramKeys": []
				}
			}
		},
		"/toolbox/ftp/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改 FTP 账户",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.FtpUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update FTP user",
				"tags": [
					"FTP"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"user",
						"path"
					],
					"formatEN": "update FTP [user][path]",
					"formatZH": "修改 FTP 账户 [user][path]",
					"paramKeys": []
				}
			}
		},
		"/toolbox/scan": {
			"post": {
				"description": "扫描系统垃圾文件",
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Scan system",
				"tags": [
					"Device"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [],
					"formatEN": "scan System Junk Files",
					"formatZH": "扫描系统垃圾文件",
					"paramKeys": []
				}
			}
		},
		"/websites": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"primaryDomain"
					],
					"formatEN": "Create website [primaryDomain]",
					"formatZH": "创建网站 [primaryDomain]",
					"paramKeys": []
				}
			}
		},
		"/websites/:id": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 id 查询网站",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteDTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search website by id",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/:id/config/:type": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 id 查询网站 nginx",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.FileInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search website nginx by id",
				"tags": [
					"Website Nginx"
				]
			}
		},
		"/websites/:id/https": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取 https 配置",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteHTTPS"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load https conf",
				"tags": [
					"Website HTTPS"
				]
			},
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 https 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteHTTPSOp"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteHTTPS"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update https conf",
				"tags": [
					"Website HTTPS"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteId",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteId"
					],
					"formatEN": "Update website https [domain] conf",
					"formatZH": "更新网站 [domain] https 配置",
					"paramKeys": []
				}
			}
		},
		"/websites/acme": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站 acme",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteAcmeAccountCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteAcmeAccountDTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website acme account",
				"tags": [
					"Website Acme"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"email"
					],
					"formatEN": "Create website acme [email]",
					"formatZH": "创建网站 acme [email]",
					"paramKeys": []
				}
			}
		},
		"/websites/acme/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站 acme",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteResourceReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website acme account",
				"tags": [
					"Website Acme"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_acme_accounts",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "email",
							"output_value": "email"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete website acme [email]",
					"formatZH": "删除网站 acme [email]",
					"paramKeys": []
				}
			}
		},
		"/websites/acme/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 acme 列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page website acme accounts",
				"tags": [
					"Website Acme"
				]
			}
		},
		"/websites/auths": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取密码访问配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxAuthReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get AuthBasic conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/auths/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新密码访问配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxAuthUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get AuthBasic conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/ca": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站 ca",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCACreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/request.WebsiteCACreate"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website ca",
				"tags": [
					"Website CA"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Create website ca [name]",
					"formatZH": "创建网站 ca [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/ca/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站 ca",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCommonReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website ca",
				"tags": [
					"Website CA"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_cas",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete website ca [name]",
					"formatZH": "删除网站 ca [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/ca/download": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "下载 CA 证书文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteResourceReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download CA file",
				"tags": [
					"Website CA"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_cas",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "download ca file [name]",
					"formatZH": "下载 CA 证书文件 [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/ca/obtain": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "自签 SSL 证书",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCAObtain"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Obtain SSL",
				"tags": [
					"Website CA"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_cas",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Obtain SSL [name]",
					"formatZH": "自签 SSL 证书 [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/ca/renew": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "续签 SSL 证书",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCAObtain"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Obtain SSL",
				"tags": [
					"Website CA"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_cas",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Obtain SSL [name]",
					"formatZH": "自签 SSL 证书 [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/ca/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 ca 列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCASearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page website ca",
				"tags": [
					"Website CA"
				]
			}
		},
		"/websites/ca/{id}": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 ca",
				"parameters": [
					{
						"description": "id",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteCADTO"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get website ca",
				"tags": [
					"Website CA"
				]
			}
		},
		"/websites/check": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "网站创建前检查",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteInstallCheckReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.WebsitePreInstallCheck"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Check before create website",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/config": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取 nginx 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxScopeReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteNginxConfig"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load nginx conf",
				"tags": [
					"Website Nginx"
				]
			}
		},
		"/websites/config/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 nginx 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxConfigUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update nginx conf",
				"tags": [
					"Website Nginx"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteId",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteId"
					],
					"formatEN": "Nginx conf update [domain]",
					"formatZH": "nginx 配置修改 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/default/html/:type": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取默认 html",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.FileInfo"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get default html",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/default/html/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新默认 html",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteHtmlUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update default html",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type"
					],
					"formatEN": "Update default html",
					"formatZH": "更新默认 html",
					"paramKeys": []
				}
			}
		},
		"/websites/default/server": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作网站日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDefaultUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Change default server",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id",
						"operate"
					],
					"formatEN": "Change default server =\u003e [domain]",
					"formatZH": "修改默认 server =\u003e [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete website [domain]",
					"formatZH": "删除网站 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/dir": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站目录配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteCommonReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get website dir",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/dir/permission": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新网站目录权限",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteUpdateDirPermission"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update Site Dir permission",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update  domain [domain] dir permission",
					"formatZH": "更新网站 [domain] 目录权限",
					"paramKeys": []
				}
			}
		},
		"/websites/dir/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新网站目录",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteUpdateDir"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update Site Dir",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update  domain [domain] dir",
					"formatZH": "更新网站 [domain] 目录",
					"paramKeys": []
				}
			}
		},
		"/websites/dns": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站 dns",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDnsAccountCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website dns account",
				"tags": [
					"Website DNS"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Create website dns [name]",
					"formatZH": "创建网站 dns [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/dns/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站 dns",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteResourceReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website dns account",
				"tags": [
					"Website DNS"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_dns_accounts",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "name",
							"output_value": "name"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete website dns [name]",
					"formatZH": "删除网站 dns [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/dns/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 dns 列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/dto.PageInfo"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page website dns accounts",
				"tags": [
					"Website DNS"
				]
			}
		},
		"/websites/dns/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新网站 dns",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDnsAccountUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update website dns account",
				"tags": [
					"Website DNS"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"name"
					],
					"formatEN": "Update website dns [name]",
					"formatZH": "更新网站 dns [name]",
					"paramKeys": []
				}
			}
		},
		"/websites/domains": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站域名",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDomainCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/model.WebsiteDomain"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website domain",
				"tags": [
					"Website Domain"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"domain"
					],
					"formatEN": "Create domain [domain]",
					"formatZH": "创建域名 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/domains/:websiteId": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过网站 id 查询域名",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "websiteId",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/model.WebsiteDomain"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search website domains by websiteId",
				"tags": [
					"Website Domain"
				]
			}
		},
		"/websites/domains/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站域名",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDomainDelete"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website domain",
				"tags": [
					"Website Domain"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_domains",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Delete domain [domain]",
					"formatZH": "删除域名 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/leech": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取防盗链配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxCommonReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get AntiLeech conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/leech/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新防盗链配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxAntiLeechUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update AntiLeech",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/list": {
			"get": {
				"description": "获取网站列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.WebsiteDTO"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List websites",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/log": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作网站日志",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteLogReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.WebsiteLog"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate website log",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id",
						"operate"
					],
					"formatEN": "[domain][operate] logs",
					"formatZH": "[domain][operate] 日志",
					"paramKeys": []
				}
			}
		},
		"/websites/nginx/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 网站 nginx 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteNginxUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update website nginx conf",
				"tags": [
					"Website Nginx"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "[domain] Nginx conf update",
					"formatZH": "[domain] Nginx 配置修改",
					"paramKeys": []
				}
			}
		},
		"/websites/operate": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "操作网站",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteOp"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Operate website",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id",
						"operate"
					],
					"formatEN": "[operate] website [domain]",
					"formatZH": "[operate] 网站 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/options": {
			"get": {
				"description": "获取网站列表",
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"type": "string"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "List website names",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/php/config": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 网站 PHP 配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsitePHPConfigUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update website php conf",
				"tags": [
					"Website PHP"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "[domain] PHP conf update",
					"formatZH": "[domain] PHP 配置修改",
					"paramKeys": []
				}
			}
		},
		"/websites/php/config/:id": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 php 配置",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/response.PHPConfig"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Load website php conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/php/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 php 配置文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsitePHPFileUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update php conf",
				"tags": [
					"Website PHP"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteId",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteId"
					],
					"formatEN": "Nginx conf update [domain]",
					"formatZH": "php 配置修改 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/php/version": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "变更 php 版本",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsitePHPVersionReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update php version",
				"tags": [
					"Website PHP"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteId",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteId"
					],
					"formatEN": "php version update [domain]",
					"formatZH": "php 版本变更 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/proxies": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取反向代理配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteProxyReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get proxy conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/proxies/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改反向代理配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteProxyConfig"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update proxy conf",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update domain [domain] proxy config",
					"formatZH": "修改网站 [domain] 反向代理配置 ",
					"paramKeys": []
				}
			}
		},
		"/websites/proxy/file": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新反向代理文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxProxyUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update proxy file",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteID",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteID"
					],
					"formatEN": "Nginx conf proxy file update [domain]",
					"formatZH": "更新反向代理文件 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/redirect": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取重定向配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteProxyReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get redirect conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/redirect/file": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新重定向文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxRedirectUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update redirect file",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteID",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteID"
					],
					"formatEN": "Nginx conf redirect file update [domain]",
					"formatZH": "更新重定向文件 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/redirect/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "修改重定向配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxRedirectReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update redirect conf",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteID",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteID"
					],
					"formatEN": "Update domain [domain] redirect config",
					"formatZH": "修改网站 [domain] 重定向理配置 ",
					"paramKeys": []
				}
			}
		},
		"/websites/rewrite": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取伪静态配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxRewriteReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Get rewrite conf",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/rewrite/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新伪静态配置",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.NginxRewriteUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update rewrite conf",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "websites",
							"input_column": "id",
							"input_value": "websiteID",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"websiteID"
					],
					"formatEN": "Nginx conf rewrite update [domain]",
					"formatZH": "伪静态配置修改 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/dto.PageResult"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page websites",
				"tags": [
					"Website"
				]
			}
		},
		"/websites/ssl": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "创建网站 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLCreate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLCreate"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Create website ssl",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"primaryDomain"
					],
					"formatEN": "Create website ssl [primaryDomain]",
					"formatZH": "创建网站 ssl [primaryDomain]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/:id": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过 id 查询 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search website ssl by id",
				"tags": [
					"Website SSL"
				]
			}
		},
		"/websites/ssl/del": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "删除网站 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteBatchDelReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Delete website ssl",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_ssls",
							"input_column": "id",
							"input_value": "ids",
							"isList": true,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"ids"
					],
					"formatEN": "Delete ssl [domain]",
					"formatZH": "删除 ssl [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/download": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "下载证书文件",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteResourceReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Download SSL  file",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_ssls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "download ssl file [domain]",
					"formatZH": "下载证书文件 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/obtain": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "申请证书",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLApply"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Apply  ssl",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_ssls",
							"input_column": "id",
							"input_value": "ID",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"ID"
					],
					"formatEN": "apply ssl [domain]",
					"formatZH": "申请证书  [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/resolve": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "解析网站 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteDNSReq"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK",
						"schema": {
							"items": {
								"$ref": "#/definitions/response.WebsiteDNSRes"
							},
							"type": "array"
						}
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Resolve website ssl",
				"tags": [
					"Website SSL"
				]
			}
		},
		"/websites/ssl/search": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "获取网站 ssl 列表分页",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLSearch"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Page website ssl",
				"tags": [
					"Website SSL"
				]
			}
		},
		"/websites/ssl/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update ssl",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [
						{
							"db": "website_ssls",
							"input_column": "id",
							"input_value": "id",
							"isList": false,
							"output_column": "primary_domain",
							"output_value": "domain"
						}
					],
					"bodyKeys": [
						"id"
					],
					"formatEN": "Update ssl config [domain]",
					"formatZH": "更新证书设置 [domain]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/upload": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "上传 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteSSLUpload"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Upload ssl",
				"tags": [
					"Website SSL"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"type"
					],
					"formatEN": "Upload ssl [type]",
					"formatZH": "上传 ssl [type]",
					"paramKeys": []
				}
			}
		},
		"/websites/ssl/website/:websiteId": {
			"get": {
				"consumes": [
					"application/json"
				],
				"description": "通过网站 id 查询 ssl",
				"parameters": [
					{
						"description": "request",
						"in": "path",
						"name": "websiteId",
						"required": true,
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Search website ssl by website id",
				"tags": [
					"Website SSL"
				]
			}
		},
		"/websites/update": {
			"post": {
				"consumes": [
					"application/json"
				],
				"description": "更新网站",
				"parameters": [
					{
						"description": "request",
						"in": "body",
						"name": "request",
						"required": true,
						"schema": {
							"$ref": "#/definitions/request.WebsiteUpdate"
						}
					}
				],
				"responses": {
					"200": {
						"description": "OK"
					}
				},
				"security": [
					{
						"ApiKeyAuth": []
					}
				],
				"summary": "Update website",
				"tags": [
					"Website"
				],
				"x-panel-log": {
					"BeforeFunctions": [],
					"bodyKeys": [
						"primaryDomain"
					],
					"formatEN": "Update website [primaryDomain]",
					"formatZH": "更新网站 [primaryDomain]",
					"paramKeys": []
				}
			}
		}
	},
	"definitions": {
		"dto.AddrRuleOperate": {
			"properties": {
				"address": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"operation": {
					"enum": [
						"add",
						"remove"
					],
					"type": "string"
				},
				"strategy": {
					"enum": [
						"accept",
						"drop"
					],
					"type": "string"
				}
			},
			"required": [
				"address",
				"operation",
				"strategy"
			],
			"type": "object"
		},
		"dto.AddrRuleUpdate": {
			"properties": {
				"newRule": {
					"$ref": "#/definitions/dto.AddrRuleOperate"
				},
				"oldRule": {
					"$ref": "#/definitions/dto.AddrRuleOperate"
				}
			},
			"type": "object"
		},
		"dto.AppInstallInfo": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"key": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.AppResource": {
			"properties": {
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.AppVersion": {
			"properties": {
				"detailId": {
					"type": "integer"
				},
				"dockerCompose": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.BackupInfo": {
			"properties": {
				"backupPath": {
					"type": "string"
				},
				"bucket": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"vars": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.BackupOperate": {
			"properties": {
				"accessKey": {
					"type": "string"
				},
				"backupPath": {
					"type": "string"
				},
				"bucket": {
					"type": "string"
				},
				"credential": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"vars": {
					"type": "string"
				}
			},
			"required": [
				"type",
				"vars"
			],
			"type": "object"
		},
		"dto.BackupSearchFile": {
			"properties": {
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.BatchDelete": {
			"properties": {
				"force": {
					"type": "boolean"
				},
				"names": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"required": [
				"names"
			],
			"type": "object"
		},
		"dto.BatchDeleteReq": {
			"properties": {
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"dto.BatchRuleOperate": {
			"properties": {
				"rules": {
					"items": {
						"$ref": "#/definitions/dto.PortRuleOperate"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.BindUser": {
			"properties": {
				"database": {
					"type": "string"
				},
				"db": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"permission": {
					"type": "string"
				},
				"username": {
					"type": "string"
				}
			},
			"required": [
				"database",
				"db",
				"password",
				"permission",
				"username"
			],
			"type": "object"
		},
		"dto.ChangeDBInfo": {
			"properties": {
				"database": {
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb",
						"postgresql"
					],
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"required": [
				"database",
				"from",
				"type",
				"value"
			],
			"type": "object"
		},
		"dto.ChangeHostGroup": {
			"properties": {
				"groupID": {
					"type": "integer"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"groupID",
				"id"
			],
			"type": "object"
		},
		"dto.ChangePasswd": {
			"properties": {
				"passwd": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ChangeRedisPass": {
			"properties": {
				"database": {
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"required": [
				"database"
			],
			"type": "object"
		},
		"dto.ClamBaseInfo": {
			"properties": {
				"freshIsActive": {
					"type": "boolean"
				},
				"freshIsExist": {
					"type": "boolean"
				},
				"freshVersion": {
					"type": "string"
				},
				"isActive": {
					"type": "boolean"
				},
				"isExist": {
					"type": "boolean"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ClamCreate": {
			"properties": {
				"description": {
					"type": "string"
				},
				"infectedDir": {
					"type": "string"
				},
				"infectedStrategy": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"spec": {
					"type": "string"
				},
				"status": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ClamDelete": {
			"properties": {
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				},
				"removeInfected": {
					"type": "boolean"
				},
				"removeRecord": {
					"type": "boolean"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"dto.ClamFileReq": {
			"properties": {
				"name": {
					"type": "string"
				},
				"tail": {
					"type": "string"
				}
			},
			"required": [
				"name"
			],
			"type": "object"
		},
		"dto.ClamLogReq": {
			"properties": {
				"clamName": {
					"type": "string"
				},
				"recordName": {
					"type": "string"
				},
				"tail": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ClamLogSearch": {
			"properties": {
				"clamID": {
					"type": "integer"
				},
				"endTime": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"startTime": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.ClamUpdate": {
			"properties": {
				"description": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"infectedDir": {
					"type": "string"
				},
				"infectedStrategy": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"spec": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ClamUpdateStatus": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"status": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.Clean": {
			"properties": {
				"name": {
					"type": "string"
				},
				"size": {
					"type": "integer"
				},
				"treeType": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.CommandInfo": {
			"properties": {
				"command": {
					"type": "string"
				},
				"groupBelong": {
					"type": "string"
				},
				"groupID": {
					"type": "integer"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.CommandOperate": {
			"properties": {
				"command": {
					"type": "string"
				},
				"groupBelong": {
					"type": "string"
				},
				"groupID": {
					"type": "integer"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"required": [
				"command",
				"name"
			],
			"type": "object"
		},
		"dto.CommonBackup": {
			"properties": {
				"detailName": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"secret": {
					"type": "string"
				},
				"type": {
					"enum": [
						"app",
						"mysql",
						"mariadb",
						"redis",
						"website",
						"postgresql"
					],
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.CommonRecover": {
			"properties": {
				"detailName": {
					"type": "string"
				},
				"file": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"secret": {
					"type": "string"
				},
				"source": {
					"enum": [
						"OSS",
						"S3",
						"SFTP",
						"MINIO",
						"LOCAL",
						"COS",
						"KODO",
						"OneDrive",
						"WebDAV"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"app",
						"mysql",
						"mariadb",
						"redis",
						"website",
						"postgresql"
					],
					"type": "string"
				}
			},
			"required": [
				"source",
				"type"
			],
			"type": "object"
		},
		"dto.ComposeCreate": {
			"properties": {
				"file": {
					"type": "string"
				},
				"from": {
					"enum": [
						"edit",
						"path",
						"template"
					],
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"template": {
					"type": "integer"
				}
			},
			"required": [
				"from"
			],
			"type": "object"
		},
		"dto.ComposeOperation": {
			"properties": {
				"name": {
					"type": "string"
				},
				"operation": {
					"enum": [
						"start",
						"stop",
						"down"
					],
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"withFile": {
					"type": "boolean"
				}
			},
			"required": [
				"name",
				"operation",
				"path"
			],
			"type": "object"
		},
		"dto.ComposeTemplateCreate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"required": [
				"name"
			],
			"type": "object"
		},
		"dto.ComposeTemplateInfo": {
			"properties": {
				"content": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ComposeTemplateUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.ComposeUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				}
			},
			"required": [
				"content",
				"name",
				"path"
			],
			"type": "object"
		},
		"dto.ContainerCommit": {
			"properties": {
				"author": {
					"type": "string"
				},
				"comment": {
					"type": "string"
				},
				"containerID": {
					"type": "string"
				},
				"containerName": {
					"type": "string"
				},
				"newImageName": {
					"type": "string"
				},
				"pause": {
					"type": "boolean"
				}
			},
			"required": [
				"containerID"
			],
			"type": "object"
		},
		"dto.ContainerListStats": {
			"properties": {
				"containerID": {
					"type": "string"
				},
				"cpuPercent": {
					"type": "number"
				},
				"cpuTotalUsage": {
					"type": "integer"
				},
				"memoryCache": {
					"type": "integer"
				},
				"memoryLimit": {
					"type": "integer"
				},
				"memoryPercent": {
					"type": "number"
				},
				"memoryUsage": {
					"type": "integer"
				},
				"percpuUsage": {
					"type": "integer"
				},
				"systemUsage": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.ContainerOperate": {
			"properties": {
				"autoRemove": {
					"type": "boolean"
				},
				"cmd": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"containerID": {
					"type": "string"
				},
				"cpuShares": {
					"type": "integer"
				},
				"entrypoint": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"env": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"exposedPorts": {
					"items": {
						"$ref": "#/definitions/dto.PortHelper"
					},
					"type": "array"
				},
				"forcePull": {
					"type": "boolean"
				},
				"image": {
					"type": "string"
				},
				"ipv4": {
					"type": "string"
				},
				"ipv6": {
					"type": "string"
				},
				"labels": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"memory": {
					"type": "number"
				},
				"name": {
					"type": "string"
				},
				"nanoCPUs": {
					"type": "number"
				},
				"network": {
					"type": "string"
				},
				"openStdin": {
					"type": "boolean"
				},
				"privileged": {
					"type": "boolean"
				},
				"publishAllPorts": {
					"type": "boolean"
				},
				"restartPolicy": {
					"type": "string"
				},
				"tty": {
					"type": "boolean"
				},
				"volumes": {
					"items": {
						"$ref": "#/definitions/dto.VolumeHelper"
					},
					"type": "array"
				}
			},
			"required": [
				"image",
				"name"
			],
			"type": "object"
		},
		"dto.ContainerOperation": {
			"properties": {
				"names": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"operation": {
					"enum": [
						"start",
						"stop",
						"restart",
						"kill",
						"pause",
						"unpause",
						"remove"
					],
					"type": "string"
				}
			},
			"required": [
				"names",
				"operation"
			],
			"type": "object"
		},
		"dto.ContainerPrune": {
			"properties": {
				"pruneType": {
					"enum": [
						"container",
						"image",
						"volume",
						"network",
						"buildcache"
					],
					"type": "string"
				},
				"withTagAll": {
					"type": "boolean"
				}
			},
			"required": [
				"pruneType"
			],
			"type": "object"
		},
		"dto.ContainerPruneReport": {
			"properties": {
				"deletedNumber": {
					"type": "integer"
				},
				"spaceReclaimed": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.ContainerRename": {
			"properties": {
				"name": {
					"type": "string"
				},
				"newName": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"newName"
			],
			"type": "object"
		},
		"dto.ContainerStats": {
			"properties": {
				"cache": {
					"type": "number"
				},
				"cpuPercent": {
					"type": "number"
				},
				"ioRead": {
					"type": "number"
				},
				"ioWrite": {
					"type": "number"
				},
				"memory": {
					"type": "number"
				},
				"networkRX": {
					"type": "number"
				},
				"networkTX": {
					"type": "number"
				},
				"shotTime": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ContainerUpgrade": {
			"properties": {
				"forcePull": {
					"type": "boolean"
				},
				"image": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"required": [
				"image",
				"name"
			],
			"type": "object"
		},
		"dto.CronjobBatchDelete": {
			"properties": {
				"cleanData": {
					"type": "boolean"
				},
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"dto.CronjobClean": {
			"properties": {
				"cleanData": {
					"type": "boolean"
				},
				"cronjobID": {
					"type": "integer"
				},
				"isDelete": {
					"type": "boolean"
				}
			},
			"required": [
				"cronjobID"
			],
			"type": "object"
		},
		"dto.CronjobCreate": {
			"properties": {
				"appID": {
					"type": "string"
				},
				"backupAccounts": {
					"type": "string"
				},
				"command": {
					"type": "string"
				},
				"containerName": {
					"type": "string"
				},
				"dbName": {
					"type": "string"
				},
				"dbType": {
					"type": "string"
				},
				"defaultDownload": {
					"type": "string"
				},
				"exclusionRules": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"retainCopies": {
					"minimum": 1,
					"type": "integer"
				},
				"script": {
					"type": "string"
				},
				"secret": {
					"type": "string"
				},
				"sourceDir": {
					"type": "string"
				},
				"spec": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"url": {
					"type": "string"
				},
				"website": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"spec",
				"type"
			],
			"type": "object"
		},
		"dto.CronjobDownload": {
			"properties": {
				"backupAccountID": {
					"type": "integer"
				},
				"recordID": {
					"type": "integer"
				}
			},
			"required": [
				"backupAccountID",
				"recordID"
			],
			"type": "object"
		},
		"dto.CronjobUpdate": {
			"properties": {
				"appID": {
					"type": "string"
				},
				"backupAccounts": {
					"type": "string"
				},
				"command": {
					"type": "string"
				},
				"containerName": {
					"type": "string"
				},
				"dbName": {
					"type": "string"
				},
				"dbType": {
					"type": "string"
				},
				"defaultDownload": {
					"type": "string"
				},
				"exclusionRules": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"retainCopies": {
					"minimum": 1,
					"type": "integer"
				},
				"script": {
					"type": "string"
				},
				"secret": {
					"type": "string"
				},
				"sourceDir": {
					"type": "string"
				},
				"spec": {
					"type": "string"
				},
				"url": {
					"type": "string"
				},
				"website": {
					"type": "string"
				}
			},
			"required": [
				"id",
				"name",
				"spec"
			],
			"type": "object"
		},
		"dto.CronjobUpdateStatus": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"status": {
					"type": "string"
				}
			},
			"required": [
				"id",
				"status"
			],
			"type": "object"
		},
		"dto.DBBaseInfo": {
			"properties": {
				"containerName": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.DBConfUpdateByFile": {
			"properties": {
				"database": {
					"type": "string"
				},
				"file": {
					"type": "string"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb",
						"postgresql",
						"redis"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"type"
			],
			"type": "object"
		},
		"dto.DaemonJsonConf": {
			"properties": {
				"cgroupDriver": {
					"type": "string"
				},
				"experimental": {
					"type": "boolean"
				},
				"fixedCidrV6": {
					"type": "string"
				},
				"insecureRegistries": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"ip6Tables": {
					"type": "boolean"
				},
				"iptables": {
					"type": "boolean"
				},
				"ipv6": {
					"type": "boolean"
				},
				"isSwarm": {
					"type": "boolean"
				},
				"liveRestore": {
					"type": "boolean"
				},
				"logMaxFile": {
					"type": "string"
				},
				"logMaxSize": {
					"type": "string"
				},
				"registryMirrors": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"status": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DaemonJsonUpdateByFile": {
			"properties": {
				"file": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DashboardBase": {
			"properties": {
				"appInstalledNumber": {
					"type": "integer"
				},
				"cpuCores": {
					"type": "integer"
				},
				"cpuLogicalCores": {
					"type": "integer"
				},
				"cpuModelName": {
					"type": "string"
				},
				"cronjobNumber": {
					"type": "integer"
				},
				"currentInfo": {
					"$ref": "#/definitions/dto.DashboardCurrent"
				},
				"databaseNumber": {
					"type": "integer"
				},
				"hostname": {
					"type": "string"
				},
				"kernelArch": {
					"type": "string"
				},
				"kernelVersion": {
					"type": "string"
				},
				"os": {
					"type": "string"
				},
				"platform": {
					"type": "string"
				},
				"platformFamily": {
					"type": "string"
				},
				"platformVersion": {
					"type": "string"
				},
				"virtualizationSystem": {
					"type": "string"
				},
				"websiteNumber": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.DashboardCurrent": {
			"properties": {
				"cpuPercent": {
					"items": {
						"type": "number"
					},
					"type": "array"
				},
				"cpuTotal": {
					"type": "integer"
				},
				"cpuUsed": {
					"type": "number"
				},
				"cpuUsedPercent": {
					"type": "number"
				},
				"diskData": {
					"items": {
						"$ref": "#/definitions/dto.DiskInfo"
					},
					"type": "array"
				},
				"gpuData": {
					"items": {
						"$ref": "#/definitions/dto.GPUInfo"
					},
					"type": "array"
				},
				"ioCount": {
					"type": "integer"
				},
				"ioReadBytes": {
					"type": "integer"
				},
				"ioReadTime": {
					"type": "integer"
				},
				"ioWriteBytes": {
					"type": "integer"
				},
				"ioWriteTime": {
					"type": "integer"
				},
				"load1": {
					"type": "number"
				},
				"load15": {
					"type": "number"
				},
				"load5": {
					"type": "number"
				},
				"loadUsagePercent": {
					"type": "number"
				},
				"memoryAvailable": {
					"type": "integer"
				},
				"memoryTotal": {
					"type": "integer"
				},
				"memoryUsed": {
					"type": "integer"
				},
				"memoryUsedPercent": {
					"type": "number"
				},
				"netBytesRecv": {
					"type": "integer"
				},
				"netBytesSent": {
					"type": "integer"
				},
				"procs": {
					"type": "integer"
				},
				"shotTime": {
					"type": "string"
				},
				"swapMemoryAvailable": {
					"type": "integer"
				},
				"swapMemoryTotal": {
					"type": "integer"
				},
				"swapMemoryUsed": {
					"type": "integer"
				},
				"swapMemoryUsedPercent": {
					"type": "number"
				},
				"timeSinceUptime": {
					"type": "string"
				},
				"uptime": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.DatabaseCreate": {
			"properties": {
				"address": {
					"type": "string"
				},
				"clientCert": {
					"type": "string"
				},
				"clientKey": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"name": {
					"maxLength": 256,
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				},
				"rootCert": {
					"type": "string"
				},
				"skipVerify": {
					"type": "boolean"
				},
				"ssl": {
					"type": "boolean"
				},
				"type": {
					"type": "string"
				},
				"username": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"required": [
				"from",
				"name",
				"type",
				"username",
				"version"
			],
			"type": "object"
		},
		"dto.DatabaseDelete": {
			"properties": {
				"deleteBackup": {
					"type": "boolean"
				},
				"forceDelete": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"dto.DatabaseInfo": {
			"properties": {
				"address": {
					"type": "string"
				},
				"clientCert": {
					"type": "string"
				},
				"clientKey": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"maxLength": 256,
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				},
				"rootCert": {
					"type": "string"
				},
				"skipVerify": {
					"type": "boolean"
				},
				"ssl": {
					"type": "boolean"
				},
				"type": {
					"type": "string"
				},
				"username": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DatabaseItem": {
			"properties": {
				"database": {
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DatabaseOption": {
			"properties": {
				"address": {
					"type": "string"
				},
				"database": {
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DatabaseSearch": {
			"properties": {
				"info": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.DatabaseUpdate": {
			"properties": {
				"address": {
					"type": "string"
				},
				"clientCert": {
					"type": "string"
				},
				"clientKey": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"password": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				},
				"rootCert": {
					"type": "string"
				},
				"skipVerify": {
					"type": "boolean"
				},
				"ssl": {
					"type": "boolean"
				},
				"type": {
					"type": "string"
				},
				"username": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"required": [
				"type",
				"username",
				"version"
			],
			"type": "object"
		},
		"dto.DeviceBaseInfo": {
			"properties": {
				"dns": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"hostname": {
					"type": "string"
				},
				"hosts": {
					"items": {
						"$ref": "#/definitions/dto.HostHelper"
					},
					"type": "array"
				},
				"localTime": {
					"type": "string"
				},
				"maxSize": {
					"type": "integer"
				},
				"ntp": {
					"type": "string"
				},
				"swapDetails": {
					"items": {
						"$ref": "#/definitions/dto.SwapHelper"
					},
					"type": "array"
				},
				"swapMemoryAvailable": {
					"type": "integer"
				},
				"swapMemoryTotal": {
					"type": "integer"
				},
				"swapMemoryUsed": {
					"type": "integer"
				},
				"timeZone": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.DiskInfo": {
			"properties": {
				"device": {
					"type": "string"
				},
				"free": {
					"type": "integer"
				},
				"inodesFree": {
					"type": "integer"
				},
				"inodesTotal": {
					"type": "integer"
				},
				"inodesUsed": {
					"type": "integer"
				},
				"inodesUsedPercent": {
					"type": "number"
				},
				"path": {
					"type": "string"
				},
				"total": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"used": {
					"type": "integer"
				},
				"usedPercent": {
					"type": "number"
				}
			},
			"type": "object"
		},
		"dto.DockerOperation": {
			"properties": {
				"operation": {
					"enum": [
						"start",
						"restart",
						"stop"
					],
					"type": "string"
				}
			},
			"required": [
				"operation"
			],
			"type": "object"
		},
		"dto.DownloadRecord": {
			"properties": {
				"fileDir": {
					"type": "string"
				},
				"fileName": {
					"type": "string"
				},
				"source": {
					"enum": [
						"OSS",
						"S3",
						"SFTP",
						"MINIO",
						"LOCAL",
						"COS",
						"KODO",
						"OneDrive",
						"WebDAV"
					],
					"type": "string"
				}
			},
			"required": [
				"fileDir",
				"fileName",
				"source"
			],
			"type": "object"
		},
		"dto.Fail2BanBaseInfo": {
			"properties": {
				"banAction": {
					"type": "string"
				},
				"banTime": {
					"type": "string"
				},
				"findTime": {
					"type": "string"
				},
				"isActive": {
					"type": "boolean"
				},
				"isEnable": {
					"type": "boolean"
				},
				"isExist": {
					"type": "boolean"
				},
				"logPath": {
					"type": "string"
				},
				"maxRetry": {
					"type": "integer"
				},
				"port": {
					"type": "integer"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.Fail2BanSearch": {
			"properties": {
				"status": {
					"enum": [
						"banned",
						"ignore"
					],
					"type": "string"
				}
			},
			"required": [
				"status"
			],
			"type": "object"
		},
		"dto.Fail2BanUpdate": {
			"properties": {
				"key": {
					"enum": [
						"port",
						"bantime",
						"findtime",
						"maxretry",
						"banaction",
						"logpath",
						"port"
					],
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"required": [
				"key"
			],
			"type": "object"
		},
		"dto.FirewallBaseInfo": {
			"properties": {
				"name": {
					"type": "string"
				},
				"pingStatus": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.FirewallOperation": {
			"properties": {
				"operation": {
					"enum": [
						"start",
						"stop",
						"restart",
						"disablePing",
						"enablePing"
					],
					"type": "string"
				}
			},
			"required": [
				"operation"
			],
			"type": "object"
		},
		"dto.ForBuckets": {
			"properties": {
				"accessKey": {
					"type": "string"
				},
				"credential": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"vars": {
					"type": "string"
				}
			},
			"required": [
				"credential",
				"type",
				"vars"
			],
			"type": "object"
		},
		"dto.ForwardRuleOperate": {
			"properties": {
				"rules": {
					"items": {
						"properties": {
							"num": {
								"type": "string"
							},
							"operation": {
								"enum": [
									"add",
									"remove"
								],
								"type": "string"
							},
							"port": {
								"type": "string"
							},
							"protocol": {
								"enum": [
									"tcp",
									"udp",
									"tcp/udp"
								],
								"type": "string"
							},
							"targetIP": {
								"type": "string"
							},
							"targetPort": {
								"type": "string"
							}
						},
						"required": [
							"operation",
							"port",
							"protocol",
							"targetPort"
						],
						"type": "object"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"dto.FtpBaseInfo": {
			"properties": {
				"isActive": {
					"type": "boolean"
				},
				"isExist": {
					"type": "boolean"
				}
			},
			"type": "object"
		},
		"dto.FtpCreate": {
			"properties": {
				"description": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"password",
				"path",
				"user"
			],
			"type": "object"
		},
		"dto.FtpLogSearch": {
			"properties": {
				"operation": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.FtpUpdate": {
			"properties": {
				"description": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"password": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"status": {
					"type": "string"
				}
			},
			"required": [
				"password",
				"path"
			],
			"type": "object"
		},
		"dto.GPUInfo": {
			"properties": {
				"fanSpeed": {
					"type": "string"
				},
				"gpuUtil": {
					"type": "string"
				},
				"index": {
					"type": "integer"
				},
				"maxPowerLimit": {
					"type": "string"
				},
				"memTotal": {
					"type": "string"
				},
				"memUsed": {
					"type": "string"
				},
				"memoryUsage": {
					"type": "string"
				},
				"performanceState": {
					"type": "string"
				},
				"powerDraw": {
					"type": "string"
				},
				"powerUsage": {
					"type": "string"
				},
				"productName": {
					"type": "string"
				},
				"temperature": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.GenerateLoad": {
			"properties": {
				"encryptionMode": {
					"enum": [
						"rsa",
						"ed25519",
						"ecdsa",
						"dsa"
					],
					"type": "string"
				}
			},
			"required": [
				"encryptionMode"
			],
			"type": "object"
		},
		"dto.GenerateSSH": {
			"properties": {
				"encryptionMode": {
					"enum": [
						"rsa",
						"ed25519",
						"ecdsa",
						"dsa"
					],
					"type": "string"
				},
				"password": {
					"type": "string"
				}
			},
			"required": [
				"encryptionMode"
			],
			"type": "object"
		},
		"dto.GroupCreate": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"type"
			],
			"type": "object"
		},
		"dto.GroupInfo": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"isDefault": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.GroupSearch": {
			"properties": {
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.GroupUpdate": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"isDefault": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.HostConnTest": {
			"properties": {
				"addr": {
					"type": "string"
				},
				"authMode": {
					"enum": [
						"password",
						"key"
					],
					"type": "string"
				},
				"passPhrase": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"port": {
					"maximum": 65535,
					"minimum": 1,
					"type": "integer"
				},
				"privateKey": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"addr",
				"port",
				"user"
			],
			"type": "object"
		},
		"dto.HostHelper": {
			"properties": {
				"host": {
					"type": "string"
				},
				"ip": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.HostOperate": {
			"properties": {
				"addr": {
					"type": "string"
				},
				"authMode": {
					"enum": [
						"password",
						"key"
					],
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"groupID": {
					"type": "integer"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"passPhrase": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"port": {
					"maximum": 65535,
					"minimum": 1,
					"type": "integer"
				},
				"privateKey": {
					"type": "string"
				},
				"rememberPassword": {
					"type": "boolean"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"addr",
				"port",
				"user"
			],
			"type": "object"
		},
		"dto.HostTree": {
			"properties": {
				"children": {
					"items": {
						"$ref": "#/definitions/dto.TreeChild"
					},
					"type": "array"
				},
				"id": {
					"type": "integer"
				},
				"label": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ImageBuild": {
			"properties": {
				"dockerfile": {
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"tags": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"required": [
				"dockerfile",
				"from",
				"name"
			],
			"type": "object"
		},
		"dto.ImageInfo": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"id": {
					"type": "string"
				},
				"isUsed": {
					"type": "boolean"
				},
				"size": {
					"type": "string"
				},
				"tags": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"dto.ImageLoad": {
			"properties": {
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"dto.ImagePull": {
			"properties": {
				"imageName": {
					"type": "string"
				},
				"repoID": {
					"type": "integer"
				}
			},
			"required": [
				"imageName"
			],
			"type": "object"
		},
		"dto.ImagePush": {
			"properties": {
				"name": {
					"type": "string"
				},
				"repoID": {
					"type": "integer"
				},
				"tagName": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"repoID",
				"tagName"
			],
			"type": "object"
		},
		"dto.ImageRepoDelete": {
			"properties": {
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"dto.ImageRepoOption": {
			"properties": {
				"downloadUrl": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ImageRepoUpdate": {
			"properties": {
				"auth": {
					"type": "boolean"
				},
				"downloadUrl": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"password": {
					"maxLength": 256,
					"type": "string"
				},
				"protocol": {
					"type": "string"
				},
				"username": {
					"maxLength": 256,
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.ImageSave": {
			"properties": {
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"tagName": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"path",
				"tagName"
			],
			"type": "object"
		},
		"dto.ImageTag": {
			"properties": {
				"sourceID": {
					"type": "string"
				},
				"targetName": {
					"type": "string"
				}
			},
			"required": [
				"sourceID",
				"targetName"
			],
			"type": "object"
		},
		"dto.InspectReq": {
			"properties": {
				"id": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"id",
				"type"
			],
			"type": "object"
		},
		"dto.LogOption": {
			"properties": {
				"logMaxFile": {
					"type": "string"
				},
				"logMaxSize": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.MonitorSearch": {
			"properties": {
				"endTime": {
					"type": "string"
				},
				"info": {
					"type": "string"
				},
				"param": {
					"enum": [
						"all",
						"cpu",
						"memory",
						"load",
						"io",
						"network"
					],
					"type": "string"
				},
				"startTime": {
					"type": "string"
				}
			},
			"required": [
				"param"
			],
			"type": "object"
		},
		"dto.MysqlDBCreate": {
			"properties": {
				"database": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"format": {
					"enum": [
						"utf8mb4",
						"utf8",
						"gbk",
						"big5"
					],
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"permission": {
					"type": "string"
				},
				"username": {
					"type": "string"
				}
			},
			"required": [
				"database",
				"format",
				"from",
				"name",
				"password",
				"permission",
				"username"
			],
			"type": "object"
		},
		"dto.MysqlDBDelete": {
			"properties": {
				"database": {
					"type": "string"
				},
				"deleteBackup": {
					"type": "boolean"
				},
				"forceDelete": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"id",
				"type"
			],
			"type": "object"
		},
		"dto.MysqlDBDeleteCheck": {
			"properties": {
				"database": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"id",
				"type"
			],
			"type": "object"
		},
		"dto.MysqlDBSearch": {
			"properties": {
				"database": {
					"type": "string"
				},
				"info": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"database",
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.MysqlLoadDB": {
			"properties": {
				"database": {
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"from",
				"type"
			],
			"type": "object"
		},
		"dto.MysqlOption": {
			"properties": {
				"database": {
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.MysqlStatus": {
			"properties": {
				"Aborted_clients": {
					"type": "string"
				},
				"Aborted_connects": {
					"type": "string"
				},
				"Bytes_received": {
					"type": "string"
				},
				"Bytes_sent": {
					"type": "string"
				},
				"Com_commit": {
					"type": "string"
				},
				"Com_rollback": {
					"type": "string"
				},
				"Connections": {
					"type": "string"
				},
				"Created_tmp_disk_tables": {
					"type": "string"
				},
				"Created_tmp_tables": {
					"type": "string"
				},
				"File": {
					"type": "string"
				},
				"Innodb_buffer_pool_pages_dirty": {
					"type": "string"
				},
				"Innodb_buffer_pool_read_requests": {
					"type": "string"
				},
				"Innodb_buffer_pool_reads": {
					"type": "string"
				},
				"Key_read_requests": {
					"type": "string"
				},
				"Key_reads": {
					"type": "string"
				},
				"Key_write_requests": {
					"type": "string"
				},
				"Key_writes": {
					"type": "string"
				},
				"Max_used_connections": {
					"type": "string"
				},
				"Open_tables": {
					"type": "string"
				},
				"Opened_files": {
					"type": "string"
				},
				"Opened_tables": {
					"type": "string"
				},
				"Position": {
					"type": "string"
				},
				"Qcache_hits": {
					"type": "string"
				},
				"Qcache_inserts": {
					"type": "string"
				},
				"Questions": {
					"type": "string"
				},
				"Run": {
					"type": "string"
				},
				"Select_full_join": {
					"type": "string"
				},
				"Select_range_check": {
					"type": "string"
				},
				"Sort_merge_passes": {
					"type": "string"
				},
				"Table_locks_waited": {
					"type": "string"
				},
				"Threads_cached": {
					"type": "string"
				},
				"Threads_connected": {
					"type": "string"
				},
				"Threads_created": {
					"type": "string"
				},
				"Threads_running": {
					"type": "string"
				},
				"Uptime": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.MysqlVariables": {
			"properties": {
				"binlog_cache_size": {
					"type": "string"
				},
				"innodb_buffer_pool_size": {
					"type": "string"
				},
				"innodb_log_buffer_size": {
					"type": "string"
				},
				"join_buffer_size": {
					"type": "string"
				},
				"key_buffer_size": {
					"type": "string"
				},
				"long_query_time": {
					"type": "string"
				},
				"max_connections": {
					"type": "string"
				},
				"max_heap_table_size": {
					"type": "string"
				},
				"query_cache_size": {
					"type": "string"
				},
				"query_cache_type": {
					"type": "string"
				},
				"read_buffer_size": {
					"type": "string"
				},
				"read_rnd_buffer_size": {
					"type": "string"
				},
				"slow_query_log": {
					"type": "string"
				},
				"sort_buffer_size": {
					"type": "string"
				},
				"table_open_cache": {
					"type": "string"
				},
				"thread_cache_size": {
					"type": "string"
				},
				"thread_stack": {
					"type": "string"
				},
				"tmp_table_size": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.MysqlVariablesUpdate": {
			"properties": {
				"database": {
					"type": "string"
				},
				"type": {
					"enum": [
						"mysql",
						"mariadb"
					],
					"type": "string"
				},
				"variables": {
					"items": {
						"$ref": "#/definitions/dto.MysqlVariablesUpdateHelper"
					},
					"type": "array"
				}
			},
			"required": [
				"database",
				"type"
			],
			"type": "object"
		},
		"dto.MysqlVariablesUpdateHelper": {
			"properties": {
				"param": {
					"type": "string"
				},
				"value": {}
			},
			"type": "object"
		},
		"dto.NetworkCreate": {
			"properties": {
				"auxAddress": {
					"items": {
						"$ref": "#/definitions/dto.SettingUpdate"
					},
					"type": "array"
				},
				"auxAddressV6": {
					"items": {
						"$ref": "#/definitions/dto.SettingUpdate"
					},
					"type": "array"
				},
				"driver": {
					"type": "string"
				},
				"gateway": {
					"type": "string"
				},
				"gatewayV6": {
					"type": "string"
				},
				"ipRange": {
					"type": "string"
				},
				"ipRangeV6": {
					"type": "string"
				},
				"ipv4": {
					"type": "boolean"
				},
				"ipv6": {
					"type": "boolean"
				},
				"labels": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"name": {
					"type": "string"
				},
				"options": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"subnet": {
					"type": "string"
				},
				"subnetV6": {
					"type": "string"
				}
			},
			"required": [
				"driver",
				"name"
			],
			"type": "object"
		},
		"dto.NginxKey": {
			"enum": [
				"index",
				"limit-conn",
				"ssl",
				"cache",
				"http-per",
				"proxy-cache"
			],
			"type": "string",
			"x-enum-varnames": [
				"Index",
				"LimitConn",
				"SSL",
				"CACHE",
				"HttpPer",
				"ProxyCache"
			]
		},
		"dto.OneDriveInfo": {
			"properties": {
				"client_id": {
					"type": "string"
				},
				"client_secret": {
					"type": "string"
				},
				"redirect_uri": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.Operate": {
			"properties": {
				"operation": {
					"type": "string"
				}
			},
			"required": [
				"operation"
			],
			"type": "object"
		},
		"dto.OperateByID": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"dto.OperationWithName": {
			"properties": {
				"name": {
					"type": "string"
				}
			},
			"required": [
				"name"
			],
			"type": "object"
		},
		"dto.OperationWithNameAndType": {
			"properties": {
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"dto.Options": {
			"properties": {
				"option": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.OsInfo": {
			"properties": {
				"diskSize": {
					"type": "integer"
				},
				"kernelArch": {
					"type": "string"
				},
				"kernelVersion": {
					"type": "string"
				},
				"os": {
					"type": "string"
				},
				"platform": {
					"type": "string"
				},
				"platformFamily": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.PageContainer": {
			"properties": {
				"excludeAppStore": {
					"type": "boolean"
				},
				"filters": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"state",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"state": {
					"enum": [
						"all",
						"created",
						"running",
						"paused",
						"restarting",
						"removing",
						"exited",
						"dead"
					],
					"type": "string"
				}
			},
			"required": [
				"order",
				"orderBy",
				"page",
				"pageSize",
				"state"
			],
			"type": "object"
		},
		"dto.PageCronjob": {
			"properties": {
				"info": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"status",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.PageInfo": {
			"properties": {
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.PageResult": {
			"properties": {
				"items": {},
				"total": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.PortHelper": {
			"properties": {
				"containerPort": {
					"type": "string"
				},
				"hostIP": {
					"type": "string"
				},
				"hostPort": {
					"type": "string"
				},
				"protocol": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.PortRuleOperate": {
			"properties": {
				"address": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"operation": {
					"enum": [
						"add",
						"remove"
					],
					"type": "string"
				},
				"port": {
					"type": "string"
				},
				"protocol": {
					"enum": [
						"tcp",
						"udp",
						"tcp/udp"
					],
					"type": "string"
				},
				"strategy": {
					"enum": [
						"accept",
						"drop"
					],
					"type": "string"
				}
			},
			"required": [
				"operation",
				"port",
				"protocol",
				"strategy"
			],
			"type": "object"
		},
		"dto.PortRuleUpdate": {
			"properties": {
				"newRule": {
					"$ref": "#/definitions/dto.PortRuleOperate"
				},
				"oldRule": {
					"$ref": "#/definitions/dto.PortRuleOperate"
				}
			},
			"type": "object"
		},
		"dto.PostgresqlBindUser": {
			"properties": {
				"database": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"superUser": {
					"type": "boolean"
				},
				"username": {
					"type": "string"
				}
			},
			"required": [
				"database",
				"name",
				"password",
				"username"
			],
			"type": "object"
		},
		"dto.PostgresqlDBCreate": {
			"properties": {
				"database": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"format": {
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"superUser": {
					"type": "boolean"
				},
				"username": {
					"type": "string"
				}
			},
			"required": [
				"database",
				"from",
				"name",
				"password",
				"username"
			],
			"type": "object"
		},
		"dto.PostgresqlDBDelete": {
			"properties": {
				"database": {
					"type": "string"
				},
				"deleteBackup": {
					"type": "boolean"
				},
				"forceDelete": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"postgresql"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"id",
				"type"
			],
			"type": "object"
		},
		"dto.PostgresqlDBDeleteCheck": {
			"properties": {
				"database": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"postgresql"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"id",
				"type"
			],
			"type": "object"
		},
		"dto.PostgresqlDBSearch": {
			"properties": {
				"database": {
					"type": "string"
				},
				"info": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"database",
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.PostgresqlLoadDB": {
			"properties": {
				"database": {
					"type": "string"
				},
				"from": {
					"enum": [
						"local",
						"remote"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"postgresql"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"from",
				"type"
			],
			"type": "object"
		},
		"dto.RecordSearch": {
			"properties": {
				"detailName": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize",
				"type"
			],
			"type": "object"
		},
		"dto.RecordSearchByCronjob": {
			"properties": {
				"cronjobID": {
					"type": "integer"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"cronjobID",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.RedisCommand": {
			"properties": {
				"command": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.RedisConf": {
			"properties": {
				"containerName": {
					"type": "string"
				},
				"database": {
					"type": "string"
				},
				"maxclients": {
					"type": "string"
				},
				"maxmemory": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				},
				"requirepass": {
					"type": "string"
				},
				"timeout": {
					"type": "string"
				}
			},
			"required": [
				"database"
			],
			"type": "object"
		},
		"dto.RedisConfPersistenceUpdate": {
			"properties": {
				"appendfsync": {
					"type": "string"
				},
				"appendonly": {
					"type": "string"
				},
				"database": {
					"type": "string"
				},
				"save": {
					"type": "string"
				},
				"type": {
					"enum": [
						"aof",
						"rbd"
					],
					"type": "string"
				}
			},
			"required": [
				"database",
				"type"
			],
			"type": "object"
		},
		"dto.RedisConfUpdate": {
			"properties": {
				"database": {
					"type": "string"
				},
				"maxclients": {
					"type": "string"
				},
				"maxmemory": {
					"type": "string"
				},
				"timeout": {
					"type": "string"
				}
			},
			"required": [
				"database"
			],
			"type": "object"
		},
		"dto.RedisPersistence": {
			"properties": {
				"appendfsync": {
					"type": "string"
				},
				"appendonly": {
					"type": "string"
				},
				"database": {
					"type": "string"
				},
				"save": {
					"type": "string"
				}
			},
			"required": [
				"database"
			],
			"type": "object"
		},
		"dto.RedisStatus": {
			"properties": {
				"connected_clients": {
					"type": "string"
				},
				"database": {
					"type": "string"
				},
				"instantaneous_ops_per_sec": {
					"type": "string"
				},
				"keyspace_hits": {
					"type": "string"
				},
				"keyspace_misses": {
					"type": "string"
				},
				"latest_fork_usec": {
					"type": "string"
				},
				"mem_fragmentation_ratio": {
					"type": "string"
				},
				"tcp_port": {
					"type": "string"
				},
				"total_commands_processed": {
					"type": "string"
				},
				"total_connections_received": {
					"type": "string"
				},
				"uptime_in_days": {
					"type": "string"
				},
				"used_memory": {
					"type": "string"
				},
				"used_memory_peak": {
					"type": "string"
				},
				"used_memory_rss": {
					"type": "string"
				}
			},
			"required": [
				"database"
			],
			"type": "object"
		},
		"dto.ResourceLimit": {
			"properties": {
				"cpu": {
					"type": "integer"
				},
				"memory": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.RuleSearch": {
			"properties": {
				"info": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"status": {
					"type": "string"
				},
				"strategy": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize",
				"type"
			],
			"type": "object"
		},
		"dto.SSHConf": {
			"properties": {
				"file": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.SSHHistory": {
			"properties": {
				"address": {
					"type": "string"
				},
				"area": {
					"type": "string"
				},
				"authMode": {
					"type": "string"
				},
				"date": {
					"type": "string"
				},
				"dateStr": {
					"type": "string"
				},
				"message": {
					"type": "string"
				},
				"port": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.SSHInfo": {
			"properties": {
				"autoStart": {
					"type": "boolean"
				},
				"listenAddress": {
					"type": "string"
				},
				"message": {
					"type": "string"
				},
				"passwordAuthentication": {
					"type": "string"
				},
				"permitRootLogin": {
					"type": "string"
				},
				"port": {
					"type": "string"
				},
				"pubkeyAuthentication": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"useDNS": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.SSHLog": {
			"properties": {
				"failedCount": {
					"type": "integer"
				},
				"logs": {
					"items": {
						"$ref": "#/definitions/dto.SSHHistory"
					},
					"type": "array"
				},
				"successfulCount": {
					"type": "integer"
				},
				"totalCount": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"dto.SSHUpdate": {
			"properties": {
				"key": {
					"type": "string"
				},
				"newValue": {
					"type": "string"
				},
				"oldValue": {
					"type": "string"
				}
			},
			"required": [
				"key"
			],
			"type": "object"
		},
		"dto.SearchClamWithPage": {
			"properties": {
				"info": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"name",
						"status",
						"created_at"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SearchForTree": {
			"properties": {
				"info": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.SearchHostWithPage": {
			"properties": {
				"groupID": {
					"type": "integer"
				},
				"info": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SearchRecord": {
			"properties": {
				"cronjobID": {
					"type": "integer"
				},
				"endTime": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"startTime": {
					"type": "string"
				},
				"status": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SearchSSHLog": {
			"properties": {
				"Status": {
					"enum": [
						"Success",
						"Failed",
						"All"
					],
					"type": "string"
				},
				"info": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"Status",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SearchTaskLogReq": {
			"properties": {
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"status": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SearchWithPage": {
			"properties": {
				"info": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"dto.SettingInfo": {
			"properties": {
				"appStoreLastModified": {
					"type": "string"
				},
				"appStoreSyncStatus": {
					"type": "string"
				},
				"appStoreVersion": {
					"type": "string"
				},
				"defaultNetwork": {
					"type": "string"
				},
				"dockerSockPath": {
					"type": "string"
				},
				"fileRecycleBin": {
					"type": "string"
				},
				"lastCleanData": {
					"type": "string"
				},
				"lastCleanSize": {
					"type": "string"
				},
				"lastCleanTime": {
					"type": "string"
				},
				"localTime": {
					"type": "string"
				},
				"monitorInterval": {
					"type": "string"
				},
				"monitorStatus": {
					"type": "string"
				},
				"monitorStoreDays": {
					"type": "string"
				},
				"ntpSite": {
					"type": "string"
				},
				"snapshotIgnore": {
					"type": "string"
				},
				"systemIP": {
					"type": "string"
				},
				"systemVersion": {
					"type": "string"
				},
				"timeZone": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.SettingUpdate": {
			"properties": {
				"key": {
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"required": [
				"key"
			],
			"type": "object"
		},
		"dto.SnapshotBatchDelete": {
			"properties": {
				"deleteWithFile": {
					"type": "boolean"
				},
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"dto.SnapshotCreate": {
			"properties": {
				"defaultDownload": {
					"type": "string"
				},
				"description": {
					"maxLength": 256,
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"secret": {
					"type": "string"
				}
			},
			"required": [
				"defaultDownload",
				"from"
			],
			"type": "object"
		},
		"dto.SnapshotImport": {
			"properties": {
				"description": {
					"maxLength": 256,
					"type": "string"
				},
				"from": {
					"type": "string"
				},
				"names": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"dto.SnapshotRecover": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"isNew": {
					"type": "boolean"
				},
				"reDownload": {
					"type": "boolean"
				},
				"secret": {
					"type": "string"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"dto.SwapHelper": {
			"properties": {
				"isNew": {
					"type": "boolean"
				},
				"path": {
					"type": "string"
				},
				"size": {
					"type": "integer"
				},
				"used": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"dto.TreeChild": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"label": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.UpdateByFile": {
			"properties": {
				"file": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.UpdateByNameAndFile": {
			"properties": {
				"file": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"dto.UpdateDescription": {
			"properties": {
				"description": {
					"maxLength": 256,
					"type": "string"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"dto.UpdateFirewallDescription": {
			"properties": {
				"address": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"port": {
					"type": "string"
				},
				"protocol": {
					"type": "string"
				},
				"strategy": {
					"enum": [
						"accept",
						"drop"
					],
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"strategy"
			],
			"type": "object"
		},
		"dto.VolumeCreate": {
			"properties": {
				"driver": {
					"type": "string"
				},
				"labels": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"name": {
					"type": "string"
				},
				"options": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"required": [
				"driver",
				"name"
			],
			"type": "object"
		},
		"dto.VolumeHelper": {
			"properties": {
				"containerDir": {
					"type": "string"
				},
				"mode": {
					"type": "string"
				},
				"sourceDir": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"files.FileInfo": {
			"properties": {
				"content": {
					"type": "string"
				},
				"extension": {
					"type": "string"
				},
				"favoriteID": {
					"type": "integer"
				},
				"gid": {
					"type": "string"
				},
				"group": {
					"type": "string"
				},
				"isDetail": {
					"type": "boolean"
				},
				"isDir": {
					"type": "boolean"
				},
				"isHidden": {
					"type": "boolean"
				},
				"isSymlink": {
					"type": "boolean"
				},
				"itemTotal": {
					"type": "integer"
				},
				"items": {
					"items": {
						"$ref": "#/definitions/files.FileInfo"
					},
					"type": "array"
				},
				"linkPath": {
					"type": "string"
				},
				"mimeType": {
					"type": "string"
				},
				"modTime": {
					"type": "string"
				},
				"mode": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"size": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"uid": {
					"type": "string"
				},
				"updateTime": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.App": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"crossVersionUpdate": {
					"type": "boolean"
				},
				"document": {
					"type": "string"
				},
				"github": {
					"type": "string"
				},
				"icon": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"key": {
					"type": "string"
				},
				"lastModified": {
					"type": "integer"
				},
				"limit": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"readMe": {
					"type": "string"
				},
				"recommend": {
					"type": "integer"
				},
				"required": {
					"type": "string"
				},
				"resource": {
					"type": "string"
				},
				"shortDescEn": {
					"type": "string"
				},
				"shortDescZh": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"tags": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"website": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.AppInstall": {
			"properties": {
				"app": {
					"$ref": "#/definitions/model.App"
				},
				"appDetailId": {
					"type": "integer"
				},
				"appId": {
					"type": "integer"
				},
				"containerName": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"dockerCompose": {
					"type": "string"
				},
				"env": {
					"type": "string"
				},
				"httpPort": {
					"type": "integer"
				},
				"httpsPort": {
					"type": "integer"
				},
				"id": {
					"type": "integer"
				},
				"message": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"param": {
					"type": "string"
				},
				"serviceName": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.Tag": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"key": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"sort": {
					"type": "integer"
				},
				"updatedAt": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.Website": {
			"properties": {
				"IPV6": {
					"type": "boolean"
				},
				"accessLog": {
					"type": "boolean"
				},
				"alias": {
					"type": "string"
				},
				"appInstallId": {
					"type": "integer"
				},
				"createdAt": {
					"type": "string"
				},
				"defaultServer": {
					"type": "boolean"
				},
				"domains": {
					"items": {
						"$ref": "#/definitions/model.WebsiteDomain"
					},
					"type": "array"
				},
				"errorLog": {
					"type": "boolean"
				},
				"expireDate": {
					"type": "string"
				},
				"ftpId": {
					"type": "integer"
				},
				"group": {
					"type": "string"
				},
				"httpConfig": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"primaryDomain": {
					"type": "string"
				},
				"protocol": {
					"type": "string"
				},
				"proxy": {
					"type": "string"
				},
				"proxyType": {
					"type": "string"
				},
				"remark": {
					"type": "string"
				},
				"rewrite": {
					"type": "string"
				},
				"runtimeID": {
					"type": "integer"
				},
				"siteDir": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"user": {
					"type": "string"
				},
				"webSiteGroupId": {
					"type": "integer"
				},
				"webSiteSSL": {
					"$ref": "#/definitions/model.WebsiteSSL"
				},
				"webSiteSSLId": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"model.WebsiteAcmeAccount": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"eabHmacKey": {
					"type": "string"
				},
				"eabKid": {
					"type": "string"
				},
				"email": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"url": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.WebsiteDnsAccount": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"model.WebsiteDomain": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"domain": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"port": {
					"type": "integer"
				},
				"updatedAt": {
					"type": "string"
				},
				"websiteId": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"model.WebsiteSSL": {
			"properties": {
				"acmeAccount": {
					"$ref": "#/definitions/model.WebsiteAcmeAccount"
				},
				"acmeAccountId": {
					"type": "integer"
				},
				"autoRenew": {
					"type": "boolean"
				},
				"caId": {
					"type": "integer"
				},
				"certURL": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"dir": {
					"type": "string"
				},
				"disableCNAME": {
					"type": "boolean"
				},
				"dnsAccount": {
					"$ref": "#/definitions/model.WebsiteDnsAccount"
				},
				"dnsAccountId": {
					"type": "integer"
				},
				"domains": {
					"type": "string"
				},
				"execShell": {
					"type": "boolean"
				},
				"expireDate": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"message": {
					"type": "string"
				},
				"nameserver1": {
					"type": "string"
				},
				"nameserver2": {
					"type": "string"
				},
				"organization": {
					"type": "string"
				},
				"pem": {
					"type": "string"
				},
				"primaryDomain": {
					"type": "string"
				},
				"privateKey": {
					"type": "string"
				},
				"provider": {
					"type": "string"
				},
				"pushDir": {
					"type": "boolean"
				},
				"shell": {
					"type": "string"
				},
				"skipDNS": {
					"type": "boolean"
				},
				"startDate": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"websites": {
					"items": {
						"$ref": "#/definitions/model.Website"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"request.AppInstallCreate": {
			"properties": {
				"advanced": {
					"type": "boolean"
				},
				"allowPort": {
					"type": "boolean"
				},
				"appDetailId": {
					"type": "integer"
				},
				"containerName": {
					"type": "string"
				},
				"cpuQuota": {
					"type": "number"
				},
				"dockerCompose": {
					"type": "string"
				},
				"editCompose": {
					"type": "boolean"
				},
				"hostMode": {
					"type": "boolean"
				},
				"memoryLimit": {
					"type": "number"
				},
				"memoryUnit": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"params": {
					"additionalProperties": true,
					"type": "object"
				},
				"pullImage": {
					"type": "boolean"
				},
				"services": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"taskID": {
					"type": "string"
				}
			},
			"required": [
				"appDetailId",
				"name"
			],
			"type": "object"
		},
		"request.AppInstalledIgnoreUpgrade": {
			"properties": {
				"detailID": {
					"type": "integer"
				},
				"operate": {
					"enum": [
						"cancel",
						"ignore"
					],
					"type": "string"
				}
			},
			"required": [
				"detailID",
				"operate"
			],
			"type": "object"
		},
		"request.AppInstalledInfo": {
			"properties": {
				"key": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"required": [
				"key"
			],
			"type": "object"
		},
		"request.AppInstalledOperate": {
			"properties": {
				"backup": {
					"type": "boolean"
				},
				"backupId": {
					"type": "integer"
				},
				"deleteBackup": {
					"type": "boolean"
				},
				"deleteDB": {
					"type": "boolean"
				},
				"detailId": {
					"type": "integer"
				},
				"dockerCompose": {
					"type": "string"
				},
				"forceDelete": {
					"type": "boolean"
				},
				"installId": {
					"type": "integer"
				},
				"operate": {
					"type": "string"
				},
				"pullImage": {
					"type": "boolean"
				},
				"taskID": {
					"type": "string"
				}
			},
			"required": [
				"installId",
				"operate"
			],
			"type": "object"
		},
		"request.AppInstalledSearch": {
			"properties": {
				"all": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"sync": {
					"type": "boolean"
				},
				"tags": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				},
				"unused": {
					"type": "boolean"
				},
				"update": {
					"type": "boolean"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.AppInstalledUpdate": {
			"properties": {
				"advanced": {
					"type": "boolean"
				},
				"allowPort": {
					"type": "boolean"
				},
				"containerName": {
					"type": "string"
				},
				"cpuQuota": {
					"type": "number"
				},
				"dockerCompose": {
					"type": "string"
				},
				"editCompose": {
					"type": "boolean"
				},
				"hostMode": {
					"type": "boolean"
				},
				"installId": {
					"type": "integer"
				},
				"memoryLimit": {
					"type": "number"
				},
				"memoryUnit": {
					"type": "string"
				},
				"params": {
					"additionalProperties": true,
					"type": "object"
				},
				"pullImage": {
					"type": "boolean"
				}
			},
			"required": [
				"installId",
				"params"
			],
			"type": "object"
		},
		"request.AppSearch": {
			"properties": {
				"name": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"recommend": {
					"type": "boolean"
				},
				"resource": {
					"type": "string"
				},
				"tags": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.DirSizeReq": {
			"properties": {
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.ExposedPort": {
			"properties": {
				"containerPort": {
					"type": "integer"
				},
				"hostPort": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"request.FavoriteCreate": {
			"properties": {
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FavoriteDelete": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.FileBatchDelete": {
			"properties": {
				"isDir": {
					"type": "boolean"
				},
				"paths": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"required": [
				"paths"
			],
			"type": "object"
		},
		"request.FileCompress": {
			"properties": {
				"dst": {
					"type": "string"
				},
				"files": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"name": {
					"type": "string"
				},
				"replace": {
					"type": "boolean"
				},
				"secret": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"dst",
				"files",
				"name",
				"type"
			],
			"type": "object"
		},
		"request.FileContentReq": {
			"properties": {
				"isDetail": {
					"type": "boolean"
				},
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FileCreate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"isDir": {
					"type": "boolean"
				},
				"isLink": {
					"type": "boolean"
				},
				"isSymlink": {
					"type": "boolean"
				},
				"linkPath": {
					"type": "string"
				},
				"mode": {
					"type": "integer"
				},
				"path": {
					"type": "string"
				},
				"sub": {
					"type": "boolean"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FileDeCompress": {
			"properties": {
				"dst": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"secret": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"dst",
				"path",
				"type"
			],
			"type": "object"
		},
		"request.FileDelete": {
			"properties": {
				"forceDelete": {
					"type": "boolean"
				},
				"isDir": {
					"type": "boolean"
				},
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FileDownload": {
			"properties": {
				"compress": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"paths": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"paths",
				"type"
			],
			"type": "object"
		},
		"request.FileEdit": {
			"properties": {
				"content": {
					"type": "string"
				},
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FileMove": {
			"properties": {
				"cover": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"newPath": {
					"type": "string"
				},
				"oldPaths": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"newPath",
				"oldPaths",
				"type"
			],
			"type": "object"
		},
		"request.FileOption": {
			"properties": {
				"containSub": {
					"type": "boolean"
				},
				"dir": {
					"type": "boolean"
				},
				"expand": {
					"type": "boolean"
				},
				"isDetail": {
					"type": "boolean"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"path": {
					"type": "string"
				},
				"search": {
					"type": "string"
				},
				"showHidden": {
					"type": "boolean"
				},
				"sortBy": {
					"type": "string"
				},
				"sortOrder": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.FilePathCheck": {
			"properties": {
				"path": {
					"type": "string"
				}
			},
			"required": [
				"path"
			],
			"type": "object"
		},
		"request.FileReadByLineReq": {
			"properties": {
				"ID": {
					"type": "integer"
				},
				"latest": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"taskID": {
					"type": "string"
				},
				"taskOperate": {
					"type": "string"
				},
				"taskType": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize",
				"type"
			],
			"type": "object"
		},
		"request.FileRename": {
			"properties": {
				"newName": {
					"type": "string"
				},
				"oldName": {
					"type": "string"
				}
			},
			"required": [
				"newName",
				"oldName"
			],
			"type": "object"
		},
		"request.FileRoleReq": {
			"properties": {
				"group": {
					"type": "string"
				},
				"mode": {
					"type": "integer"
				},
				"paths": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"sub": {
					"type": "boolean"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"group",
				"mode",
				"paths",
				"user"
			],
			"type": "object"
		},
		"request.FileRoleUpdate": {
			"properties": {
				"group": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"sub": {
					"type": "boolean"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"group",
				"path",
				"user"
			],
			"type": "object"
		},
		"request.FileWget": {
			"properties": {
				"ignoreCertificate": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"url": {
					"type": "string"
				}
			},
			"required": [
				"name",
				"path",
				"url"
			],
			"type": "object"
		},
		"request.HostToolConfig": {
			"properties": {
				"content": {
					"type": "string"
				},
				"operate": {
					"enum": [
						"get",
						"set"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"supervisord"
					],
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"request.HostToolCreate": {
			"properties": {
				"configPath": {
					"type": "string"
				},
				"serviceName": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"request.HostToolLogReq": {
			"properties": {
				"type": {
					"enum": [
						"supervisord"
					],
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"request.HostToolReq": {
			"properties": {
				"operate": {
					"enum": [
						"status",
						"restart",
						"start",
						"stop"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"supervisord"
					],
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"request.NewAppInstall": {
			"properties": {
				"advanced": {
					"type": "boolean"
				},
				"allowPort": {
					"type": "boolean"
				},
				"appDetailID": {
					"type": "integer"
				},
				"containerName": {
					"type": "string"
				},
				"cpuQuota": {
					"type": "number"
				},
				"dockerCompose": {
					"type": "string"
				},
				"editCompose": {
					"type": "boolean"
				},
				"hostMode": {
					"type": "boolean"
				},
				"memoryLimit": {
					"type": "number"
				},
				"memoryUnit": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"params": {
					"additionalProperties": true,
					"type": "object"
				},
				"pullImage": {
					"type": "boolean"
				}
			},
			"type": "object"
		},
		"request.NginxAntiLeechUpdate": {
			"properties": {
				"blocked": {
					"type": "boolean"
				},
				"cache": {
					"type": "boolean"
				},
				"cacheTime": {
					"type": "integer"
				},
				"cacheUint": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"extends": {
					"type": "string"
				},
				"logEnable": {
					"type": "boolean"
				},
				"noneRef": {
					"type": "boolean"
				},
				"return": {
					"type": "string"
				},
				"serverNames": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"extends",
				"return",
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxAuthReq": {
			"properties": {
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxAuthUpdate": {
			"properties": {
				"operate": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"remark": {
					"type": "string"
				},
				"username": {
					"type": "string"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"operate",
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxCommonReq": {
			"properties": {
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxConfigFileUpdate": {
			"properties": {
				"backup": {
					"type": "boolean"
				},
				"content": {
					"type": "string"
				}
			},
			"required": [
				"content"
			],
			"type": "object"
		},
		"request.NginxConfigUpdate": {
			"properties": {
				"operate": {
					"enum": [
						"add",
						"update",
						"delete"
					],
					"type": "string"
				},
				"params": {},
				"scope": {
					"$ref": "#/definitions/dto.NginxKey"
				},
				"websiteId": {
					"type": "integer"
				}
			},
			"required": [
				"operate"
			],
			"type": "object"
		},
		"request.NginxProxyUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"content",
				"name",
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxRedirectReq": {
			"properties": {
				"domains": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"enable": {
					"type": "boolean"
				},
				"keepPath": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"operate": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"redirect": {
					"type": "string"
				},
				"redirectRoot": {
					"type": "boolean"
				},
				"target": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"name",
				"operate",
				"redirect",
				"target",
				"type",
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxRedirectUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"content",
				"name",
				"websiteID"
			],
			"type": "object"
		},
		"request.NginxRewriteReq": {
			"properties": {
				"name": {
					"type": "string"
				},
				"websiteId": {
					"type": "integer"
				}
			},
			"required": [
				"name",
				"websiteId"
			],
			"type": "object"
		},
		"request.NginxRewriteUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"websiteId": {
					"type": "integer"
				}
			},
			"required": [
				"name",
				"websiteId"
			],
			"type": "object"
		},
		"request.NginxScopeReq": {
			"properties": {
				"scope": {
					"$ref": "#/definitions/dto.NginxKey"
				},
				"websiteId": {
					"type": "integer"
				}
			},
			"required": [
				"scope"
			],
			"type": "object"
		},
		"request.NodeModuleReq": {
			"properties": {
				"ID": {
					"type": "integer"
				}
			},
			"required": [
				"ID"
			],
			"type": "object"
		},
		"request.NodePackageReq": {
			"properties": {
				"codeDir": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.PHPExtensionsCreate": {
			"properties": {
				"extensions": {
					"type": "string"
				},
				"name": {
					"type": "string"
				}
			},
			"required": [
				"extensions",
				"name"
			],
			"type": "object"
		},
		"request.PHPExtensionsDelete": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.PHPExtensionsSearch": {
			"properties": {
				"all": {
					"type": "boolean"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.PHPExtensionsUpdate": {
			"properties": {
				"extensions": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"extensions",
				"id"
			],
			"type": "object"
		},
		"request.PortUpdate": {
			"properties": {
				"key": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"request.ProcessReq": {
			"properties": {
				"PID": {
					"type": "integer"
				}
			},
			"required": [
				"PID"
			],
			"type": "object"
		},
		"request.RecycleBinReduce": {
			"properties": {
				"from": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"rName": {
					"type": "string"
				}
			},
			"required": [
				"from",
				"rName"
			],
			"type": "object"
		},
		"request.RuntimeCreate": {
			"properties": {
				"appDetailId": {
					"type": "integer"
				},
				"clean": {
					"type": "boolean"
				},
				"codeDir": {
					"type": "string"
				},
				"exposedPorts": {
					"items": {
						"$ref": "#/definitions/request.ExposedPort"
					},
					"type": "array"
				},
				"image": {
					"type": "string"
				},
				"install": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"params": {
					"additionalProperties": true,
					"type": "object"
				},
				"port": {
					"type": "integer"
				},
				"resource": {
					"type": "string"
				},
				"source": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.RuntimeDelete": {
			"properties": {
				"forceDelete": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"request.RuntimeOperate": {
			"properties": {
				"ID": {
					"type": "integer"
				},
				"operate": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.RuntimeSearch": {
			"properties": {
				"name": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"status": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.RuntimeUpdate": {
			"properties": {
				"clean": {
					"type": "boolean"
				},
				"codeDir": {
					"type": "string"
				},
				"exposedPorts": {
					"items": {
						"$ref": "#/definitions/request.ExposedPort"
					},
					"type": "array"
				},
				"id": {
					"type": "integer"
				},
				"image": {
					"type": "string"
				},
				"install": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"params": {
					"additionalProperties": true,
					"type": "object"
				},
				"port": {
					"type": "integer"
				},
				"rebuild": {
					"type": "boolean"
				},
				"source": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.SearchUploadWithPage": {
			"properties": {
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"path": {
					"type": "string"
				}
			},
			"required": [
				"page",
				"pageSize",
				"path"
			],
			"type": "object"
		},
		"request.SupervisorProcessConfig": {
			"properties": {
				"command": {
					"type": "string"
				},
				"dir": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"numprocs": {
					"type": "string"
				},
				"operate": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"request.SupervisorProcessFileReq": {
			"properties": {
				"content": {
					"type": "string"
				},
				"file": {
					"enum": [
						"out.log",
						"err.log",
						"config"
					],
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"operate": {
					"enum": [
						"get",
						"clear",
						"update"
					],
					"type": "string"
				}
			},
			"required": [
				"file",
				"name",
				"operate"
			],
			"type": "object"
		},
		"request.WebsiteAcmeAccountCreate": {
			"properties": {
				"eabHmacKey": {
					"type": "string"
				},
				"eabKid": {
					"type": "string"
				},
				"email": {
					"type": "string"
				},
				"keyType": {
					"enum": [
						"P256",
						"P384",
						"2048",
						"3072",
						"4096",
						"8192"
					],
					"type": "string"
				},
				"type": {
					"enum": [
						"letsencrypt",
						"zerossl",
						"buypass",
						"google"
					],
					"type": "string"
				}
			},
			"required": [
				"email",
				"keyType",
				"type"
			],
			"type": "object"
		},
		"request.WebsiteBatchDelReq": {
			"properties": {
				"ids": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"required": [
				"ids"
			],
			"type": "object"
		},
		"request.WebsiteCACreate": {
			"properties": {
				"city": {
					"type": "string"
				},
				"commonName": {
					"type": "string"
				},
				"country": {
					"type": "string"
				},
				"keyType": {
					"enum": [
						"P256",
						"P384",
						"2048",
						"3072",
						"4096",
						"8192"
					],
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"organization": {
					"type": "string"
				},
				"organizationUint": {
					"type": "string"
				},
				"province": {
					"type": "string"
				}
			},
			"required": [
				"commonName",
				"country",
				"keyType",
				"name",
				"organization"
			],
			"type": "object"
		},
		"request.WebsiteCAObtain": {
			"properties": {
				"autoRenew": {
					"type": "boolean"
				},
				"description": {
					"type": "string"
				},
				"dir": {
					"type": "string"
				},
				"domains": {
					"type": "string"
				},
				"execShell": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"enum": [
						"P256",
						"P384",
						"2048",
						"3072",
						"4096",
						"8192"
					],
					"type": "string"
				},
				"pushDir": {
					"type": "boolean"
				},
				"renew": {
					"type": "boolean"
				},
				"shell": {
					"type": "string"
				},
				"sslID": {
					"type": "integer"
				},
				"time": {
					"type": "integer"
				},
				"unit": {
					"type": "string"
				}
			},
			"required": [
				"domains",
				"id",
				"keyType",
				"time",
				"unit"
			],
			"type": "object"
		},
		"request.WebsiteCASearch": {
			"properties": {
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.WebsiteCommonReq": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsiteCreate": {
			"properties": {
				"IPV6": {
					"type": "boolean"
				},
				"alias": {
					"type": "string"
				},
				"appID": {
					"type": "integer"
				},
				"appInstall": {
					"$ref": "#/definitions/request.NewAppInstall"
				},
				"appInstallID": {
					"type": "integer"
				},
				"appType": {
					"enum": [
						"new",
						"installed"
					],
					"type": "string"
				},
				"ftpPassword": {
					"type": "string"
				},
				"ftpUser": {
					"type": "string"
				},
				"otherDomains": {
					"type": "string"
				},
				"port": {
					"type": "integer"
				},
				"primaryDomain": {
					"type": "string"
				},
				"proxy": {
					"type": "string"
				},
				"proxyType": {
					"type": "string"
				},
				"remark": {
					"type": "string"
				},
				"runtimeID": {
					"type": "integer"
				},
				"taskID": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"webSiteGroupID": {
					"type": "integer"
				}
			},
			"required": [
				"alias",
				"primaryDomain",
				"type",
				"webSiteGroupID"
			],
			"type": "object"
		},
		"request.WebsiteDNSReq": {
			"properties": {
				"acmeAccountId": {
					"type": "integer"
				},
				"domains": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"required": [
				"acmeAccountId",
				"domains"
			],
			"type": "object"
		},
		"request.WebsiteDefaultUpdate": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"request.WebsiteDelete": {
			"properties": {
				"deleteApp": {
					"type": "boolean"
				},
				"deleteBackup": {
					"type": "boolean"
				},
				"forceDelete": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsiteDnsAccountCreate": {
			"properties": {
				"authorization": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"authorization",
				"name",
				"type"
			],
			"type": "object"
		},
		"request.WebsiteDnsAccountUpdate": {
			"properties": {
				"authorization": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"authorization",
				"id",
				"name",
				"type"
			],
			"type": "object"
		},
		"request.WebsiteDomainCreate": {
			"properties": {
				"domains": {
					"type": "string"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"domains",
				"websiteID"
			],
			"type": "object"
		},
		"request.WebsiteDomainDelete": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsiteHTTPSOp": {
			"properties": {
				"SSLProtocol": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"algorithm": {
					"type": "string"
				},
				"certificate": {
					"type": "string"
				},
				"certificatePath": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"hsts": {
					"type": "boolean"
				},
				"httpConfig": {
					"enum": [
						"HTTPSOnly",
						"HTTPAlso",
						"HTTPToHTTPS"
					],
					"type": "string"
				},
				"importType": {
					"type": "string"
				},
				"privateKey": {
					"type": "string"
				},
				"privateKeyPath": {
					"type": "string"
				},
				"type": {
					"enum": [
						"existed",
						"auto",
						"manual"
					],
					"type": "string"
				},
				"websiteId": {
					"type": "integer"
				},
				"websiteSSLId": {
					"type": "integer"
				}
			},
			"required": [
				"websiteId"
			],
			"type": "object"
		},
		"request.WebsiteHtmlUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"content",
				"type"
			],
			"type": "object"
		},
		"request.WebsiteInstallCheckReq": {
			"properties": {
				"InstallIds": {
					"items": {
						"type": "integer"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"request.WebsiteLogReq": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"logType": {
					"type": "string"
				},
				"operate": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"id",
				"logType",
				"operate"
			],
			"type": "object"
		},
		"request.WebsiteNginxUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"content",
				"id"
			],
			"type": "object"
		},
		"request.WebsiteOp": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"operate": {
					"type": "string"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsitePHPConfigUpdate": {
			"properties": {
				"disableFunctions": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"id": {
					"type": "integer"
				},
				"params": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"scope": {
					"type": "string"
				},
				"uploadMaxSize": {
					"type": "string"
				}
			},
			"required": [
				"id",
				"scope"
			],
			"type": "object"
		},
		"request.WebsitePHPFileUpdate": {
			"properties": {
				"content": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				}
			},
			"required": [
				"content",
				"id",
				"type"
			],
			"type": "object"
		},
		"request.WebsitePHPVersionReq": {
			"properties": {
				"retainConfig": {
					"type": "boolean"
				},
				"runtimeID": {
					"type": "integer"
				},
				"websiteID": {
					"type": "integer"
				}
			},
			"required": [
				"runtimeID",
				"websiteID"
			],
			"type": "object"
		},
		"request.WebsiteProxyConfig": {
			"properties": {
				"cache": {
					"type": "boolean"
				},
				"cacheTime": {
					"type": "integer"
				},
				"cacheUnit": {
					"type": "string"
				},
				"content": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"filePath": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"match": {
					"type": "string"
				},
				"modifier": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"operate": {
					"type": "string"
				},
				"proxyHost": {
					"type": "string"
				},
				"proxyPass": {
					"type": "string"
				},
				"replaces": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"sni": {
					"type": "boolean"
				}
			},
			"required": [
				"id",
				"match",
				"name",
				"operate",
				"proxyHost",
				"proxyPass"
			],
			"type": "object"
		},
		"request.WebsiteProxyReq": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsiteResourceReq": {
			"properties": {
				"id": {
					"type": "integer"
				}
			},
			"required": [
				"id"
			],
			"type": "object"
		},
		"request.WebsiteSSLApply": {
			"properties": {
				"ID": {
					"type": "integer"
				},
				"nameservers": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"skipDNSCheck": {
					"type": "boolean"
				}
			},
			"required": [
				"ID"
			],
			"type": "object"
		},
		"request.WebsiteSSLCreate": {
			"properties": {
				"acmeAccountId": {
					"type": "integer"
				},
				"apply": {
					"type": "boolean"
				},
				"autoRenew": {
					"type": "boolean"
				},
				"description": {
					"type": "string"
				},
				"dir": {
					"type": "string"
				},
				"disableCNAME": {
					"type": "boolean"
				},
				"dnsAccountId": {
					"type": "integer"
				},
				"execShell": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"nameserver1": {
					"type": "string"
				},
				"nameserver2": {
					"type": "string"
				},
				"otherDomains": {
					"type": "string"
				},
				"primaryDomain": {
					"type": "string"
				},
				"provider": {
					"type": "string"
				},
				"pushDir": {
					"type": "boolean"
				},
				"shell": {
					"type": "string"
				},
				"skipDNS": {
					"type": "boolean"
				}
			},
			"required": [
				"acmeAccountId",
				"primaryDomain",
				"provider"
			],
			"type": "object"
		},
		"request.WebsiteSSLSearch": {
			"properties": {
				"acmeAccountID": {
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				}
			},
			"required": [
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.WebsiteSSLUpdate": {
			"properties": {
				"acmeAccountId": {
					"type": "integer"
				},
				"apply": {
					"type": "boolean"
				},
				"autoRenew": {
					"type": "boolean"
				},
				"description": {
					"type": "string"
				},
				"dir": {
					"type": "string"
				},
				"disableCNAME": {
					"type": "boolean"
				},
				"dnsAccountId": {
					"type": "integer"
				},
				"execShell": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"nameserver1": {
					"type": "string"
				},
				"nameserver2": {
					"type": "string"
				},
				"otherDomains": {
					"type": "string"
				},
				"primaryDomain": {
					"type": "string"
				},
				"provider": {
					"type": "string"
				},
				"pushDir": {
					"type": "boolean"
				},
				"shell": {
					"type": "string"
				},
				"skipDNS": {
					"type": "boolean"
				}
			},
			"required": [
				"id",
				"primaryDomain",
				"provider"
			],
			"type": "object"
		},
		"request.WebsiteSSLUpload": {
			"properties": {
				"certificate": {
					"type": "string"
				},
				"certificatePath": {
					"type": "string"
				},
				"description": {
					"type": "string"
				},
				"privateKey": {
					"type": "string"
				},
				"privateKeyPath": {
					"type": "string"
				},
				"sslID": {
					"type": "integer"
				},
				"type": {
					"enum": [
						"paste",
						"local"
					],
					"type": "string"
				}
			},
			"required": [
				"type"
			],
			"type": "object"
		},
		"request.WebsiteSearch": {
			"properties": {
				"name": {
					"type": "string"
				},
				"order": {
					"enum": [
						"null",
						"ascending",
						"descending"
					],
					"type": "string"
				},
				"orderBy": {
					"enum": [
						"primary_domain",
						"type",
						"status",
						"created_at",
						"expire_date"
					],
					"type": "string"
				},
				"page": {
					"type": "integer"
				},
				"pageSize": {
					"type": "integer"
				},
				"websiteGroupId": {
					"type": "integer"
				}
			},
			"required": [
				"order",
				"orderBy",
				"page",
				"pageSize"
			],
			"type": "object"
		},
		"request.WebsiteUpdate": {
			"properties": {
				"IPV6": {
					"type": "boolean"
				},
				"expireDate": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"primaryDomain": {
					"type": "string"
				},
				"remark": {
					"type": "string"
				},
				"webSiteGroupID": {
					"type": "integer"
				}
			},
			"required": [
				"id",
				"primaryDomain"
			],
			"type": "object"
		},
		"request.WebsiteUpdateDir": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"siteDir": {
					"type": "string"
				}
			},
			"required": [
				"id",
				"siteDir"
			],
			"type": "object"
		},
		"request.WebsiteUpdateDirPermission": {
			"properties": {
				"group": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"user": {
					"type": "string"
				}
			},
			"required": [
				"group",
				"id",
				"user"
			],
			"type": "object"
		},
		"response.AppDTO": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"crossVersionUpdate": {
					"type": "boolean"
				},
				"document": {
					"type": "string"
				},
				"github": {
					"type": "string"
				},
				"icon": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"installed": {
					"type": "boolean"
				},
				"key": {
					"type": "string"
				},
				"lastModified": {
					"type": "integer"
				},
				"limit": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"readMe": {
					"type": "string"
				},
				"recommend": {
					"type": "integer"
				},
				"required": {
					"type": "string"
				},
				"resource": {
					"type": "string"
				},
				"shortDescEn": {
					"type": "string"
				},
				"shortDescZh": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"tags": {
					"items": {
						"$ref": "#/definitions/model.Tag"
					},
					"type": "array"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"versions": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"website": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.AppDetailDTO": {
			"properties": {
				"appId": {
					"type": "integer"
				},
				"createdAt": {
					"type": "string"
				},
				"dockerCompose": {
					"type": "string"
				},
				"downloadCallBackUrl": {
					"type": "string"
				},
				"downloadUrl": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"hostMode": {
					"type": "boolean"
				},
				"id": {
					"type": "integer"
				},
				"ignoreUpgrade": {
					"type": "boolean"
				},
				"image": {
					"type": "string"
				},
				"lastModified": {
					"type": "integer"
				},
				"lastVersion": {
					"type": "string"
				},
				"params": {},
				"status": {
					"type": "string"
				},
				"update": {
					"type": "boolean"
				},
				"updatedAt": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.AppInstalledCheck": {
			"properties": {
				"app": {
					"type": "string"
				},
				"appInstallId": {
					"type": "integer"
				},
				"containerName": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"httpPort": {
					"type": "integer"
				},
				"httpsPort": {
					"type": "integer"
				},
				"installPath": {
					"type": "string"
				},
				"isExist": {
					"type": "boolean"
				},
				"lastBackupAt": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.AppParam": {
			"properties": {
				"edit": {
					"type": "boolean"
				},
				"key": {
					"type": "string"
				},
				"labelEn": {
					"type": "string"
				},
				"labelZh": {
					"type": "string"
				},
				"multiple": {
					"type": "boolean"
				},
				"required": {
					"type": "boolean"
				},
				"rule": {
					"type": "string"
				},
				"showValue": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"value": {},
				"values": {}
			},
			"type": "object"
		},
		"response.AppService": {
			"properties": {
				"config": {},
				"from": {
					"type": "string"
				},
				"label": {
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.FileInfo": {
			"properties": {
				"content": {
					"type": "string"
				},
				"extension": {
					"type": "string"
				},
				"favoriteID": {
					"type": "integer"
				},
				"gid": {
					"type": "string"
				},
				"group": {
					"type": "string"
				},
				"isDetail": {
					"type": "boolean"
				},
				"isDir": {
					"type": "boolean"
				},
				"isHidden": {
					"type": "boolean"
				},
				"isSymlink": {
					"type": "boolean"
				},
				"itemTotal": {
					"type": "integer"
				},
				"items": {
					"items": {
						"$ref": "#/definitions/files.FileInfo"
					},
					"type": "array"
				},
				"linkPath": {
					"type": "string"
				},
				"mimeType": {
					"type": "string"
				},
				"modTime": {
					"type": "string"
				},
				"mode": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				},
				"size": {
					"type": "integer"
				},
				"type": {
					"type": "string"
				},
				"uid": {
					"type": "string"
				},
				"updateTime": {
					"type": "string"
				},
				"user": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.FileTree": {
			"properties": {
				"children": {
					"items": {
						"$ref": "#/definitions/response.FileTree"
					},
					"type": "array"
				},
				"extension": {
					"type": "string"
				},
				"id": {
					"type": "string"
				},
				"isDir": {
					"type": "boolean"
				},
				"name": {
					"type": "string"
				},
				"path": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.IgnoredApp": {
			"properties": {
				"detailID": {
					"type": "integer"
				},
				"icon": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.NginxParam": {
			"properties": {
				"name": {
					"type": "string"
				},
				"params": {
					"items": {
						"type": "string"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"response.NginxStatus": {
			"properties": {
				"accepts": {
					"type": "string"
				},
				"active": {
					"type": "string"
				},
				"handled": {
					"type": "string"
				},
				"reading": {
					"type": "string"
				},
				"requests": {
					"type": "string"
				},
				"waiting": {
					"type": "string"
				},
				"writing": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.PHPConfig": {
			"properties": {
				"disableFunctions": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"params": {
					"additionalProperties": {
						"type": "string"
					},
					"type": "object"
				},
				"uploadMaxSize": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.PHPExtensionsDTO": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"extensions": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteAcmeAccountDTO": {
			"properties": {
				"createdAt": {
					"type": "string"
				},
				"eabHmacKey": {
					"type": "string"
				},
				"eabKid": {
					"type": "string"
				},
				"email": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"url": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteCADTO": {
			"properties": {
				"city": {
					"type": "string"
				},
				"commonName": {
					"type": "string"
				},
				"country": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"csr": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"keyType": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"organization": {
					"type": "string"
				},
				"organizationUint": {
					"type": "string"
				},
				"privateKey": {
					"type": "string"
				},
				"province": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteDNSRes": {
			"properties": {
				"domain": {
					"type": "string"
				},
				"err": {
					"type": "string"
				},
				"resolve": {
					"type": "string"
				},
				"value": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteDTO": {
			"properties": {
				"IPV6": {
					"type": "boolean"
				},
				"accessLog": {
					"type": "boolean"
				},
				"accessLogPath": {
					"type": "string"
				},
				"alias": {
					"type": "string"
				},
				"appInstallId": {
					"type": "integer"
				},
				"appName": {
					"type": "string"
				},
				"createdAt": {
					"type": "string"
				},
				"defaultServer": {
					"type": "boolean"
				},
				"domains": {
					"items": {
						"$ref": "#/definitions/model.WebsiteDomain"
					},
					"type": "array"
				},
				"errorLog": {
					"type": "boolean"
				},
				"errorLogPath": {
					"type": "string"
				},
				"expireDate": {
					"type": "string"
				},
				"ftpId": {
					"type": "integer"
				},
				"group": {
					"type": "string"
				},
				"httpConfig": {
					"type": "string"
				},
				"id": {
					"type": "integer"
				},
				"primaryDomain": {
					"type": "string"
				},
				"protocol": {
					"type": "string"
				},
				"proxy": {
					"type": "string"
				},
				"proxyType": {
					"type": "string"
				},
				"remark": {
					"type": "string"
				},
				"rewrite": {
					"type": "string"
				},
				"runtimeID": {
					"type": "integer"
				},
				"runtimeName": {
					"type": "string"
				},
				"siteDir": {
					"type": "string"
				},
				"sitePath": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"type": {
					"type": "string"
				},
				"updatedAt": {
					"type": "string"
				},
				"user": {
					"type": "string"
				},
				"webSiteGroupId": {
					"type": "integer"
				},
				"webSiteSSL": {
					"$ref": "#/definitions/model.WebsiteSSL"
				},
				"webSiteSSLId": {
					"type": "integer"
				}
			},
			"type": "object"
		},
		"response.WebsiteHTTPS": {
			"properties": {
				"SSL": {
					"$ref": "#/definitions/model.WebsiteSSL"
				},
				"SSLProtocol": {
					"items": {
						"type": "string"
					},
					"type": "array"
				},
				"algorithm": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"hsts": {
					"type": "boolean"
				},
				"httpConfig": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteLog": {
			"properties": {
				"content": {
					"type": "string"
				},
				"enable": {
					"type": "boolean"
				},
				"end": {
					"type": "boolean"
				},
				"path": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"response.WebsiteNginxConfig": {
			"properties": {
				"enable": {
					"type": "boolean"
				},
				"params": {
					"items": {
						"$ref": "#/definitions/response.NginxParam"
					},
					"type": "array"
				}
			},
			"type": "object"
		},
		"response.WebsitePreInstallCheck": {
			"properties": {
				"appName": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"status": {
					"type": "string"
				},
				"version": {
					"type": "string"
				}
			},
			"type": "object"
		}
	}
}`

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
}