# Lockbox

Lockbox is a command-line tool for managing team GPG keys in Git repositories. It provides a secure way to share and manage GPG public keys within a team, storing them in a `.lockbox` directory at the root of your Git repository.

## Features

- Easy team member key management
- Interactive key selection
- Secure key storage in Git repository
- Support for multiple GPG keys per user
- Cross-platform support (Linux, macOS, Windows)

## Installation

### Using Homebrew (macOS)

```bash
brew install yourusername/tap/lockbox
```

### From Binary Releases

Download the appropriate binary for your platform from the [releases page](https://github.com/yourusername/lockbox/releases).

## Usage

### Initialize Lockbox

Create a `.lockbox` directory in your Git repository:

```bash
lockbox init
```

### Team Management Commands

#### Add a Team Member

```bash
# Interactive mode
lockbox team add

# Add your own public key
lockbox team add --me

# Add from a file
lockbox team add --file path/to/key.asc

# Add by key ID
lockbox team add --id ABC123

# Add by fingerprint
lockbox team add --fingerprint ABC123...
```

#### Remove a Team Member

```bash
# Interactive mode
lockbox team remove

# Remove your own key
lockbox team remove --me

# Remove by key ID
lockbox team remove --id ABC123

# Remove by fingerprint
lockbox team remove --fingerprint ABC123...
```

#### List Team Members

```bash
lockbox team list
```

## Development

### Prerequisites

- Go 1.21 or later
- `libgpgme-dev` (Linux) or `gpgme` (macOS)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/lockbox.git
cd lockbox

# Build
make build

# Install locally
make install
```

### Running Tests

```bash
make test
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
