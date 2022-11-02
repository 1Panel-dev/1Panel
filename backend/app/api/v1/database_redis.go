package v1

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/terminal"
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (b *BaseApi) LoadRedisStatus(c *gin.Context) {
	data, err := redisService.LoadStatus()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

func (b *BaseApi) LoadRedisConf(c *gin.Context) {
	data, err := redisService.LoadConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

func (b *BaseApi) LoadPersistenceConf(c *gin.Context) {
	data, err := redisService.LoadPersistenceConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

func (b *BaseApi) UpdateRedisConf(c *gin.Context) {
	var req dto.RedisConfUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := redisService.UpdateConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) UpdateRedisConfByFile(c *gin.Context) {
	var req dto.RedisConfUpdateByFile
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	redisInfo, err := redisService.LoadConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	path := fmt.Sprintf("/opt/1Panel/data/apps/redis/%s/conf/redis.conf", redisInfo.Name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) RedisExec(c *gin.Context) {
	redisConf, err := redisService.LoadConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	client, err := docker.NewDockerClient()
	if wshandleError(wsConn, errors.WithMessage(err, "New docker client failed.")) {
		return
	}

	auth := "redis-cli"
	if len(redisConf.Requirepass) != 0 {
		auth = fmt.Sprintf("redis-cli -a %s --no-auth-warning", redisConf.Requirepass)
	}
	conf := types.ExecConfig{Tty: true, Cmd: []string{"bash"}, AttachStderr: true, AttachStdin: true, AttachStdout: true, User: "root"}
	ir, err := client.ContainerExecCreate(context.TODO(), redisConf.ContainerName, conf)
	if wshandleError(wsConn, errors.WithMessage(err, "failed to set exec conf.")) {
		return
	}
	hr, err := client.ContainerExecAttach(c, ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if wshandleError(wsConn, errors.WithMessage(err, "failed to set up the connection.")) {
		return
	}
	defer hr.Close()

	sws, err := terminal.NewExecConn(cols, rows, wsConn, hr.Conn, auth)
	if wshandleError(wsConn, err) {
		return
	}

	quitChan := make(chan bool, 3)
	ctx, cancel := context.WithCancel(context.Background())
	sws.Start(ctx, quitChan)
	<-quitChan
	cancel()

	if wshandleError(wsConn, err) {
		return
	}
}
