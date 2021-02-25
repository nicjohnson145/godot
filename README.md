# godot
Still another dotfiles manager, this time in go

This project mostly functioned as a learning exercise for teaching myself Go. I had written a custom
dotfiles manager in Python, so rewriting in Go seemed feasible since I knew my personal requirements
, thus making it a "solved" problem for me.

## Installation
Compiled binaries can be found on the releases page [here](https://github.com/nicjohnson145/godot/releases/latest)
. Optionally, clone this repository and run `go install`


## Setup

godot depends on 2 distinct configuration files.

### ~/.config/godot/config.json

This config contains only 2 values

Value | Meaning | Optional
------|---------|---------
target | This is the "name" of this computer, used to track what files it will recieve | True, defaults to current hostname
dotfiles_root | The root of the repository godot should use to look for templates and manage its settings | True, defaults to `~/dotfiles`
package_managers | What installation methods should attempt to be used when bootstrapping, valid values are `apt`, `brew`, `git` | True, will try to infer package manager from `GOOS`

Example:

```
{
    "target": "desktop",
    "dotfiles_root": "/home/njohnson/my_dotfiles",
    "package_maanagers": ["apt", "git"]
}
```


### <dotfiles_root>/config.json

This is managed by godot, and contains what files are under its control, and what hosts should get
what files. While this file can be managed by hand, it's better to let godot do it.

## Bootstrapping

Godot can bootstrap an environment (install packages, clone repositories, etc). Godot only assumes
that `git` is installed. It will attempt to use the configured package manager if the bootstrap item
can be installed that way, and will fall back to `git clone` if the bootstrap item cannot be installed
through a package manager, or git is the only available installation option. 

For example, suppose `pyenv` is configured to be installed via Homebrew and through a git checkout.
On a system where Homebrew is available, godot will use that to install `pyenv`. If Homebrew is not
available, godot will fall back to doing a `git clone`. Suppose `ripgrep` is only configured to be
installed via apt-get. On a system where apt-get is not available, godot will error, indicating that
it was unable to install `ripgrep`, as the only configured option was unavailable.

#### A note about apt

Since `apt get install <blarg>` requires elevated permissions, godot requires that the user can run
at minimum `sudo apt install` without a password prompt

## Templating

The files stored in the dotfiles repository will be evaluated as go templates. Information about
Go's templating can be found [here](https://golang.org/pkg/text/template/#hdr-Actions). Godot
defines supplements that with the following

<table>
<tr>
    <td>Value</td>
    <td>Type</td>
    <td>Meaning</td>
    <td>Example Usage</td>
</tr>
<tr>
    <td>Target</td>
    <td>variable</td>
    <td>The name of the current target</td>
    <td>Export an environment variable with the current target name <br /><code>export GODOT_TARGET="{{ .Target }}"</code> </td>
</tr>
<tr>
    <td>Submodules</td>
    <td>variable</td>
    <td>A special directory in the dotfiles repo for using git submodules</td>
    <td><code>export PATH="{{ .Submodules }}/fzf/bin:${PATH}"</code> </td>
</tr>
<tr>
    <td>Home</td>
    <td>variable</td>
    <td>Path to the current users home directory</td>
    <td><code>export PATH={{ .Home }}/bin:${PATH}</code></td>
</tr>
<tr>
    <td>oneOf</td>
    <td>function</td>
    <td>shorthand for evaulating if the current target is in a list</td>
    <td><pre>
{{ if oneOf . "work" "home" }}
export FOO="bar"
{{ end }}</pre></td>
</tr>
<tr>
    <td>notOneOf</td>
    <td>function</td>
    <td>The inverse of <code>oneOf</code>, evaluates if the target is not one of the list</td>
    <td></td>
</tr>
</table

