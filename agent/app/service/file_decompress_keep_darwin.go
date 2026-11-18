package service

import "golang.org/x/sys/unix"

func decompressKeepMountID(_ int) (uint64, error) {
	return 0, nil
}

func renameDecompressKeepDirectory(srcFD int, src string, dstFD int, dst string) error {
	return unix.RenameatxNp(srcFD, src, dstFD, dst, unix.RENAME_EXCL)
}
