package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	hostPort := flag.String("hostport", ":2222", "host:port to listen on")
	flag.Parse()

	s := &http.Server{
		Addr:           *hostPort,
		Handler:        http.HandlerFunc(handler),
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "multipart parse: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()
	tempDir, err := ioutil.TempDir("", "rst2pdf-web-")
	if err != nil {
		http.Error(w, "create tempdir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	for _, fileHeaders := range r.MultipartForm.File {
		for _, hdr := range fileHeaders {
			fn := filepath.Join(tempDir, relativize(hdr.Filename))
			src, err := hdr.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("opening %s: %v", hdr, err), http.StatusBadRequest)
				return
			}
			if err = saveTo(fn, src); err != nil {
				http.Error(w, fmt.Sprintf("writing out %s: %v", fn, err), http.StatusInternalServerError)
				return
			}
			log.Printf("saved %s", fn)
		}
	}
	args := r.Form["arg"]

	out, err := ioutil.TempFile(tempDir, "out-")
	if err != nil {
		http.Error(w, "creating out tempfile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	c := exec.Command("rst2pdf", args...)
	c.Dir = tempDir
	c.Stdout = out
	var errBuf bytes.Buffer
	c.Stderr = &errBuf

	if err = c.Run(); err != nil {
		http.Error(w, fmt.Sprintf("running %s: %v\n%s", c, err, errBuf.String()),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	if err = out.Close(); err != nil {
		http.Error(w, "closing "+out.Name()+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, out.Name())
}

func saveTo(fn string, r io.ReadCloser) error {
	defer r.Close()
	fh, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = io.Copy(fh, r)
	return err
}

func relativize(fn string) string {
	for i, r := range fn {
		if r != '.' && r != '/' {
			return fn[i:]
		}
	}
	return fn
}
