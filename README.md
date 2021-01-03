# godot
Still another dotfiles manager, this time in go

This project mostly functioned as a learning exercise for teaching myself Go. I had written a custom
dotfiles manager in Python, so rewriting in Go seemed feasible since I knew my personal requirements
, thus making it a "solved" problem for me.

## Setup

godot depends on 2 distinct configuration files.

### ~/.config/godot/config.json

This config contains only 2 values

Value | Meaning | Optional
------|---------|---------
target | This is the "name" of this computer, used to track what files it will recieve | False
dotfiles_root | The root of the repository godot should use to look for templates and manage its settings | True, defaults to `~/dotfiles`

Example:

```
{
	"target": "desktop",
	"dotfiles_root": "/home/njohnson/my_dotfiles"
}
```


### <dotfiles_root>/config.json

This is managed by godot, and contains what files are under its control, and what hosts should get
what files
