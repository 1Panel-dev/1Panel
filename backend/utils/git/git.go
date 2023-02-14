package git

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/antihax/optional"
	"github.com/google/go-github/github"
	"net/http"
	"time"
)

type RepoInfo struct {
	RepoType     string
	Version      string
	ReleaseNote  string
	CreatedAt    string
	DownloadPath string
}

type RepoContent struct {
	RepoType     string
	Content      []byte
	DownloadPath string
}

var gitRepoTypes = []string{"gitee", "github"}

func CheckAndGetInfo(owner, repoName string) (*RepoInfo, error) {
	for _, repoType := range gitRepoTypes {
		url := fmt.Sprintf("https://%s.com/%s/%s", repoType, owner, repoName)
		if checkValid(url) {
			res, err := getLatestRepoInfo(repoType, owner, repoName)
			if err == nil {
				return res, nil
			} else {
				global.LOG.Errorf("get %s last release version  failed %s", repoType, err.Error())
			}
		} else {
			global.LOG.Errorf("get %s remote repo [%s] failed", repoType, url)
		}
	}
	return nil, errors.New("all remote repo get failed")
}

func CheckAndGetContent(owner, repoName, branch, path string) (*RepoContent, error) {
	for _, repoType := range gitRepoTypes {
		url := fmt.Sprintf("https://%s.com/%s/%s", repoType, owner, repoName)
		if checkValid(url) {
			res, err := getContentFromBranch(repoType, owner, repoName, branch, path)
			if err == nil {
				return res, nil
			} else {
				global.LOG.Errorf("get %s last version failed %s", repoType, err.Error())
			}
		} else {
			global.LOG.Errorf("check %s valid [%s] failed", repoType, url)
		}
	}
	return nil, errors.New("all repo  get failed")
}

func checkValid(addr string) bool {
	timeout := 2 * time.Second
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	if _, err := client.Get(addr); err != nil {
		return false
	}
	return true
}

func getLatestRepoInfo(repoType, owner, repoName string) (*RepoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var repoInfo RepoInfo
	repoInfo.RepoType = repoType
	if repoType == "gitee" {
		client := gitee.NewAPIClient(gitee.NewConfiguration())
		stats, res, err := client.RepositoriesApi.GetV5ReposOwnerRepoReleasesLatest(ctx, owner, repoName, &gitee.GetV5ReposOwnerRepoReleasesLatestOpts{})
		if res.StatusCode != 200 || err != nil {
			return nil, err
		}
		repoInfo.Version = stats.Name
		repoInfo.ReleaseNote = stats.Body
		repoInfo.CreatedAt = stats.CreatedAt.Format("2006-01-02 15:04:05")
		repoInfo.DownloadPath = fmt.Sprintf("https://gitee.com/%s/%s/releases/download/%s/", owner, repoName, repoInfo.Version)
	} else {
		client := github.NewClient(nil)
		stats, res, err := client.Repositories.GetLatestRelease(ctx, owner, repoName)
		if res.StatusCode != 200 || err != nil {
			return nil, err
		}
		repoInfo.Version = *stats.Name
		repoInfo.ReleaseNote = *stats.Body
		repoInfo.CreatedAt = stats.PublishedAt.Add(8 * time.Hour).Format("2006-01-02 15:04:05")
		repoInfo.DownloadPath = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/", owner, repoName, repoInfo.Version)
	}
	return &repoInfo, nil
}

func getContentFromBranch(repoType, owner, repoName, branch, path string) (*RepoContent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var repoContent RepoContent
	repoContent.RepoType = repoType
	if repoType == "gitee" {
		client := gitee.NewAPIClient(gitee.NewConfiguration())
		content, res, err := client.RepositoriesApi.GetV5ReposOwnerRepoContentsPath(ctx, owner, repoName, path, &gitee.GetV5ReposOwnerRepoContentsPathOpts{
			Ref: optional.NewString(branch),
		})
		if res.StatusCode != 200 || err != nil {
			return nil, err
		}
		bs64Bytes, err := base64.StdEncoding.DecodeString(content.Content)
		if err != nil {
			return nil, err
		}
		repoContent.Content = bs64Bytes
		repoContent.DownloadPath = fmt.Sprintf("https://gitee.com/%s/%s/releases/download/", owner, repoName)
		return &repoContent, nil
	} else {
		client := github.NewClient(nil)
		content, _, res, err := client.Repositories.GetContents(ctx, owner, repoName, path, &github.RepositoryContentGetOptions{
			Ref: branch,
		})
		if res.StatusCode != 200 || err != nil {
			return nil, err
		}
		bs64Bytes, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			return nil, err
		}
		repoContent.Content = bs64Bytes
		repoContent.DownloadPath = fmt.Sprintf("https://github.com/%s/%s/releases/download", owner, repoName)
		return &repoContent, nil
	}
}
