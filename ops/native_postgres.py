"""PostgreSQL bootstrap helpers for native source-mode development."""

from __future__ import annotations

import os
import socket
import subprocess

from .config import ROOT_DIR


def load_env():
    try:
        from dotenv import load_dotenv
    except ModuleNotFoundError as exc:
        if exc.name != "dotenv":
            raise
        raise SystemExit("Missing python-dotenv. Run tasks with 'uv run inv ...'.") from exc
    load_dotenv(str(ROOT_DIR / ".env"), override=True)


def env(name: str, default: str) -> str:
    load_env()
    return os.environ.get(name, default)


def port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def psql_command(sql: str, *, database: str, user: str, password: str) -> subprocess.CompletedProcess[str]:
    host = env("DB_HOST", "127.0.0.1")
    port = env("DB_PORT", "5432")
    proc_env = {**os.environ, "PGPASSWORD": password}
    return subprocess.run(
        ["psql", "-v", "ON_ERROR_STOP=1", "-h", host, "-p", port, "-U", user, "-d", database, "-c", sql],
        capture_output=True,
        text=True,
        env=proc_env,
    )


def emit_psql_result(result: subprocess.CompletedProcess[str]):
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        for line in result.stderr.splitlines():
            if "NOTICE:" not in line:
                print(line)


def _sql_literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def _sql_identifier(value: str) -> str:
    if not value.replace("_", "").isalnum():
        raise SystemExit(f"Unsupported PostgreSQL identifier: {value!r}")
    return '"' + value.replace('"', '""') + '"'


def bootstrap_database() -> bool:
    """Create/update the app database and role through the local superuser."""
    load_env()
    host = os.environ.get("DB_HOST", "127.0.0.1")
    port = int(os.environ.get("DB_PORT", "5432"))
    if not port_open(port, host=host):
        print(f"  PostgreSQL is not reachable at {host}:{port}")
        return False

    super_user = os.environ.get("POSTGRES_USER", "postgres")
    super_password = os.environ.get("POSTGRES_PASSWORD", "")
    app_user = os.environ.get("DB_USER", "mycelis")
    app_password = os.environ.get("DB_PASSWORD", "")
    app_db = os.environ.get("DB_NAME", "cortex")
    if not super_password:
        raise SystemExit("POSTGRES_PASSWORD is required in .env for native PostgreSQL bootstrap.")
    if not app_password:
        raise SystemExit("DB_PASSWORD is required in .env for native PostgreSQL bootstrap.")

    app_user_lit = _sql_literal(app_user)
    app_user_ident = _sql_identifier(app_user)
    app_password_lit = _sql_literal(app_password)
    app_db_lit = _sql_literal(app_db)
    app_db_ident = _sql_identifier(app_db)
    role_sql = (
        "DO $$ BEGIN "
        f"IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = {app_user_lit}) THEN "
        f"CREATE ROLE {app_user_ident} LOGIN CREATEDB PASSWORD {app_password_lit}; "
        f"ELSE ALTER ROLE {app_user_ident} WITH LOGIN CREATEDB PASSWORD {app_password_lit}; "
        "END IF; END $$;"
    )
    result = psql_command(role_sql, database="postgres", user=super_user, password=super_password)
    if result.returncode != 0:
        emit_psql_result(result)
        return False

    check = psql_command(
        f"SELECT 1 FROM pg_database WHERE datname = {app_db_lit};",
        database="postgres",
        user=super_user,
        password=super_password,
    )
    if check.returncode != 0:
        emit_psql_result(check)
        return False
    if "1" not in check.stdout.split():
        create = psql_command(
            f"CREATE DATABASE {app_db_ident} OWNER {app_user_ident};",
            database="postgres",
            user=super_user,
            password=super_password,
        )
        if create.returncode != 0:
            emit_psql_result(create)
            return False
    else:
        owner = psql_command(
            f"ALTER DATABASE {app_db_ident} OWNER TO {app_user_ident};",
            database="postgres",
            user=super_user,
            password=super_password,
        )
        if owner.returncode != 0:
            emit_psql_result(owner)
            return False

    grant = psql_command(
        f"GRANT ALL PRIVILEGES ON DATABASE {app_db_ident} TO {app_user_ident};",
        database="postgres",
        user=super_user,
        password=super_password,
    )
    if grant.returncode != 0:
        emit_psql_result(grant)
        return False
    if not ensure_extensions(app_db):
        return False
    return True


def ensure_extensions(dbname: str | None = None) -> bool:
    """Create extensions that need bootstrap privileges before app migrations."""
    load_env()
    super_user = os.environ.get("POSTGRES_USER", "postgres")
    super_password = os.environ.get("POSTGRES_PASSWORD", "")
    target_db = dbname or os.environ.get("DB_NAME", "cortex")
    if not super_password:
        return True
    result = psql_command(
        "CREATE EXTENSION IF NOT EXISTS vector; CREATE EXTENSION IF NOT EXISTS pg_trgm;",
        database=target_db,
        user=super_user,
        password=super_password,
    )
    if result.returncode != 0:
        emit_psql_result(result)
        return False
    return True
