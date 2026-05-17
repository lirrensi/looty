# Looty Agent Notes

## Known Issues

### pnpm-workspace.yaml breaks CI
If `pnpm install --frozen-lockfile` fails with:
```
ERROR packages field missing or empty
```

**Cause:** A malformed `pnpm-workspace.yaml` file exists in `web/`. This file is invalid (not a proper workspace config) and breaks pnpm.

**Fix:** Delete `web/pnpm-workspace.yaml`. This file should never exist in the project.

## Prevention
Ensure `web/pnpm-workspace.yaml` is added to `.gitignore` or never committed. It should never exist in the repo.

---

## Releasing (MAKE RELEASE)

When the user says "MAKE RELEASE", do this EXACT sequence. No shortcuts. No assumptions.

### 1. Read the VERSION file
```bash
cat VERSION
# Expected output: 1.4.0 (or whatever the current version is)
```

### 2. Commit everything
```bash
git add -A
git commit -m "Release: <description of changes>"
git push origin main
```

### 3. Create and push the tag
The tag MUST match the VERSION file with a `v` prefix:
```bash
# If VERSION contains "1.4.0":
git tag v1.4.0 -m "Looty v1.4.0"
git push origin v1.4.0
```

### 4. Verify the release was created
Wait 2-3 minutes for the workflow to complete, then verify:
```bash
# Check via GitHub API
curl -s https://api.github.com/repos/lirrensi/looty/releases/latest | python -c "import sys,json; r=json.load(sys.stdin); print('Tag:', r.get('tag_name', 'NO RELEASE')); print('Assets:', len(r.get('assets', [])))"

# Should output something like:
# Tag: v1.4.0
# Assets: 6
```

If `Tag:` shows anything other than the expected version, the release failed.

### CRITICAL RULES

1. **Tag format:** Always use `v` prefix: `v1.4.0`, NOT `1.4.0`
2. **Never manually set `tag_name` in the workflow** — the action auto-detects from the git ref
3. **Always push the tag separately:** `git push origin v1.4.0` — this is what triggers the release
4. **Verify after pushing:** Don't assume it worked. Check the API.

### Common Failures

| Symptom | Cause | Fix |
|---------|-------|-----|
| Release workflow runs but no release appears | `tag_name` was overridden in workflow to `1.4.0` instead of `v1.4.0` | Remove `tag_name` from workflow, let action auto-detect |
| Workflow doesn't run at all | Tag not pushed to remote | Must run `git push origin v1.4.0` |
| pnpm install fails | `web/pnpm-workspace.yaml` exists | Delete it, add to `.gitignore` |
| Old binaries in new release | Tag moved but workflow cached old artifacts | Delete tag remotely and recreate: `git push origin :refs/tags/v1.4.0 && git tag -d v1.4.0 && git tag v1.4.0 -m "Looty v1.4.0" && git push origin v1.4.0` |

### Emergency: Recreate a Broken Release

If a release is broken (wrong binaries, missing assets, etc.):

```bash
# 1. Delete the broken release (via GitHub UI or API)
# 2. Delete the tag locally and remotely
git tag -d v1.4.0
git push origin :refs/tags/v1.4.0

# 3. Ensure the fix is on main
git push origin main

# 4. Recreate the tag at the latest commit
git tag v1.4.0 -m "Looty v1.4.0"
git push origin v1.4.0

# 5. Verify
sleep 120
curl -s https://api.github.com/repos/lirrensi/looty/releases/latest | python -c "import sys,json; r=json.load(sys.stdin); print('Tag:', r.get('tag_name')); print('Assets:', len(r.get('assets', [])))"
```
