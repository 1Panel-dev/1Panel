package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

const decompressKeepDirFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW | unix.O_CLOEXEC

type decompressKeepIdentity struct{ dev, ino uint64 }
type decompressKeepFilesystem struct{ dev, mount uint64 }
type decompressKeepLink struct {
	stage *decompressKeepStage
	name  string
}
type decompressKeepStage struct {
	fd, parent int
	name       string
	identity   decompressKeepIdentity
	entries    map[string]bool
}
type decompressKeepState struct {
	stages      map[decompressKeepFilesystem]*decompressKeepStage
	localStages map[int]*decompressKeepStage
	hardlinks   map[decompressFileIdentity][]decompressKeepLink
}

func newDecompressKeepState() *decompressKeepState {
	return &decompressKeepState{
		stages:      make(map[decompressKeepFilesystem]*decompressKeepStage),
		localStages: make(map[int]*decompressKeepStage),
		hardlinks:   make(map[decompressFileIdentity][]decompressKeepLink),
	}
}

func copyDecompressTreeKeep(ctx context.Context, src, dst string) (retErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	parent, name, err := openDecompressKeepParent(ctx, dst)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	state := newDecompressKeepState()
	defer func() { retErr = errors.Join(retErr, state.close()) }()
	fd, created, err := state.directory(parent, name)
	if err != nil {
		return err
	}
	if fd < 0 {
		return fmt.Errorf("decompression destination is not a directory: %s", dst)
	}
	defer func() {
		// A canceled task retains its new directories. Restore their public
		// permissions now; a retry correctly treats them as existing entries.
		if created {
			retErr = errors.Join(retErr, unix.Fchmod(fd, 0755))
		}
		retErr = errors.Join(retErr, unix.Close(fd))
	}()
	return state.children(ctx, src, fd)
}

func copyDecompressEntryKeep(ctx context.Context, src, dst string) (retErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	parent, name, err := openDecompressKeepParent(ctx, dst)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	state := newDecompressKeepState()
	defer func() { retErr = errors.Join(retErr, state.close()) }()
	if err := state.entry(ctx, src, parent, name); err != nil {
		return err
	}
	return ctx.Err()
}

func openDecompressKeepParent(ctx context.Context, dst string) (int, string, error) {
	abs, err := filepath.Abs(dst)
	if err != nil {
		return -1, "", err
	}
	parent, name := filepath.Dir(abs), filepath.Base(abs)
	var missing []string
	for {
		if err := ctx.Err(); err != nil {
			return -1, "", err
		}
		resolved, resolveErr := filepath.EvalSymlinks(parent)
		if resolveErr == nil {
			parent = resolved
			break
		}
		if !os.IsNotExist(resolveErr) || parent == filepath.Dir(parent) {
			return -1, "", resolveErr
		}
		missing = append(missing, filepath.Base(parent))
		parent = filepath.Dir(parent)
	}
	for i := len(missing) - 1; i >= 0; i-- {
		parent = filepath.Join(parent, missing[i])
	}
	fd, err := unix.Open(string(filepath.Separator), decompressKeepDirFlags, 0)
	if err != nil {
		return -1, "", err
	}
	for _, component := range strings.Split(strings.Trim(parent, string(filepath.Separator)), string(filepath.Separator)) {
		if err := ctx.Err(); err != nil {
			_ = unix.Close(fd)
			return -1, "", err
		}
		if component == "" {
			continue
		}
		next, openErr := unix.Openat(fd, component, decompressKeepDirFlags, 0)
		if openErr == unix.ENOENT {
			if mkdirErr := unix.Mkdirat(fd, component, 0755); mkdirErr != nil && mkdirErr != unix.EEXIST {
				_ = unix.Close(fd)
				return -1, "", mkdirErr
			}
			next, openErr = unix.Openat(fd, component, decompressKeepDirFlags, 0)
		}
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, "", openErr
		}
		fd = next
	}
	return fd, name, nil
}

func decompressKeepRandomName() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return ".decompress-keep-" + hex.EncodeToString(random[:]), nil
}

func decompressKeepFDIdentity(fd int) (decompressKeepIdentity, error) {
	var info unix.Stat_t
	if err := unix.Fstat(fd, &info); err != nil {
		return decompressKeepIdentity{}, err
	}
	return decompressKeepIdentity{uint64(info.Dev), uint64(info.Ino)}, nil
}

func (s *decompressKeepState) staging(parent int) (*decompressKeepStage, error) {
	identity, err := decompressKeepFDIdentity(parent)
	if err != nil {
		return nil, err
	}
	mount, err := decompressKeepMountID(parent)
	if err != nil {
		return nil, err
	}
	filesystem := decompressKeepFilesystem{dev: identity.dev, mount: mount}
	if stage := s.stages[filesystem]; stage != nil {
		return stage, nil
	}
	stage, err := newDecompressKeepStage(parent)
	if err != nil {
		return nil, err
	}
	s.stages[filesystem] = stage
	return stage, nil
}

func (s *decompressKeepState) localStaging(parent int) (*decompressKeepStage, error) {
	if stage := s.localStages[parent]; stage != nil {
		return stage, nil
	}
	if _, err := s.staging(parent); err != nil {
		return nil, err
	}
	stage, err := newDecompressKeepStage(parent)
	if err != nil {
		return nil, err
	}
	s.localStages[parent] = stage
	return stage, nil
}

func newDecompressKeepStage(parent int) (*decompressKeepStage, error) {
	name, err := decompressKeepRandomName()
	if err != nil {
		return nil, err
	}
	if err := unix.Mkdirat(parent, name, 0700); err != nil {
		return nil, err
	}
	var created unix.Stat_t
	if err := unix.Fstatat(parent, name, &created, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return nil, err
	}
	if created.Mode&unix.S_IFMT != unix.S_IFDIR || created.Uid != uint32(os.Geteuid()) || created.Mode&0777 != 0700 {
		return nil, fmt.Errorf("decompression staging directory changed")
	}
	createdIdentity := decompressKeepIdentity{uint64(created.Dev), uint64(created.Ino)}
	registered := false
	defer func() {
		if !registered {
			removeDecompressKeepStageDirectory(parent, name, createdIdentity)
		}
	}()
	fd, err := unix.Openat(parent, name, decompressKeepDirFlags, 0)
	if err != nil {
		return nil, err
	}
	var stat unix.Stat_t
	if err = unix.Fstat(fd, &stat); err != nil || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0777 != 0700 ||
		(decompressKeepIdentity{uint64(stat.Dev), uint64(stat.Ino)}) != createdIdentity {
		_ = unix.Close(fd)
		if err == nil {
			err = fmt.Errorf("decompression staging directory changed")
		}
		return nil, err
	}
	parentCopy, err := unix.Dup(parent)
	if err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	unix.CloseOnExec(parentCopy)
	stage := &decompressKeepStage{fd: fd, parent: parentCopy, name: name,
		identity: decompressKeepIdentity{uint64(stat.Dev), uint64(stat.Ino)}, entries: make(map[string]bool)}
	registered = true
	return stage, nil
}

func removeDecompressKeepStageDirectory(parent int, name string, identity decompressKeepIdentity) error {
	var current unix.Stat_t
	if err := unix.Fstatat(parent, name, &current, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil
		}
		return err
	}
	if (decompressKeepIdentity{uint64(current.Dev), uint64(current.Ino)}) != identity {
		return nil
	}
	return unix.Unlinkat(parent, name, unix.AT_REMOVEDIR)
}

func (stage *decompressKeepStage) close() (retErr error) {
	for name, directory := range stage.entries {
		flags := 0
		if directory {
			flags = unix.AT_REMOVEDIR
		}
		if err := unix.Unlinkat(stage.fd, name, flags); err != unix.ENOENT {
			retErr = errors.Join(retErr, err)
		}
	}
	retErr = errors.Join(retErr, removeDecompressKeepStageDirectory(stage.parent, stage.name, stage.identity))
	return errors.Join(retErr, unix.Close(stage.fd), unix.Close(stage.parent))
}

func (s *decompressKeepState) closeLocal(parent int) error {
	if stage := s.localStages[parent]; stage != nil {
		delete(s.localStages, parent)
		return stage.close()
	}
	return nil
}

func (s *decompressKeepState) close() (retErr error) {
	for parent := range s.localStages {
		retErr = errors.Join(retErr, s.closeLocal(parent))
	}
	for _, stage := range s.stages {
		retErr = errors.Join(retErr, stage.close())
	}
	clear(s.stages)
	return retErr
}

func (s *decompressKeepState) directory(parent int, name string) (int, bool, error) {
	var existing unix.Stat_t
	err := unix.Fstatat(parent, name, &existing, unix.AT_SYMLINK_NOFOLLOW)
	if err == nil {
		if existing.Mode&unix.S_IFMT != unix.S_IFDIR {
			return -1, false, nil
		}
		fd, err := unix.Openat(parent, name, decompressKeepDirFlags, 0)
		if errors.Is(err, unix.ENOTDIR) || errors.Is(err, unix.ELOOP) {
			return -1, false, nil
		}
		return fd, false, err
	}
	if !errors.Is(err, unix.ENOENT) {
		return -1, false, err
	}
	stage, err := s.localStaging(parent)
	if err != nil {
		return -1, false, err
	}
	privateName, err := decompressKeepRandomName()
	if err != nil {
		return -1, false, err
	}
	if err := unix.Mkdirat(stage.fd, privateName, 0700); err != nil {
		return -1, false, err
	}
	stage.entries[privateName] = true
	fd, err := unix.Openat(stage.fd, privateName, decompressKeepDirFlags, 0)
	if err != nil {
		return -1, false, err
	}
	if err = renameDecompressKeepDirectory(stage.fd, privateName, parent, name); err != nil {
		_ = unix.Close(fd)
		if err == unix.EEXIST || err == unix.ENOTEMPTY {
			if err := unix.Fstatat(parent, name, &existing, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return -1, false, err
			}
			if existing.Mode&unix.S_IFMT != unix.S_IFDIR {
				return -1, false, nil
			}
			fd, err := unix.Openat(parent, name, decompressKeepDirFlags, 0)
			if err == unix.ENOTDIR || err == unix.ELOOP {
				return -1, false, nil
			}
			return fd, false, err
		}
		return -1, false, err
	}
	delete(stage.entries, privateName)
	return fd, true, nil
}

func (s *decompressKeepState) children(ctx context.Context, src string, dst int) (retErr error) {
	defer func() { retErr = errors.Join(retErr, s.closeLocal(dst)) }()
	if err := ctx.Err(); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := s.entry(ctx, filepath.Join(src, entry.Name()), dst, entry.Name()); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (s *decompressKeepState) entry(ctx context.Context, src string, parent int, name string) (retErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		fd, created, err := s.directory(parent, name)
		if err != nil || fd < 0 {
			return err
		}
		defer func() {
			if created {
				stamp := unix.NsecToTimespec(info.ModTime().UnixNano())
				retErr = errors.Join(retErr, decompressKeepFDMetadata(fd, info),
					unix.UtimesNanoAt(fd, ".", []unix.Timespec{stamp, stamp}, 0))
			}
			retErr = errors.Join(retErr, unix.Close(fd))
		}()
		return s.children(ctx, src, fd)
	}
	var existing unix.Stat_t
	if err := unix.Fstatat(parent, name, &existing, unix.AT_SYMLINK_NOFOLLOW); err == nil {
		return nil
	} else if !errors.Is(err, unix.ENOENT) {
		return err
	}
	if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("unsupported archive entry type: %s", src)
	}
	identity, hardlink := decompressHardlinkIdentity(info)
	if hardlink && info.Mode().IsRegular() {
		for _, link := range s.hardlinks[identity] {
			err := unix.Linkat(link.stage.fd, link.name, parent, name, 0)
			if err == nil || errors.Is(err, unix.EEXIST) {
				return nil
			}
			if !errors.Is(err, unix.EXDEV) {
				return err
			}
		}
	}
	stage, err := s.localStaging(parent)
	if err != nil {
		return err
	}
	privateName, err := decompressKeepRandomName()
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		if err := unix.Symlinkat(target, stage.fd, privateName); err != nil {
			return err
		}
		stage.entries[privateName] = false
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			if err := unix.Fchownat(stage.fd, privateName, int(stat.Uid), int(stat.Gid), unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
		}
	} else {
		fd, err := unix.Openat(stage.fd, privateName, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
		if err != nil {
			return err
		}
		stage.entries[privateName] = false
		if err := copyDecompressKeepFile(ctx, src, fd, info); err != nil {
			return err
		}
		stamp := unix.NsecToTimespec(info.ModTime().UnixNano())
		if err := unix.UtimesNanoAt(stage.fd, privateName, []unix.Timespec{stamp, stamp}, 0); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := unix.Linkat(stage.fd, privateName, parent, name, 0); err != nil {
		if errors.Is(err, unix.EEXIST) {
			_ = unix.Unlinkat(stage.fd, privateName, 0)
			delete(stage.entries, privateName)
			return nil
		}
		return err
	}
	if hardlink && info.Mode().IsRegular() {
		anchors, err := s.staging(parent)
		if err != nil {
			return err
		}
		anchorName, err := decompressKeepRandomName()
		if err != nil {
			return err
		}
		if err := unix.Linkat(stage.fd, privateName, anchors.fd, anchorName, 0); err != nil {
			return err
		}
		anchors.entries[anchorName] = false
		s.hardlinks[identity] = append(s.hardlinks[identity], decompressKeepLink{anchors, anchorName})
	}
	if err := unix.Unlinkat(stage.fd, privateName, 0); err != nil {
		return err
	}
	delete(stage.entries, privateName)
	return nil
}

func decompressKeepFDMetadata(fd int, info os.FileInfo) error {
	var ownerErr error
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		ownerErr = unix.Fchown(fd, int(stat.Uid), int(stat.Gid))
	}
	return errors.Join(ownerErr, unix.Fchmod(fd, uint32(info.Mode().Perm())))
}

func copyDecompressKeepFile(ctx context.Context, src string, fd int, info os.FileInfo) (retErr error) {
	dst := os.NewFile(uintptr(fd), "decompression staging file")
	defer func() {
		if err := dst.Close(); retErr == nil {
			retErr = err
		}
	}()
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()
	buffer := make([]byte, 64*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := input.Read(buffer)
		if err := ctx.Err(); err != nil {
			return err
		}
		if n > 0 {
			written, err := dst.Write(buffer[:n])
			if err != nil {
				return err
			}
			if written != n {
				return io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return decompressKeepFDMetadata(fd, info)
}
