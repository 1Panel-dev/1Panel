package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/psutil"
	"github.com/shirou/gopsutil/v4/disk"
)

// diskStatTimeout bounds how long stat/statfs may block on a single mount. A
// healthy filesystem answers in microseconds; anything slower is treated as a
// stale mount (unreachable NFS/CIFS server, stuck FUSE daemon, ...).
const diskStatTimeout = 3 * time.Second

// diskMountInfoFile is the mount table of our own mount namespace, the one
// stat resolves paths in. df reads the same file.
const diskMountInfoFile = "/proc/self/mountinfo"

// minDiskTotal hides filesystems too small to matter, pseudo ones reporting no
// blocks included. The df-based code skipped every size df -h printed in K.
const minDiskTotal = 1 << 20

// Reasons a mount is reported as failed, sent to the frontend as
// dto.DiskInfo.ErrorType.
const (
	diskErrTimeout = "timeout"
	diskErrDenied  = "denied"
	diskErrMissing = "missing"
	diskErrOther   = "error"
)

var errDiskStatTimeout = errors.New("stat timed out, the mount may be stale")

var diskMountExcludes = map[string]struct{}{
	"/mnt/cdrom": {}, "/boot": {}, "/boot/efi": {}, "/dev": {}, "/dev/shm": {},
	"/run/lock": {}, "/run": {}, "/run/shm": {}, "/run/user": {},
}

// skipDiskFsTypes are pseudo and per-user filesystems that never hold user data.
var skipDiskFsTypes = map[string]struct{}{
	"autofs": {}, "binfmt_misc": {}, "bpf": {}, "cgroup": {}, "cgroup2": {},
	"configfs": {}, "debugfs": {}, "devpts": {}, "devtmpfs": {}, "efivarfs": {},
	"fuse.gvfsd-fuse": {}, "fuse.lxcfs": {}, "fuse.portal": {}, "fusectl": {},
	"futexfs": {}, "hugetlbfs": {}, "inotifyfs": {}, "mqueue": {}, "nfsd": {},
	"nsfs": {}, "overlay": {}, "pipefs": {}, "proc": {}, "pstore": {}, "ramfs": {},
	"resctrl": {}, "rpc_pipefs": {}, "securityfs": {}, "selinuxfs": {}, "sockfs": {},
	"squashfs": {}, "sysfs": {}, "tmpfs": {}, "tracefs": {}, "usbfs": {},
}

// remoteDiskFsTypes and smbDiskFsTypes follow the filesystems df treats as remote.
var remoteDiskFsTypes = map[string]struct{}{
	"acfs": {}, "afs": {}, "auristorfs": {}, "coda": {}, "fhgfs": {}, "gpfs": {},
	"ibrix": {}, "ocfs2": {}, "vxfs": {},
}

var smbDiskFsTypes = map[string]struct{}{"cifs": {}, "smb3": {}, "smbfs": {}}

// diskMount is one entry of the mount table. It is also the key in-flight stat
// calls are tracked by, so it must stay comparable.
type diskMount struct {
	Path   string // mount point
	Type   string // filesystem type
	Device string // mount source, as df prints it
	Root   string // directory of the filesystem mounted here; "/" unless a bind mount or subvolume
}

// remote reports whether the mount is served over the network, roughly as df
// decides it. It only widens the dedup key, so a wrong guess shows one row more.
func (m diskMount) remote() bool {
	if strings.Contains(m.Device, ":") && !strings.HasPrefix(m.Device, "/") {
		return true
	}
	if _, ok := smbDiskFsTypes[m.Type]; ok && strings.HasPrefix(m.Device, "//") {
		return true
	}
	_, ok := remoteDiskFsTypes[m.Type]
	return ok
}

type diskStat struct {
	Dev   uint64
	Usage *disk.UsageStat
}

type diskUsage struct {
	diskMount
	Usage *disk.UsageStat
}

type diskFailure struct {
	diskMount
	Err error
}

// readDiskMounts lists mounts from the kernel mount table rather than running
// df, which statfs-es every mount and hangs as soon as one of them is stale.
// Reading the table is plain procfs parsing and touches no mount point.
func readDiskMounts() ([]diskMount, error) {
	data, err := os.ReadFile(diskMountInfoFile)
	if err != nil {
		return nil, fmt.Errorf("read mount table: %w", err)
	}
	return parseDiskMountInfo(string(data))
}

// parseDiskMountInfo parses proc_pid_mountinfo(5) lines such as
//
//	36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
//
// The source is kept as the kernel reports it, which is what df prints: device
// mapper names are not resolved to /dev/dm-N, and mounts sharing a device
// number keep their own source. A line that doesn't fit the format fails the
// whole table: a mount we can't read must not go missing unnoticed.
func parseDiskMountInfo(data string) ([]diskMount, error) {
	var mounts []diskMount
	for i, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		head, tail, ok := strings.Cut(line, " - ")
		fields := strings.Fields(head)
		// The source may be empty, so the tail is not split with Fields.
		source := strings.SplitN(tail, " ", 3)
		if !ok || len(fields) < 5 || len(source) < 2 {
			return nil, fmt.Errorf("unrecognized line %d in mount table %s: %q", i+1, diskMountInfoFile, line)
		}
		mounts = append(mounts, diskMount{
			Path:   unescapeMountField(fields[4]),
			Type:   source[0],
			Device: unescapeMountField(source[1]),
			Root:   unescapeMountField(fields[3]),
		})
	}
	return mounts, nil
}

// unescapeMountField decodes the octal escapes the kernel writes in mount
// table fields: \040 for a space, \011 tab, \012 newline, \134 backslash.
func unescapeMountField(field string) string {
	if !strings.Contains(field, `\`) {
		return field
	}
	var b strings.Builder
	for i := 0; i < len(field); i++ {
		if field[i] == '\\' && i+3 < len(field) {
			if c, err := strconv.ParseUint(field[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(c))
				i += 3
				continue
			}
		}
		b.WriteByte(field[i])
	}
	return b.String()
}

// topDiskMounts keeps one entry per mount point. When a path is mounted over,
// the last entry is the filesystem stat reaches.
func topDiskMounts(mounts []diskMount) []diskMount {
	var tops []diskMount
	index := make(map[string]int, len(mounts))
	for _, mount := range mounts {
		if i, ok := index[mount.Path]; ok {
			tops[i] = mount
			continue
		}
		index[mount.Path] = len(tops)
		tops = append(tops, mount)
	}
	return tops
}

// filterDiskMounts drops pseudo filesystems and paths that are not worth
// showing. keepRootOverlay keeps an overlay mounted at / (overlayroot, live
// systems), where it is the root disk rather than a container layer.
func filterDiskMounts(mounts []diskMount, keepRootOverlay bool) []diskMount {
	var kept []diskMount
	for _, mount := range topDiskMounts(mounts) {
		if isExcludedDiskMount(mount.Path) {
			continue
		}
		if _, skip := skipDiskFsTypes[mount.Type]; skip {
			if !keepRootOverlay || mount.Type != "overlay" || mount.Path != "/" {
				continue
			}
		}
		kept = append(kept, mount)
	}
	return kept
}

func isExcludedDiskMount(path string) bool {
	if strings.HasPrefix(path, "/snap") || strings.HasPrefix(path, "/run/user/") || len(strings.Split(path, "/")) > 10 {
		return true
	}
	if strings.Contains(path, "docker") || strings.Contains(path, "podman") ||
		strings.Contains(path, "containerd") || strings.HasPrefix(path, "/var/lib/containers") {
		return true
	}
	_, ok := diskMountExcludes[path]
	return ok
}

// lookupDiskMount returns the current mount entry for path, so that callers
// knowing only a path share in-flight stat calls with the ones listing mounts.
func lookupDiskMount(path string) (diskMount, error) {
	path = filepath.Clean(path)
	mounts, err := readDiskMounts()
	if err != nil {
		return diskMount{}, err
	}
	for _, mount := range topDiskMounts(mounts) {
		if mount.Path == path {
			return mount, nil
		}
	}
	return diskMount{Path: path}, nil
}

// diskErrorType names why a mount could not be read. Every error is reported;
// the type only picks the wording shown to the user.
func diskErrorType(err error) string {
	switch {
	case errors.Is(err, errDiskStatTimeout):
		return diskErrTimeout
	case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EPERM):
		return diskErrDenied
	case isDiskMissingErr(err):
		return diskErrMissing
	default:
		return diskErrOther
	}
}

func isDiskMissingErr(err error) bool {
	return errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.ENOTDIR)
}

// statDisk retries on EINTR: FUSE and network filesystems return it when a
// signal (e.g. SIGCHLD from a finished command) interrupts the call, and it
// says nothing about the mount. os.Stat retries by itself, statfs does not.
func statDisk(path string) (diskStat, error) {
	info, err := os.Stat(path)
	if err != nil {
		return diskStat{}, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return diskStat{}, fmt.Errorf("stat %s: no device number on this platform", path)
	}
	usage, err := retryDiskEINTR(func() (*disk.UsageStat, error) {
		return psutil.DISK.GetUsage(path, false)
	})
	if err != nil {
		return diskStat{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	return diskStat{Dev: uint64(st.Dev), Usage: usage}, nil
}

func retryDiskEINTR(usageFn func() (*disk.UsageStat, error)) (*disk.UsageStat, error) {
	for {
		usage, err := usageFn()
		if !errors.Is(err, syscall.EINTR) {
			return usage, err
		}
	}
}

// preferDiskMount reports whether mount should replace cur as the entry shown
// for the filesystem they share. The preferences are df's: a real device name
// over a placeholder, the whole filesystem over a bind mount of one of its
// subdirectories, then the mount point nearer the root. Unlike df the result
// does not depend on the order the mounts are listed in.
func preferDiskMount(mount, cur diskMount) bool {
	if named, curNamed := strings.Contains(mount.Device, "/"), strings.Contains(cur.Device, "/"); named != curNamed {
		return named
	}
	if len(mount.Root) != len(cur.Root) {
		return len(mount.Root) < len(cur.Root)
	}
	return len(mount.Path) < len(cur.Path)
}

// diskStatGuard runs stat calls in the background and waits at most timeout
// for them. A syscall stuck on a stale mount cannot be interrupted, so at most
// one call per mount is kept in flight: concurrent callers share it, and once it
// has outlived the timeout, later callers fail immediately instead of piling up
// more stuck threads. The call is forgotten when it finally returns, so a
// recovered mount is stat-ed normally on the next request. Calls are keyed by
// the whole mount entry rather than the path, so a different filesystem
// mounted at the same path isn't blamed for a call stuck on the old one.
type diskStatGuard struct {
	timeout time.Duration
	statFn  func(path string) (diskStat, error)

	mu    sync.Mutex
	calls map[diskMount]*diskStatCall
	// failures and tableErr hold what was last logged, so that a state polled
	// every few seconds is logged when it changes, not each time.
	failures map[diskMount]string
	tableErr string
}

type diskStatCall struct {
	start time.Time
	done  chan struct{}
	stat  diskStat
	err   error
}

var diskStats = newDiskStatGuard(diskStatTimeout, statDisk)

func newDiskStatGuard(timeout time.Duration, statFn func(path string) (diskStat, error)) *diskStatGuard {
	return &diskStatGuard{
		timeout:  timeout,
		statFn:   statFn,
		calls:    make(map[diskMount]*diskStatCall),
		failures: make(map[diskMount]string),
	}
}

func (g *diskStatGuard) stat(mount diskMount) (diskStat, error) {
	g.mu.Lock()
	call, ok := g.calls[mount]
	if !ok {
		call = &diskStatCall{start: time.Now(), done: make(chan struct{})}
		g.calls[mount] = call
		go g.run(mount, call)
	}
	g.mu.Unlock()

	timer := time.NewTimer(max(g.timeout-time.Since(call.start), 0))
	defer timer.Stop()
	select {
	case <-call.done:
		return call.stat, call.err
	case <-timer.C:
		// Both may be ready at once; a result that is there wins over the timeout.
		select {
		case <-call.done:
			return call.stat, call.err
		default:
			return diskStat{}, errDiskStatTimeout
		}
	}
}

func (g *diskStatGuard) run(mount diskMount, call *diskStatCall) {
	call.stat, call.err = g.statFn(mount.Path)
	g.mu.Lock()
	delete(g.calls, mount)
	g.mu.Unlock()
	close(call.done)
}

// diskDedupKey identifies the filesystem a mount shows. Remote mounts also
// carry their source: exports of one server filesystem share a device number
// yet are mounted on purpose, and df lists each of them.
type diskDedupKey struct {
	dev    uint64
	source string
}

// collect stats mounts concurrently. Every mount that hangs or fails is
// returned in failures. Filesystems below minDiskTotal are dropped, and
// several mounts of one filesystem (bind mounts) collapse to a single entry,
// as df does.
func (g *diskStatGuard) collect(mounts []diskMount) ([]diskUsage, []diskFailure) {
	stats := make([]diskStat, len(mounts))
	errs := make([]error, len(mounts))
	var wg sync.WaitGroup
	for i := range mounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats[i], errs[i] = g.stat(mounts[i])
		}()
	}
	wg.Wait()

	var (
		usages   []diskUsage
		failures []diskFailure
		seen     = make(map[diskDedupKey]int)
	)
	for i, mount := range mounts {
		if errs[i] != nil {
			failures = append(failures, diskFailure{diskMount: mount, Err: errs[i]})
			continue
		}
		if stats[i].Usage.Total < minDiskTotal {
			continue
		}
		key := diskDedupKey{dev: stats[i].Dev}
		if mount.remote() {
			key.source = mount.Device
		}
		item := diskUsage{diskMount: mount, Usage: stats[i].Usage}
		if idx, ok := seen[key]; ok {
			if preferDiskMount(mount, usages[idx].diskMount) {
				usages[idx] = item
			}
			continue
		}
		seen[key] = len(usages)
		usages = append(usages, item)
	}
	return usages, failures
}

// logFailures logs a mount when it starts failing, when its error changes and
// when it stops failing.
func (g *diskStatGuard) logFailures(mounts []diskMount, failures []diskFailure) {
	failing := make(map[diskMount]string, len(failures))
	for _, failure := range failures {
		failing[failure.diskMount] = failure.Err.Error()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, mount := range mounts {
		msg, isFailing := failing[mount]
		last, wasFailing := g.failures[mount]
		switch {
		case isFailing && msg != last:
			g.failures[mount] = msg
			global.LOG.Errorf("load disk info from %s failed, err: %s", mount.Path, msg)
		case !isFailing && wasFailing:
			delete(g.failures, mount)
			global.LOG.Infof("load disk info from %s no longer fails", mount.Path)
		}
	}
}

// logTableError logs a mount table error when it appears, changes or clears.
func (g *diskStatGuard) logTableError(err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if msg == g.tableErr {
		return
	}
	g.tableErr = msg
	if err != nil {
		global.LOG.Errorf("load disk info failed, err: %v", err)
		return
	}
	global.LOG.Info("the mount table is readable again")
}

// dropUnmounted leaves out the failures of mounts that were unmounted between
// listing and stat. It is the one failure not reported, and only once the mount
// table confirms the entry is gone; if the table can't be read they all stay.
func dropUnmounted(failures []diskFailure, readTable func() ([]diskMount, error)) []diskFailure {
	var (
		loaded  bool
		current map[diskMount]struct{}
	)
	mounted := func(mount diskMount) bool {
		if !loaded {
			loaded = true
			if table, err := readTable(); err == nil {
				current = make(map[diskMount]struct{}, len(table))
				for _, item := range topDiskMounts(table) {
					current[item] = struct{}{}
				}
			}
		}
		if current == nil {
			return true
		}
		_, ok := current[mount]
		return ok
	}

	var kept []diskFailure
	for _, failure := range failures {
		if isDiskMissingErr(failure.Err) && !mounted(failure.diskMount) {
			continue
		}
		kept = append(kept, failure)
	}
	return kept
}

// loadDiskInfoWith lists the mounts worth showing. It fails closed: an
// unreadable mount table is an error rather than an empty list, and a mount
// whose usage can't be loaded stays in the list with ErrorType and Error set
// and zero sizes, so every caller and the frontend can tell it from a healthy one.
func loadDiskInfoWith(keepRootOverlay bool) ([]dto.DiskInfo, error) {
	table, err := readDiskMounts()
	diskStats.logTableError(err)
	if err != nil {
		return nil, err
	}
	mounts := filterDiskMounts(table, keepRootOverlay)
	usages, failures := diskStats.collect(mounts)
	failures = dropUnmounted(failures, readDiskMounts)
	diskStats.logFailures(mounts, failures)
	return buildDiskInfos(usages, failures), nil
}

func buildDiskInfos(usages []diskUsage, failures []diskFailure) []dto.DiskInfo {
	datas := make([]dto.DiskInfo, 0, len(usages)+len(failures))
	for _, item := range usages {
		datas = append(datas, dto.DiskInfo{
			Path:              item.Path,
			Type:              item.Type,
			Device:            item.Device,
			Total:             item.Usage.Total,
			Free:              item.Usage.Free,
			Used:              item.Usage.Used,
			UsedPercent:       item.Usage.UsedPercent,
			InodesTotal:       item.Usage.InodesTotal,
			InodesUsed:        item.Usage.InodesUsed,
			InodesFree:        item.Usage.InodesFree,
			InodesUsedPercent: item.Usage.InodesUsedPercent,
		})
	}
	for _, item := range failures {
		datas = append(datas, dto.DiskInfo{
			Path:      item.Path,
			Type:      item.Type,
			Device:    item.Device,
			ErrorType: diskErrorType(item.Err),
			Error:     item.Err.Error(),
		})
	}
	sort.SliceStable(datas, func(i, j int) bool {
		return datas[i].Path < datas[j].Path
	})
	return datas
}
