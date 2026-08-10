#!/bin/sh
set -eu

expected_version="go version go1.25.12 "
actual_version="$(go version)"
case "$actual_version" in
  "$expected_version"*) ;;
  *)
    echo "expected go1.25.12 toolchain, got: $actual_version" >&2
    exit 1
    ;;
esac

unformatted="$(gofmt -l cmd internal)"
if [ -n "$unformatted" ]; then
  echo "gofmt drift:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go run ./cmd/foigent-integration-control schemas check --dir contracts/v1
go run ./cmd/foigent-integration-control authorization verify \
  --manifest fixtures/v1/valid/manifest.json \
  --unirouter-ack fixtures/v1/valid/unirouter-ack.json \
  --xinsight-ack fixtures/v1/valid/xinsight-ack.json \
  --authorization fixtures/v1/valid/authorization.json \
  --signature fixtures/v1/valid/authorization-signature.json \
  --trust-set fixtures/v1/valid/authorization-trust-set.json
git diff --check
git diff --cached --check

remote_policy="${N1_REMOTE_POLICY:-none}"
case "$remote_policy" in
  none)
    if [ -n "$(git remote)" ]; then
      echo "N1 local repository unexpectedly has a remote" >&2
      exit 1
    fi
    ;;
  github-origin)
    if [ "$(git remote)" != "origin" ]; then
      echo "durable-distribution CI requires exactly one remote named origin" >&2
      exit 1
    fi
    case "$(git remote get-url origin)" in
      https://github.com/foigent/integration-control|https://github.com/foigent/integration-control.git) ;;
      *)
        echo "unexpected durable-distribution origin URL" >&2
        exit 1
        ;;
    esac
    ;;
  *)
    echo "unknown N1_REMOTE_POLICY: $remote_policy" >&2
    exit 1
    ;;
esac
