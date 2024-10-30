from typing import Optional, List, Tuple, Any
import click
import gnupg
from pathlib import Path

from lockbox.utils.click import MutuallyExclusiveOption


def get_lockbox_gpg() -> gnupg.GPG:
    """Get GPG instance using .lockbox as homedir"""
    current: Path = Path.cwd()
    while current != current.parent:
        if (current / ".git").exists():
            lockbox_dir: Path = current / ".lockbox"
            if not lockbox_dir.exists():
                raise click.ClickException(
                    "Lockbox directory not initialized. Run 'lockbox init' first."
                )
            return gnupg.GPG(homedir=str(lockbox_dir))
        current = current.parent
    raise click.ClickException("Not in a git repository")


def get_user_gpg() -> gnupg.GPG:
    """Get GPG instance using user's default GPG directory"""
    return gnupg.GPG()


@click.group()
def team() -> None:
    """Manage team members' GPG keys"""
    pass


@team.command()
@click.option(
    "--me",
    is_flag=True,
    cls=MutuallyExclusiveOption,
    help="Add your own public key to the team",
    mutually_exclusive=["file", "id", "fingerprint"],
)
@click.option(
    "--file",
    type=click.Path(exists=True),
    cls=MutuallyExclusiveOption,
    help="Import public key from a file",
    mutually_exclusive=["me", "id", "fingerprint"],
)
@click.option(
    "--id",
    cls=MutuallyExclusiveOption,
    help="Import public key by key ID",
    mutually_exclusive=["me", "file", "fingerprint"],
)
@click.option(
    "--fingerprint",
    cls=MutuallyExclusiveOption,
    help="Import public key by fingerprint",
    mutually_exclusive=["me", "file", "id"],
)
def add(
    me: bool, file: Optional[str], id: Optional[str], fingerprint: Optional[str]
) -> None:
    """Add a team member's public key.

    If no options are provided, enters interactive mode to search and select a key.
    """
    user_gpg: gnupg.GPG = get_user_gpg()
    lockbox_gpg: gnupg.GPG = get_lockbox_gpg()

    # Handle --me flag
    if me:
        secret_keys: list[dict] = user_gpg.list_keys(secret=True)
        if not secret_keys:
            raise click.ClickException("No private keys found in your keyring")

        key_data = secret_keys[0]
        key_id = key_data["keyid"]

    # Handle --file option
    elif file:
        with open(file, "r") as f:
            key_data = f.read()
        import_result: Any = lockbox_gpg.import_keys(key_data)
        if import_result.count == 0:
            raise click.ClickException("Failed to import key from file")
        click.echo("Successfully added key from file")
        return

    # Handle --id option
    elif id:
        key_id = id

    # Handle --fingerprint option
    elif fingerprint:
        key_id = fingerprint

    # Interactive mode
    else:
        public_keys: List[dict] = user_gpg.list_keys()
        if not public_keys:
            raise click.ClickException("No public keys found in your keyring")

        search: str = click.prompt(
            "Enter name or email to search (or press Enter to list all)"
        )

        matching_keys: list[dict] = []
        for key in public_keys:
            uids: List[str] = key.get("uids", [])
            if not search:  # If no search term, include all keys
                matching_keys.append(key)
            else:  # Otherwise, search in UIDs
                for uid in uids:
                    if search.lower() in uid.lower():
                        matching_keys.append(key)
                        break

        if not matching_keys:
            raise click.ClickException("No matching keys found")

        if len(matching_keys) == 1:
            key_data = matching_keys[0]
        else:
            click.echo("\nMatching keys:")
            for idx, key in enumerate(matching_keys):
                uid: str = key.get("uids", ["<no uid>"])[0]
                click.echo(f"{idx + 1}. {uid} ({key['keyid']})")

            choice: int = click.prompt(
                "\nEnter the number of the key to add",
                type=click.IntRange(1, len(matching_keys)),
            )
            key_data = matching_keys[choice - 1]

        key_id = key_data["keyid"]

    # Export and import the selected key
    if not file:  # file was handled earlier
        public_key: str = user_gpg.export_keys(key_id)
        if not public_key:
            raise click.ClickException(f"Failed to export key {key_id}")

        import_result = lockbox_gpg.import_keys(public_key)
        if import_result.count == 0:
            raise click.ClickException("Failed to import key")

        # Get the UID for the success message
        imported_keys: List[dict] = lockbox_gpg.list_keys()
        for key in imported_keys:
            if key["keyid"] == key_id:
                uid = key.get("uids", ["<no uid>"])[0]
                click.echo(f"Successfully added key for {uid}")
                return

        click.echo(f"Successfully added key {key_id}")


@team.command()
@click.option(
    "--me",
    is_flag=True,
    cls=MutuallyExclusiveOption,
    help="Remove your own public key from the team",
    mutually_exclusive=["id", "fingerprint"],
)
@click.option(
    "--id",
    cls=MutuallyExclusiveOption,
    help="Remove public key by key ID",
    mutually_exclusive=["me", "fingerprint"],
)
@click.option(
    "--fingerprint",
    cls=MutuallyExclusiveOption,
    help="Remove public key by fingerprint",
    mutually_exclusive=["me", "id"],
)
def remove(me: bool, id: Optional[str], fingerprint: Optional[str]) -> None:
    """Remove a team member's public key.

    If no options are provided, enters interactive mode to select a key to remove.
    """
    lockbox_gpg: gnupg.GPG = get_lockbox_gpg()
    user_gpg: gnupg.GPG = get_user_gpg()
    team_keys: List[dict] = lockbox_gpg.list_keys()

    # Handle --me flag
    if me:
        user_keys: List[dict] = user_gpg.list_keys()
        if not user_keys:
            raise click.ClickException("No public keys found in your keyring")

        matching_keys: List[dict] = []
        for team_key in team_keys:
            for user_key in user_keys:
                if team_key["fingerprint"] == user_key["fingerprint"]:
                    matching_keys.append(team_key)

        if not matching_keys:
            raise click.ClickException("None of your keys were found in the team")

        key_to_remove: dict = matching_keys[0]
        if len(matching_keys) > 1:
            click.echo("Multiple keys found. Please select one to remove:")
            for idx, key in enumerate(matching_keys):
                uid = key.get("uids", ["<no uid>"])[0]
                click.echo(f"{idx + 1}. {uid} ({key['keyid']})")

            choice: int = click.prompt(
                "Enter the number of the key to remove",
                type=click.IntRange(1, len(matching_keys)),
            )
            key_to_remove = matching_keys[choice - 1]

        key_id: str = key_to_remove["keyid"]

    # Handle --id option
    elif id:
        key_id = id

    # Handle --fingerprint option
    elif fingerprint:
        key_id = fingerprint

    # Interactive mode
    else:
        if not team_keys:
            raise click.ClickException("No team members found")

        # Create list of choices for the interactive prompt
        choices: List[Tuple[str, dict]] = []
        for key in team_keys:
            uid = key.get("uids", ["<no uid>"])[0]
            display: str = f"{uid} ({key['keyid']})"
            choices.append((display, key))

        click.echo("\nSelect a team member to remove:")
        for idx, (display, _) in enumerate(choices, 1):
            click.echo(f"{idx}. {display}")

        choice = click.prompt(
            "\nEnter the number of the key to remove",
            type=click.IntRange(1, len(choices)),
        )

        _, key_to_remove = choices[choice - 1]
        key_id = key_to_remove["keyid"]

    # Perform the removal
    result: Any = lockbox_gpg.delete_keys(key_id)
    if result.status != "ok":
        raise click.ClickException(f"Failed to remove key {key_id}")

    # Get the UID for the success message
    uid: Optional[str] = None
    for key in team_keys:
        if key["keyid"] == key_id:
            uid = key.get("uids", ["<no uid>"])[0]
            break

    if uid:
        click.echo(f"Successfully removed key for {uid}")
    else:
        click.echo(f"Successfully removed key {key_id}")


@team.command(name="list")
def list_keys() -> None:
    """List all team members' public keys."""
    lockbox_gpg: gnupg.GPG = get_lockbox_gpg()

    public_keys: List[dict] = lockbox_gpg.list_keys()
    if not public_keys:
        click.echo("No team members found")
        return

    click.echo("Team members:")
    for key in public_keys:
        uids: List[str] = key.get("uids", [])
        uid_str: str = uids[0] if uids else "No UID"
        click.echo(f"- {key['keyid']}: {uid_str}")
