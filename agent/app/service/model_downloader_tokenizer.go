package service

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

const tokenizerDownloadMaxSize = 256 << 20

type tokenizerDownloadMetadata struct {
	Source   string `json:"source"`
	RepoID   string `json:"repoID"`
	Revision string `json:"revision"`
}

func filterTokenizerDownloadFiles(items []dto.ModelFile) []dto.ModelFile {
	result := make([]dto.ModelFile, 0)
	for _, item := range items {
		name := item.Name
		if !filepath.IsLocal(name) || strings.ContainsAny(name, "\\\x00\r\n") || filepath.ToSlash(filepath.Clean(name)) != name {
			continue
		}
		if strings.Contains(name, "/") {
			if strings.HasPrefix(name, "chat_templates/") && strings.HasSuffix(name, ".jinja") {
				result = append(result, item)
			}
			continue
		}
		switch name {
		case "config.json", "tokenizer_config.json", "tokenizer.json", "special_tokens_map.json", "added_tokens.json", "vocab.json", "vocab.txt", "merges.txt", "tekken.json":
			result = append(result, item)
		default:
			if strings.HasSuffix(name, ".model") || strings.HasPrefix(name, "tokenizer.model") || strings.HasSuffix(name, ".tiktoken") || strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".jinja") {
				result = append(result, item)
			}
		}
	}
	return result
}

func validateTokenizerDownloadFiles(items []dto.ModelFile) error {
	if len(items) == 0 || len(items) > 512 {
		return newModelDownloaderInputError("tokenizer must contain between 1 and 512 files")
	}
	names := make(map[string]bool, len(items))
	var size int64
	vocabulary := false
	for _, item := range items {
		if names[item.Name] || item.Size < 0 || item.Size > tokenizerDownloadMaxSize-size {
			return newModelDownloaderInputError("invalid tokenizer files or total size exceeds 256 MiB")
		}
		size += item.Size
		names[item.Name] = true
		if item.Name == "tokenizer.json" || item.Name == "tekken.json" || item.Name == "vocab.txt" || strings.HasSuffix(item.Name, ".model") || strings.HasPrefix(item.Name, "tokenizer.model") || strings.HasSuffix(item.Name, ".tiktoken") {
			vocabulary = true
		}
	}
	if !vocabulary && !(names["vocab.json"] && names["merges.txt"]) {
		return newModelDownloaderInputError("repository has no supported tokenizer vocabulary in its root directory")
	}
	return nil
}

func validateDownloadedTokenizer(root string) error {
	var items []dto.ModelFile
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		stat, err := entry.Info()
		if err != nil {
			return err
		}
		if !stat.Mode().IsRegular() || stat.Size() > tokenizerDownloadMaxSize {
			return newModelDownloaderInputError("invalid tokenizer file")
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		items = append(items, dto.ModelFile{Name: filepath.ToSlash(name), Size: stat.Size()})
		if strings.HasSuffix(name, ".json") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !json.Valid(content) {
				return newModelDownloaderInputError("invalid tokenizer JSON: %s", name)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return validateTokenizerDownloadFiles(items)
}

func listDownloadedTokenizers(info string) ([]dto.LocalModelItem, error) {
	items := make([]dto.LocalModelItem, 0)
	root := filepath.Join(global.CONF.Base.InstallDir, "1panel/ai/tokenizer")
	directories, err := listModelDownloadDirectories(root)
	if err != nil {
		return nil, err
	}
	for _, item := range directories {
		if _, err := normalizeModelDownloaderRepoID(item.Name); err != nil {
			continue
		}
		path, err := filepath.EvalSymlinks(item.Path)
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(filepath.Join(path, ".1panel-tokenizer"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var metadata tokenizerDownloadMetadata
		if err := json.Unmarshal(content, &metadata); err != nil {
			return nil, err
		}
		if info != "" && !strings.Contains(strings.ToLower(metadata.RepoID), strings.ToLower(info)) {
			continue
		}
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		size, err := files.NewFileOp().GetDirSize(path)
		if err != nil {
			return nil, err
		}
		items = append(items, dto.LocalModelItem{Name: item.Name, Path: path, RepoID: metadata.RepoID, Source: metadata.Source, Revision: metadata.Revision, DownloadType: "tokenizer", Size: size, SizeFormatted: formatBytes(size), CreatedAt: stat.ModTime().Format("2006-01-02 15:04:05")})
	}
	return items, nil
}
