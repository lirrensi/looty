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