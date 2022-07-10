package fs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

type StorageDir struct {
	Host string
	path string
	tmp  string
}

func NewStorageDir(host, path, tmp string) (*StorageDir, error) {
	if _, err := os.Stat(path); err != nil {
		err = os.MkdirAll(path, 0o755)
		if err != nil {
			return nil, fmt.Errorf("make config dir: %w", err)
		}
	}
	return &StorageDir{
		Host: host,
		path: path,
		tmp:  tmp,
	}, nil
}

func (d *StorageDir) makePath(t time.Time) (year int, week int, day int, err error) {
	year, week = t.ISOWeek()
	dir := fmt.Sprintf("%s/%d/%d", d.path, year, week)
	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return
	}
	day = (int(t.Weekday()) + 6) % 7
	return
}

func (d *StorageDir) dayFile(t time.Time) (string, error) {
	y, w, wd, err := d.makePath(t)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/%d/%d/%d.k", d.path, y, w, wd)
	return path, nil
}

type SegmentFile struct {
	Time time.Time
	Host string
	Path string
}


func (d *StorageDir) NewSegmentFile(t time.Time) (*SegmentFile, error) {
	y, w, wd, err := d.makePath(t)
	if err != nil {
		return nil, err
	}

	// make tmp file
	dst, err := ioutil.TempFile(d.tmp, fmt.Sprintf("%d-%d-%d-*.k", y, w, wd))
	if err != nil {
		return nil, fmt.Errorf("make: %w", err)
	}

	// write header
	_, err = WriteHeader(dst, t, d.Host)
	if err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

	// copy the current day file to the temp file
	f, unlock, err := d.openExclusive(t)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	defer unlock()
	_, err = io.Copy(dst, f)
	if err != nil {
		return nil, fmt.Errorf("copy: %w", err)
	}

	return &SegmentFile{
		Time: t,
		Host: d.Host,
		Path: dst.Name(),
	}, nil
}

func (f *SegmentFile) Read() (*bytes.Buffer, error) {
	src, err := os.Open(f.Path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer src.Close()
	return ReadSegment(src)
}

func (d *StorageDir) Write(t time.Time, data io.Reader, overwrite bool) error {
	src, unlock, err := d.openExclusive(t)
	if err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	defer unlock()

	var buf bytes.Buffer
	if !overwrite {
		// read it to memory
		_, err = io.Copy(&buf, src)
		if err != nil {
			return fmt.Errorf("read src: %w", err)
		}
	}

	// truncate day file
	dst, err := os.Create(src.Name())
	if err != nil {
		return fmt.Errorf("trunc: %w", err)
	}
	defer dst.Close()

	// write new data
	_, err = io.Copy(dst, data)
	if err != nil {
		return fmt.Errorf("write new data: %w", err)
	}

	// write old data
	_, err = io.Copy(dst, &buf)
	if err != nil {
		return fmt.Errorf("write old data: %w", err)
	}
	return nil
}

func (d *StorageDir) openExclusive(t time.Time) (*os.File, func(), error) {
	srcPath, err := d.dayFile(t)
	if err != nil {
		return nil, nil, fmt.Errorf("get src: %w", err)
	}
	// make it if it doesn't exist
	if _, err := os.Stat(srcPath); err != nil {
		f, err := os.Create(srcPath)
		if err != nil {
			return nil,nil, fmt.Errorf("create: %w", err)
		}
		f.Close()
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	fd := int(f.Fd())
	err = unix.Flock(fd, unix.LOCK_EX)
	if err != nil {
		return nil, nil, err
	}
	return f, func() {
		f.Close()
		unix.Flock(fd, unix.LOCK_UN)
	}, nil
}
