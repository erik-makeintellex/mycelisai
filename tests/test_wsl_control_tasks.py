from __future__ import annotations

from invoke import Context

from ops import wsl_control


def test_wsl_down_terminates_configured_distro(monkeypatch, capsys):
    calls: list[list[str]] = []
    monkeypatch.setattr(wsl_control, "_wsl_available", lambda: True)
    monkeypatch.setenv("MYCELIS_WSL_PROOF_DISTRO", "proof-root")

    def fake_run(command, capture_output=True, text=True, check=False):
        calls.append(command)
        return wsl_control.subprocess.CompletedProcess(command, 0, "", "")

    monkeypatch.setattr(wsl_control.subprocess, "run", fake_run)

    wsl_control.down.body(Context())

    assert calls == [["wsl.exe", "--terminate", "proof-root"]]
    assert "WSL distro stopped: proof-root" in capsys.readouterr().out
