# Lockbox

Lockbox is a command-line tool for securely sharing secrets within a team. It uses modern encryption (age) to ensure only team members can decrypt shared secrets.

## Features

- Built-in encryption (no external GPG required)
- Team member key management
- Secure secret sharing
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

### Set Up Your Identity

Before you can encrypt or decrypt secrets, you need to set up your identity:

```bash
lockbox team init --name "Your Name"
```

This will:
- Generate your key pair
- Store your private key locally
- Add your public key to the team

### Team Management

Add a team member:
```bash
lockbox team add --name "Team Member" --key "age1..."
```

Remove a team member:
```bash
lockbox team remove --key "age1..."
```

List team members:
```bash
lockbox team list
```

### Encrypting and Decrypting Secrets

(Coming soon)

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
│   └── commands/         # CLI commands
├── .github/              # GitHub Actions workflows
└── Formula/              # Homebrew formula
```

### Making Changes

1. Create a new branch:
```bash
git checkout -b feature/your-feature
```

2. Make your changes and test:
```bash
make build
./bin/lockbox --help
```

3. Run tests and linter:
```bash
make test
make lint
```

4. Commit and push:
```bash
git commit -m "Add your feature"
git push origin feature/your-feature
```

### Release Process

1. Tag a new version:
```bash
git tag v0.1.0
git push origin v0.1.0
```

2. GitHub Actions will automatically:
- Run tests
- Build binaries for all platforms
- Create a GitHub release
- Update the Homebrew formula

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
