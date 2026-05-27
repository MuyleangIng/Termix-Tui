# Release

Pushing a version tag creates a GitHub Release through GoReleaser.

```bash
git checkout main
git pull
go test ./...
git tag -a v0.1.0 -m "Termix v0.1.0 first public release"
git push origin v0.1.0
```

Expected assets:

- `termix_Windows_x86_64.zip`
- `termix_Windows_arm64.zip`
- `termix_Linux_x86_64.tar.gz`
- `termix_Linux_arm64.tar.gz`
- `termix_Darwin_x86_64.tar.gz`
- `termix_Darwin_arm64.tar.gz`
- `checksums.txt`
