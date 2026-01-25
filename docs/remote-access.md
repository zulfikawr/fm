# Remote Access (SSH/SFTP)

`fm` allows you to manage files on remote servers as easily as local ones using SSH/SFTP.

## Connecting to a Remote Server

There are two primary ways to initiate a remote connection:

### 1. Command Line Flags

You can start `fm` directly into a remote session:

```bash
# Using a host defined in ~/.ssh/config
fm -r my-server

# Using user@host
fm --remote user@192.168.1.50

# Using a specific identity file
fm -r user@hostname /path/to/id_rsa
```

### 2. In-App "Go to" Command

While inside `fm`, press `g` to open the "Go to" prompt. You will be asked to choose between **Local** (`l`) or **Remote** (`r`). If you choose Remote, you can enter an address like `user@host` or a pre-defined SSH alias. `fm` also supports **Tab autocompletion** for paths and filenames when navigating or specifying PEM keys.

## Authentication Methods

`fm` supports several authentication methods:
- **SSH Agent:** If you have an SSH agent running, `fm` will use it automatically.
- **Identity Files:** Standard keys like `~/.ssh/id_rsa`, `~/.ssh/id_ed25519`, etc., are checked by default.
- **Password Auth:** If key-based auth fails, `fm` will interactively prompt you for a password.
- **Custom Keys:** You can specify a `.pem` or private key file via the CLI, or enter its path when prompted in the TUI. Autocompletion is available for local PEM paths.

## Security Features

- **Host Key Verification:** When connecting to a new server, `fm` will show you the host's fingerprint and ask for confirmation before adding it to your `known_hosts` file.
- **Encrypted Transfers:** All data transferred via SFTP is encrypted using standard SSH protocols.

## Remote Workflow

Once connected, you can perform most standard operations (copy, move, delete, create) on the remote server just like you would locally. Note that some features like Git integration may have limited functionality on remote filesystems depending on the server configuration.
