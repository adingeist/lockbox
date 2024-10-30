import click
import tomllib
from pathlib import Path


def get_version():
    pyproject_path = Path(__file__).parent.parent.parent / "pyproject.toml"
    with open(pyproject_path, "rb") as f:
        data = tomllib.load(f)
    return data["tool"]["poetry"]["version"]


@click.group()
@click.version_option(version=get_version(), prog_name="Lockbox")
def lockbox():
    """Command line interface for managing the lockbox."""
    pass


@lockbox.command()
def init():
    """Initialize a .gpg directory in the git project root."""
    # Find git root by walking up directories looking for .git
    current = Path.cwd()
    while current != current.parent:
        if (current / ".git").exists():
            gpg_dir = current / ".gpg"
            if not gpg_dir.exists():
                gpg_dir.mkdir()
                click.echo(f"Created GPG directory at {gpg_dir}")
            else:
                click.echo(f"GPG directory already exists at {gpg_dir}")
            return
        current = current.parent

    click.echo("Error: Not in a git repository", err=True)
    raise click.Abort()


if __name__ == "__main__":
    lockbox()
