package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kevinwylder/k/fs"
)

type Server struct {
	data *fs.StorageDir
	http.ServeMux
}

func NewServer(data *fs.StorageDir) *Server {
	s := &Server{
		data: data,
	}
	s.HandleFunc("/day", s.handleDay)
	return s
}

func (s *Server) handleDay(w http.ResponseWriter, r *http.Request) {
	timeStr := r.URL.Query().Get("t")
	unixTime, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "failed to parse time: %v", err)
		return
	}
	t := time.Unix(unixTime, 0)
	switch r.Method {
	case "GET":
		s.handleDayDownload(w, t)
	case "POST":
		s.handleDayAppend(r, t)
		s.handleDayDownload(w, t)
	default:
		w.WriteHeader(400)
		fmt.Fprintf(w, "unknown method: %s", r.Method)
	}
}

func (s *Server) handleDayDownload(w http.ResponseWriter, t time.Time) {
	log.Println("handle request for", t.Format("2006-01-02"))
	src, err := s.data.Read(t)
	if err != nil {
		log.Println("read", err)
		return
	}
	defer src.Close()
	io.Copy(w, src)
}

func (s *Server) handleDayAppend(r *http.Request, t time.Time) {
	err := s.data.Write(t, r.Body, false)
	if err != nil {
		log.Println("append", err)
		return
	}
}
