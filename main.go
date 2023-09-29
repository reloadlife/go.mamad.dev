package main

import (
	"bytes"
	`errors`
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	
	log "github.com/sirupsen/logrus"
)

var tpl = template.Must(template.New("html").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Host}} {{.VCS}} {{.URL}}">
</head>
</html>
`))

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":80"
	}
	
	vcs := os.Getenv("VCS_TYPE")
	if vcs == "" {
		vcs = "git"
	}
	vcsURL := os.Getenv("VCS_URL")
	if vcsURL == "" {
		log.Warnf("VCS_URL env not specified (eg: https://github.com/username), using the default: https://github.com/reloadlife")
		vcsURL = "https://github.com/reloadlife"
	}
	
	u, err := url.Parse(vcsURL)
	if err != nil {
		log.Fatalf("invalid vcs url: %v", err)
	}
	
	if u.Scheme != "https" {
		log.Fatalf("vcs url scheme must be https")
	}
	
	muxHttpHandler := http.NewServeMux()
	muxHttpHandler.Handle("/ping", health())
	muxHttpHandler.Handle("/", redirectToPkg(vcs, u))
	
	log.Printf("Listening on: %s", addr)
	if cert, key := os.Getenv("CERT"),
		os.Getenv("CERT_KEY"); cert != "" && key != "" {
		err = http.ListenAndServeTLS(addr, cert, key, muxHttpHandler)
	} else {
		err = http.ListenAndServe(addr, muxHttpHandler)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen error: %+v", err)
	}
	
	log.Infof("server shutdown successfully")
}

func health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}
}

func redirectToPkg(vcs string, vcsURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		u, err := url.Parse(fmt.Sprintf("https://%s%s", vcsURL.Host, path.Join(vcsURL.Path, r.URL.Path)))
		if err != nil {
			http.Error(w, fmt.Sprintf("error building vcs url: %v", err), http.StatusInternalServerError)
			return
		}
		
		if r.URL.Query().Get("go-get") != "1" || len(r.URL.Path) < 2 {
			http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
			return
		}
		
		data := struct {
			Host string
			VCS  string
			URL  string
		}{
			path.Join(r.Host, r.URL.Path),
			vcs,
			u.String(),
		}
		
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, &data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("cache-Control", "no-store")
		_, _ = w.Write(buf.Bytes())
	}
}
