package client

import (
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

	"github.com/go-resty/resty/v2"
)

type googleDriveClient struct {
	accessToken string
}

func NewGoogleDriveClient(vars map[string]interface{}) (*googleDriveClient, error) {
	accessToken, err := RefreshGoogleToken("refresh_token", "accessToken", vars)
	if err != nil {
		return nil, err
	}
	return &googleDriveClient{accessToken: accessToken}, nil
}

func (g *googleDriveClient) ListBuckets() ([]interface{}, error) {
	return nil, nil
}

func (g *googleDriveClient) Exist(pathItem string) (bool, error) {
	pathItem = path.Join("root", pathItem)
	if _, err := g.loadFileWithName(pathItem); err != nil {
		return false, err
	}
	return true, nil
}

func (g *googleDriveClient) Size(pathItem string) (int64, error) {
	pathItem = path.Join("root", pathItem)
	fileInfo, err := g.loadFileWithName(pathItem)
	if err != nil {
		return 0, err
	}
	size, _ := strconv.ParseInt(fileInfo.Size, 10, 64)
	return size, nil
}

func (g *googleDriveClient) Delete(pathItem string) (bool, error) {
	pathItem = path.Join("root", pathItem)
	fileInfo, err := g.loadFileWithName(pathItem)
	if err != nil {
		return false, err
	}
	if len(fileInfo.ID) == 0 {
		return false, fmt.Errorf("no such file %s", pathItem)
	}
	res, err := g.googleRequest("https://www.googleapis.com/drive/v3/files/"+fileInfo.ID, http.MethodDelete, nil, nil)
	if err != nil {
		return false, err
	}
	fmt.Println(string(res))
	return true, nil
}

func (g *googleDriveClient) Upload(src, target string) (bool, error) {
	target = path.Join("/root", target)
	parentID := "root"
	var err error
	if path.Dir(target) != "/root" {
		parentID, err = g.mkdirWithPath(path.Dir(target))
		if err != nil {
			return false, err
		}
	}
	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	data := map[string]interface{}{
		"name":    fileInfo.Name(),
		"parents": []string{parentID},
	}
	urlItem := "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true"
	client := resty.New()
	client.SetProxy("http://127.0.0.1:7890")
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+g.accessToken).
		SetBody(data).
		Post(urlItem)
	if err != nil {
		return false, err
	}
	uploadUrl := resp.Header().Get("location")
	if _, err := g.googleRequest(uploadUrl, http.MethodPut, func(req *resty.Request) {
		req.SetHeader("Content-Length", strconv.FormatInt(fileInfo.Size(), 10)).SetBody(file)
	}, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (g *googleDriveClient) Download(src, target string) (bool, error) {
	src = path.Join("/root", src)
	fileInfo, err := g.loadFileWithName(src)
	if err != nil {
		return false, err
	}
	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media&acknowledgeAbuse=true", fileInfo.ID)
	if err := g.handleDownload(url, target); err != nil {
		return false, err
	}
	return true, nil
}

func (g *googleDriveClient) ListObjects(src string) ([]string, error) {
	if len(src) == 0 || src == "root" || src == "/root" {
		src = "root"
	} else {
		src = path.Join("/root", src)
	}
	fileInfos, err := g.loadDirWithPath(src)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range fileInfos {
		names = append(names, item.Name)
	}
	return names, nil
}

type googleFileResp struct {
	Files []googleFile `json:"files"`
}
type googleFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
}

func (g *googleDriveClient) mkdirWithPath(target string) (string, error) {
	pathItems := strings.Split(target, "/")
	var (
		fileInfos []googleFile
		err       error
	)
	parentID := "root"
	for i := 0; i < len(pathItems); i++ {
		if len(pathItems[i]) == 0 {
			continue
		}
		fileInfos, err = g.loadFileWithParentID(parentID)
		if err != nil {
			return "", err
		}
		isEnd := false
		if i == len(pathItems)-2 {
			isEnd = true
		}
		exist := false
		for _, item := range fileInfos {
			if item.Name == pathItems[i+1] {
				parentID = item.ID
				if isEnd {
					return item.ID, nil
				} else {
					exist = true
				}
			}
		}
		if !exist {
			parentID, err = g.mkdir(parentID, pathItems[i+1])
			if err != nil {
				return parentID, err
			}
			if isEnd {
				return parentID, nil
			}
		}
	}
	return "", errors.New("mkdir failed.")
}

type googleMkdirRes struct {
	ID string `json:"id"`
}

func (g *googleDriveClient) mkdir(parentID, name string) (string, error) {
	data := map[string]interface{}{
		"name":     name,
		"parents":  []string{parentID},
		"mimeType": "application/vnd.google-apps.folder",
	}
	res, err := g.googleRequest("https://www.googleapis.com/drive/v3/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	if err != nil {
		return "", err
	}
	var mkdirResp googleMkdirRes
	if err := json.Unmarshal(res, &mkdirResp); err != nil {
		return "", err
	}
	return mkdirResp.ID, nil
}

func (g *googleDriveClient) handleDownload(urlItem string, target string) error {
	req, err := http.NewRequest(http.MethodGet, urlItem, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+g.accessToken)
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("handle download with url failed, code: %v", response.StatusCode)
	}
	if _, err := os.Stat(path.Dir(target)); err != nil {
		_ = os.MkdirAll(path.Dir(target), os.ModePerm)
	}
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, response.Body); err != nil {
		return err
	}

	return nil
}

func (g *googleDriveClient) loadFileWithName(pathItem string) (googleFile, error) {
	pathItems := strings.Split(pathItem, "/")
	var (
		fileInfos []googleFile
		err       error
	)
	parentID := "root"
	for i := 0; i < len(pathItems); i++ {
		if len(pathItems[i]) == 0 {
			continue
		}
		fileInfos, err = g.loadFileWithParentID(parentID)
		if err != nil {
			return googleFile{}, err
		}
		isEnd := false
		if i == len(pathItems)-2 {
			isEnd = true
		}
		exist := false
		for _, item := range fileInfos {
			if item.Name == pathItems[i+1] {
				if isEnd {
					return item, nil
				} else {
					parentID = item.ID
					exist = true
				}
			}
		}
		if !exist {
			return googleFile{}, errors.New("no such file or dir")
		}

	}
	return googleFile{}, errors.New("no such file or dir")
}

func (g *googleDriveClient) loadFileWithParentID(parentID string) ([]googleFile, error) {
	query := map[string]string{
		"fields": "files(id,name,mimeType,size)",
		"q":      fmt.Sprintf("'%s' in parents and trashed = false", parentID),
	}

	res, err := g.googleRequest("https://www.googleapis.com/drive/v3/files", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, nil)
	if err != nil {
		return nil, err
	}
	var fileResp googleFileResp
	if err := json.Unmarshal(res, &fileResp); err != nil {
		return nil, err
	}
	return fileResp.Files, nil
}

func (g googleDriveClient) loadDirWithPath(path string) ([]googleFile, error) {
	pathItems := strings.Split(path, "/")
	var (
		fileInfos []googleFile
		err       error
	)
	parentID := "root"
	for i := 0; i < len(pathItems); i++ {
		if len(pathItems[i]) == 0 {
			continue
		}
		fileInfos, err = g.loadFileWithParentID(parentID)
		if err != nil {
			return fileInfos, err
		}
		if i == len(pathItems)-1 {
			return fileInfos, nil
		}
		exist := false
		for _, item := range fileInfos {
			if item.Name == pathItems[i+1] {
				parentID = item.ID
				exist = true
			}
		}
		if !exist {
			return nil, errors.New("no such file or dir")
		}
	}
	return fileInfos, errors.New("no such file or dir")
}

type reqCallback func(req *resty.Request)

func (g *googleDriveClient) googleRequest(urlItem, method string, callback reqCallback, resp interface{}) ([]byte, error) {
	client := resty.New()
	client.SetProxy("http://127.0.0.1:7890")
	req := client.R()
	req.SetHeader("Authorization", "Bearer "+g.accessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(req)
	}
	res, err := req.Execute(method, urlItem)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() == 401 {
		// refresh and retry
		return nil, fmt.Errorf("request for %s failed, code: %v", urlItem, res.StatusCode())
	}
	if res.StatusCode() > 300 {
		return nil, fmt.Errorf("request for %s failed, err: %v", urlItem, res.StatusCode())
	}
	return res.Body(), nil
}

type googleTokenRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshGoogleToken(grantType string, tokenType string, varMap map[string]interface{}) (string, error) {
	client := resty.New()
	client.SetProxy("http://127.0.0.1:7890")
	data := map[string]interface{}{
		"client_id":     loadParamFromVars("client_id", varMap),
		"client_secret": loadParamFromVars("client_secret", varMap),
		"redirect_uri":  loadParamFromVars("redirect_uri", varMap),
	}
	if grantType == "refresh_token" {
		data["grant_type"] = "refresh_token"
		data["refresh_token"] = loadParamFromVars("refresh_token", varMap)
	} else {
		data["grant_type"] = "authorization_code"
		data["code"] = loadParamFromVars("code", varMap)
	}
	urlItem := "https://www.googleapis.com/oauth2/v4/token"
	resp, err := client.R().
		SetBody(data).
		Post(urlItem)
	if err != nil {
		return "", fmt.Errorf("load account token failed, err: %v", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("load account token failed, code: %v", resp.StatusCode())
	}
	var respItem googleTokenRes
	if err := json.Unmarshal(resp.Body(), &respItem); err != nil {
		return "", err
	}
	fmt.Println(respItem)
	if tokenType == "accessToken" {
		return respItem.AccessToken, nil
	}
	return respItem.RefreshToken, nil
}
