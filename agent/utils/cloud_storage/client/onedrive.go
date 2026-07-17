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

	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

const oneDriveGlobalBaseURL = "https://graph.microsoft.com/v1.0/"
const oneDriveChinaBaseURL = "https://microsoftgraph.chinacloudapi.cn/v1.0/"

type oneDriveClient struct {
	client  *http.Client
	baseURL *url.URL
	token   string
}

func NewOneDriveClient(vars map[string]interface{}) (*oneDriveClient, error) {
	token, err := RefreshToken("refresh_token", "accessToken", vars)
	if err != nil {
		return nil, err
	}

	baseURL := oneDriveGlobalBaseURL
	if loadParamFromVars("isCN", vars) == "true" {
		baseURL = oneDriveChinaBaseURL
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse OneDrive base URL failed: %w", err)
	}
	return &oneDriveClient{client: http.DefaultClient, baseURL: parsedBaseURL, token: token}, nil
}

func (o oneDriveClient) ListBuckets() ([]interface{}, error) { return nil, nil }

func (o oneDriveClient) Exist(itemPath string) (bool, error) {
	_, err := o.loadIDByPath(normalizeDrivePath(itemPath))
	if isOneDriveNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func (o oneDriveClient) Size(itemPath string) (int64, error) {
	var item DriveItem
	if err := o.getDriveItem(context.Background(), normalizeDrivePath(itemPath), &item); err != nil {
		return 0, err
	}
	return item.Size, nil
}

func (o oneDriveClient) Delete(itemPath string) (bool, error) {
	itemPath = normalizeDrivePath(itemPath)
	if err := o.doJSON(context.Background(), http.MethodDelete, "me/drive/root:"+escapeDrivePath(itemPath), nil, nil); err != nil {
		return false, fmt.Errorf("delete OneDrive file failed: %w", err)
	}
	return true, nil
}

func (o oneDriveClient) Upload(ctx context.Context, src, target string) (bool, error) {
	target = normalizeDrivePath(target)
	parentPath := path.Dir(target)
	if _, err := o.loadIDByPath(parentPath); err != nil {
		if !isOneDriveNotFound(err) {
			return false, err
		}
		if err := o.createFolder(parentPath); err != nil {
			return false, fmt.Errorf("create directory before upload failed: %w", err)
		}
	}
	folderID, err := o.loadIDByPath(parentPath)
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
	if fileInfo.Size() < 4*1024*1024 {
		return o.upSmall(ctx, src, folderID)
	}
	return o.upBig(ctx, src, folderID, fileInfo.Size())
}

func (o oneDriveClient) Download(src, target string) (bool, error) {
	var item DriveItem
	if err := o.getDriveItem(context.Background(), normalizeDrivePath(src), &item); err != nil {
		return false, err
	}
	if item.DownloadURL == "" {
		return false, errors.New("OneDrive download URL is missing")
	}
	resp, err := o.client.Get(item.DownloadURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, newOneDriveHTTPError(resp)
	}
	out, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer out.Close()
	_, err = io.CopyBuffer(out, resp.Body, make([]byte, 2*1024*1024))
	return err == nil, err
}

func (o *oneDriveClient) ListObjects(prefix string) ([]string, error) {
	folderID, err := o.loadIDByPath(normalizeDrivePath(prefix))
	if err != nil {
		return nil, err
	}
	var items oneDriveItemsResponse
	endpoint := fmt.Sprintf("me/drive/items/%s/children", url.PathEscape(folderID))
	if err := o.doJSON(context.Background(), http.MethodGet, endpoint, nil, &items); err != nil {
		return nil, fmt.Errorf("list OneDrive files failed: %w", err)
	}
	result := make([]string, 0, len(items.Value))
	for _, item := range items.Value {
		result = append(result, item.Name)
	}
	return result, nil
}

func (o *oneDriveClient) loadIDByPath(itemPath string) (string, error) {
	var item DriveItem
	if err := o.getDriveItem(context.Background(), itemPath, &item); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (o *oneDriveClient) getDriveItem(ctx context.Context, itemPath string, result *DriveItem) error {
	endpoint := "me/drive/root"
	if itemPath != "/" {
		endpoint += ":" + escapeDrivePath(itemPath)
	}
	if err := o.doJSON(ctx, http.MethodGet, endpoint, nil, result); err != nil {
		return fmt.Errorf("get OneDrive item failed: %w", err)
	}
	return nil
}

func (o *oneDriveClient) createFolder(parent string) error {
	if parent == "/" {
		return nil
	}
	parentID, err := o.loadIDByPath(path.Dir(parent))
	if err != nil {
		if !isOneDriveNotFound(err) {
			return err
		}
		if err := o.createFolder(path.Dir(parent)); err != nil {
			return err
		}
		parentID, err = o.loadIDByPath(path.Dir(parent))
		if err != nil {
			return err
		}
	}
	body := struct {
		Name   string                 `json:"name"`
		Folder map[string]interface{} `json:"folder"`
	}{Name: path.Base(parent), Folder: map[string]interface{}{}}
	endpoint := fmt.Sprintf("me/drive/items/%s/children", url.PathEscape(parentID))
	return o.doJSON(context.Background(), http.MethodPost, endpoint, body, nil)
}

func (o *oneDriveClient) upSmall(ctx context.Context, srcPath, folderID string) (bool, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	endpoint := fmt.Sprintf("me/drive/items/%s:/%s:/content?@microsoft.graph.conflictBehavior=rename", url.PathEscape(folderID), url.PathEscape(path.Base(srcPath)))
	req, err := o.newGraphRequest(ctx, http.MethodPut, endpoint, file)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", files.GetMimeType(srcPath))
	if err := o.do(req, nil); err != nil {
		return false, fmt.Errorf("upload OneDrive file failed: %w", err)
	}
	return true, nil
}

func (o *oneDriveClient) upBig(ctx context.Context, srcPath, folderID string, fileSize int64) (bool, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	body := struct {
		Item struct {
			ConflictBehavior string `json:"@microsoft.graph.conflictBehavior"`
		} `json:"item"`
	}{}
	body.Item.ConflictBehavior = "rename"
	var session oneDriveUploadSession
	endpoint := fmt.Sprintf("me/drive/items/%s:/%s:/createUploadSession", url.PathEscape(folderID), url.PathEscape(path.Base(srcPath)))
	if err := o.doJSON(ctx, http.MethodPost, endpoint, body, &session); err != nil {
		return false, fmt.Errorf("create OneDrive upload session failed: %w", err)
	}

	const chunkSize int64 = 5 * 1024 * 1024
	reader := bufio.NewReader(file)
	buffer := make([]byte, chunkSize)
	for offset := int64(0); offset < fileSize; {
		length, readErr := io.ReadFull(reader, buffer)
		if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
			return false, readErr
		}
		if length == 0 {
			return false, io.ErrUnexpectedEOF
		}
		if err := o.uploadChunk(ctx, session.UploadURL, offset, fileSize, buffer[:length]); err != nil {
			return false, err
		}
		offset += int64(length)
	}
	return true, nil
}

func (o *oneDriveClient) uploadChunk(ctx context.Context, uploadURL string, offset, total int64, chunk []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Length", strconv.Itoa(len(chunk)))
	req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, offset+int64(len(chunk))-1, total))
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return newOneDriveHTTPError(resp)
	}
	return nil
}

func (o *oneDriveClient) doJSON(ctx context.Context, method, endpoint string, body, result interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := o.newGraphRequest(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return o.do(req, result)
}

func (o *oneDriveClient) newGraphRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	apiURL, err := o.baseURL.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, apiURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.token)
	return req, nil
}

func (o *oneDriveClient) do(req *http.Request, result interface{}) error {
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newOneDriveHTTPError(resp)
	}
	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

type DriveItem struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	DownloadURL string `json:"@microsoft.graph.downloadUrl"`
	Size        int64  `json:"size"`
}

type oneDriveItemsResponse struct {
	Value []DriveItem `json:"value"`
}

type oneDriveUploadSession struct {
	UploadURL string `json:"uploadUrl"`
}

type oneDriveError struct {
	Details struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	statusCode int
}

func (e *oneDriveError) Error() string {
	if e.Details.Code != "" || e.Details.Message != "" {
		return fmt.Sprintf("OneDrive API error: %s: %s", e.Details.Code, e.Details.Message)
	}
	return fmt.Sprintf("OneDrive API returned HTTP %d", e.statusCode)
}

func newOneDriveHTTPError(resp *http.Response) error {
	apiErr := &oneDriveError{statusCode: resp.StatusCode}
	if err := json.NewDecoder(resp.Body).Decode(apiErr); err != nil {
		return fmt.Errorf("OneDrive API returned HTTP %d", resp.StatusCode)
	}
	return apiErr
}

func isOneDriveNotFound(err error) bool {
	var apiErr *oneDriveError
	return errors.As(err, &apiErr) && (apiErr.statusCode == http.StatusNotFound || apiErr.Details.Code == "itemNotFound")
}

func normalizeDrivePath(itemPath string) string { return "/" + strings.TrimPrefix(itemPath, "/") }

func escapeDrivePath(itemPath string) string {
	parts := strings.Split(strings.TrimPrefix(itemPath, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return "/" + strings.Join(parts, "/")
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
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	if isCN == "true" {
		tokenURL = "https://login.chinacloudapi.cn/common/oauth2/v2.0/token"
	}
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create access token request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request access token failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", newOneDriveHTTPError(resp)
	}
	var tokenMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenMap); err != nil {
		return "", fmt.Errorf("decode token response failed: %w", err)
	}
	key := "refresh_token"
	if tokenType == "accessToken" {
		key = "access_token"
	}
	token, ok := tokenMap[key].(string)
	if !ok {
		return "", fmt.Errorf("no %s in token response", key)
	}
	return token, nil
}
