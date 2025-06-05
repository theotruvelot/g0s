# Developing g0s

- [Developing g0s](#developing-g0s)
  - [Getting started](#getting-started)
    - [Install dependencies](#install-dependencies)
  - [Local development](#local-development)
    - [Fork the repo](#fork-the-repo)
    - [Clone the repo](#clone-the-repo)
    - [Project structure](#project-structure)
  - [Running the components](#running-the-components)
    - [Agent](#agent)
    - [Server](#server)
    - [TUI](#tui)
  - [Create a pull request](#create-a-pull-request)
  - [Common tasks](#common-tasks)
    - [Running tests](#running-tests)
    - [Building the project](#building-the-project)
  - [Community channels](#community-channels)

## Getting started

Thank you for your interest in g0s (pronounced "ghost") and your willingness to contribute! g0s is a terminal-based server management tool built with Go — fast, intuitive, and lightweight.

### Install dependencies

You will need to install and configure the following dependencies on your machine:

- [Git](https://git-scm.com/)
- [Go](https://golang.org/doc/install) version 1.24 or higher
- [Make](https://www.gnu.org/software/make/) (optional, for using the Makefile)

## Local development

### Fork the repo

To contribute code to g0s, you must fork the g0s repo.

### Clone the repo

1. Clone your GitHub forked repo:

   ```sh
   git clone https://github.com/<your-username>/g0s.git
   ```

2. Go to the g0s directory:
   ```sh
   cd g0s
   ```

3. Add the upstream repository:
   ```sh
   git remote add upstream https://github.com/theotruvelot/g0s.git
   ```

### Project structure

The project is organized into three main components:

```
g0s/
├── cmd/
│   ├── agent/
│   │   └── main.go                 # Entry point of the agent
│   ├── cli/
│   │   └── main.go                 # Entry point of the CLI/TUI
│   └── server/
│       └── main.go                 # Entry point of the server
├── internal/
│   ├── agent/                      # Agent logic
│   │   ├── collector/              # Collect metrics and information
│   │   ├── monitor/                # Monitor services
│   │   └── executor/               # Execute commands
│   ├── cli/                        # CLI logic
│   │   ├── tui/                    # TUI components
│   │   ├── commands/               # Command definitions
│   │   └── prompt/                 # Command prompt
│   ├── server/                     # Server logic
│   │   ├── api/                    # HTTP/gRPC API
│   │   ├── handler/                # Request handlers
│   │   ├── middleware/             # Middlewares (auth, logging, etc.)
│   └── domain/                     # Shared business logic and types
│       ├── models/                 # Common data structures
│       ├── services/               # Shared business services
│       ├── config/                 # Shared configuration
│       └── errors/                 # Shared error types
├── pkg/                            # Potential public libraries
│   ├── logger/                     # Logging utility
│   ├── client/                     # API client
│   ├── protocol/                   # Communication protocol between components
│   └── utils/                      # Generic utilities
```

## Running the components

### Agent

To run the agent in development mode:

```sh
make run-agent-dev SERVER=<server-url> TOKEN=<your-token>
```

Or manually:

```sh
go run ./cmd/agent/main.go
```

### Server

To run the server in development mode:

```sh
go run ./cmd/server/main.go
```

### TUI

To run the Terminal UI in development mode:

```sh
go run ./cmd/cli/main.go
```

## Create a pull request

1. Create a new branch for your changes:
   ```sh
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit them:
   ```sh
   git add .
   git commit -m "Description of your changes"
   ```

3. Push to your fork:
   ```sh
   git push origin feature/your-feature-name
   ```

4. Open a pull request through the GitHub interface.

After making any changes, open a pull request. Once you submit your pull request, the maintainers will review it with you.

## Common tasks

### Running tests

Run all tests:

```sh
go test ./...
```

Run tests for a specific component:

```sh
go test ./agent/...
go test ./server/...
go test ./tui/...
```

### Building the project

Build all components:

```sh
go build ./cmd/...
```

Build a specific component:

```sh
# Build the agent
make build-agent

# Build the server
make build-server

# Build the CLI
make build-cli

# Or manually
go build ./cmd/agent/main.go

# Build the server
go build ./cmd/server/main.go

# Build the TUI
go build ./cmd/cli/main.go
```

## Community channels

If you get stuck somewhere or have any questions, feel free to open an issue or join our community channels.