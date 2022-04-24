# gomodurl

> Like "gomodules", but _url_ the end.

Host vanity domain imports for Go modules with remote and dynamic configuration.

## Usage

Before running and configuring, here are some things worth noting:

- Sample configuration can be found at [gomodurl.json](/gomodurl.json), which is referenced in the examples below.
- The program caches HTTP responses, so it would be useful although not necessary to bind persistent volume to container's `/home/nonroot/.cache`.

### Docker

```bash
docker run --rm \
  -v $HOME/.cache:/home/nonroot/.cache \
  -e GOMODURL_CONFIG=https://raw.githubusercontent.com/wilsonehusin/gomodurl/main/gomodurl.json \
  -p 8000:8000 \
  --it ghcr.io/wilsonehusin/gomodurl:latest
```

### Deploying on Fly.io

This repository has continuous deployment to Fly, so [fly.toml](/fly.toml) is available for reference.

```bash
flyctl launch --image ghcr.io/wilsonehusin/gomodurl
```
