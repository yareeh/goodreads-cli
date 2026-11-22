#!/usr/bin/env bash
# Install this repo's git hooks into the current clone. Run once per clone /
# host (also via `make hooks`):
#
#   scripts/install-hooks.sh
#
# Hooks are symlinked, so later changes to scripts/hooks/* take effect without
# reinstalling. Works from any worktree (hooks live in the shared git dir) and
# leaves pre-commit's own .git/hooks/pre-commit alone.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
hooks_dir="$(git rev-parse --path-format=absolute --git-common-dir)/hooks"
mkdir -p "$hooks_dir"

for src in "$root"/scripts/hooks/*; do
	name="$(basename "$src")"
	dest="$hooks_dir/$name"
	chmod +x "$src"
	if [ -e "$dest" ] && [ ! -L "$dest" ]; then
		mv "$dest" "$dest.bak"
		echo "moved existing $name hook to $dest.bak"
	fi
	# Point at the main checkout's copy so the hook survives worktree removal.
	target="$(git rev-parse --path-format=absolute --git-common-dir)/../scripts/hooks/$name"
	[ -e "$target" ] || target="$src"
	ln -sf "$target" "$dest"
	echo "installed $name -> $target"
done
