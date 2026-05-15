from ops import interface_env, interface_workspace


def test_build_playwright_env_infers_native_core_workspace_root(monkeypatch, tmp_path):
    monkeypatch.setattr(interface_workspace, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(interface_workspace, "repo_local_core_listener_active", lambda: True)
    monkeypatch.setattr(interface_env, "_task_env", lambda extra=None: dict(extra or {}))
    monkeypatch.setenv("MYCELIS_WORKSPACE", "./workspace")
    monkeypatch.delenv("MYCELIS_BACKEND_WORKSPACE_ROOT", raising=False)
    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", raising=False)
    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_PROBE", raising=False)

    env = interface_env._build_playwright_env(live_backend=True, port=4311)

    assert env["PLAYWRIGHT_LIVE_BACKEND"] == "1"
    assert env["MYCELIS_BACKEND_WORKSPACE_ROOT"] == str((tmp_path / "core" / "workspace").resolve())


def test_build_playwright_env_preserves_explicit_and_k8s_workspace_roots(monkeypatch, tmp_path):
    monkeypatch.setattr(interface_workspace, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(interface_workspace, "repo_local_core_listener_active", lambda: True)
    monkeypatch.setattr(interface_env, "_task_env", lambda extra=None: dict(extra or {}))
    monkeypatch.setenv("MYCELIS_WORKSPACE", "./workspace")
    monkeypatch.setenv("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", str(tmp_path / "explicit-workspace"))
    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_PROBE", raising=False)

    explicit_env = interface_env._build_playwright_env(live_backend=True, port=4311)

    assert explicit_env["PLAYWRIGHT_BACKEND_WORKSPACE_ROOT"] == str(tmp_path / "explicit-workspace")
    assert "MYCELIS_BACKEND_WORKSPACE_ROOT" not in explicit_env

    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", raising=False)
    monkeypatch.setenv("PLAYWRIGHT_BACKEND_WORKSPACE_PROBE", "k8s")

    k8s_env = interface_env._build_playwright_env(live_backend=True, port=4311)

    assert "MYCELIS_BACKEND_WORKSPACE_ROOT" not in k8s_env


def test_build_playwright_env_leaves_workspace_unset_without_repo_local_core(monkeypatch, tmp_path):
    monkeypatch.setattr(interface_workspace, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(interface_workspace, "repo_local_core_listener_active", lambda: False)
    monkeypatch.setattr(interface_env, "_task_env", lambda extra=None: dict(extra or {}))
    monkeypatch.setenv("MYCELIS_WORKSPACE", "./workspace")
    monkeypatch.delenv("MYCELIS_BACKEND_WORKSPACE_ROOT", raising=False)
    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", raising=False)
    monkeypatch.delenv("PLAYWRIGHT_BACKEND_WORKSPACE_PROBE", raising=False)

    env = interface_env._build_playwright_env(live_backend=True, port=4311)

    assert "MYCELIS_BACKEND_WORKSPACE_ROOT" not in env
