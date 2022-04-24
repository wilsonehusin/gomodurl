package gomodurl

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/bindl-dev/httpcache"
	"github.com/bindl-dev/httpcache/diskcache"
)

var (
	httpClient *http.Client

	httpClientOnce sync.Once
)

func initHTTPClient() {
	baseDir, err := os.UserCacheDir()
	if err != nil {
		logger.Printf("warn: finding cache directory: %v", err.Error())
		baseDir = os.TempDir()
	}
	dir := filepath.Join(baseDir, "gomodurl")

	httpClient = httpcache.NewTransport(diskcache.New(dir)).Client()
	logger.Printf("info: caching http responses in '%s'", dir)
}

func HTTPClient() *http.Client {
	httpClientOnce.Do(initHTTPClient)
	return httpClient
}
