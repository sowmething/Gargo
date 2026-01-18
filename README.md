# Gargo
Remote Go compiler written in Go.

# Go Remote Build & Sync Tool

This project is a simple automation tool that uploads a local directory to a remote server via SFTP, runs a Go build on the remote machine, and downloads the compiled binary back to the local machine.

---

## Features

- Uploads a local directory to a remote server
- Runs `go mod init` / `go mod tidy` on the remote server
- Builds the project for a specified OS/ARCH
- Supports optional pre-build and post-build hooks
- Downloads the compiled binary to the local machine

---

## Configuration (config.ini)

Before running the tool, create a `config.ini` file like this:

```ini
[project]
name = myapp
directory = /home/user/myapp
inputdir = mycodes,true

[remote]
host = 192.168.1.10
port = 22
user = root
password = mypassword
expectedremote = linux/windows/any

[build]
os = linux/windows
arch = amd64/386/arm/arm64
CGO_ENABLED = 0
GOGC = 20
core = 1
small = true/false
before = echo "before build"
after = echo "after build"
```

## Field explanation

| Field               | Description                                                 |
| ------------------- | ----------------------------------------------------------- |
| `project.name`      | Output binary name                                          |
| `project.directory` | Remote directory to upload and build in                     |
| `project.inputdir`  | Local directory to upload. Use `,true` for recursive upload |
| `remote.*`          | SSH connection settings                                     |
| `build.os/arch`     | Target OS and architecture (Windows is not supported FOR NOW|
| `build.CGO_ENABLED` | CGO_ENABLED value (default `0`)                             |
| `build.GOGC`        | GOGC value (default `20`)                                   |
| `build.core`        | Number of CPU cores for build (default `1`)                 |
| `build.small`       | If `true`, builds with `-trimpath -ldflags="-s -w"`         |
| `build.before`      | Commands to run before build (comma separated)              |
| `build.after`       | Commands to run after build (comma separated)               |

## Build / Installing
# Build:
Install the source code
Open command prompt on the same directory
type "go build -trimpath -ldflags="-s -w" -o Gargo.exe"
Done.

# Using precompiled files
Locate to bin directory and there should be some files. To decide which one to run,
| File name           | OS, ARCH                                                    |
| ------------------- | ----------------------------------------------------------- |
| `Gargox64.exe`      | For Windows, x64 arch                                       |
| `Gargox86.exe`      | For Windows, x86 arch                                       |
| `Gargoarm.exe`      | For Windows, arm arch                                       |
| `Gargoarm64.exe`    | For Windows, arm64 arch                                     |
| `Gargox64`          | For Linux, x64 arch                                         |
| `Gargox86`          | For Linux, x86 arch                                         |
| `Gargoarm`          | For Linux, arm arch                                         |
| `Gargoarm64`        | For Linux, arm64 arch                                       |

## Usage
Run the tool.

## Troubleshooting
Make sure the remote server has Go installed
Make sure the remote directory exists or can be created
Ensure correct SSH credentials
