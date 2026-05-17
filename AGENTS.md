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

## Releasing

To create a new release:

1. **Read the current version** from `VERSION` file
2. **Commit all changes** you want in this release
3. **Create and push the tag** (this triggers the release workflow):
   ```bash
   git add -A
   git commit -m "Release: <description>"
   git push origin main

   # Create tag with same version as VERSION file content
   git tag v<VERSION> -m "Looty v<VERSION>"
   git push origin v<VERSION>
   ```

**Critical:** The tag MUST be pushed with `git push origin v<VERSION>` — the release workflow only runs when the tag exists on the remote. Pushing the tag is what triggers `softprops/action-gh-release` to create the release automatically.