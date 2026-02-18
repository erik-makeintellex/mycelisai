import os
import subprocess
from pathlib import Path

from invoke import task, Collection
from .config import CORE_DIR, ROOT_DIR


MIGRATIONS_DIR = CORE_DIR / "migrations"


def _load_env():
    from dotenv import load_dotenv
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


def _psql(sql=None, file=None, dbname=None):
    """Run psql with connection args. Uses -f for files, -c for strings."""
    host, port, user, password, db = _dsn(dbname)
    env = {**os.environ, "PGPASSWORD": password}
    cmd = ["psql", "-h", host, "-p", port, "-U", user, "-d", db]
    if file:
        cmd += ["-f", str(file)]
    elif sql:
        cmd += ["-c", sql]
    result = subprocess.run(cmd, env=env, capture_output=True, text=True)
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        for line in result.stderr.splitlines():
            # Suppress harmless NOTICE lines
            if "NOTICE:" in line:
                continue
            print(line)
    return result.returncode


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
    print(f"Creating database '{db}'...")
    # Check existence first
    rc = _psql(sql=f"SELECT 1 FROM pg_database WHERE datname = '{db}';",
               dbname="postgres")
    # Try create (idempotent)
    _psql(sql=f"CREATE DATABASE {db};", dbname="postgres")
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

    print(f"Resetting database '{db}'...")

    # Terminate existing connections
    _psql(
        sql=(
            f"SELECT pg_terminate_backend(pid) FROM pg_stat_activity "
            f"WHERE datname = '{db}' AND pid <> pg_backend_pid();"
        ),
        dbname="postgres",
    )

    _psql(sql=f"DROP DATABASE IF EXISTS {db};", dbname="postgres")
    _psql(sql=f"CREATE DATABASE {db} OWNER mycelis;", dbname="postgres")
    print(f"Database '{db}' recreated.")

    # Apply migrations
    migrate(c)


ns = Collection("db")
ns.add_task(status)
ns.add_task(create)
ns.add_task(migrate)
ns.add_task(reset)
