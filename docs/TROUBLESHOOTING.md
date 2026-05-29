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

## Missing MesloLGM Nerd Font

This is a warning, not a startup failure. Use:

```bash
termix fonts list
termix fonts install "MesloLGM Nerd Font" --yes
termix fonts apply "MesloLGM Nerd Font"
```

On Apple Terminal, restart Terminal after applying. If needed, open `Terminal > Settings > Profiles > Text > Font` and choose `MesloLGM Nerd Font Mono`.

## Windows Terminal F1 popup

Termix uses `?` and `h` for help. If Windows Terminal still opens Help, unbind F1 in Windows Terminal settings:

```json
{
  "command": "unbound",
  "keys": "f1"
}
```
