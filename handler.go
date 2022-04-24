package gomodurl

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/xid"
)

var (
	//go:embed view/package.html.tmpl
	packageTmplRaw string
	packageTmpl    *template.Template
)

func init() {
	packageTmpl = template.Must(template.New("html").Parse(packageTmplRaw))
}

func Handler(ctx context.Context, configPath string) (http.Handler, error) {
	config, err := readLocalOrRemoteFile(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("retrieving config: %w", err)
	}

	sources, err := ParseSources(config)
	if err != nil {
		return nil, fmt.Errorf("parsing sources: %w", err)
	}

	packages := NewGoPackageList()
	for _, src := range sources {
		pkgs, err := src.Repositories(ctx, HTTPClient())
		if err != nil {
			log.Printf("error: retrieving repositories: %s", err.Error())
			continue
		}
		packages.Add(pkgs...)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		host, _, _ := strings.Cut(req.Host, ":")
		path := strings.TrimPrefix(req.URL.Path, "/")
		path = strings.TrimSuffix(path, "/")

		id := xid.New().String()
		log.Printf("[%s] info: looking up for '%s/%s'", id, host, path)
		pkg := packages.Lookup(host, path)
		if pkg == nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Unknown host '%s' and path '%s' combination.\n[%s]\n", host, path, id)
			return
		}
		err := packageTmpl.Execute(w, pkg)
		if err != nil {
			log.Printf("[%s] error rendering template: %v", id, err.Error())
		}
	}), nil
}

func readLocalOrRemoteFile(ctx context.Context, path string) ([]byte, error) {
	var configSrc io.ReadCloser
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("generating request to '%s': %w", path, err)
		}
		resp, err := HTTPClient().Do(req)
		if err != nil {
			return nil, fmt.Errorf("sending request to '%s': %w", path, err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d from '%s'", resp.StatusCode, path)
		}
		configSrc = resp.Body
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening config '%s': %w", path, err)
		}
		configSrc = f
	}

	var config bytes.Buffer
	_, err := io.Copy(&config, configSrc)
	//nolint:errcheck // don't really care if closing fails
	_ = configSrc.Close()
	if err != nil {
		return nil, fmt.Errorf("reading config '%s': %w", path, err)
	}

	return config.Bytes(), nil
}
