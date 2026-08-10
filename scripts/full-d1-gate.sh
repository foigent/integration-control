#!/bin/sh
set -eu

script_dir="$(CDPATH= cd "$(dirname "$0")" && pwd)"
. "$script_dir/n1-distribution-lib.sh"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 NEW_GATE_WORK_DIRECTORY" >&2
  exit 2
fi

gate_dir="$1"
n1_require_absolute_path "$gate_dir"
if [ -e "$gate_dir" ] || [ -L "$gate_dir" ]; then
  n1_fail "gate work directory already exists; choose a new path: $gate_dir"
fi
n1_create_new_directory "$gate_dir"
mkdir -p "$gate_dir/n1-gocache"

repo_root="$(git -C "$script_dir/.." rev-parse --show-toplevel)"
n1_assert_git_coordinates "$repo_root"
git -C "$repo_root" diff --exit-code "$N1_RELEASE_COMMIT" -- "releases/$N1_VERSION"
for executable_script in \
  "$script_dir/full-d1-gate.sh" \
  "$script_dir/full-n1-gate.sh" \
  "$script_dir/n1-distribution-lib.sh" \
  "$script_dir/qualify-n1-release.sh" \
  "$script_dir/rebuild-n1-candidate.sh" \
  "$script_dir/stage-n1-release.sh"
do
  [ -x "$executable_script" ] || n1_fail "distribution script is not executable: $executable_script"
done

workflow="$repo_root/.github/workflows/n1-distribution-qualification.yml"
[ -f "$workflow" ] && [ ! -L "$workflow" ] || n1_fail "distribution workflow is missing"
grep -Fq 'permissions:' "$workflow" || n1_fail "workflow permissions declaration is missing"
grep -Fq '  contents: read' "$workflow" || n1_fail "workflow does not pin read-only contents permission"
grep -Fq 'actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683' "$workflow" || n1_fail "checkout action pin drift"
grep -Fq 'actions/setup-go@0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32' "$workflow" || n1_fail "setup-go action pin drift"
if grep -Eq 'pull_request_target|upload-artifact|gh[[:space:]]+release|curl|wget' "$workflow"; then
  n1_fail "workflow exceeds read-only local qualification boundary"
fi
if grep -En 'rm[[:space:]]+-[^[:space:]]*r|git[[:space:]]+clean|find.*-de[l]ete' "$script_dir"/*.sh >/dev/null; then
  n1_fail "distribution scripts contain a recursive or bulk deletion command"
fi

go_command="$(n1_resolve_go_command)"
go_bin_dir="$(dirname "$go_command")"
(
  cd "$repo_root"
  PATH="$go_bin_dir:$PATH" GOENV=off GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly \
    GOCACHE="$gate_dir/n1-gocache" N1_REMOTE_POLICY="${N1_REMOTE_POLICY:-none}" ./scripts/full-n1-gate.sh
)

"$script_dir/rebuild-n1-candidate.sh" "$gate_dir/rebuild"
rebuilt_assets="$gate_dir/rebuild/release-assets"
published_assets="$gate_dir/simulated-release"
"$script_dir/stage-n1-release.sh" "$rebuilt_assets" "$published_assets"
"$script_dir/stage-n1-release.sh" "$rebuilt_assets" "$published_assets"
"$script_dir/qualify-n1-release.sh" "$published_assets" "$gate_dir/consumer"

mkdir -p "$gate_dir/conflict-release"
printf '%s\n' 'intentional-conflict-fixture' >"$gate_dir/conflict-release/$N1_DARWIN_ASSET"
if "$script_dir/stage-n1-release.sh" "$rebuilt_assets" "$gate_dir/conflict-release" >/dev/null 2>&1; then
  n1_fail "conflicting existing asset was not rejected"
fi

"$script_dir/stage-n1-release.sh" "$rebuilt_assets" "$gate_dir/extra-release" >/dev/null
printf '%s\n' 'intentional-extra-fixture' >"$gate_dir/extra-release/unexpected.asset"
if "$script_dir/qualify-n1-release.sh" "$gate_dir/extra-release" "$gate_dir/extra-consumer" >/dev/null 2>&1; then
  n1_fail "unexpected release asset was not rejected"
fi

if [ "${D1_REQUIRE_CLEAN:-1}" = "1" ]; then
  [ -z "$(git -C "$repo_root" status --porcelain)" ] || n1_fail "repository is not clean"
fi
git -C "$repo_root" diff --check
git -C "$repo_root" diff --cached --check

echo "status=PASS"
echo "gate=full_d1_local_distribution"
echo "double_build_equal=true"
echo "idempotent_exact_restage=true"
echo "conflicting_overwrite_rejected=true"
echo "extra_asset_rejected=true"
echo "qualified_release_manifest_sha256=$N1_MANIFEST_SHA256"
echo "retained_gate_directory=$gate_dir"
