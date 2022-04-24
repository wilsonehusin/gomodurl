package gomodurl

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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

func HTTPHandler(ctx context.Context, configPath string) (http.Handler, error) {
	packages, err := loadPackages(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}
	h := &Handler{
		packages:   packages,
		ctx:        ctx,
		configPath: configPath,
	}
	go h.daemon()
	return h, nil
}

type Handler struct {
	packages *GoPackageList

	ctx        context.Context
	configPath string
	lock       sync.RWMutex
}

func (h *Handler) daemon() {
	t := time.NewTicker(90 * time.Minute)

	for {
		select {
		case <-h.ctx.Done():
			logger.Printf("handler exit")
			return
		case <-t.C:
			h.reload()
		}

	}
}

func (h *Handler) reload() {
	h.lock.Lock()
	defer h.lock.Unlock()

	id := xid.New().String()
	p, err := loadPackages(h.ctx, h.configPath)
	if err != nil {
		logger.Printf("[%s] error: reloading config: %s", id, err.Error())
		return
	}

	h.packages = p
	logger.Printf("[%s] info: reloaded config", id)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	id := xid.New().String()
	reqLog := sublogger(id)
	host, _, _ := strings.Cut(req.Host, ":")
	path := strings.TrimPrefix(req.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")

	reqLog.Printf("[%s] info: looking up for '%s/%s'", id, host, path)
	pkg := h.packages.Lookup(host, path)
	if pkg == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Unknown host '%s' and path '%s' combination.\n[%s]\n", host, path, id)
		return
	}
	err := packageTmpl.Execute(w, pkg)
	if err != nil {
		reqLog.Printf("[%s] error rendering template: %v", id, err.Error())
	}
}

func loadPackages(ctx context.Context, configPath string) (*GoPackageList, error) {
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
			logger.Printf("error: retrieving repositories: %s", err.Error())
			continue
		}
		packages.Add(pkgs...)
	}

	return packages, nil
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
