# godot
Still another dotfiles manager, this time in go

This project mostly functioned as a learning exercise for teaching myself Go. I had written a custom
dotfiles manager in Python, so rewriting in Go seemed feasible since I knew my personal requirements,
thus making it a "solved" problem for me.

## Installation

Compiled binaries can be found on the releases page [here](https://github.com/nicjohnson145/godot/releases/latest).

## Setup/Usage

Godot requires a Personal Access Token exported as `GITHUB_PAT`. instructions on generating a PAT
can be found [here](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token).
Optionally, if using the Hashicorp Vault integrations, the PAT can be pulled from Vault instead.

Godot requires 2 config files, `~/.config/godot/config.yaml` and a separate `config.yaml` at the
root of your dotfiles repo

The user config file (`~/.config/godot/config.yaml`) has the following options:

| Option | Description | Required | Default |
| ------ | ----------- | -------- | ------- |
| binary-dir | Where to place binaries downloaded from github releases | No | `~/bin` |
| github-user | Your github username | Yes | - |
| target | The name of the target of this particular computer | Yes | - |
| dotfiles-url | The url of your dotfiles repo | No | `https://github.com/<github-user>/dotfiles` |
| clone-location | The location you wish godot to clone its copy of your dotfiles repo (note this is separate from your own usage & clone) | No | `~/.config/godot/dotfiles` |
| build-location | Where to place the rendered config files to symlink against | No | `~/.config/godot/rendered` |
| package-manager | The package manager to use when installing system packages (currently only supports `apt` & `brew`) | No | OS specific |
| vault-config | All Hashicorp Vault related configurations. See the section on Vault for details | No | - |

The `config.yaml` is the actual configuration of what packages, config files, etc you want
installed. An example configuration is given below to get you started

```
github-releases:
- name: bat
  repo: sharkdp/bat
  is-archive: true
  tag: v0.20.0
  mac-pattern: '.*x86_64-apple-darwin.*'
  linux-pattern: '.*x86_64-unknown-linux-musl.*'

- name: fd
  repo: sharkdp/fd
  is-archive: true
  tag: v8.3.2
  mac-pattern: '.*x86_64-apple-darwin.*'
  linux-pattern: '.*x86_64-unknown-linux-musl.*'

git-repos:
- name: diff-so-fancy
  url: https://github.com/so-fancy/diff-so-fancy
  location: ~/github/diff-so-fancy
  commit: a673cb4d2707f64d92b86498a2f5f71c8e2643d5

config-files:
- name: dot_fdignore 
  destination: ~/.config/fd/ignore

- name: dot_gitconfig 
  destination: ~/.gitconfig

system-packages:
- name: tmux
  apt: tmux
  brew: tmux
- name: git
  apt: git
  brew: git

bundles:
- name: fd-bundle
  github-releases:
  - fd
  config-files:
  - dot_fdignore
targets:
  wsl:
    bundles:
    - fd-bundle
    system-packages:
    - tmux
    - git
    github-releases:
    - bat
    config-files:
    - dot_gitconfig
```

#### A note about apt

Since `apt get install <blarg>` requires elevated permissions, godot requires that the user can run
at minimum `sudo apt install` without a password prompt


### Hashicorp Vault Integrations

Godot has limited ability to pull values from Hashicorp Vault. These features are gated through the
`vault-config` section in `~/.config/godot/config.yaml`.

| Option | Description | Required | Default |
| ------ | ----------- | -------- | ------- |
| address | Address of the vault server | Yes | - |
| token-path | Path to the token file for vault | No | `~/.vault-token` |
| pat-from-vault | Instruct godot to pull the Github PAT from vault, requires `github-pat-config` to be set | No | - |
| github-pat-config | The path in vault & the key where the github pat is stored | No | - |

A full example of enabling Vault, as well as pulling the PAT from vault is given below

```
github-user: foobar
target: foo
vault-config:
  address: https://vault.some-domain.com
  pat-from-vault: true
  github-pat-config:
    path: secrets/some/path
    key: github-pat
```

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
        <td>Export an environment variable with the current target name
        <br /><code>export GODOT_TARGET="{{ .Target  }}"</code> </td>
    </tr>
    <tr>
        <td>Submodules</td>
        <td>variable</td>
        <td>A special directory in the dotfiles repo for using git submodules</td>
        <td><code>export PATH="{{ .Submodules}}/fzf/bin:${PATH}"</code> </td>
    </tr>
    <tr>
        <td>Home</td>
        <td>variable</td>
        <td>Path to the current users home directory</td>
        <td><code>export PATH={{ .Home}}/bin:${PATH}</code></td>
    </tr>
    <tr>
        <td>oneOf</td>
        <td>function</td>
        <td>shorthand for evaulating if the current target is in a list</td>
        <td>
            <pre>
{{ if oneOf . "work" "home"  }}
export FOO="bar"
{{ end }}
            </pre>
        </td>
    </tr>
    <tr>
        <td>notOneOf</td>
        <td>function</td>
        <td>The inverse of <code>oneOf</code>, evaluates if the target is not one of the list</td>
        <td></td>
    </tr>
    <tr>
        <td>VaultLookup</td>
        <td>function</td>
        <td>Lookup a Key/Value secret stored in Vault. Requires that vault configuration is set up</td>
        <td>
            <pre>
{{ VaultLookup "secrets/super-secrets" "my-secret-password" }}
            </pre>
        </td>
    </tr>
</table
