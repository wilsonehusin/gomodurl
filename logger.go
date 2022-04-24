package gomodurl

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	logPrefix = " [ ~ g o m o d u r l  ] "
)

var (
	logger *log.Logger

	logFlags = log.Lshortfile
)

func init() {
	if os.Getenv("GOMODURL_DEV") != "" {
		logFlags |= log.LstdFlags
	}
	DisableLogger()
}

func sublogger(id string) *log.Logger {
	return log.New(logger.Writer(), fmt.Sprintf(" [%s] ", id), logFlags)
}

func EnableLogger(w io.Writer) {
	logger = log.New(w, logPrefix, logFlags)
}

func DisableLogger() {
	logger = log.New(io.Discard, logPrefix, logFlags)
}
