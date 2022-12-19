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

```yaml
executors:
  bat:
    type: github-release
    spec:
      repo: sharkdp/bat
      is-archive: true
      tag: v0.20.0
      mac-pattern: '.*x86_64-apple-darwin.*'
      linux-pattern: '.*x86_64-unknown-linux-musl.*'
  fd:
    type: github-release
    spec:
      repo: sharkdp/fd
      is-archive: true
      tag: v8.3.2
      mac-pattern: '.*x86_64-apple-darwin.*'
      linux-pattern: '.*x86_64-unknown-linux-musl.*'
  diff-so-fancy:
    type: git-repo
    spec:
      url: https://github.com/so-fancy/diff-so-fancy
      location: ~/github/diff-so-fancy
      ref:
        commit: a673cb4d2707f64d92b86498a2f5f71c8e2643d5
  dot_fdignore:
    type: config-file
    spec:
      template-name: dot_fdignore
      destination: ~/.config/fd/ignore
  dot_gitconfig:
    type: config-file
    spec:
      template-name: dot_gitconfig
      destination: ~/.gitconfig
  tmux:
    type: sys-package
    spec:
      apt: tmux
      brew: tmux
  fd-bundle:
    type: bundle
    spec:
      items:
      - fd
      - dot_fdignore
targets:
  work:
  - fd-bundle
  - tmux
  - git
  - bat
  - diff-so-fancy
```

## Executors

There are several types of configuration that godot can manage, they are as follows:

### Config Files

```go
type ConfigFile struct {
	Name         string `yaml:"-"`
	TemplateName string `yaml:"template-name" mapstructure:"template-name"`
	Destination  string `yaml:"destination" mapstructure:"destination"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| template-name | the name of the template in the templates folder of the dotfiles repo | Yes |
| destination | where the symlink to the rendered config file should be created | Yes |

### Git Repo

```go
type GitRepo struct {
	Name        string `yaml:"-"`
	URL         string `yaml:"url" mapstructure:"url"`
	Location    string `yaml:"location" mapstructure:"location"`
	Private     bool   `yaml:"private" mapstructure:"private"`
	TrackLatest bool   `yaml:"track-latest" mapstructure:"track-latest"`
	Ref         Ref    `yaml:"ref" mapstructure:"ref"`
}

type Ref struct {
	Commit string `yaml:"commit" mapstructure:"commit"`
	Tag    string `yaml:"tag" mapstructure:"tag"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| url | the url of the git repository | Yes |
| location | where to clone the repository | Yes |
| private | is this a private repo | No |
| track-latest | should this repo be kept up to date with the latest changes | No |
| ref.commit | ensure that this commit SHA is checked out when the repo is cloned | No |
| ref.tag | ensure that this tag is checked out when the repo is cloned | No |

### Github Release

```go
type GithubRelease struct {
	Name           string `yaml:"-"`
	Repo           string `yaml:"repo" mapstructure:"repo"`
	Tag            string `yaml:"tag" mapstructure:"tag"`
	IsArchive      bool   `yaml:"is-archive" mapstructure:"is-archive"`
	Regex          string `yaml:"regex" mapstructure:"regex"`
	MacPattern     string `yaml:"mac-pattern" mapstructure:"mac-pattern"`
	LinuxPattern   string `yaml:"linux-pattern" mapstructure:"linux-pattern"`
	WindowsPattern string `yaml:"windows-pattern" mapstructure:"windows-pattern"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| repo | which repository hosts the binary | Yes |
| tag | what release to download (can be "LATEST") | Yes |
| is-archive | indicate if the binary is packaged as an archive. Normally this can be auto detected | No |
| regex | a regex to find the binary when unpacking an archive release. Only required if multiple files in the archive are executable | No |
| mac-pattern | a regex of which asset link to download when running on mac | No |
| linux-pattern | a regex of which asset link to download when running on linux | No |
| windows-pattern | a regex of which asset link to download when running on windows | No |

### System Package

```go
type SystemPackage struct {
	Name     string `yaml:"-"`
	AptName  string `yaml:"apt" mapstructure:"apt"`
	BrewName string `yaml:"brew" mapstructure:"brew"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| apt | the name of the package when running `apt install` | No |
| brew | the name of the package when running `brew install` | No |

#### A note about apt

Since `apt get install <blarg>` requires elevated permissions, godot requires that the user can run
at minimum `sudo apt install` without a password prompt, otherwise the tool will prompt you during
execution

### Url Download

```go
type UrlDownload struct {
	Name       string `yaml:"-"`
	Tag        string `yaml:"tag" mapstructure:"tag"`
	MacUrl     string `yaml:"mac-url" mapstructure:"mac-url"`
	LinuxUrl   string `yaml:"linux-url" mapstructure:"linux-url"`
	WindowsUrl string `yaml:"windows-url" mapstructure:"windows-url"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| tag | the tag to download, will be available as a template var for the urls. Al | No |
| mac-url | the url to download from when running on mac | No |
| linux-url | the url to download from when running on linux | No |
| windows-url | the url to download from when running on windows | No |

### Bundle

```go
type Bundle struct {
	Name  string `yaml:"-"`
	Items []string `yaml:"items"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| items | the names of any executors that should be installed as part of this bundle | Yes |

### Golang

*Note:* this executor is currently only available when running on linux

```go
type Golang struct {
	Name string `yaml:"-"`
	Version string `yaml:"version" mapstructure:"version"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| version | the version of go to install | Yes |

### Go Install

```go
type GoInstall struct {
	Name string `yaml:"-"`
	Package string `yaml:"package" mapstructure:"package"`
	Version string `yaml:"version" mapstructure:"version"`
}
```

| Field | Description | Required |
| ------| ----------- | -------- |
| package | the `go get` path of the module | Yes |
| version | the version of the module to install, defaults to latest | No |

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
    <tr>
        <td>IsInstalled</td>
        <td>function</td>
        <td>Check if another piece of configuration is installed</td>
        <td>
            <pre>
                {{ if IsInstalled "some-tool" }}
                some-tool -init-env
                {{ end }}
            </pre>
        </td>
    </tr>
</table
