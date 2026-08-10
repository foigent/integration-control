#!/bin/sh

# Exact immutable N1 candidate coordinates. This file is sourced by the local
# distribution scripts; it performs no network, product, signing or runtime
# action.
N1_VERSION="v0.1.0-candidate.1"
N1_SOURCE_COMMIT="12f1ede8ae85e9d311a8150694eea54ff6f02950"
N1_SOURCE_TREE="e6945fa071afd6a0b9317d3ec862cbcbf47a0154"
N1_RELEASE_COMMIT="c9e5a77dd0416654d8e46efb09f18a6ecca026f8"
N1_RELEASE_TREE="325863188f625c07e9f10a7a3eae67f9574de198"
N1_TAG="v0.1.0-candidate.1"
N1_ARCHIVE_ROOT="foigent-integration-control-v0.1.0-candidate.1"

N1_DARWIN_ASSET="foigent-integration-control-v0.1.0-candidate.1-darwin-arm64"
N1_DARWIN_SIZE="4034066"
N1_DARWIN_SHA256="ba574ac25abd063914d785a34d097b14a1cf09021a953c40803414900c1f3dc1"
N1_LINUX_ASSET="foigent-integration-control-v0.1.0-candidate.1-linux-amd64"
N1_LINUX_SIZE="4132939"
N1_LINUX_SHA256="b37a9b92ef237844327911b9a4b1a1266b2970f3acbc50de428aa7d518a1fbb8"
N1_SOURCE_ASSET="foigent-integration-control-v0.1.0-candidate.1-source.tar"
N1_SOURCE_SIZE="235520"
N1_SOURCE_SHA256="360cbc70cbbe84361c23e9f89f3aeed2689b235197b2ebcb7b436187c9aedfef"
N1_MANIFEST_ASSET="foigent.integration-control.release.v1.json"
N1_MANIFEST_SIZE="2070"
N1_MANIFEST_SHA256="0604100b8cb8c8a728fc091d1c9269eafbc3cb70449222ab6c10657bca33d3e5"
N1_SUMS_ASSET="SHA256SUMS"
N1_SUMS_SIZE="528"
N1_SUMS_SHA256="d1536bf3f68bf35293470491c57cfc78f784c61f4d6291a3b9b9ba1b3f34e848"

n1_fail() {
  echo "N1 distribution error: $*" >&2
  exit 1
}

n1_sha256_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  n1_fail "neither shasum nor sha256sum is available"
}

n1_size_file() {
  wc -c <"$1" | tr -d '[:space:]'
}

n1_require_absolute_path() {
  case "$1" in
    /*) ;;
    *) n1_fail "an absolute path is required: $1" ;;
  esac
}

n1_create_new_directory() {
  n1_new_directory="$1"
  n1_require_absolute_path "$n1_new_directory"
  n1_parent_directory="$(dirname "$n1_new_directory")"
  [ -d "$n1_parent_directory" ] && [ ! -L "$n1_parent_directory" ] || n1_fail "parent must be an existing non-symlink directory: $n1_parent_directory"
  mkdir "$n1_new_directory" || n1_fail "refused to reuse work directory: $n1_new_directory"
}

n1_resolve_go_command() {
  n1_base_go="${N1_GO_COMMAND:-$(command -v go || true)}"
  [ -n "$n1_base_go" ] && [ -x "$n1_base_go" ] || n1_fail "a local Go executable is required"
  n1_host_os="$(GOENV=off GOTOOLCHAIN=local "$n1_base_go" env GOOS)"
  n1_host_arch="$(GOENV=off GOTOOLCHAIN=local "$n1_base_go" env GOARCH)"
  if [ "$(GOENV=off GOTOOLCHAIN=local "$n1_base_go" version 2>/dev/null || true)" = "go version go1.25.12 $n1_host_os/$n1_host_arch" ]; then
    printf '%s\n' "$n1_base_go"
    return
  fi
  n1_mod_cache="$(GOENV=off GOTOOLCHAIN=local "$n1_base_go" env GOMODCACHE)"
  n1_toolchain_go="$n1_mod_cache/golang.org/toolchain@v0.0.1-go1.25.12.$n1_host_os-$n1_host_arch/bin/go"
  [ -x "$n1_toolchain_go" ] || n1_fail "go1.25.12 is not already installed locally; automatic download is disabled"
  [ "$(GOENV=off GOTOOLCHAIN=local "$n1_toolchain_go" version)" = "go version go1.25.12 $n1_host_os/$n1_host_arch" ] || n1_fail "local go1.25.12 toolchain identity drift"
  printf '%s\n' "$n1_toolchain_go"
}

n1_assert_file() {
  n1_path="$1"
  n1_expected_size="$2"
  n1_expected_sha="$3"
  [ -f "$n1_path" ] || n1_fail "missing regular file: $n1_path"
  [ ! -L "$n1_path" ] || n1_fail "symbolic links are not accepted: $n1_path"
  n1_actual_size="$(n1_size_file "$n1_path")"
  [ "$n1_actual_size" = "$n1_expected_size" ] || n1_fail "size mismatch for $n1_path: got $n1_actual_size, expected $n1_expected_size"
  n1_actual_sha="$(n1_sha256_file "$n1_path")"
  [ "$n1_actual_sha" = "$n1_expected_sha" ] || n1_fail "SHA-256 mismatch for $n1_path: got $n1_actual_sha, expected $n1_expected_sha"
}

n1_assert_flat_assets() {
  n1_asset_dir="$1"
  [ -d "$n1_asset_dir" ] || n1_fail "asset directory does not exist: $n1_asset_dir"
  [ ! -L "$n1_asset_dir" ] || n1_fail "asset directory must not be a symbolic link: $n1_asset_dir"
  n1_entry_count="$(find "$n1_asset_dir" ! -path "$n1_asset_dir" -prune -print | wc -l | tr -d '[:space:]')"
  [ "$n1_entry_count" = "5" ] || n1_fail "asset directory must contain exactly five entries, found $n1_entry_count: $n1_asset_dir"
  n1_assert_file "$n1_asset_dir/$N1_DARWIN_ASSET" "$N1_DARWIN_SIZE" "$N1_DARWIN_SHA256"
  n1_assert_file "$n1_asset_dir/$N1_LINUX_ASSET" "$N1_LINUX_SIZE" "$N1_LINUX_SHA256"
  n1_assert_file "$n1_asset_dir/$N1_SOURCE_ASSET" "$N1_SOURCE_SIZE" "$N1_SOURCE_SHA256"
  n1_assert_file "$n1_asset_dir/$N1_MANIFEST_ASSET" "$N1_MANIFEST_SIZE" "$N1_MANIFEST_SHA256"
  n1_assert_file "$n1_asset_dir/$N1_SUMS_ASSET" "$N1_SUMS_SIZE" "$N1_SUMS_SHA256"
}

n1_copy_create_or_verify() {
  n1_copy_source="$1"
  n1_copy_destination="$2"
  n1_copy_size="$3"
  n1_copy_sha="$4"
  n1_assert_file "$n1_copy_source" "$n1_copy_size" "$n1_copy_sha"
  if [ -e "$n1_copy_destination" ] || [ -L "$n1_copy_destination" ]; then
    n1_assert_file "$n1_copy_destination" "$n1_copy_size" "$n1_copy_sha"
    return
  fi
  # noclobber makes publication create-only. An interrupted write is retained
  # as a detectable conflict rather than being silently replaced on retry.
  if ! (set -C; cat "$n1_copy_source" >"$n1_copy_destination"); then
    n1_fail "refused to overwrite destination: $n1_copy_destination"
  fi
  n1_assert_file "$n1_copy_destination" "$n1_copy_size" "$n1_copy_sha"
}

n1_assert_git_coordinates() {
  n1_repo_root="$1"
  [ "$(git -C "$n1_repo_root" rev-parse "$N1_SOURCE_COMMIT^{tree}")" = "$N1_SOURCE_TREE" ] || n1_fail "source tree drift"
  [ "$(git -C "$n1_repo_root" rev-parse "$N1_RELEASE_COMMIT^{tree}")" = "$N1_RELEASE_TREE" ] || n1_fail "release tree drift"
  [ "$(git -C "$n1_repo_root" rev-parse "$N1_RELEASE_COMMIT^")" = "$N1_SOURCE_COMMIT" ] || n1_fail "release parent drift"
  [ "$(git -C "$n1_repo_root" cat-file -t "refs/tags/$N1_TAG")" = "commit" ] || n1_fail "candidate tag is not lightweight"
  [ "$(git -C "$n1_repo_root" rev-parse "refs/tags/$N1_TAG^{commit}")" = "$N1_RELEASE_COMMIT" ] || n1_fail "candidate tag target drift"
  git -C "$n1_repo_root" merge-base --is-ancestor "$N1_RELEASE_COMMIT" HEAD || n1_fail "current HEAD does not descend from exact release R"
}
