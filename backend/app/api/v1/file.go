package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	websocket2 "github.com/1Panel-dev/1Panel/backend/utils/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// @Tags File
// @Summary List files
// @Description 获取文件列表
// @Accept json
// @Param request body request.FileOption true "request"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Router /files/search [post]
func (b *BaseApi) ListFiles(c *gin.Context) {
	var req request.FileOption
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	files, err := fileService.GetFileList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, files)
}

// @Tags File
// @Summary Page file
// @Description 分页获取上传文件
// @Accept json
// @Param request body request.SearchUploadWithPage true "request"
// @Success 200 {array} response.FileInfo
// @Security ApiKeyAuth
// @Router /files/upload/search [post]
func (b *BaseApi) SearchUploadWithPage(c *gin.Context) {
	var req request.SearchUploadWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, files, err := fileService.SearchUploadWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: files,
		Total: total,
	})
}

// @Tags File
// @Summary Load files tree
// @Description 加载文件树
// @Accept json
// @Param request body request.FileOption true "request"
// @Success 200 {array} response.FileTree
// @Security ApiKeyAuth
// @Router /files/tree [post]
func (b *BaseApi) GetFileTree(c *gin.Context) {
	var req request.FileOption
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	tree, err := fileService.GetFileTree(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, tree)
}

// @Tags File
// @Summary Create file
// @Description 创建文件/文件夹
// @Accept json
// @Param request body request.FileCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建文件/文件夹 [path]","formatEN":"Create dir or file [path]"}
func (b *BaseApi) CreateFile(c *gin.Context) {
	var req request.FileCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.Create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Delete file
// @Description 删除文件/文件夹
// @Accept json
// @Param request body request.FileDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/del [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除文件/文件夹 [path]","formatEN":"Delete dir or file [path]"}
func (b *BaseApi) DeleteFile(c *gin.Context) {
	var req request.FileDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.Delete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Batch delete file
// @Description 批量删除文件/文件夹
// @Accept json
// @Param request body request.FileBatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/batch/del [post]
// @x-panel-log {"bodyKeys":["paths"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量删除文件/文件夹 [paths]","formatEN":"Batch delete dir or file [paths]"}
func (b *BaseApi) BatchDeleteFile(c *gin.Context) {
	var req request.FileBatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.BatchDelete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Change file mode
// @Description 修改文件权限
// @Accept json
// @Param request body request.FileCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/mode [post]
// @x-panel-log {"bodyKeys":["path","mode"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改权限 [paths] => [mode]","formatEN":"Change mode [paths] => [mode]"}
func (b *BaseApi) ChangeFileMode(c *gin.Context) {
	var req request.FileCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.ChangeMode(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Change file owner
// @Description 修改文件用户/组
// @Accept json
// @Param request body request.FileRoleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/owner [post]
// @x-panel-log {"bodyKeys":["path","user","group"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改用户/组 [paths] => [user]/[group]","formatEN":"Change owner [paths] => [user]/[group]"}
func (b *BaseApi) ChangeFileOwner(c *gin.Context) {
	var req request.FileRoleUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := fileService.ChangeOwner(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Compress file
// @Description 压缩文件
// @Accept json
// @Param request body request.FileCompress true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/compress [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"压缩文件 [name]","formatEN":"Compress file [name]"}
func (b *BaseApi) CompressFile(c *gin.Context) {
	var req request.FileCompress
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.Compress(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Decompress file
// @Description 解压文件
// @Accept json
// @Param request body request.FileDeCompress true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/decompress [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"解压 [path]","formatEN":"Decompress file [path]"}
func (b *BaseApi) DeCompressFile(c *gin.Context) {
	var req request.FileDeCompress
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := fileService.DeCompress(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Load file content
// @Description 获取文件内容
// @Accept json
// @Param request body request.FileContentReq true "request"
// @Success 200 {object} response.FileInfo
// @Security ApiKeyAuth
// @Router /files/content [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"获取文件内容 [path]","formatEN":"Load file content [path]"}
func (b *BaseApi) GetContent(c *gin.Context) {
	var req request.FileContentReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	info, err := fileService.GetContent(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags File
// @Summary Update file content
// @Description 更新文件内容
// @Accept json
// @Param request body request.FileEdit true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/save [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新文件内容 [path]","formatEN":"Update file content [path]"}
func (b *BaseApi) SaveContent(c *gin.Context) {
	var req request.FileEdit
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := fileService.SaveContent(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Upload file
// @Description 上传文件
// @Param file formData file true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/upload [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"上传文件 [path]","formatEN":"Upload file [path]"}
func (b *BaseApi) UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	files := form.File["file"]
	paths := form.Value["path"]
	if len(paths) == 0 || !strings.Contains(paths[0], "/") {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error paths in request"))
		return
	}
	dir := path.Dir(paths[0])
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, fmt.Errorf("mkdir %s failed, err: %v", dir, err))
			return
		}
	}
	success := 0
	failures := make(buserr.MultiErr)
	for _, file := range files {
		if err := c.SaveUploadedFile(file, path.Join(paths[0], file.Filename)); err != nil {
			e := fmt.Errorf("upload [%s] file failed, err: %v", file.Filename, err)
			failures[file.Filename] = e
			global.LOG.Error(e)
			continue
		}
		success++
	}
	if success == 0 {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, failures)
	} else {
		helper.SuccessWithMsg(c, fmt.Sprintf("%d files upload success", success))
	}
}

// @Tags File
// @Summary Check file exist
// @Description 检测文件是否存在
// @Accept json
// @Param request body request.FilePathCheck true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/check [post]
func (b *BaseApi) CheckFile(c *gin.Context) {
	var req request.FilePathCheck
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if _, err := os.Stat(req.Path); err != nil {
		helper.SuccessWithData(c, false)
		return
	}
	helper.SuccessWithData(c, true)
}

// @Tags File
// @Summary Change file name
// @Description 修改文件名称
// @Accept json
// @Param request body request.FileRename true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/rename [post]
// @x-panel-log {"bodyKeys":["oldName","newName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重命名 [oldName] => [newName]","formatEN":"Rename [oldName] => [newName]"}
func (b *BaseApi) ChangeFileName(c *gin.Context) {
	var req request.FileRename
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := fileService.ChangeName(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Wget file
// @Description 下载远端文件
// @Accept json
// @Param request body request.FileWget true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/wget [post]
// @x-panel-log {"bodyKeys":["url","path","name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"下载 url => [path]/[name]","formatEN":"Download url => [path]/[name]"}
func (b *BaseApi) WgetFile(c *gin.Context) {
	var req request.FileWget
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	key, err := fileService.Wget(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, response.FileWgetRes{
		Key: key,
	})
}

// @Tags File
// @Summary Move file
// @Description 移动文件
// @Accept json
// @Param request body request.FileMove true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/move [post]
// @x-panel-log {"bodyKeys":["oldPaths","newPath"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"移动文件 [oldPaths] => [newPath]","formatEN":"Move [oldPaths] => [newPath]"}
func (b *BaseApi) MoveFile(c *gin.Context) {
	var req request.FileMove
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := fileService.MvFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Download file
// @Description 下载文件
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Router /files/download [get]
func (b *BaseApi) Download(c *gin.Context) {
	filePath := c.Query("path")
	file, err := os.Open(filePath)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
	}
	info, _ := file.Stat()
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(info.Name()))
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
}

// @Tags File
// @Summary Chunk Download file
// @Description 分片下载下载文件
// @Accept json
// @Param request body request.FileDownload true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/chunkdownload [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"下载文件 [name]","formatEN":"Download file [name]"}
func (b *BaseApi) DownloadChunkFiles(c *gin.Context) {
	var req request.FileChunkDownload
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.Path) {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrPathNotFound, nil)
		return
	}
	filePath := req.Path
	fstFile, err := fileOp.OpenFile(filePath)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	info, err := fstFile.Stat()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if info.IsDir() {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrFileDownloadDir, err)
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", req.Name))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Writer.Header().Set("Accept-Ranges", "bytes")

	if c.Request.Header.Get("Range") != "" {
		rangeHeader := c.Request.Header.Get("Range")
		rangeArr := strings.Split(rangeHeader, "=")[1]
		rangeParts := strings.Split(rangeArr, "-")

		startPos, _ := strconv.ParseInt(rangeParts[0], 10, 64)

		var endPos int64
		if rangeParts[1] == "" {
			endPos = info.Size() - 1
		} else {
			endPos, _ = strconv.ParseInt(rangeParts[1], 10, 64)
		}

		c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startPos, endPos, info.Size()))
		c.Writer.WriteHeader(http.StatusPartialContent)

		buffer := make([]byte, 1024*1024)
		file, err := os.Open(filePath)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		defer file.Close()

		_, _ = file.Seek(startPos, 0)
		reader := io.LimitReader(file, endPos-startPos+1)
		_, err = io.CopyBuffer(c.Writer, reader, buffer)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
	} else {
		c.File(filePath)
	}
}

// @Tags File
// @Summary Load file size
// @Description 获取文件夹大小
// @Accept json
// @Param request body request.DirSizeReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/size [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"获取文件夹大小 [path]","formatEN":"Load file size [path]"}
func (b *BaseApi) Size(c *gin.Context) {
	var req request.DirSizeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := fileService.DirSize(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

func mergeChunks(fileName string, fileDir string, dstDir string, chunkCount int) error {
	op := files.NewFileOp()
	dstDir = strings.TrimSpace(dstDir)
	if _, err := os.Stat(dstDir); err != nil && os.IsNotExist(err) {
		if err = op.CreateDir(dstDir, os.ModePerm); err != nil {
			return err
		}
	}
	targetFile, err := os.Create(filepath.Join(dstDir, fileName))
	if err != nil {
		return err
	}
	defer targetFile.Close()

	for i := 0; i < chunkCount; i++ {
		chunkPath := filepath.Join(fileDir, fmt.Sprintf("%s.%d", fileName, i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}
		_, err = targetFile.Write(chunkData)
		if err != nil {
			return err
		}
	}

	return files.NewFileOp().DeleteDir(fileDir)
}

// @Tags File
// @Summary ChunkUpload file
// @Description 分片上传文件
// @Param file formData file true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/chunkupload [post]
func (b *BaseApi) UploadChunkFiles(c *gin.Context) {
	var err error
	fileForm, err := c.FormFile("chunk")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	uploadFile, err := fileForm.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	chunkIndex, err := strconv.Atoi(c.PostForm("chunkIndex"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	chunkCount, err := strconv.Atoi(c.PostForm("chunkCount"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	fileOp := files.NewFileOp()
	tmpDir := path.Join(global.CONF.System.TmpDir, "upload")
	if !fileOp.Stat(tmpDir) {
		if err := fileOp.CreateDir(tmpDir, 0755); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
	}
	filename := c.PostForm("filename")
	fileDir := filepath.Join(tmpDir, filename)
	if chunkIndex == 0 {
		if fileOp.Stat(fileDir) {
			_ = fileOp.DeleteDir(fileDir)
		}
		_ = os.MkdirAll(fileDir, 0755)
	}
	filePath := filepath.Join(fileDir, filename)

	defer func() {
		if err != nil {
			_ = os.Remove(fileDir)
		}
	}()
	var (
		emptyFile *os.File
		chunkData []byte
	)

	emptyFile, err = os.Create(filePath)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	defer emptyFile.Close()

	chunkData, err = io.ReadAll(uploadFile)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, buserr.WithMap(constant.ErrFileUpload, map[string]interface{}{"name": filename, "detail": err.Error()}, err))
		return
	}

	chunkPath := filepath.Join(fileDir, fmt.Sprintf("%s.%d", filename, chunkIndex))
	err = os.WriteFile(chunkPath, chunkData, 0644)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, buserr.WithMap(constant.ErrFileUpload, map[string]interface{}{"name": filename, "detail": err.Error()}, err))
		return
	}

	if chunkIndex+1 == chunkCount {
		err = mergeChunks(filename, fileDir, c.PostForm("path"), chunkCount)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, buserr.WithMap(constant.ErrFileUpload, map[string]interface{}{"name": filename, "detail": err.Error()}, err))
			return
		}
		helper.SuccessWithData(c, true)
	} else {
		return
	}
}

var wsUpgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (b *BaseApi) Ws(c *gin.Context) {
	ws, err := wsUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	wsClient := websocket2.NewWsClient("fileClient", ws)
	go wsClient.Read()
	go wsClient.Write()
}

func (b *BaseApi) Keys(c *gin.Context) {
	res := &response.FileProcessKeys{}
	keys, err := global.CACHE.PrefixScanKey("file-wget-")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	res.Keys = keys
	helper.SuccessWithData(c, res)
}

// @Tags File
// @Summary Read file by Line
// @Description 按行读取日志文件
// @Param request body request.FileReadByLineReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/read [post]
func (b *BaseApi) ReadFileByLine(c *gin.Context) {
	var req request.FileReadByLineReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := fileService.ReadLogByLine(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags File
// @Summary Batch change file mode and owner
// @Description 批量修改文件权限和用户/组
// @Accept json
// @Param request body request.FileRoleReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/batch/role [post]
// @x-panel-log {"bodyKeys":["paths","mode","user","group"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量修改文件权限和用户/组 [paths] => [mode]/[user]/[group]","formatEN":"Batch change file mode and owner [paths] => [mode]/[user]/[group]"}
func (b *BaseApi) BatchChangeModeAndOwner(c *gin.Context) {
	var req request.FileRoleReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := fileService.BatchChangeModeAndOwner(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
	}
	helper.SuccessWithOutData(c)
}
