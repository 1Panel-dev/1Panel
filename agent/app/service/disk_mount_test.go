package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/sirupsen/logrus"
)

const testDiskTotal = 100 << 20

func setupDiskTestLog() {
	if global.LOG == nil {
		global.LOG = logrus.New()
		global.LOG.SetOutput(io.Discard)
	}
}

type diskLogRecorder struct {
	mu       sync.Mutex
	messages []string
}

func (r *diskLogRecorder) Levels() []logrus.Level { return logrus.AllLevels }

func (r *diskLogRecorder) Fire(entry *logrus.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, entry.Message)
	return nil
}

func (r *diskLogRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.messages)
}

// recordDiskLogs swaps the global logger for one that records, until the test ends.
func recordDiskLogs(t *testing.T) *diskLogRecorder {
	t.Helper()
	recorder := &diskLogRecorder{}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.AddHook(recorder)
	oldLogger := global.LOG
	global.LOG = logger
	t.Cleanup(func() { global.LOG = oldLogger })
	return recorder
}

func usagePaths(usages []diskUsage) []string {
	var paths []string
	for _, item := range usages {
		paths = append(paths, item.Path)
	}
	return paths
}

func failurePaths(failures []diskFailure) []string {
	var paths []string
	for _, item := range failures {
		paths = append(paths, item.Path)
	}
	return paths
}

func TestParseDiskMountInfo(t *testing.T) {
	data := `22 1 8:1 / / rw,relatime shared:1 - ext4 /dev/sda1 rw
23 22 0:21 / /proc rw,nosuid shared:5 - proc proc rw
24 22 253:0 / /home rw,relatime shared:30 - xfs /dev/mapper/vg-home rw,attr2
25 22 0:45 / /mnt/nas\040share rw,relatime shared:31 - cifs //nas/My\040Share rw,vers=3.0
26 22 0:46 / /mnt/c rw,noatime - 9p C:\134 rw,aname=drvfs
27 22 0:47 / /mnt/photo rw - nfs4 nas:/volume1/photo rw
28 22 0:47 / /mnt/video rw - nfs4 nas:/volume1/video rw
29 22 8:1 /data/www /www rw shared:1 - ext4 /dev/sda1 rw
30 22 0:50 / /mnt/empty rw - tmpfs  rw

31 22 0:51 / /mnt/noopts rw - ext4 /dev/sdb1
`
	want := []diskMount{
		{ID: 22, ParentID: 1, Path: "/", Type: "ext4", Device: "/dev/sda1", Root: "/"},
		{ID: 23, ParentID: 22, Path: "/proc", Type: "proc", Device: "proc", Root: "/"},
		{ID: 24, ParentID: 22, Path: "/home", Type: "xfs", Device: "/dev/mapper/vg-home", Root: "/"},
		{ID: 25, ParentID: 22, Path: "/mnt/nas share", Type: "cifs", Device: "//nas/My Share", Root: "/"},
		{ID: 26, ParentID: 22, Path: "/mnt/c", Type: "9p", Device: `C:\`, Root: "/"},
		{ID: 27, ParentID: 22, Path: "/mnt/photo", Type: "nfs4", Device: "nas:/volume1/photo", Root: "/"},
		{ID: 28, ParentID: 22, Path: "/mnt/video", Type: "nfs4", Device: "nas:/volume1/video", Root: "/"},
		{ID: 29, ParentID: 22, Path: "/www", Type: "ext4", Device: "/dev/sda1", Root: "/data/www"},
		{ID: 30, ParentID: 22, Path: "/mnt/empty", Type: "tmpfs", Device: "", Root: "/"},
		{ID: 31, ParentID: 22, Path: "/mnt/noopts", Type: "ext4", Device: "/dev/sdb1", Root: "/"},
	}
	got, err := parseDiskMountInfo(data)
	if err != nil {
		t.Fatalf("parseDiskMountInfo() err = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDiskMountInfo() =\n%+v\nwant\n%+v", got, want)
	}
}

// A table we can't fully read must fail as a whole instead of losing mounts.
func TestParseDiskMountInfoRejectsUnknownFormat(t *testing.T) {
	good := "22 1 8:1 / / rw,relatime shared:1 - ext4 /dev/sda1 rw\n"
	cases := map[string]string{
		"invalid mount ID":  good + "x 22 0:21 / /proc rw - proc proc rw\n",
		"zero mount ID":     good + "0 22 0:21 / /proc rw - proc proc rw\n",
		"invalid parent ID": good + "23 x 0:21 / /proc rw - proc proc rw\n",
		"no separator":      good + "/dev/sdb1 /data ext4 rw 0 0\n",
		"too few fields":    good + "23 22 0:21 / - proc proc rw\n",
		"no source":         good + "24 22 0:22 / /sys rw - sysfs\n",
		"garbage":           good + "not a mountinfo line\n",
		"only a bad header": "Filesystem Type Size Used Avail Use% Mounted on\n",
	}
	for name, data := range cases {
		mounts, err := parseDiskMountInfo(data)
		if err == nil || mounts != nil {
			t.Errorf("%s: parseDiskMountInfo() = %+v, %v, want an error and no mounts", name, mounts, err)
		}
	}
	if mounts, err := parseDiskMountInfo(""); err != nil || len(mounts) != 0 {
		t.Errorf("empty table: parseDiskMountInfo() = %+v, %v, want no mounts and no error", mounts, err)
	}
}

func TestUnescapeMountField(t *testing.T) {
	cases := map[string]string{
		"/plain":         "/plain",
		`/a\040b`:        "/a b",
		`/tab\011nl\012`: "/tab\tnl\n",
		`C:\134`:         `C:\`,
		`/trailing\04`:   `/trailing\04`,
		`/not\9octal`:    `/not\9octal`,
		`/big\777`:       `/big\777`,
	}
	for in, want := range cases {
		if got := unescapeMountField(in); got != want {
			t.Errorf("unescapeMountField(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDiskMountRemote(t *testing.T) {
	cases := []struct {
		mount diskMount
		want  bool
	}{
		{diskMount{Type: "nfs4", Device: "nas:/export"}, true},
		{diskMount{Type: "fuse.sshfs", Device: "user@host:/"}, true},
		{diskMount{Type: "cifs", Device: "//nas/share"}, true},
		{diskMount{Type: "ocfs2", Device: "/dev/sdc1"}, true},
		{diskMount{Type: "ext4", Device: "/dev/sda1"}, false},
		{diskMount{Type: "ext4", Device: "/dev/disk/by-path/pci-0000:00:1f.2-ata-1"}, false},
		{diskMount{Type: "zfs", Device: "pool/data"}, false},
	}
	for _, c := range cases {
		if got := c.mount.remote(); got != c.want {
			t.Errorf("%+v remote() = %v, want %v", c.mount, got, c.want)
		}
	}
}

func TestFilterDiskMounts(t *testing.T) {
	mounts := []diskMount{
		{ID: 1, ParentID: 0, Path: "/", Type: "ext4", Device: "/dev/sda1"},
		{ID: 2, ParentID: 1, Path: "/run", Type: "tmpfs", Device: "tmpfs"},
		{ID: 3, ParentID: 1, Path: "/boot", Type: "ext4", Device: "/dev/sda2"},
		{ID: 4, ParentID: 1, Path: "/var/lib/docker/overlay2/abc/merged", Type: "overlay", Device: "overlay"},
		{ID: 5, ParentID: 1, Path: "/var/lib/docker", Type: "ext4", Device: "/dev/sdb1"},
		{ID: 6, ParentID: 1, Path: "/snap/core/123", Type: "squashfs", Device: "/dev/loop0"},
		{ID: 7, ParentID: 1, Path: "/run/user/1000/gvfs", Type: "fuse.gvfsd-fuse", Device: "gvfsd-fuse"},
		{ID: 8, ParentID: 1, Path: "/run/user/1000/doc", Type: "fuse.portal", Device: "portal"},
		{ID: 9, ParentID: 1, Path: "/run/user/1000/media", Type: "ext4", Device: "/dev/sdc1"},
		{ID: 10, ParentID: 1, Path: "/var/lib/lxcfs", Type: "fuse.lxcfs", Device: "lxcfs"},
		{ID: 11, ParentID: 1, Path: "/mnt/nas", Type: "nfs4", Device: "nas:/export"},
		{ID: 12, ParentID: 1, Path: "/mnt/ssh", Type: "fuse.sshfs", Device: "user@host:/"},
		{ID: 13, ParentID: 1, Path: "/data", Type: "ext4", Device: "/dev/sdd1"},
		{ID: 14, ParentID: 13, Path: "/data", Type: "nfs", Device: "nas:/data"},
		{ID: 15, ParentID: 1, Path: "/a/b/c/d/e/f/g/h/i/j", Type: "ext4", Device: "/dev/sde1"},
		{ID: 16, ParentID: 1, Path: "/merged", Type: "overlay", Device: "overlay"},
	}
	want := []diskMount{
		{ID: 1, ParentID: 0, Path: "/", Type: "ext4", Device: "/dev/sda1"},
		{ID: 11, ParentID: 1, Path: "/mnt/nas", Type: "nfs4", Device: "nas:/export"},
		{ID: 12, ParentID: 1, Path: "/mnt/ssh", Type: "fuse.sshfs", Device: "user@host:/"},
		{ID: 14, ParentID: 13, Path: "/data", Type: "nfs", Device: "nas:/data"},
	}
	for _, keepRootOverlay := range []bool{false, true} {
		if got := filterDiskMounts(mounts, keepRootOverlay); !reflect.DeepEqual(got, want) {
			t.Fatalf("filterDiskMounts(keepRootOverlay=%v) =\n%+v\nwant\n%+v", keepRootOverlay, got, want)
		}
	}
}

func TestFilterDiskMountsRootOverlay(t *testing.T) {
	mounts := []diskMount{
		{ID: 1, ParentID: 0, Path: "/", Type: "overlay", Device: "overlayroot"},
		{ID: 2, ParentID: 1, Path: "/merged", Type: "overlay", Device: "overlay"},
		{ID: 3, ParentID: 1, Path: "/data", Type: "ext4", Device: "/dev/sdb1"},
	}
	data := diskMount{ID: 3, ParentID: 1, Path: "/data", Type: "ext4", Device: "/dev/sdb1"}
	if got, want := filterDiskMounts(mounts, false), []diskMount{data}; !reflect.DeepEqual(got, want) {
		t.Fatalf("filterDiskMounts(keepRootOverlay=false) = %+v, want %+v", got, want)
	}
	want := []diskMount{{ID: 1, ParentID: 0, Path: "/", Type: "overlay", Device: "overlayroot"}, data}
	if got := filterDiskMounts(mounts, true); !reflect.DeepEqual(got, want) {
		t.Fatalf("filterDiskMounts(keepRootOverlay=true) = %+v, want %+v", got, want)
	}
}

func TestFilterDiskMountsSkipsPseudoFsMountedOver(t *testing.T) {
	mounts := []diskMount{
		{ID: 1, ParentID: 0, Path: "/data", Type: "ext4", Device: "/dev/sdb1"},
		{ID: 2, ParentID: 1, Path: "/data", Type: "tmpfs", Device: "tmpfs"},
	}
	if got := filterDiskMounts(mounts, false); len(got) != 0 {
		t.Fatalf("filterDiskMounts() = %+v, want none", got)
	}
}

// The parent relationships, not line order, determine which mounts are reachable.
func TestTopDiskMountsVisibility(t *testing.T) {
	cases := []struct {
		name  string
		table string
		want  []uint64
	}{
		{
			name: "stack hides old children but keeps new children",
			table: "1 1 8:1 / / rw - ext4 /dev/sda1 rw\n" +
				"20 1 8:2 / /data rw - ext4 /dev/sdb1 rw\n" +
				"21 20 8:3 / /data/old rw - ext4 /dev/sdc1 rw\n" +
				"22 21 8:4 / /data/old/deep rw - ext4 /dev/sdd1 rw\n" +
				"30 20 0:30 / /data rw - nfs4 nas:/data rw\n" +
				"31 30 8:5 / /data/live rw - ext4 /dev/sde1 rw\n" +
				"32 31 0:32 / /data/live rw - tmpfs tmpfs rw\n" +
				"33 32 8:6 / /data/live/nested rw - ext4 /dev/sdf1 rw\n",
			want: []uint64{1, 30, 32, 33},
		},
		{
			name: "later ancestor hides descendants attached to the underlying root",
			table: "1 1 8:1 / / rw - ext4 /dev/sda1 rw\n" +
				"10 1 8:2 / /data/old rw - ext4 /dev/sdb1 rw\n" +
				"11 10 8:3 / /data/old/deep rw - ext4 /dev/sdc1 rw\n" +
				"20 1 0:20 / /data rw - nfs4 nas:/data rw\n" +
				"21 1 8:4 / /database rw - ext4 /dev/sdd1 rw\n",
			want: []uint64{1, 20, 21},
		},
		{
			name: "same path on the visible parent replaces a hidden descendant",
			table: "1 1 8:1 / / rw - ext4 /dev/sda1 rw\n" +
				"10 1 8:2 / /data/child rw - ext4 /dev/sdb1 rw\n" +
				"20 1 0:20 / /data rw - nfs4 nas:/data rw\n" +
				"30 20 8:3 / /data/child rw - ext4 /dev/sdc1 rw\n",
			want: []uint64{1, 20, 30},
		},
		{
			name: "stacked namespace root",
			table: "1 1 8:1 / / rw - ext4 /dev/sda1 rw\n" +
				"2 1 8:2 / /old rw - ext4 /dev/sdb1 rw\n" +
				"3 1 0:3 / / rw - overlay overlay rw\n" +
				"4 3 8:3 / /new rw - ext4 /dev/sdc1 rw\n",
			want: []uint64{3, 4},
		},
		{
			name: "parent outside a chroot is absent from the table",
			table: "10 999 8:1 /jail / rw - ext4 /dev/sda1 rw\n" +
				"11 10 8:2 / /data rw - ext4 /dev/sdb1 rw\n",
			want: []uint64{10, 11},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mounts, err := parseDiskMountInfo(tc.table)
			if err != nil {
				t.Fatal(err)
			}
			// Rotating and reversing also place older hidden entries after their replacements.
			for order := 0; order < 2*len(mounts); order++ {
				if order == len(mounts) {
					for i, j := 0, len(mounts)-1; i < j; i, j = i+1, j-1 {
						mounts[i], mounts[j] = mounts[j], mounts[i]
					}
				}
				mounts = append(mounts[1:], mounts[0])
				var want []diskMount
				for _, mount := range mounts {
					for _, id := range tc.want {
						if mount.ID == id {
							want = append(want, mount)
						}
					}
				}
				if got := topDiskMounts(mounts); !reflect.DeepEqual(got, want) {
					t.Fatalf("order %d: visible mounts = %+v, want %+v", order, got, want)
				}
			}
		})
	}
}

func TestDiskStatGuardSameSourceRemount(t *testing.T) {
	before, err := parseDiskMountInfo("71 22 0:71 / /mnt/nas rw - fuse.rclone remote: rw\n")
	if err != nil {
		t.Fatal(err)
	}
	// The source, filesystem type, path and root are unchanged; only the kernel identity is new.
	after, err := parseDiskMountInfo("99 22 0:99 / /mnt/nas rw - fuse.rclone remote: rw\n")
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	defer close(release)
	var calls atomic.Int32
	g := newDiskStatGuard(50*time.Millisecond, func(path string) (diskStat, error) {
		if calls.Add(1) == 1 {
			<-release
		}
		return diskStat{Dev: 99, Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	if _, err := g.stat(before[0]); !errors.Is(err, errDiskStatTimeout) {
		t.Fatalf("old mount: %v, want timeout", err)
	}
	if stat, err := g.stat(after[0]); err != nil || stat.Dev != 99 {
		t.Fatalf("replacement mount: %+v, %v, want a fresh successful stat", stat, err)
	}
	if _, err := g.stat(before[0]); !errors.Is(err, errDiskStatTimeout) {
		t.Fatalf("old mount after replacement: %v, want timeout", err)
	}
	if n := calls.Load(); n != 2 {
		t.Fatalf("statFn called %d times, want one per mount instance", n)
	}
}

func TestDiskErrorType(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{errDiskStatTimeout, diskErrTimeout},
		{syscall.EACCES, diskErrDenied},
		{syscall.EPERM, diskErrDenied},
		{&fs.PathError{Op: "stat", Path: "/x", Err: syscall.EACCES}, diskErrDenied},
		{syscall.ENOENT, diskErrMissing},
		{syscall.ENOTDIR, diskErrMissing},
		{syscall.EIO, diskErrOther},
		{syscall.ENOTCONN, diskErrOther},
		{syscall.ESTALE, diskErrOther},
		{fmt.Errorf("statfs /x: %w", syscall.EHOSTDOWN), diskErrOther},
		{errors.New("stat /x: no device number on this platform"), diskErrOther},
	}
	for _, c := range cases {
		if got := diskErrorType(c.err); got != c.want {
			t.Errorf("diskErrorType(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}

func TestRetryDiskEINTR(t *testing.T) {
	calls := 0
	usage, err := retryDiskEINTR(func() (*disk.UsageStat, error) {
		calls++
		if calls < 3 {
			return nil, syscall.EINTR
		}
		return &disk.UsageStat{Total: 7}, nil
	})
	if err != nil || usage == nil || usage.Total != 7 || calls != 3 {
		t.Fatalf("retryDiskEINTR() = %+v, %v after %d calls, want the third result", usage, err, calls)
	}

	calls = 0
	if _, err := retryDiskEINTR(func() (*disk.UsageStat, error) {
		calls++
		return nil, syscall.ESTALE
	}); !errors.Is(err, syscall.ESTALE) || calls != 1 {
		t.Fatalf("retryDiskEINTR() err = %v after %d calls, want ESTALE after 1", err, calls)
	}
}

func TestPreferDiskMount(t *testing.T) {
	storage := diskMount{Path: "/storage", Device: "/dev/sda1", Root: "/"}
	bind := diskMount{Path: "/www", Device: "/dev/sda1", Root: "/data/www"}
	root := diskMount{Path: "/", Device: "/dev/sda1", Root: "/"}
	placeholder := diskMount{Path: "/a", Device: "rootfs", Root: "/"}
	named := diskMount{Path: "/longer", Device: "/dev/root", Root: "/"}
	cases := []struct {
		name       string
		mount, cur diskMount
		want       bool
	}{
		{"shorter path of the same root wins", root, storage, true},
		{"longer path of the same root loses", storage, root, false},
		{"bind of a subdirectory loses despite its shorter path", bind, storage, false},
		{"whole filesystem wins over a bind listed first", storage, bind, true},
		{"real device name wins over a placeholder", named, placeholder, true},
		{"placeholder loses despite its shorter path", placeholder, named, false},
	}
	for _, c := range cases {
		if got := preferDiskMount(c.mount, c.cur); got != c.want {
			t.Errorf("%s: preferDiskMount() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestDiskStatGuardStuckCall(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	g := newDiskStatGuard(50*time.Millisecond, func(path string) (diskStat, error) {
		if calls.Add(1) == 1 {
			<-release
		}
		return diskStat{Dev: 1, Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	nas := diskMount{Path: "/mnt/nas", Type: "nfs4", Device: "nas:/export"}

	start := time.Now()
	if _, err := g.stat(nas); !errors.Is(err, errDiskStatTimeout) {
		t.Fatalf("first stat err = %v, want timeout", err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("first stat returned after %s, want it to wait for the timeout", elapsed)
	}

	start = time.Now()
	for i := 0; i < 5; i++ {
		if _, err := g.stat(nas); !errors.Is(err, errDiskStatTimeout) {
			t.Fatalf("repeated stat err = %v, want timeout", err)
		}
	}
	// Far below the 5 x 50ms that waiting out the timeout again would take.
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("repeated stats on a known stuck mount took %s, want immediate failure", elapsed)
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("statFn called %d times while stuck, want 1", n)
	}

	close(release)
	deadline := time.Now().Add(time.Second)
	for {
		g.mu.Lock()
		_, inflight := g.calls[nas]
		g.mu.Unlock()
		if !inflight {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("stuck call was not forgotten after it returned")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := g.stat(nas); err != nil {
		t.Fatalf("stat after recovery err = %v, want nil", err)
	}
	if n := calls.Load(); n != 2 {
		t.Fatalf("statFn called %d times, want 2", n)
	}
}

// A call that outlived the timeout but has a result by the time a caller looks
// must hand out that result, not a timeout.
func TestDiskStatGuardPrefersResultOverTimeout(t *testing.T) {
	g := newDiskStatGuard(50*time.Millisecond, func(path string) (diskStat, error) {
		t.Error("statFn must not be called for a call already in the table")
		return diskStat{}, nil
	})
	mount := diskMount{Path: "/mnt/nas"}
	call := &diskStatCall{
		start: time.Now().Add(-time.Second),
		done:  make(chan struct{}),
		stat:  diskStat{Dev: 9, Usage: &disk.UsageStat{Total: testDiskTotal}},
	}
	close(call.done)
	g.calls[mount] = call
	for i := 0; i < 200; i++ {
		stat, err := g.stat(mount)
		if err != nil || stat.Dev != 9 {
			t.Fatalf("stat #%d = %+v, %v, want the finished call's result", i, stat, err)
		}
		time.Sleep(50 * time.Microsecond)
	}
}

func TestDiskStatGuardRemountedPath(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	var calls atomic.Int32
	g := newDiskStatGuard(50*time.Millisecond, func(path string) (diskStat, error) {
		if calls.Add(1) == 1 {
			<-release
		}
		return diskStat{Dev: 2, Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	oldMount := diskMount{Path: "/mnt/nas", Type: "nfs4", Device: "old:/export"}
	newMount := diskMount{Path: "/mnt/nas", Type: "nfs4", Device: "new:/export"}

	if _, err := g.stat(oldMount); !errors.Is(err, errDiskStatTimeout) {
		t.Fatalf("stat on the stuck mount err = %v, want timeout", err)
	}
	if _, err := g.stat(newMount); err != nil {
		t.Fatalf("stat on the new mount at the same path err = %v, want nil", err)
	}
	if _, err := g.stat(oldMount); !errors.Is(err, errDiskStatTimeout) {
		t.Fatalf("repeated stat on the stuck mount err = %v, want timeout", err)
	}
	if n := calls.Load(); n != 2 {
		t.Fatalf("statFn called %d times, want 2", n)
	}
}

func TestDiskStatGuardSharesInflightCall(t *testing.T) {
	var calls atomic.Int32
	g := newDiskStatGuard(time.Second, func(path string) (diskStat, error) {
		calls.Add(1)
		time.Sleep(30 * time.Millisecond)
		return diskStat{Dev: 1, Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := g.stat(diskMount{Path: "/"}); err != nil {
				t.Errorf("stat err = %v", err)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("statFn called %d times for concurrent callers, want 1", n)
	}
}

// Every mount that can't be read comes back as a failure, whatever the error.
func TestDiskStatGuardCollect(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	results := map[string]diskStat{
		"/www":  {Dev: 1, Usage: &disk.UsageStat{Total: testDiskTotal}},
		"/":     {Dev: 1, Usage: &disk.UsageStat{Total: testDiskTotal}},
		"/home": {Dev: 2, Usage: &disk.UsageStat{Total: 2 * testDiskTotal}},
		"/zero": {Dev: 3, Usage: &disk.UsageStat{Total: 0}},
		"/tiny": {Dev: 4, Usage: &disk.UsageStat{Total: 512 << 10}},
	}
	errs := map[string]error{
		"/mnt/nas":  syscall.ENOTCONN,
		"/mnt/priv": syscall.EACCES,
		"/gone":     syscall.ENOENT,
	}
	g := newDiskStatGuard(50*time.Millisecond, func(path string) (diskStat, error) {
		if path == "/mnt/stuck" {
			<-release
		}
		if err, ok := errs[path]; ok {
			return diskStat{}, err
		}
		return results[path], nil
	})

	var mounts []diskMount
	for _, path := range []string{"/www", "/", "/home", "/zero", "/tiny", "/mnt/nas", "/mnt/priv", "/gone", "/mnt/stuck"} {
		mounts = append(mounts, diskMount{Path: path})
	}
	usages, failures := g.collect(mounts)

	if got, want := usagePaths(usages), []string{"/", "/home"}; !reflect.DeepEqual(got, want) {
		t.Errorf("usage paths = %v, want %v", got, want)
	}
	if got, want := failurePaths(failures), []string{"/mnt/nas", "/mnt/priv", "/gone", "/mnt/stuck"}; !reflect.DeepEqual(got, want) {
		t.Errorf("failed paths = %v, want %v", got, want)
	}
	wantTypes := []string{diskErrOther, diskErrDenied, diskErrMissing, diskErrTimeout}
	for i, failure := range failures {
		if got := diskErrorType(failure.Err); i < len(wantTypes) && got != wantTypes[i] {
			t.Errorf("%s error type = %q, want %q", failure.Path, got, wantTypes[i])
		}
	}
}

func TestDiskStatGuardCollectDedup(t *testing.T) {
	mounts := []diskMount{
		// A bind mount of a subdirectory listed before the filesystem itself.
		{Path: "/www", Type: "ext4", Device: "/dev/sda1", Root: "/data/www"},
		{Path: "/storage", Type: "ext4", Device: "/dev/sda1", Root: "/"},
		// Two exports of one server filesystem, and the first one mounted twice.
		{Path: "/mnt/photo", Type: "nfs4", Device: "nas:/volume1/photo", Root: "/"},
		{Path: "/mnt/video", Type: "nfs4", Device: "nas:/volume1/video", Root: "/"},
		{Path: "/mnt/photo-again", Type: "nfs4", Device: "nas:/volume1/photo", Root: "/"},
		// Two shares that the server reports under one device number.
		{Path: "/mnt/docs", Type: "cifs", Device: "//nas/My Docs", Root: "/"},
		{Path: "/mnt/media", Type: "cifs", Device: "//nas/media", Root: "/"},
	}
	devs := map[string]uint64{
		"/www": 1, "/storage": 1,
		"/mnt/photo": 5, "/mnt/video": 5, "/mnt/photo-again": 5,
		"/mnt/docs": 6, "/mnt/media": 6,
	}
	g := newDiskStatGuard(time.Second, func(path string) (diskStat, error) {
		return diskStat{Dev: devs[path], Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	usages, failures := g.collect(mounts)

	want := []string{"/storage", "/mnt/photo", "/mnt/video", "/mnt/docs", "/mnt/media"}
	if got := usagePaths(usages); !reflect.DeepEqual(got, want) {
		t.Errorf("usage paths = %v, want %v", got, want)
	}
	if len(failures) != 0 {
		t.Errorf("failures = %+v, want none", failures)
	}
}

func TestDiskStatGuardLogsFailureOnce(t *testing.T) {
	recorder := recordDiskLogs(t)
	var broken atomic.Bool
	broken.Store(true)
	g := newDiskStatGuard(time.Second, func(path string) (diskStat, error) {
		if path == "/mnt/ssh" && broken.Load() {
			return diskStat{}, syscall.ENOTCONN
		}
		return diskStat{Dev: 1, Usage: &disk.UsageStat{Total: testDiskTotal}}, nil
	})
	mounts := []diskMount{
		{Path: "/", Type: "ext4", Device: "/dev/sda1"},
		{Path: "/mnt/ssh", Type: "fuse.sshfs", Device: "user@host:/"},
	}
	poll := func(times int, wantFailed []string) {
		t.Helper()
		for i := 0; i < times; i++ {
			_, failures := g.collect(mounts)
			g.logFailures(mounts, failures)
			if got := failurePaths(failures); !reflect.DeepEqual(got, wantFailed) {
				t.Fatalf("failed = %v, want %v", got, wantFailed)
			}
		}
	}

	poll(5, []string{"/mnt/ssh"})
	if n := recorder.count(); n != 1 {
		t.Fatalf("logged %d times while the mount kept failing, want 1: %v", n, recorder.messages)
	}
	broken.Store(false)
	poll(3, nil)
	if n := recorder.count(); n != 2 {
		t.Fatalf("logged %d times after recovery, want 2: %v", n, recorder.messages)
	}
	broken.Store(true)
	poll(2, []string{"/mnt/ssh"})
	if n := recorder.count(); n != 3 {
		t.Fatalf("logged %d times after failing again, want 3: %v", n, recorder.messages)
	}
}

func TestDiskStatGuardLogsTableErrorOnce(t *testing.T) {
	recorder := recordDiskLogs(t)
	g := newDiskStatGuard(time.Second, statDisk)
	tableErr := errors.New("read mount table: open /proc/self/mountinfo: no such file or directory")
	for i := 0; i < 5; i++ {
		g.logTableError(tableErr)
	}
	if n := recorder.count(); n != 1 {
		t.Fatalf("logged %d times while the table stayed unreadable, want 1", n)
	}
	g.logTableError(nil)
	g.logTableError(nil)
	if n := recorder.count(); n != 2 {
		t.Fatalf("logged %d times after the table became readable, want 2", n)
	}
}

func TestDropUnmounted(t *testing.T) {
	gone := diskMount{ID: 1, Path: "/mnt/gone", Type: "nfs4", Device: "nas:/gone"}
	shadowed := diskMount{ID: 2, Path: "/mnt/shadowed", Type: "ext4", Device: "/dev/sdb1"}
	stale := diskMount{ID: 3, Path: "/mnt/stale", Type: "nfs4", Device: "nas:/stale"}
	failures := []diskFailure{
		{diskMount: gone, Err: &fs.PathError{Op: "stat", Path: gone.Path, Err: syscall.ENOENT}},
		{diskMount: shadowed, Err: syscall.ENOTDIR},
		{diskMount: stale, Err: errDiskStatTimeout},
	}

	reads := 0
	table := func() ([]diskMount, error) {
		reads++
		return []diskMount{shadowed, stale}, nil
	}
	if got, want := failurePaths(dropUnmounted(failures, table)), []string{"/mnt/shadowed", "/mnt/stale"}; !reflect.DeepEqual(got, want) {
		t.Errorf("kept = %v, want %v: only the mount gone from the table may be dropped", got, want)
	}
	if reads != 1 {
		t.Errorf("mount table read %d times, want once", reads)
	}

	unreadable := func() ([]diskMount, error) { return nil, errors.New("read mount table: permission denied") }
	if got := dropUnmounted(failures, unreadable); len(got) != len(failures) {
		t.Errorf("kept %d of %d failures with an unreadable table, want all of them", len(got), len(failures))
	}

	reads = 0
	onlyStale := []diskFailure{{diskMount: stale, Err: errDiskStatTimeout}}
	if got := dropUnmounted(onlyStale, table); len(got) != 1 || reads != 0 {
		t.Errorf("kept %d failures after %d table reads, want 1 and no read", len(got), reads)
	}
}

func TestBuildDiskInfos(t *testing.T) {
	usages := []diskUsage{
		{diskMount: diskMount{Path: "/mnt/video", Type: "nfs4", Device: "nas:/video"}, Usage: &disk.UsageStat{Total: 300, Free: 100, Used: 200, UsedPercent: 66.6}},
		{diskMount: diskMount{Path: "/", Type: "ext4", Device: "/dev/sda1"}, Usage: &disk.UsageStat{Total: 100, Free: 40, Used: 60, UsedPercent: 60}},
	}
	failures := []diskFailure{
		{diskMount: diskMount{Path: "/mnt/nas", Type: "nfs4", Device: "nas:/backup"}, Err: errDiskStatTimeout},
		{diskMount: diskMount{Path: "/home/u/mnt", Type: "fuse.sshfs", Device: "u@h:/"}, Err: &fs.PathError{Op: "stat", Path: "/home/u/mnt", Err: syscall.EACCES}},
	}
	got := buildDiskInfos(usages, failures)
	want := []dto.DiskInfo{
		{Path: "/", Type: "ext4", Device: "/dev/sda1", Total: 100, Free: 40, Used: 60, UsedPercent: 60},
		{Path: "/home/u/mnt", Type: "fuse.sshfs", Device: "u@h:/", ErrorType: diskErrDenied, Error: "stat /home/u/mnt: permission denied"},
		{Path: "/mnt/nas", Type: "nfs4", Device: "nas:/backup", ErrorType: diskErrTimeout, Error: errDiskStatTimeout.Error()},
		{Path: "/mnt/video", Type: "nfs4", Device: "nas:/video", Total: 300, Free: 100, Used: 200, UsedPercent: 66.6},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildDiskInfos() =\n%+v\nwant\n%+v", got, want)
	}
	if empty := buildDiskInfos(nil, nil); empty == nil || len(empty) != 0 {
		t.Fatalf("buildDiskInfos(nil, nil) = %#v, want an empty non-nil list", empty)
	}
}

func TestDiskStatGuardCollectRealFS(t *testing.T) {
	setupDiskTestLog()
	g := newDiskStatGuard(diskStatTimeout, statDisk)
	missing := diskMount{Path: "/nonexistent-1panel-mount"}
	usages, failures := g.collect([]diskMount{{Path: "/"}, missing})
	if len(usages) != 1 || usages[0].Path != "/" || usages[0].Usage.Total == 0 {
		t.Fatalf("usages = %+v, want / with a non-zero size", usages)
	}
	if len(failures) != 1 || diskErrorType(failures[0].Err) != diskErrMissing {
		t.Fatalf("failures = %+v, want the missing path reported as missing", failures)
	}

	if _, err := os.Stat(diskMountInfoFile); err != nil {
		t.Skipf("no mount table to read on this platform: %v", err)
	}
	if kept := dropUnmounted(failures, readDiskMounts); len(kept) != 0 {
		t.Fatalf("kept = %+v, want a path absent from the mount table dropped", kept)
	}
	m, err := lookupDiskMount("/./")
	if err != nil || m.Path != "/" || m.Type == "" {
		t.Fatalf("lookupDiskMount(\"/./\") = %+v, %v, want the root mount entry", m, err)
	}
	infos, err := loadDiskInfoWith(false)
	if err != nil {
		t.Fatalf("loadDiskInfoWith() err = %v", err)
	}
	for _, info := range infos {
		if strings.TrimSpace(info.Path) == "" || (info.ErrorType == "") != (info.Error == "") {
			t.Fatalf("inconsistent entry %+v", info)
		}
	}
	t.Logf("disk infos on this host: %+v", infos)
}
