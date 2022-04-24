# gomodurl

> Like "gomodules", but _url_ the end.

Host vanity domain imports for Go modules.

## Usage

Sample configuration can be found at [gomodurl.json](/gomodurl.json)

```sh
docker run --rm \
  -v $HOME/.cache:/home/nonroot/.cache \ # Cache HTTP responses
  -e GOMODURL_CONFIG=https://raw.githubusercontent.com/wilsonehusin/gomodurl/main/gomodurl.json \ # or any local or remote path
  -p 8000:8000
  --it ghcr.io/wilsonehusin/gomodurl:latest
```
