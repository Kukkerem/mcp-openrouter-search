#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

usage() {
  echo "Usage: $0 <version>"
  echo "  version: semver tag, e.g. v0.1.4 or 0.1.4"
  exit 1
}

[[ $# -eq 1 ]] || usage

VERSION="${1#v}"  # strip leading 'v' for metadata
TAG="v${VERSION}"

cd "$REPO_ROOT"

# Sanity checks
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "error: working directory has uncommitted changes" >&2
  exit 1
fi

if git tag | grep -qx "$TAG"; then
  echo "error: tag $TAG already exists" >&2
  exit 1
fi

echo "Releasing $TAG..."

# Patch version in flake.nix (inside the let binding)
sed -i "s/version = \"[^\"]*\";/version = \"${VERSION}\";/" flake.nix

# Verify the substitution landed
if ! grep -q "version = \"${VERSION}\";" flake.nix; then
  echo "error: failed to update version in flake.nix" >&2
  git checkout flake.nix
  exit 1
fi

git add flake.nix
git commit -m "chore: bump flake version to ${TAG}"
git tag "$TAG"
git push origin HEAD "$TAG"

echo "Done. Tag $TAG pushed."
echo "GitHub Actions release workflow will run automatically."
