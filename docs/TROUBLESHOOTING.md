# Troubleshooting

## CONFIG NOT FOUND

Run:

```bash
termix repair
```

## Theme cache empty

Run:

```bash
termix themes update
termix cache rebuild
```

## Missing CaskaydiaCove Nerd Font

This is a warning, not a startup failure. Use:

```bash
termix fonts list
termix fonts apply "Cascadia Code"
```

## Windows Terminal F1 popup

Termix uses `?` and `h` for help. If Windows Terminal still opens Help, unbind F1 in Windows Terminal settings:

```json
{
  "command": "unbound",
  "keys": "f1"
}
```
