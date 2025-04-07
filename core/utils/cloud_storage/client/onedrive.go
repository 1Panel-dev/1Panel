package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
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
	"time"

	odsdk "github.com/goh-chunlin/go-onedrive/onedrive"
	"golang.org/x/oauth2"
)

type oneDriveClient struct {
	client odsdk.Client
}

func NewOneDriveClient(vars map[string]interface{}) (*oneDriveClient, error) {
	token, err := RefreshToken("refresh_token", "accessToken", vars)
	if err != nil {
		return nil, err
	}
	isCN := loadParamFromVars("isCN", vars)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := odsdk.NewClient(tc)
	if isCN == "true" {
		client.BaseURL, _ = url.Parse("https://microsoftgraph.chinacloudapi.cn/v1.0/")
	}
	return &oneDriveClient{client: *client}, nil
}

func (o oneDriveClient) ListBuckets() ([]interface{}, error) {
	return nil, nil
}

func (o oneDriveClient) Upload(src, target string) (bool, error) {
	target = "/" + strings.TrimPrefix(target, "/")
	if _, err := o.loadIDByPath(path.Dir(target)); err != nil {
		if !strings.Contains(err.Error(), "itemNotFound") {
			return false, err
		}
		if err := o.createFolder(path.Dir(target)); err != nil {
			return false, fmt.Errorf("create dir before upload failed, err: %v", err)
		}
	}

	ctx := context.Background()
	folderID, err := o.loadIDByPath(path.Dir(target))
	if err != nil {
		return false, err
	}
	fileInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return false, errors.New("only file is allowed to be uploaded here")
	}
	var isOk bool
	if fileInfo.Size() < 4*1024*1024 {
		isOk, err = o.upSmall(src, folderID, fileInfo.Size())
	} else {
		isOk, err = o.upBig(ctx, src, folderID, fileInfo.Size())
	}
	return isOk, err
}

func (o oneDriveClient) Delete(path string) (bool, error) {
	path = "/" + strings.TrimPrefix(path, "/")
	req, err := o.client.NewRequest("DELETE", fmt.Sprintf("me/drive/root:%s", path), nil)
	if err != nil {
		return false, fmt.Errorf("new request for delete file failed, err: %v \n", err)
	}
	if err := o.client.Do(context.Background(), req, false, nil); err != nil {
		return false, fmt.Errorf("do request for delete file failed, err: %v \n", err)
	}

	return true, nil
}

func (o *oneDriveClient) loadIDByPath(path string) (string, error) {
	pathItem := "root:" + path
	if path == "/" {
		pathItem = "root"
	}
	req, err := o.client.NewRequest("GET", fmt.Sprintf("me/drive/%s", pathItem), nil)
	if err != nil {
		return "", fmt.Errorf("new request for file id failed, err: %v", err)
	}
	var driveItem *odsdk.DriveItem
	if err := o.client.Do(context.Background(), req, false, &driveItem); err != nil {
		return "", fmt.Errorf("do request for file id failed, err: %v", err)
	}
	return driveItem.Id, nil
}

func RefreshToken(grantType string, tokenType string, varMap map[string]interface{}) (string, error) {
	data := url.Values{}
	isCN := loadParamFromVars("isCN", varMap)
	data.Set("client_id", loadParamFromVars("client_id", varMap))
	data.Set("client_secret", loadParamFromVars("client_secret", varMap))
	if grantType == "refresh_token" {
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", loadParamFromVars("refresh_token", varMap))
	} else {
		data.Set("grant_type", "authorization_code")
		data.Set("code", loadParamFromVars("code", varMap))
	}
	data.Set("redirect_uri", loadParamFromVars("redirect_uri", varMap))
	client := &http.Client{}
	defer client.CloseIdleConnections()
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	if isCN == "true" {
		url = "https://login.chinacloudapi.cn/common/oauth2/v2.0/token"
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("new http post client for access token failed, err: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request for access token failed, err: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read data from response body failed, err: %v", err)
	}

	tokenMap := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &tokenMap); err != nil {
		return "", fmt.Errorf("unmarshal data from response body failed, err: %v", err)
	}
	if tokenType == "accessToken" {
		accessToken, ok := tokenMap["access_token"].(string)
		if !ok {
			return "", errors.New("no such access token in response")
		}
		tokenMap = nil
		return accessToken, nil
	}
	refreshToken, ok := tokenMap["refresh_token"].(string)
	if !ok {
		return "", errors.New("no such access token in response")
	}
	tokenMap = nil
	return refreshToken, nil
}

func (o *oneDriveClient) createFolder(parent string) error {
	if _, err := o.loadIDByPath(path.Dir(parent)); err != nil {
		if !strings.Contains(err.Error(), "itemNotFound") {
			return err
		}
		_ = o.createFolder(path.Dir(parent))
	}
	item2, err := o.loadIDByPath(path.Dir(parent))
	if err != nil {
		return err
	}
	if _, err := o.client.DriveItems.CreateNewFolder(context.Background(), "", item2, path.Base(parent)); err != nil {
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

func (o *oneDriveClient) NewSessionFileUploadRequest(absoluteUrl string, grandOffset, grandTotalSize int64, byteReader *bytes.Reader) (*http.Request, error) {
	apiUrl, err := o.client.BaseURL.Parse(absoluteUrl)
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

func (o *oneDriveClient) upSmall(srcPath, folderID string, fileSize int64) (bool, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, fileSize)
	_, _ = file.Read(buffer)
	fileReader := bytes.NewReader(buffer)
	apiURL := fmt.Sprintf("me/drive/items/%s:/%s:/content?@microsoft.graph.conflictBehavior=rename", url.PathEscape(folderID), path.Base(srcPath))

	mimeType := getMimeType(srcPath)
	req, err := o.client.NewFileUploadRequest(apiURL, mimeType, fileReader)
	if err != nil {
		return false, err
	}
	var response *DriveItem
	if err := o.client.Do(context.Background(), req, false, &response); err != nil {
		return false, fmt.Errorf("do request for list failed, err: %v", err)
	}
	return true, nil
}

func (o *oneDriveClient) upBig(ctx context.Context, srcPath, folderID string, fileSize int64) (bool, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	apiURL := fmt.Sprintf("me/drive/items/%s:/%s:/createUploadSession", url.PathEscape(folderID), path.Base(srcPath))
	sessionCreationRequestInside := NewUploadSessionCreationRequest{
		ConflictBehavior: "rename",
	}

	sessionCreationRequest := struct {
		Item        NewUploadSessionCreationRequest `json:"item"`
		DeferCommit bool                            `json:"deferCommit"`
	}{sessionCreationRequestInside, false}

	sessionCreationReq, err := o.client.NewRequest("POST", apiURL, sessionCreationRequest)
	if err != nil {
		return false, err
	}

	var sessionCreationResp *NewUploadSessionCreationResponse
	err = o.client.Do(ctx, sessionCreationReq, false, &sessionCreationResp)
	if err != nil {
		return false, fmt.Errorf("session creation failed %w", err)
	}

	fileSessionUploadUrl := sessionCreationResp.UploadURL

	sizePerSplit := int64(5 * 1024 * 1024)
	buffer := make([]byte, 5*1024*1024)
	splitCount := fileSize / sizePerSplit
	if fileSize%sizePerSplit != 0 {
		splitCount += 1
	}
	bfReader := bufio.NewReader(file)
	httpClient := http.Client{
		Timeout: time.Minute * 10,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for splitNow := int64(0); splitNow < splitCount; splitNow++ {
		length, err := bfReader.Read(buffer)
		if err != nil {
			return false, err
		}
		if int64(length) < sizePerSplit {
			bufferLast := buffer[:length]
			buffer = bufferLast
		}
		sessionFileUploadReq, err := o.NewSessionFileUploadRequest(fileSessionUploadUrl, splitNow*sizePerSplit, fileSize, bytes.NewReader(buffer))
		if err != nil {
			return false, err
		}
		res, err := httpClient.Do(sessionFileUploadReq)
		if err != nil {
			return false, err
		}
		res.Body.Close()
		if res.StatusCode != 201 && res.StatusCode != 202 && res.StatusCode != 200 {
			data, _ := io.ReadAll(res.Body)
			return false, errors.New(string(data))
		}
	}
	return true, nil
}

func getMimeType(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return ""
	}
	mimeType := http.DetectContentType(buffer)
	return mimeType
}
