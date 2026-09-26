from __future__ import annotations

import pytest

from ops import lifecycle_first_boot_state as state


def fake_psql(rows: dict[str, str]):
    def psql(*args: str) -> str:
        sql = args[-1]
        return rows[sql]

    return psql


def test_count_tables_parses_pipe_rows():
    psql = fake_psql({
        "SELECT 'groups', COUNT(*)::bigint FROM groups ORDER BY 1;": "groups|0",
    })

    counts = state.count_tables(psql, ("groups",))

    assert counts == {"groups": 0}


def test_count_tables_rejects_unsafe_table_names():
    with pytest.raises(SystemExit, match="Invalid first-boot table-count contract"):
        state.count_tables(fake_psql({}), ("groups; drop table x",))


def test_count_tables_requires_every_name_to_come_back():
    psql = fake_psql({
        "SELECT 'groups', COUNT(*)::bigint FROM groups ORDER BY 1;": "",
    })

    with pytest.raises(SystemExit, match="did not receive every table count"):
        state.count_tables(psql, ("groups",))


def test_assert_empty_user_state_reports_dirty_tables(monkeypatch):
    monkeypatch.setattr(state, "invocation_tables", lambda _psql: ())
    monkeypatch.setattr(state, "count_tables", lambda _psql, _tables: {"groups": 2})

    with pytest.raises(SystemExit, match=r"created user state: \{'groups': 2\}"):
        state.assert_empty_user_state(lambda *_a: "", ("groups",), "first boot")


def test_assert_empty_user_state_passes_when_clean(monkeypatch):
    monkeypatch.setattr(state, "invocation_tables", lambda _psql: ())
    monkeypatch.setattr(state, "count_tables", lambda _psql, _tables: {"groups": 0})

    state.assert_empty_user_state(lambda *_a: "", ("groups",), "first boot")


def test_assert_schema_compatible_fails_on_mismatch(monkeypatch):
    monkeypatch.setattr(state, "SCHEMA_COMPATIBILITY_CHECKS", (("widgets table", "SELECT 1"),))

    with pytest.raises(SystemExit, match="widgets table"):
        state.assert_schema_compatible(lambda *_a: "0")


def test_assert_nats_empty_flags_retained_streams(monkeypatch):
    class FakeResponse:
        status = 200

        def __enter__(self):
            return self

        def __exit__(self, *_exc):
            return False

        def read(self):
            return b'{"streams": 1, "messages": 4}'

    monkeypatch.setattr(state.urllib.request, "urlopen", lambda *_a, **_k: FakeResponse())
    monkeypatch.setattr(state.json, "load", lambda _fp: {"streams": 1, "messages": 4})

    with pytest.raises(SystemExit, match="retained streams or messages"):
        state.assert_nats_empty(8222)
