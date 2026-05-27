# Contributing to Termix

Termix is an open source project founded by Ing Muyleang.

## Report Issues

Use the GitHub issue templates:

- Bug report
- Feature request

Include your OS, terminal, shell, Termix version, and the exact command or TUI screen involved.

## Development

```powershell
go mod tidy
go test ./...
go build -o bin/termix.exe .
```

Keep pull requests focused and explain the user workflow being improved.

## Feedback

Send feedback to `muyleanging@gmail.com` or use the feedback form inside the Termix TUI.
