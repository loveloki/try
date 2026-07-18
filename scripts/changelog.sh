#!/usr/bin/env bash
# 生成 Release body：自 HEAD 最近的祖先 tag 起，按 goreleaser 规则过滤 commit。
# 用法: ./scripts/changelog.sh [当前 tag，如 v0.4.0]
set -euo pipefail

CURRENT="${1:-${GITHUB_REF_NAME:-}}"
if [[ -n "$CURRENT" && "$CURRENT" != v* ]]; then
	CURRENT="v${CURRENT}"
fi

HEAD_COMMIT=$(git rev-parse HEAD)
PREV=""
while IFS= read -r t; do
	[[ -z "$t" || "$t" == "$CURRENT" ]] && continue
	# 必须是 HEAD 的真祖先（排除打在同一 commit 上的旧 tag）
	tip=$(git rev-parse "$t^{commit}" 2>/dev/null) || continue
	[[ "$tip" == "$HEAD_COMMIT" ]] && continue
	if git merge-base --is-ancestor "$t" HEAD 2>/dev/null; then
		PREV="$t"
		break
	fi
done < <(git tag --sort=-version:refname)

echo "## Changelog"
echo
if [[ -z "$PREV" ]]; then
	git log --pretty=format:'* %s (%h)' --reverse HEAD
else
	git log --pretty=format:'* %s (%h)' --reverse "${PREV}..HEAD"
fi | grep -vE '^\* (docs|test|chore)(\([^)]*\))?:' || true
echo
