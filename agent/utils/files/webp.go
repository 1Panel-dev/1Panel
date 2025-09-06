package files

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

//type Webp struct {
//}

var (
	listenerMu sync.Mutex
	listeners  = make(map[uint]context.CancelFunc)
)

// 转换成webp
func ConvertToWebP(imagepath, webpdir, srcpath string, quality int) error {
	rel, _ := filepath.Rel(srcpath, imagepath)

	webppath := filepath.Join(webpdir, rel)

	_ = os.MkdirAll(filepath.Dir(webppath), 0777)
	cmd := exec.Command("cwebp", "-q", fmt.Sprintf("%d", quality), imagepath, "-o", webppath+".webp")
	_ = cmd.Run()

	return nil

}

func NewWebp(srcpath string, quality int, id uint, status int) error {

	Dir := []string{srcpath}
	ctx, ch := context.WithCancel(context.Background())
	webpdir := filepath.Join(srcpath, "../webpdir")

	if _, err := os.Stat(srcpath); err != nil {
		return fmt.Errorf("srcpath error: %w", err)
	}
	if status == 0 {
		stoplistening(id)
		return nil
	} else if status == 1 {
		if walk_err := filepath.Walk(srcpath, func(path string, info os.FileInfo, err error) error {
			if status == 1 {
				//开启备份和转换成webp
				if err = ConvertToWebP(path, webpdir, srcpath, quality); err != nil { //进行图片转换
					return err
				}
				if info.IsDir() {
					Dir = append(Dir, path)
				}
			}
			return nil
		}); walk_err != nil {
			return walk_err
		}
		if _, ok := listeners[id]; !ok {
			if err := ListeningDiry(Dir, srcpath, webpdir, ctx, quality); err != nil { //监听目录变化
				ch()
				return err
			}
			listenerMu.Lock()
			listeners[id] = ch //目录监听
			listenerMu.Unlock()
		}

	}

	return nil
}

func stoplistening(id uint) {
	// 对应网站目录监听停止
	listenerMu.Lock()
	defer listenerMu.Unlock()
	if cancel, ok := listeners[id]; ok {
		cancel()
		delete(listeners, id)
	}
}

// 同步删除备份
func remove(path string, srcpath, webpdir string) error {
	relPath, err := filepath.Rel(srcpath, path)
	if err != nil {
		return err
	}
	BakcupWebpDir := filepath.Join(webpdir, relPath)

	if err = os.Remove(BakcupWebpDir + ".webp"); err != nil {
		return err
	}
	return nil
}

// 监听文件文件变化
func ListeningDiry(srcdir []string, srcpath, webdir string, ctx context.Context, quality int) error {

	watcher, err := fsnotify.NewWatcher()

	go func(ctx context.Context) {

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {

					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write { //创建或者写入

					ext := strings.ToLower(filepath.Ext(event.Name))
					info, _ := os.Stat(event.Name)
					if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
						_ = ConvertToWebP(event.Name, webdir, srcpath, quality)
					} else if info.IsDir() {

						_ = watcher.Add(event.Name)
					}
				} else if event.Op&(fsnotify.Remove) == fsnotify.Remove {

					_ = remove(event.Name, srcpath, webdir)
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			case <-ctx.Done():

				watcher.Close()
				return

			}
		}

	}(ctx)
	for _, path := range srcdir {

		err1 := watcher.Add(path)
		if err != nil {
			return err1
		}
	}

	return nil
}
