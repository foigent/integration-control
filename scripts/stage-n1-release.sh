#!/bin/sh
set -eu

script_dir="$(CDPATH= cd "$(dirname "$0")" && pwd)"
. "$script_dir/n1-distribution-lib.sh"

if [ "$#" -ne 2 ]; then
  echo "usage: $0 EXACT_INPUT_ASSET_DIRECTORY DESTINATION_DIRECTORY" >&2
  exit 2
fi

input_dir="$1"
destination_dir="$2"
n1_require_absolute_path "$input_dir"
n1_require_absolute_path "$destination_dir"
n1_assert_flat_assets "$input_dir"

if [ -L "$destination_dir" ]; then
  n1_fail "destination directory must not be a symbolic link: $destination_dir"
fi
if [ -e "$destination_dir" ] && [ ! -d "$destination_dir" ]; then
  n1_fail "destination exists and is not a directory: $destination_dir"
fi
if [ ! -e "$destination_dir" ]; then
  n1_create_new_directory "$destination_dir"
fi

unknown_entries="$(find "$destination_dir" ! -path "$destination_dir" -prune \
  ! \( -name "$N1_DARWIN_ASSET" -o -name "$N1_LINUX_ASSET" -o -name "$N1_SOURCE_ASSET" \
  -o -name "$N1_MANIFEST_ASSET" -o -name "$N1_SUMS_ASSET" \) -print)"
[ -z "$unknown_entries" ] || n1_fail "destination contains an unexpected entry: $unknown_entries"

n1_copy_create_or_verify "$input_dir/$N1_DARWIN_ASSET" "$destination_dir/$N1_DARWIN_ASSET" "$N1_DARWIN_SIZE" "$N1_DARWIN_SHA256"
n1_copy_create_or_verify "$input_dir/$N1_LINUX_ASSET" "$destination_dir/$N1_LINUX_ASSET" "$N1_LINUX_SIZE" "$N1_LINUX_SHA256"
n1_copy_create_or_verify "$input_dir/$N1_SOURCE_ASSET" "$destination_dir/$N1_SOURCE_ASSET" "$N1_SOURCE_SIZE" "$N1_SOURCE_SHA256"
n1_copy_create_or_verify "$input_dir/$N1_MANIFEST_ASSET" "$destination_dir/$N1_MANIFEST_ASSET" "$N1_MANIFEST_SIZE" "$N1_MANIFEST_SHA256"
n1_copy_create_or_verify "$input_dir/$N1_SUMS_ASSET" "$destination_dir/$N1_SUMS_ASSET" "$N1_SUMS_SIZE" "$N1_SUMS_SHA256"
n1_assert_flat_assets "$destination_dir"

echo "status=PASS"
echo "publish_mode=create_only_exact_existing_is_noop"
echo "release_assets=$destination_dir"
