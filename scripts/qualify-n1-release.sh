#!/bin/sh
set -eu

script_dir="$(CDPATH= cd "$(dirname "$0")" && pwd)"
. "$script_dir/n1-distribution-lib.sh"

if [ "$#" -ne 2 ]; then
  echo "usage: $0 EXACT_ASSET_DIRECTORY NEW_WORK_DIRECTORY" >&2
  exit 2
fi

asset_dir="$1"
work_dir="$2"
n1_require_absolute_path "$asset_dir"
n1_require_absolute_path "$work_dir"
n1_assert_flat_assets "$asset_dir"
if [ -e "$work_dir" ] || [ -L "$work_dir" ]; then
  n1_fail "qualification work directory already exists; choose a new path: $work_dir"
fi
n1_create_new_directory "$work_dir"
mkdir -p "$work_dir/source" "$work_dir/bin"
tar -xf "$asset_dir/$N1_SOURCE_ASSET" -C "$work_dir/source"

case "$(uname -s):$(uname -m)" in
  Darwin:arm64)
    host_asset="$N1_DARWIN_ASSET"
    ;;
  Linux:x86_64|Linux:amd64)
    host_asset="$N1_LINUX_ASSET"
    ;;
  *)
    n1_fail "unsupported qualification host; use Darwin arm64 or Linux amd64"
    ;;
esac

tool="$work_dir/bin/foigent-integration-control"
cp "$asset_dir/$host_asset" "$tool"
chmod 0700 "$tool"
source_root="$work_dir/source/$N1_ARCHIVE_ROOT"
[ -d "$source_root" ] || n1_fail "source archive root is missing: $source_root"

"$tool" release verify --release "$asset_dir/$N1_MANIFEST_ASSET"
"$tool" schemas check --dir "$source_root/contracts/v1"
"$tool" authorization verify \
  --manifest "$source_root/fixtures/v1/valid/manifest.json" \
  --unirouter-ack "$source_root/fixtures/v1/valid/unirouter-ack.json" \
  --xinsight-ack "$source_root/fixtures/v1/valid/xinsight-ack.json" \
  --authorization "$source_root/fixtures/v1/valid/authorization.json" \
  --signature "$source_root/fixtures/v1/valid/authorization-signature.json" \
  --trust-set "$source_root/fixtures/v1/valid/authorization-trust-set.json"

echo "status=PASS"
echo "qualified_version=$N1_VERSION"
echo "qualified_release_manifest_sha256=$N1_MANIFEST_SHA256"
echo "host_asset=$host_asset"
echo "retained_work_directory=$work_dir"
