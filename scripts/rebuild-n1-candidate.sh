#!/bin/sh
set -eu

script_dir="$(CDPATH= cd "$(dirname "$0")" && pwd)"
. "$script_dir/n1-distribution-lib.sh"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 NEW_WORK_DIRECTORY" >&2
  exit 2
fi

work_dir="$1"
n1_require_absolute_path "$work_dir"
if [ -e "$work_dir" ] || [ -L "$work_dir" ]; then
  n1_fail "work directory already exists; choose a new path: $work_dir"
fi
n1_create_new_directory "$work_dir"

repo_root="$(git -C "$script_dir/.." rev-parse --show-toplevel)"
n1_assert_git_coordinates "$repo_root"

go_command="$(n1_resolve_go_command)"
actual_go_version="$(GOTOOLCHAIN=local "$go_command" version)"
case "$actual_go_version" in
  "go version go1.25.12 "*) ;;
  *) n1_fail "expected go1.25.12 toolchain, got: $actual_go_version" ;;
esac

mkdir -p "$work_dir/archive-a" "$work_dir/archive-b" "$work_dir/tree-a" "$work_dir/tree-b" \
  "$work_dir/build-a" "$work_dir/build-b" "$work_dir/gocache-a" "$work_dir/gocache-b" \
  "$work_dir/release-assets" "$work_dir/release-inputs"

archive_a="$work_dir/archive-a/$N1_SOURCE_ASSET"
archive_b="$work_dir/archive-b/$N1_SOURCE_ASSET"
LC_ALL=C TZ=UTC git -C "$repo_root" archive --format=tar --prefix="$N1_ARCHIVE_ROOT/" --output="$archive_a" "$N1_SOURCE_COMMIT"
LC_ALL=C TZ=UTC git -C "$repo_root" archive --format=tar --prefix="$N1_ARCHIVE_ROOT/" --output="$archive_b" "$N1_SOURCE_COMMIT"
cmp "$archive_a" "$archive_b"
n1_assert_file "$archive_a" "$N1_SOURCE_SIZE" "$N1_SOURCE_SHA256"
n1_assert_file "$archive_b" "$N1_SOURCE_SIZE" "$N1_SOURCE_SHA256"

tar -xf "$archive_a" -C "$work_dir/tree-a"
tar -xf "$archive_b" -C "$work_dir/tree-b"

build_one() {
  n1_source_dir="$1"
  n1_cache_dir="$2"
  n1_output_dir="$3"
  (
    cd "$n1_source_dir"
    GOENV=off GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly \
      GOCACHE="$n1_cache_dir" CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
      "$go_command" build -trimpath -ldflags=-buildid= -o "$n1_output_dir/$N1_DARWIN_ASSET" ./cmd/foigent-integration-control
    GOENV=off GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly \
      GOCACHE="$n1_cache_dir" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      "$go_command" build -trimpath -ldflags=-buildid= -o "$n1_output_dir/$N1_LINUX_ASSET" ./cmd/foigent-integration-control
  )
}

build_one "$work_dir/tree-a/$N1_ARCHIVE_ROOT" "$work_dir/gocache-a" "$work_dir/build-a"
build_one "$work_dir/tree-b/$N1_ARCHIVE_ROOT" "$work_dir/gocache-b" "$work_dir/build-b"

cmp "$work_dir/build-a/$N1_DARWIN_ASSET" "$work_dir/build-b/$N1_DARWIN_ASSET"
cmp "$work_dir/build-a/$N1_LINUX_ASSET" "$work_dir/build-b/$N1_LINUX_ASSET"
n1_assert_file "$work_dir/build-a/$N1_DARWIN_ASSET" "$N1_DARWIN_SIZE" "$N1_DARWIN_SHA256"
n1_assert_file "$work_dir/build-b/$N1_DARWIN_ASSET" "$N1_DARWIN_SIZE" "$N1_DARWIN_SHA256"
n1_assert_file "$work_dir/build-a/$N1_LINUX_ASSET" "$N1_LINUX_SIZE" "$N1_LINUX_SHA256"
n1_assert_file "$work_dir/build-b/$N1_LINUX_ASSET" "$N1_LINUX_SIZE" "$N1_LINUX_SHA256"

manifest_input="$work_dir/release-inputs/$N1_MANIFEST_ASSET"
sums_input="$work_dir/release-inputs/$N1_SUMS_ASSET"
git -C "$repo_root" show "$N1_RELEASE_COMMIT:releases/$N1_VERSION/$N1_MANIFEST_ASSET" >"$manifest_input"
git -C "$repo_root" show "$N1_RELEASE_COMMIT:releases/$N1_VERSION/$N1_SUMS_ASSET" >"$sums_input"
n1_assert_file "$manifest_input" "$N1_MANIFEST_SIZE" "$N1_MANIFEST_SHA256"
n1_assert_file "$sums_input" "$N1_SUMS_SIZE" "$N1_SUMS_SHA256"

asset_dir="$work_dir/release-assets"
n1_copy_create_or_verify "$work_dir/build-a/$N1_DARWIN_ASSET" "$asset_dir/$N1_DARWIN_ASSET" "$N1_DARWIN_SIZE" "$N1_DARWIN_SHA256"
n1_copy_create_or_verify "$work_dir/build-a/$N1_LINUX_ASSET" "$asset_dir/$N1_LINUX_ASSET" "$N1_LINUX_SIZE" "$N1_LINUX_SHA256"
n1_copy_create_or_verify "$archive_a" "$asset_dir/$N1_SOURCE_ASSET" "$N1_SOURCE_SIZE" "$N1_SOURCE_SHA256"
n1_copy_create_or_verify "$manifest_input" "$asset_dir/$N1_MANIFEST_ASSET" "$N1_MANIFEST_SIZE" "$N1_MANIFEST_SHA256"
n1_copy_create_or_verify "$sums_input" "$asset_dir/$N1_SUMS_ASSET" "$N1_SUMS_SIZE" "$N1_SUMS_SHA256"
n1_assert_flat_assets "$asset_dir"

echo "status=PASS"
echo "source_commit=$N1_SOURCE_COMMIT"
echo "source_tree=$N1_SOURCE_TREE"
echo "release_commit=$N1_RELEASE_COMMIT"
echo "release_tree=$N1_RELEASE_TREE"
echo "tag=$N1_TAG"
echo "double_build_equal=true"
echo "release_assets=$asset_dir"
echo "retained_work_directory=$work_dir"
