# dyndns-go ![Go](https://img.shields.io/badge/Go-1.20%2B-blue) ![Go](https://img.shields.io/badge/For_unifi_users-only-red)

## Project Overview

dyndns-go is a lightweight, fast, and configurable Dynamic DNS client written in Go. It automatically updates DNS records for your domains, making it ideal for home labs, self-hosters, and anyone needing reliable dynamic DNS updates. Unifi only.

Managing dynamic IP addresses can be a hassle for self-hosted services, remote access, and home networks. dyndns-go was created to provide a simple, robust, and open-source solution for keeping your DNS records up-to-date, with minimal configuration and maximum reliability.

## Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Disclaimer](#disclaimer)
- [Currently Supported Registrars](#currently-supported-registrars)
- [Installation](#installation)
- [Build from Source](#build-from-source)
- [Usage](#usage)
- [Configuration File](#configuration-file)
- [Automation](#automation)
- [Contributions](#contributions)

# Features

- Fast and lightweight Dynamic DNS client written in Go
- Supports Strato registrar
- Simple configuration via JSON file
- Secure API key and secret management
- Command-line interface for easy automation
- Designed for Unifi users
- Can be scheduled with cron or other automation tools
- Open-source and community-driven

## Disclaimer

This tool is in no way associated with any DNS registrar, including Strato AG or Unifi. You use this tool at your own sole responsibility. Always ensure you comply with your registrar's terms of service and API usage policies.

## Currently Supported Registrars

- Strato
- (Extensible: support for other registrars can be added via code contributions)

## Installation

To install dyndns-go, download the appropriate binary for your operating system from the [releases](https://github.com/harrybawsac/dyndns-go/releases) page. Place the binary in a directory included in your system's `PATH`.

### Prerequisites

- Go 1.20 or newer (for building from source)
- Internet access

## Build from Source

1. Clone the repository:
   ```sh
   git clone https://github.com/harrybawsac/dyndns-go.git
   ```
2. Change into the project directory:
   ```sh
   cd dyndns-go
   ```
3. Build the binary:
   ```sh
   go build -o dyndns-go main.go
   ```
4. Move the binary to a directory in your `PATH`:
   ```sh
   mv dyndns-go /usr/local/bin/
   ```
5. Verify the installation:
   ```sh
   dyndns-go
   ```

## Usage

Run dyndns-go from the command line. Example:

```sh
dyndns-go -config config.json -storage storage.json
```

### Command Line Options

| Option      | Description                              |
| ----------- | ---------------------------------------- |
| `--config`  | Path to configuration file (JSON format) |
| `--storage` | Path to storage file (JSON format)       |

## Configuration File

The configuration file must be in JSON format and include your registrar credentials and API endpoint. Example:

```json
{
  "user": "yourstratodomain.com",
  "password": "asdfasdfasdfasfd",
  "host": "test.yourstratodomain.com",
  "unifiSiteManagerApiKey": "asdfasdfasdfasfd",
  "unifiSiteManagerHostId": "asdfasdfasdfasfd",
  "updateIpv4": true,
  "updateIpv6": true
}
```

### Configuration Fields

- `user`: Your Strato domain username
- `password`: Your Strato domain password
- `host`: The hostname to update
- `unifiSiteManagerApiKey`: Your Unifi Site Manager API key
- `unifiSiteManagerHostId`: Your Unifi Site Manager host ID
- `updateIpv4`: Whether to update the IPv4 address
- `updateIpv6`: Whether to update the IPv6 address

Refer to your registrar's documentation for details on obtaining API credentials. Same goes for Unifi Site Manager credentials.

## Automation

You can automate DNS updates using cron or other scheduling tools. Example cron job to update every day at 6am:

```sh
0 6 * * * /usr/local/bin/dyndns-go -config /path/to/config.json -storage /path/to/storage.json
```

This ensures your DNS records stay up-to-date even if your IP changes.

## Contributions

Contributions are welcome! To contribute:

1. Fork the repository on GitHub.
2. Create a new branch for your feature or bugfix.
3. Make your changes and commit them with clear messages.
4. Open a pull request describing your changes.

### Guidelines

- Follow Go best practices and code style.
- Write clear, concise documentation and comments.
- Add tests for new features or bugfixes.
- Be respectful and constructive in discussions.

### Issues & Support

If you find a bug or have a feature request, open an issue on GitHub. For questions, join the community chat or discussion board.
