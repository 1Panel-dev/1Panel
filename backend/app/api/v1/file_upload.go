package v1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/gin-gonic/gin"
)

func mergeChunks(fileName string, fileDir string, dstDir string, chunkCount int) error {
	//fileInfoList, err := ioutil.ReadDir(fileDir)
	//if err != nil {
	//	return err
	//}

	targetFile, err := os.Create(filepath.Join(dstDir, fileName))
	if err != nil {
		return err
	}
	defer targetFile.Close()

	for i := 0; i < chunkCount; i++ {
		chunkPath := filepath.Join(fileDir, fmt.Sprintf("%s.%d", fileName, i))
		chunkData, err := ioutil.ReadFile(chunkPath)
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

func (b *BaseApi) UploadChunkFiles(c *gin.Context) {
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
	if err := fileOp.CreateDir("uploads", 0755); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	//fileID := uuid.New().String()
	filename := c.PostForm("filename")
	fileDir := filepath.Join(global.CONF.System.DataDir, "upload", filename)

	_ = os.MkdirAll(fileDir, 0755)
	filePath := filepath.Join(fileDir, filename)

	emptyFile, err := os.Create(filePath)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	emptyFile.Close()

	chunkData, err := ioutil.ReadAll(uploadFile)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	chunkPath := filepath.Join(fileDir, fmt.Sprintf("%s.%d", filename, chunkIndex))
	err = ioutil.WriteFile(chunkPath, chunkData, 0644)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if chunkIndex+1 == chunkCount {
		err = mergeChunks(filename, fileDir, c.PostForm("path"), chunkCount)
		if err != nil {
			fmt.Println(err.Error())
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrAppDelete, err)
			return
		}
		helper.SuccessWithData(c, true)
	} else {
		return
	}
}
