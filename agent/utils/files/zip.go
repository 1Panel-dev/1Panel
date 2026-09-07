package files

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	cZip "github.com/klauspost/compress/zip"
	"github.com/spf13/afero"
)

type ZipArchiver struct {
}

func NewZipArchiver() ShellArchiver {
	return &ZipArchiver{}
}

func (z ZipArchiver) Extract(ctx context.Context, filePath, dstDir string, secret string) error {
	return z.ExtractWithOptions(ctx, filePath, dstDir, secret, false)
}

func (z ZipArchiver) ExtractWithOptions(ctx context.Context, filePath, dstDir, _ string, preserveOwner bool) error {
	if err := checkCmdAvailability("unzip"); err != nil {
		return err
	}
	args := []string{"-qo"}
	if preserveOwner {
		args = append(args, "-X")
	}
	args = append(args, filePath, "-d", dstDir)
	return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run("unzip", args...)
}

func (z ZipArchiver) Compress(ctx context.Context, sourcePaths []string, dstFile string, _ string) error {
	var err error
	tmpFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s%s.zip", common.RandStr(50), time.Now().Format(constant.DateTimeSlimLayout)))
	op := NewFileOp()
	defer func() {
		_ = op.DeleteFile(tmpFile)
		if err != nil {
			_ = op.DeleteFile(dstFile)
		}
	}()
	baseDir := path.Dir(sourcePaths[0])
	relativePaths := make([]string, len(sourcePaths))
	for i, sp := range sourcePaths {
		relativePaths[i] = path.Base(sp)
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(baseDir), cmd.WithContext(ctx))
	args := append([]string{"-qr", tmpFile}, relativePaths...)
	if err = cmdMgr.Run("zip", args...); err != nil {
		return err
	}
	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}

func normalizeZipEntry(header cZip.FileHeader) (cZip.FileHeader, error) {
	if header.NonUTF8 && header.Flags == 0 {
		name, err := decodeGBK(header.Name)
		if err != nil {
			return header, err
		}
		header.Name = name
	}
	name := strings.ReplaceAll(header.Name, `\`, "/")
	if name == "" || strings.ContainsRune(name, 0) || strings.HasPrefix(name, "/") {
		return header, fmt.Errorf("invalid ZIP path: %q", header.Name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || strings.Contains(part, ":") {
			return header, fmt.Errorf("invalid ZIP path: %q", header.Name)
		}
	}
	mode := header.Mode()
	if strings.HasSuffix(name, "/") && !mode.IsDir() {
		if mode&fs.ModeType != 0 || header.UncompressedSize64 != 0 {
			return header, fmt.Errorf("invalid ZIP directory: %q", header.Name)
		}
		header.SetMode(mode | fs.ModeDir | (mode.Perm()&0444)>>2)
	}
	header.Name = name
	isDir := header.Mode().IsDir()
	name = path.Clean(name)
	if name == "." && !isDir {
		return header, fmt.Errorf("invalid ZIP file path: %q", header.Name)
	}
	if isDir {
		name += "/"
	}
	header.Name = name
	return header, nil
}

func inspectZipPaths(ctx context.Context, input afero.File) (bool, error) {
	info, err := input.Stat()
	if err != nil {
		return false, err
	}
	reader, err := cZip.NewReader(input, info.Size())
	if err != nil {
		return false, err
	}
	compatible := false
	for _, entry := range reader.File {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		name := entry.Name
		if entry.NonUTF8 && entry.Flags == 0 {
			name, err = decodeGBK(name)
			if err != nil {
				return false, err
			}
		}
		compatible = compatible || strings.Contains(name, `\`)
	}
	if !compatible {
		return false, nil
	}
	type zipPathEntry struct {
		original string
		dir      bool
	}
	entries := make(map[string]zipPathEntry, len(reader.File))
	for _, entry := range reader.File {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		header, err := normalizeZipEntry(entry.FileHeader)
		if err != nil {
			return false, err
		}
		name := path.Clean(header.Name)
		if previous, exists := entries[name]; exists {
			return false, fmt.Errorf("conflicting ZIP paths: %q and %q", previous.original, entry.Name)
		}
		entries[name] = zipPathEntry{original: entry.Name, dir: header.Mode().IsDir()}
	}
	for name, entry := range entries {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if previous, exists := entries[parent]; exists && !previous.dir {
				return false, fmt.Errorf("conflicting ZIP paths: %q and %q", previous.original, entry.original)
			}
		}
	}
	return true, nil
}

func (f FileOp) decompressZipWithPathCompatibility(ctx context.Context, srcFile, dst string, options DecompressOptions) (bool, error) {
	input, err := f.Fs.Open(srcFile)
	if err != nil {
		return false, err
	}
	compatible, inspectErr := inspectZipPaths(ctx, input)
	if inspectErr == nil && compatible {
		_, inspectErr = f.extractArchiveWithSDK(ctx, input, dst, getFormat(Zip), options)
	}
	closeErr := input.Close()
	if inspectErr != nil {
		return false, inspectErr
	}
	return compatible, closeErr
}
