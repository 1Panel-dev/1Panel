package log

import (
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/1Panel-dev/1Panel/backend/global"
)

type Writer struct {
	m             Manager
	file          *os.File
	absPath       string
	fire          chan string
	cf            *Config
	rollingfilech chan string
}

type AsynchronousWriter struct {
	Writer
	ctx     chan int
	queue   chan []byte
	errChan chan error
	closed  int32
	wg      sync.WaitGroup
}

func (w *AsynchronousWriter) Close() error {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		close(w.ctx)
		w.onClose()

		func() {
			defer func() {
				if r := recover(); r != nil {
					global.LOG.Error(r)
				}
			}()
			w.m.Close()
		}()
		return w.file.Close()
	}
	return ErrClosed
}

func (w *AsynchronousWriter) onClose() {
	var err error
	for {
		select {
		case b := <-w.queue:
			if _, err = w.file.Write(b); err != nil {
				select {
				case w.errChan <- err:
				default:
					_asyncBufferPool.Put(&b)
					return
				}
			}
			_asyncBufferPool.Put(&b)
		default:
			return
		}
	}
}

var _asyncBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, BufferSize)
	},
}

func NewWriterFromConfig(c *Config) (RollingWriter, error) {
	if c.LogPath == "" || c.FileName == "" {
		return nil, ErrInvalidArgument
	}
	if err := os.MkdirAll(c.LogPath, 0700); err != nil {
		return nil, err
	}
	filepath := FilePath(c)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	if err := dupWrite(file); err != nil {
		return nil, err
	}
	mng, err := NewManager(c)
	if err != nil {
		return nil, err
	}

	var rollingWriter RollingWriter
	writer := Writer{
		m:       mng,
		file:    file,
		absPath: filepath,
		fire:    mng.Fire(),
		cf:      c,
	}
	if c.MaxRemain > 0 {
		writer.rollingfilech = make(chan string, c.MaxRemain)
		dir, err := os.ReadDir(c.LogPath)
		if err != nil {
			mng.Close()
			return nil, err
		}

		files := make([]string, 0, 10)
		for _, fi := range dir {
			if fi.IsDir() {
				continue
			}

			fileName := c.FileName
			if strings.Contains(fi.Name(), fileName) && strings.Contains(fi.Name(), c.LogSuffix) {
				start := strings.Index(fi.Name(), "-")
				end := strings.Index(fi.Name(), c.LogSuffix)
				name := fi.Name()
				if start > 0 && end > 0 {
					_, err := time.Parse(c.TimeTagFormat, name[start+1:end])
					if err == nil {
						files = append(files, fi.Name())
					}
				}
			}
		}
		sort.Slice(files, func(i, j int) bool {
			t1Start := strings.Index(files[i], "-")
			t1End := strings.Index(files[i], c.LogSuffix)
			t2Start := strings.Index(files[i], "-")
			t2End := strings.Index(files[i], c.LogSuffix)
			t1, _ := time.Parse(c.TimeTagFormat, files[i][t1Start+1:t1End])
			t2, _ := time.Parse(c.TimeTagFormat, files[j][t2Start+1:t2End])
			return t1.Before(t2)
		})

		for _, file := range files {
		retry:
			select {
			case writer.rollingfilech <- path.Join(c.LogPath, file):
			default:
				writer.DoRemove()
				goto retry
			}
		}
	}

	wr := &AsynchronousWriter{
		ctx:     make(chan int),
		queue:   make(chan []byte, QueueSize),
		errChan: make(chan error, QueueSize),
		wg:      sync.WaitGroup{},
		closed:  0,
		Writer:  writer,
	}

	wr.wg.Add(1)
	go wr.writer()
	wr.wg.Wait()
	rollingWriter = wr

	return rollingWriter, nil
}

func (w *AsynchronousWriter) writer() {
	var err error
	w.wg.Done()
	for {
		select {
		case filename := <-w.fire:
			if err = w.Reopen(filename); err != nil && len(w.errChan) < cap(w.errChan) {
				w.errChan <- err
			}
		case b := <-w.queue:
			if _, err = w.file.Write(b); err != nil && len(w.errChan) < cap(w.errChan) {
				w.errChan <- err
			}
			_asyncBufferPool.Put(&b)
		case <-w.ctx:
			return
		}
	}
}

func (w *Writer) DoRemove() {
	file := <-w.rollingfilech
	if err := os.Remove(file); err != nil {
		log.Println("error in remove log file", file, err)
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	var ok = false
	for !ok {
		select {
		case filename := <-w.fire:
			if err := w.Reopen(filename); err != nil {
				return 0, err
			}
		default:
			ok = true
		}
	}

	fp := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.file)))
	file := (*os.File)(fp)
	return file.Write(b)
}

func (w *Writer) Reopen(file string) error {
	fileInfo, err := w.file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	w.file.Close()
	if err := os.Rename(w.absPath, file); err != nil {
		return err
	}
	newFile, err := os.OpenFile(w.absPath, DefaultFileFlag, DefaultFileMode)
	if err != nil {
		return err
	}

	w.file = newFile

	go func() {
		if w.cf.MaxRemain > 0 {
		retry:
			select {
			case w.rollingfilech <- file:
			default:
				w.DoRemove()
				goto retry
			}
		}
	}()
	return nil
}
