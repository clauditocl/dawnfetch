<h1 align="center">dawnfetch</h1>

<p align="center">
  a fast, cross-platform system info CLI with beautiful theme support
</p>

<p align="center">
  <a href="https://github.com/almightynan/dawnfetch/actions/workflows/release-build.yml"><img src="https://img.shields.io/github/actions/workflow/status/almightynan/dawnfetch/release-build.yml?label=release%20build" alt="release build status"></a>
  <a href="https://github.com/almightynan/dawnfetch/releases"><img src="https://img.shields.io/github/v/release/almightynan/dawnfetch?display_name=tag" alt="latest release"></a>
  <img src="https://img.shields.io/badge/go-1.20+-00ADD8?logo=go&logoColor=white" alt="go version">
  <a href="https://github.com/almightynan/dawnfetch/stargazers"><img src="https://img.shields.io/github/stars/almightynan/dawnfetch" alt="stars"></a>
</p>

<div align="center">
  <a href="#installation">Installation</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="#npm-publish-automation">NPM</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="#usage">Usage</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="#themes">Themes</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="#build">Build</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="#benchmarking">Benchmarking</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://github.com/almightynan/dawnfetch/issues">Issues</a>
</div>

---

## What Is dawnfetch?

`dawnfetch` is a fast system info tool inspired by neofetch/fastfetch, with:

- cross-platform support (`linux`, `macos`, `windows`)
- built-in themed palettes (`themes.json`)
- large ascii logo set (`ascii/`)
- interactive theme preview tui (`preview-theme`)
- optional image logos (`png`, `jpg/jpeg`, `webp`, `gif`, `bmp`, `tiff`)


## Installation

### Linux and macOS

```bash
curl -fsSL https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.sh | bash
```

This installs:

- binary: `~/.local/bin/dawnfetch`
- assets: `~/.local/share/dawnfetch/themes.json` and `~/.local/share/dawnfetch/ascii/`

### Windows (PowerShell)

```powershell
powershell -c "irm https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1 | iex"
```

### npm

```bash
npm i -g @almightynan/dawnfetch
```

The npm package runs the same platform installer (`cli/install.sh` or `cli/install.ps1`) during `postinstall`.

### Windows 7 / PowerShell 2 fallback

```powershell
$f="$env:TEMP\dawnfetch-install.ps1";(New-Object Net.WebClient).DownloadFile('https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1',$f);powershell -File $f
```

If you are in `cmd.exe`:

```bat
powershell -c "$f=$env:TEMP+'\dawnfetch-install.ps1';(New-Object Net.WebClient).DownloadFile('https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1',$f);powershell -File $f"
```

---

## Usage

### Quick start

```bash
dawnfetch
```

From source:

```bash
go run .
```

Build local binary:

```bash
go build -o dawnfetch .
./dawnfetch
```

### Main commands

| Command | Description |
|---|---|
| `dawnfetch` | default output |
| `dawnfetch --help` | show help |
| `dawnfetch --version` | show version |
| `dawnfetch --list-themes` | list available themes |
| `dawnfetch preview-theme` | interactive theme preview |
| `dawnfetch set-default-theme <name>` | save default theme |
| `dawnfetch doctor` | diagnostics |

### Common flags

| Flag | Description |
|---|---|
| `--theme <name>` | use theme for current run |
| `--themes <path>` | use custom palettes file |
| `--full` | collect fuller/slower info |
| `--image <path>` | use image as logo |
| `--no-logo` | hide logo |
| `--no-color` | disable ansi colors |


## Themes

`dawnfetch` loads palettes from `themes.json`.

Examples:

```bash
dawnfetch --list-themes
dawnfetch --theme transgender
dawnfetch preview-theme
dawnfetch set-default-theme nonbinary
```


## Logos
All logos in `/ascii` are from the `fastfetch-cli/fastfetch` repository.

## Config

Config file name is fixed:

```text
dawnfetch_config.json
```

Load/save behavior:

- portable/local binary: prefers config next to executable when writable
- system install: prefers user config path
  - linux/macos: `~/.config/dawnfetch/dawnfetch_config.json` (or `$XDG_CONFIG_HOME/dawnfetch/dawnfetch_config.json`)
  - windows: `%AppData%\dawnfetch\dawnfetch_config.json`

---

## Build

Build scripts are in `build/`:

```bash
./build/build-linux.sh
./build/build-macos.sh
```

PowerShell for Windows:

```powershell
./build/build-windows.ps1
```
