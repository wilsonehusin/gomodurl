package gomodurl

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Source interface {
	Repositories(context.Context, *http.Client) ([]*GoPackage, error)
}

func ParseSources(b []byte) ([]Source, error) {
	raw := map[string][]json.RawMessage{}

	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	sources := []Source{}
	for provider, specs := range raw {
		log.Printf("provider %s has %d packages", provider, len(specs))
		switch strings.ToLower(provider) {
		case "github":
			for _, spec := range specs {
				s, err := ParseGitHubSource(spec)
				if err != nil {
					log.Printf("warn: unable to parse github spec: %v", err.Error())
				} else {
					sources = append(sources, s)
				}
			}
		default:
			log.Printf("warn: unknown provider: %s", provider)
		}
	}

	log.Printf("info: found %d sources", len(sources))
	return sources, nil
}
