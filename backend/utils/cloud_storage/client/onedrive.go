package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	odsdk "github.com/goh-chunlin/go-onedrive/onedrive"
	"golang.org/x/oauth2"
)

type oneDriveClient struct {
	Vars   map[string]interface{}
	client odsdk.Client
}

func NewOneDriveClient(vars map[string]interface{}) (*oneDriveClient, error) {
	token := ""
	if _, ok := vars["accessToken"]; ok {
		token = vars["accessToken"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	ctx := context.Background()

	newToken, err := refreshToken(token)
	if err != nil {
		return nil, err
	}
	_ = global.DB.Model(&model.Group{}).Where("type = ?", "OneDrive").Updates(map[string]interface{}{"credential": newToken}).Error

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: newToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := odsdk.NewClient(tc)
	return &oneDriveClient{client: *client}, nil
}

func (onedrive oneDriveClient) ListBuckets() ([]interface{}, error) {
	return nil, nil
}

func (onedrive oneDriveClient) Exist(path string) (bool, error) {
	path = "/" + strings.TrimPrefix(path, "/")
	fileID, err := onedrive.loadIDByPath(path)
	if err != nil {
		return false, err
	}

	return len(fileID) != 0, nil
}

func (onedrive oneDriveClient) Delete(path string) (bool, error) {
	path = "/" + strings.TrimPrefix(path, "/")
	req, err := onedrive.client.NewRequest("DELETE", fmt.Sprintf("me/drive/root:%s", path), nil)
	if err != nil {
		return false, fmt.Errorf("new request for delete file failed, err: %v \n", err)
	}
	if err := onedrive.client.Do(context.Background(), req, false, nil); err != nil {
		return false, fmt.Errorf("do request for delete file failed, err: %v \n", err)
	}

	return true, nil
}

func (onedrive oneDriveClient) Upload(src, target string) (bool, error) {
	target = "/" + strings.TrimPrefix(target, "/")
	if _, err := onedrive.loadIDByPath(path.Dir(target)); err != nil {
		if !strings.Contains(err.Error(), "itemNotFound") {
			return false, err
		}
		if err := onedrive.createFolder(path.Dir(target)); err != nil {
			return false, fmt.Errorf("create dir before upload failed, err: %v", err)
		}
	}

	ctx := context.Background()
	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return false, errors.New("Only file is allowed to be uploaded here.")
	}
	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	folderID, err := onedrive.loadIDByPath(path.Dir(target))
	if err != nil {
		return false, err
	}
	apiURL := fmt.Sprintf("me/drive/items/%s:/%s:/createUploadSession", url.PathEscape(folderID), fileName)
	sessionCreationRequestInside := NewUploadSessionCreationRequest{
		ConflictBehavior: "rename",
	}

	sessionCreationRequest := struct {
		Item        NewUploadSessionCreationRequest `json:"item"`
		DeferCommit bool                            `json:"deferCommit"`
	}{sessionCreationRequestInside, false}

	sessionCreationReq, err := onedrive.client.NewRequest("POST", apiURL, sessionCreationRequest)
	if err != nil {
		return false, err
	}

	var sessionCreationResp *NewUploadSessionCreationResponse
	err = onedrive.client.Do(ctx, sessionCreationReq, false, &sessionCreationResp)
	if err != nil {
		return false, fmt.Errorf("session creation failed %w", err)
	}

	fileSessionUploadUrl := sessionCreationResp.UploadURL

	sizePerSplit := int64(3200 * 1024)
	buffer := make([]byte, 3200*1024)
	splitCount := fileSize / sizePerSplit
	if fileSize%sizePerSplit != 0 {
		splitCount += 1
	}
	bfReader := bufio.NewReader(file)
	var fileUploadResp *UploadSessionUploadResponse
	for splitNow := int64(0); splitNow < splitCount; splitNow++ {
		length, err := bfReader.Read(buffer)
		if err != nil {
			return false, err
		}
		if int64(length) < sizePerSplit {
			bufferLast := buffer[:length]
			buffer = bufferLast
		}
		sessionFileUploadReq, err := onedrive.NewSessionFileUploadRequest(fileSessionUploadUrl, splitNow*sizePerSplit, fileSize, bytes.NewReader(buffer))
		if err != nil {
			return false, err
		}
		if err := onedrive.client.Do(ctx, sessionFileUploadReq, false, &fileUploadResp); err != nil {
			return false, err
		}
	}
	if fileUploadResp.Id == "" {
		return false, errors.New("something went wrong. file upload incomplete. consider upload the file in a step-by-step manner")
	}

	return true, nil
}

func (onedrive oneDriveClient) Download(src, target string) (bool, error) {
	src = "/" + strings.TrimPrefix(src, "/")
	req, err := onedrive.client.NewRequest("GET", fmt.Sprintf("me/drive/root:%s", src), nil)
	if err != nil {
		return false, fmt.Errorf("new request for file id failed, err: %v", err)
	}
	var driveItem *odsdk.DriveItem
	if err := onedrive.client.Do(context.Background(), req, false, &driveItem); err != nil {
		return false, fmt.Errorf("do request for file id failed, err: %v", err)
	}

	resp, err := http.Get(driveItem.DownloadURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	out, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer out.Close()
	buffer := make([]byte, 2*1024*1024)

	_, err = io.CopyBuffer(out, resp.Body, buffer)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (onedrive *oneDriveClient) ListObjects(prefix string) ([]interface{}, error) {
	prefix = "/" + strings.TrimPrefix(prefix, "/")
	folderID, err := onedrive.loadIDByPath(prefix)
	if err != nil {
		return nil, err
	}

	req, err := onedrive.client.NewRequest("GET", fmt.Sprintf("me/drive/items/%s/children", folderID), nil)
	if err != nil {
		return nil, fmt.Errorf("new request for list failed, err: %v", err)
	}
	var driveItems *odsdk.OneDriveDriveItemsResponse
	if err := onedrive.client.Do(context.Background(), req, false, &driveItems); err != nil {
		return nil, fmt.Errorf("do request for list failed, err: %v", err)
	}
	for _, item := range driveItems.DriveItems {
		return nil, fmt.Errorf("id: %v, name: %s \n", item.Id, item.Name)
	}

	var itemList []interface{}
	for _, item := range driveItems.DriveItems {
		itemList = append(itemList, item.Name)
	}
	return itemList, nil
}

func (onedrive *oneDriveClient) loadIDByPath(path string) (string, error) {
	pathItem := "root:" + path
	if path == "/" {
		pathItem = "root"
	}
	req, err := onedrive.client.NewRequest("GET", fmt.Sprintf("me/drive/%s", pathItem), nil)
	if err != nil {
		return "", fmt.Errorf("new request for file id failed, err: %v", err)
	}
	var driveItem *odsdk.DriveItem
	if err := onedrive.client.Do(context.Background(), req, false, &driveItem); err != nil {
		return "", fmt.Errorf("do request for file id failed, err: %v", err)
	}
	return driveItem.Id, nil
}

func refreshToken(oldToken string) (string, error) {
	data := url.Values{}
	data.Set("client_id", global.CONF.System.OneDriveID)
	data.Set("client_secret", global.CONF.System.OneDriveSc)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", oldToken)
	data.Set("redirect_uri", constant.OneDriveRedirectURI)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://login.microsoftonline.com/common/oauth2/v2.0/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("new http post client for access token failed, err: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request for access token failed, err: %v", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read data from response body failed, err: %v", err)
	}
	defer resp.Body.Close()

	tokenMap := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &tokenMap); err != nil {
		return "", fmt.Errorf("unmarshal data from response body failed, err: %v", err)
	}
	accessToken, ok := tokenMap["access_token"].(string)
	if !ok {
		return "", errors.New("no such access token in response")
	}
	return accessToken, nil
}

func (onedrive *oneDriveClient) createFolder(parent string) error {
	if _, err := onedrive.loadIDByPath(path.Dir(parent)); err != nil {
		if !strings.Contains(err.Error(), "itemNotFound") {
			return err
		}
		_ = onedrive.createFolder(path.Dir(parent))
	}
	item2, err := onedrive.loadIDByPath(path.Dir(parent))
	if err != nil {
		return err
	}
	if _, err := onedrive.client.DriveItems.CreateNewFolder(context.Background(), "", item2, path.Base(parent)); err != nil {
		return err
	}
	return nil
}

type NewUploadSessionCreationRequest struct {
	ConflictBehavior string `json:"@microsoft.graph.conflictBehavior,omitempty"`
}
type NewUploadSessionCreationResponse struct {
	UploadURL          string `json:"uploadUrl"`
	ExpirationDateTime string `json:"expirationDateTime"`
}
type UploadSessionUploadResponse struct {
	ExpirationDateTime string   `json:"expirationDateTime"`
	NextExpectedRanges []string `json:"nextExpectedRanges"`
	DriveItem
}
type DriveItem struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	DownloadURL string `json:"@microsoft.graph.downloadUrl"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
	WebURL      string `json:"webUrl"`
}

func (onedrive *oneDriveClient) NewSessionFileUploadRequest(absoluteUrl string, grandOffset, grandTotalSize int64, byteReader *bytes.Reader) (*http.Request, error) {
	apiUrl, err := onedrive.client.BaseURL.Parse(absoluteUrl)
	if err != nil {
		return nil, err
	}
	absoluteUrl = apiUrl.String()
	contentLength := byteReader.Size()
	req, err := http.NewRequest("PUT", absoluteUrl, byteReader)
	req.Header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	preliminaryLength := grandOffset
	preliminaryRange := grandOffset + contentLength - 1
	if preliminaryRange >= grandTotalSize {
		preliminaryRange = grandTotalSize - 1
		preliminaryLength = preliminaryRange - grandOffset + 1
	}
	req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", preliminaryLength, preliminaryRange, grandTotalSize))

	return req, err
}
