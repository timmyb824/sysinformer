---
description: Automatically bump version in main.go based on last commit, push to main, and trigger release via bash script.
---

1. Look up the last git commit message.
   - Run: `git log -1 --pretty=%B`
2. Determine the version bump type using the following chart:

   ```text
   major_tags = [":boom:", "BREAKING_CHANGE"]
   minor_tags = ["feat"]
   patch_tags = ["fix", "perf", "style", "docs", "ci", "test", "refactor", "chore", "build"]
   ```

   - If the commit message contains any of the `major_tags`, bump the major version (X.0.0).
   - If it contains any of the `minor_tags`, bump the minor version (X.Y.0).
   - If it contains any of the `patch_tags`, bump the patch version (X.Y.Z).
   - If multiple tags are present, use the highest precedence (major > minor > patch).

3. Open `main.go` and update the `var Version = "vX.Y.Z"` line to the new version.
4. Stage and commit this change:
   - Run: `git add main.go`
   - Run: `git commit -m "chore: update to version vX.Y.Z" -n`
   - Run: `git push origin main`
5. Push the new version using the bash script:
   - Run: `./tag_and_push.sh vX.Y.Z`

// turbo-all

**Notes:**

- This workflow assumes that `main.go` contains the version in the format: `var Version = "vX.Y.Z"`.
- Make sure you have permissions to push to main and execute the script.
- If the commit message does not match any tag, do not bump the version.
