# Lockbox

Lockbox is a command-line tool for securely sharing secrets within a team. It uses modern encryption (age) to ensure only team members can decrypt shared secrets.

## Features

- Built-in encryption (no external dependencies)
- Team member key management
- Personal key management across repositories
- Interactive command-line interface
- Git repository integration
- Cross-platform support (Linux, macOS, Windows)

## Installation

### Using Homebrew (macOS)

```bash
brew install yourusername/tap/lockbox
```

### From Binary Releases

Download the appropriate binary for your platform from the [releases page](https://github.com/yourusername/lockbox/releases).

## Usage

### Initialize Repository

Create a `.lockbox` directory in your Git repository:

```bash
lockbox init
```

### Managing Personal Keys

Your personal keys are stored in `~/.lockbox` and can be used across multiple repositories.

Create a new key:
```bash
lockbox key add
# Follow the interactive prompts
```

List your keys:
```bash
lockbox key list
```

Remove a key:
```bash
lockbox key remove
# Select key to remove from the list
```

### Team Management

Add a team member:
```bash
lockbox team add
# Choose to add from:
# 1. Your personal keys
# 2. A public key file
```

Remove a team member:
```bash
lockbox team remove
# Select team member to remove from the list
```

List team members:
```bash
lockbox team list
```

### Encrypting and Decrypting Secrets

Encrypt a file:
```bash
lockbox secret encrypt
# Enter file path and confirm encryption
# File will be encrypted for all team members
```

Decrypt a file:
```bash
lockbox secret decrypt
# Select which personal key to use for decryption
```

## Key Management

Lockbox uses two locations for key storage:
- `~/.lockbox/`: Stores your personal private/public key pairs
- `./.lockbox/`: Stores team members' public keys for the current repository

This means:
1. Your private keys are safely stored in your home directory
2. You can use the same keys across multiple repositories
3. Each repository maintains its own team member list

## Development

### Prerequisites

- Go 1.21 or later
- Make

### Setting Up Development Environment

1. Clone the repository:
```bash
git clone https://github.com/yourusername/lockbox.git
cd lockbox
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
```

The binary will be available in `bin/lockbox`.

### Development Commands

```bash
# Build the project
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean

# Build for all platforms
make build-all
```

### Project Structure

```
lockbox/
├── cmd/                    # Application entrypoints
│   └── lockbox/           # Main CLI application
├── internal/              # Private application code
│   ├── crypto/           # Encryption operations
│   ├── git/              # Git utilities
│   ├── output/           # Colored output formatting
│   ├── prompt/           # Interactive prompts
│   └── commands/         # CLI commands
├── .github/              # GitHub Actions workflows
└── Formula/              # Homebrew formula
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
