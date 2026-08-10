package service

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
)

type fileTransferLockSet struct {
	mu    sync.Mutex
	paths map[string][]string
}

func newFileTransferLocks() *fileTransferLockSet {
	return &fileTransferLockSet{paths: make(map[string][]string)}
}

func (s *fileTransferLockSet) Acquire(taskID string, transferPaths []string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, activePaths := range s.paths {
		for _, activePath := range activePaths {
			for _, transferPath := range transferPaths {
				if fileTransferPathsOverlap(activePath, transferPath) {
					return false
				}
			}
		}
	}
	s.paths[taskID] = transferPaths
	return true
}

func (s *fileTransferLockSet) Release(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.paths, taskID)
}

func getFileTransferPaths(req request.FileMove) []string {
	paths := make([]string, 0, 1+len(req.OldPaths)+len(req.CoverPaths))
	paths = append(paths, req.NewPath)
	paths = append(paths, req.OldPaths...)
	paths = append(paths, req.CoverPaths...)

	unique := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, item := range paths {
		item = filepath.Clean(item)
		if _, ok := unique[item]; ok {
			continue
		}
		unique[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func fileTransferPathsOverlap(first, second string) bool {
	return first == second || strings.HasPrefix(first, second+string(filepath.Separator)) || strings.HasPrefix(second, first+string(filepath.Separator))
}

func aggregateFileMoveErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var errString strings.Builder
	for _, err := range errs {
		errString.WriteString(err.Error())
		errString.WriteByte('\n')
	}
	return errors.New(errString.String())
}
