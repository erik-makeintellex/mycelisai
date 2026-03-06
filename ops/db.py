import os
import subprocess
from pathlib import Path

from invoke import task, Collection
from .config import CORE_DIR, ROOT_DIR


MIGRATIONS_DIR = CORE_DIR / "migrations"


def _load_env():
    try:
        from dotenv import load_dotenv
    except ModuleNotFoundError as exc:
        if exc.name != "dotenv":
            raise
        raise SystemExit(
            "Missing python-dotenv in the current invoke environment. "
            "Run tasks with 'uv run inv ...' or '.\\.venv\\Scripts\\inv.exe ...'; "
            "do not use 'uvx --from invoke inv ...'."
        ) from exc
    load_dotenv(str(ROOT_DIR / ".env"))


def _dsn(dbname=None):
    """Build psql connection args from .env."""
    _load_env()
    host = os.getenv("DB_HOST", "127.0.0.1")
    port = os.getenv("DB_PORT", "5432")
    user = os.getenv("DB_USER", "mycelis")
    password = os.getenv("DB_PASSWORD", "password")
    db = dbname or os.getenv("DB_NAME", "cortex")
    return host, port, user, password, db


def _run_psql(sql=None, file=None, dbname=None):
    """Run psql and return the completed process."""
    host, port, user, password, db = _dsn(dbname)
    env = {**os.environ, "PGPASSWORD": password}
    cmd = ["psql", "-h", host, "-p", port, "-U", user, "-d", db]
    if file:
        cmd += ["-f", str(file)]
    elif sql:
        cmd += ["-c", sql]
    return subprocess.run(cmd, env=env, capture_output=True, text=True)


def _emit_psql_output(result):
    """Print psql output while suppressing NOTICE noise."""
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        for line in result.stderr.splitlines():
            # Suppress harmless NOTICE lines
            if "NOTICE:" in line:
                continue
            print(line)


def _psql(sql=None, file=None, dbname=None):
    """Run psql with connection args. Uses -f for files, -c for strings."""
    result = _run_psql(sql=sql, file=file, dbname=dbname)
    _emit_psql_output(result)
    return result.returncode


def _require_postgres(dbname="postgres"):
    """Fail fast when the local PostgreSQL bridge is unavailable."""
    host, port, _user, _password, _db = _dsn(dbname)
    rc = _psql(sql="SELECT 1;", dbname=dbname)
    if rc != 0:
        raise SystemExit(
            f"Cannot connect to PostgreSQL at {host}:{port}. "
            "Start the bridge with 'uv run inv k8s.bridge' or use "
            "'uv run inv lifecycle.memory-restart', which now restores the bridge before reset."
        )


@task
def status(c):
    """Show database tables and row counts."""
    sql = (
        "SELECT tablename FROM pg_tables WHERE schemaname = 'public' "
        "ORDER BY tablename;"
    )
    rc = _psql(sql=sql)
    if rc != 0:
        print("Cannot connect. Is the bridge running? (inv k8s.bridge)")
        raise SystemExit(1)


@task
def create(c):
    """Create the cortex database if it doesn't exist."""
    _load_env()
    db = os.getenv("DB_NAME", "cortex")
    _require_postgres()
    print(f"Creating database '{db}'...")
    result = _run_psql(
        sql=f"SELECT 1 FROM pg_database WHERE datname = '{db}';",
        dbname="postgres",
    )
    _emit_psql_output(result)
    if result.returncode != 0:
        raise SystemExit(f"Failed checking whether database '{db}' exists.")
    if "1" not in result.stdout.split():
        rc = _psql(sql=f"CREATE DATABASE {db};", dbname="postgres")
        if rc != 0:
            raise SystemExit(f"Failed creating database '{db}'.")
    print(f"Database '{db}' ready.")


@task
def migrate(c):
    """Apply all migrations in order to the cortex database."""
    _load_env()
    db = os.getenv("DB_NAME", "cortex")

    # Ensure DB exists
    create(c)

    files = sorted(MIGRATIONS_DIR.glob("*.sql"))
    if not files:
        print("No migration files found.")
        return

    print(f"Applying {len(files)} migrations to '{db}'...")
    failures = []
    for f in files:
        print(f"  {f.name}...", end=" ")
        rc = _psql(file=f)
        if rc == 0:
            print("OK")
        else:
            print("ERRORS (non-fatal, continuing)")
            failures.append(f.name)

    if failures:
        print(f"\nMigrations with errors: {', '.join(failures)}")
        print("(Some errors are expected on re-runs due to IF NOT EXISTS guards)")
    else:
        print("\nAll migrations applied cleanly.")


@task
def reset(c):
    """Drop and recreate cortex database, then run all migrations."""
    _load_env()
    db = os.getenv("DB_NAME", "cortex")
    _require_postgres()

    print(f"Resetting database '{db}'...")

    # Terminate existing connections
    rc = _psql(
        sql=(
            f"SELECT pg_terminate_backend(pid) FROM pg_stat_activity "
            f"WHERE datname = '{db}' AND pid <> pg_backend_pid();"
        ),
        dbname="postgres",
    )
    if rc != 0:
        raise SystemExit(f"Failed terminating existing connections to '{db}'.")

    rc = _psql(sql=f"DROP DATABASE IF EXISTS {db};", dbname="postgres")
    if rc != 0:
        raise SystemExit(f"Failed dropping database '{db}'.")

    rc = _psql(sql=f"CREATE DATABASE {db} OWNER mycelis;", dbname="postgres")
    if rc != 0:
        raise SystemExit(f"Failed creating database '{db}'.")
    print(f"Database '{db}' recreated.")

    # Apply migrations
    migrate(c)


ns = Collection("db")
ns.add_task(status)
ns.add_task(create)
ns.add_task(migrate)
ns.add_task(reset)
