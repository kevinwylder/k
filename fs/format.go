package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"
)

const (
	HeaderIndicator 	  = "#>> Note opened"
	HeaderFormat          = HeaderIndicator + " at %s on %s\n"
	HeaderTimestampFormat = "3:04 pm"
)

func WriteHeader(dst io.Writer, t time.Time, host string) (int, error) {
	return fmt.Fprintf(dst, HeaderFormat, t.Format(HeaderTimestampFormat), host)
}

func ReadSegment(src io.Reader) (*bytes.Buffer, error) {
	buf := bufio.NewReader(src)
	peek := make([]byte, len(HeaderIndicator))
	_, err := buf.Read(peek)
	if err != nil {
		return nil, fmt.Errorf("peek: %w", err)
	}
	header := []byte(HeaderIndicator)
	if !bytes.Equal(header, peek) {
		return nil, fmt.Errorf("`%s` not header", string(peek))
	}
	dst := bytes.NewBuffer(peek)
	for {
		line, err := buf.ReadSlice('\n')
		if bytes.HasPrefix(line, header) {
			return dst, nil
		}
		if err != nil {
			if err == io.EOF {
				return dst, nil
			}
			return nil, err
		}
		_, err = dst.Write(line)
		if err != nil {
			return nil, fmt.Errorf("copy: %w", err)
		}
	}
}
