# Pinned tool versions in test image

| Tool | Version | Source |
| --- | --- | --- |
| Go | 1.27.4 | golang:1.27.4-bookworm base image |
| Bun | latest | bun.sh installer |
| TypeScript | 5.x (latest 5) | bun devDependency |
| playwright-go | 0.6000.0 | go install |
| Chromium | (bundled with playwright) | playwright install |
| xmllint | (system libxml2) | apt-get |
