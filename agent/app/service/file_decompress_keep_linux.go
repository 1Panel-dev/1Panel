package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func decompressKeepMountID(fd int) (uint64, error) {
	var stat unix.Statx_t
	if err := unix.Statx(fd, "", unix.AT_EMPTY_PATH, unix.STATX_MNT_ID, &stat); err == nil && stat.Mask&unix.STATX_MNT_ID != 0 {
		return stat.Mnt_id, nil
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/self/fdinfo/%d", fd))
	if err != nil {
		return 0, fmt.Errorf("determine decompression destination mount: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if value, ok := strings.CutPrefix(line, "mnt_id:"); ok {
			return strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		}
	}
	return 0, fmt.Errorf("decompression destination mount identity is unavailable")
}

func renameDecompressKeepDirectory(srcFD int, src string, dstFD int, dst string) error {
	return unix.Renameat2(srcFD, src, dstFD, dst, unix.RENAME_NOREPLACE)
}
