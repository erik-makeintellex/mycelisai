import os

from invoke import task

from .db import _load_env, _psql, _require_postgres, _run_psql


RUNTIME_CONTEXT_TABLES = [
    "team_interactions",
    "team_status_events",
    "team_work_items",
    "proof_artifacts",
    "execution_contracts",
    "confirm_tokens",
    "intent_proofs",
    "mission_events",
    "mission_runs",
    "conversation_turns",
    "temp_memory_channels",
    "agent_state",
]


def _table_exists(table):
    result = _run_psql(sql=f"SELECT to_regclass('public.{table}');")
    if result.returncode != 0:
        return False
    return table in result.stdout


def _table_count(table):
    result = _run_psql(sql=f"SELECT COUNT(*) FROM {table};")
    if result.returncode != 0:
        return None
    for token in result.stdout.split():
        if token.isdigit():
            return int(token)
    return None


@task(
    help={
        "yes": "Actually clear rows. Without this flag, only prints target row counts.",
        "include_memory_vectors": "Also clear context_vectors long-memory embeddings. Default keeps retained memory vectors.",
    }
)
def clear_runtime_context(c, yes=False, include_memory_vectors=False):
    """
    Clear volatile Soma/team runtime context before fresh UI or live-runtime proof.

    This keeps schema, auth, capability manifests, providers, and deployment
    configuration intact. It intentionally avoids long-memory vectors unless
    --include-memory-vectors is supplied.
    """
    _load_env()
    _require_postgres(dbname=os.getenv("DB_NAME", "cortex"))
    tables = list(RUNTIME_CONTEXT_TABLES)
    if include_memory_vectors:
        tables.append("context_vectors")

    existing = [table for table in tables if _table_exists(table)]
    if not existing:
        print("No runtime context tables found.")
        return

    print("Runtime context target counts:")
    for table in existing:
        count = _table_count(table)
        print(f"  {table}: {count if count is not None else 'unknown'}")

    if not yes:
        print("\nDry run only. Re-run with --yes to clear volatile Soma/team context.")
        if not include_memory_vectors:
            print("Retained context_vectors are kept. Add --include-memory-vectors for a full memory reset.")
        return

    print("\nClearing volatile Soma/team runtime context...")
    for table in existing:
        rc = _psql(sql=f"DELETE FROM {table};")
        if rc != 0:
            raise SystemExit(f"Failed clearing {table}.")
    print("Runtime context cleared.")
