from __future__ import annotations

from pathlib import Path

CHART_DIR = Path(__file__).resolve().parents[1] / "charts" / "mycelis-core"


def _extract_top_level_block(yaml_text: str, key: str) -> dict[str, str]:
    lines = yaml_text.splitlines()
    in_block = False
    block: dict[str, str] = {}
    for line in lines:
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if not line.startswith(" "):
            if stripped == f"{key}:":
                in_block = True
                continue
            if in_block:
                break
        elif in_block and ":" in stripped:
            k, v = stripped.split(":", 1)
            block[k.strip()] = v.strip().strip('"').strip("'")
    return block


def test_values_yaml_defines_oidc_block():
    values_path = CHART_DIR / "values.yaml"
    oidc = _extract_top_level_block(values_path.read_text(encoding="utf-8"), "oidc")

    assert oidc == {
        "issuer": "",
        "audience": "",
        "roleClaim": "role",
        "defaultRole": "viewer",
        "existingSecret": "",
        "hs256SecretKey": "mycelis-oidc-hs256-secret",
    }


def test_values_enterprise_yaml_defines_oidc_block():
    values_path = CHART_DIR / "values-enterprise.yaml"
    oidc = _extract_top_level_block(values_path.read_text(encoding="utf-8"), "oidc")

    assert oidc == {
        "issuer": "",
        "audience": "mycelis-core",
        "roleClaim": "role",
        "defaultRole": "viewer",
        "existingSecret": "",
        "hs256SecretKey": "mycelis-oidc-hs256-secret",
    }


def test_deployment_template_wires_oidc_env_vars_and_secret_ref():
    deployment_path = CHART_DIR / "templates" / "deployment.yaml"
    deployment_text = deployment_path.read_text(encoding="utf-8")

    assert "{{- if and .Values.oidc (or .Values.oidc.issuer .Values.oidc.existingSecret) }}" in deployment_text
    assert "- name: MYCELIS_OIDC_ISSUER" in deployment_text
    assert "value: {{ .Values.oidc.issuer | quote }}" in deployment_text
    assert "- name: MYCELIS_OIDC_AUDIENCE" in deployment_text
    assert 'value: {{ default "" .Values.oidc.audience | quote }}' in deployment_text
    assert "- name: MYCELIS_OIDC_ROLE_CLAIM" in deployment_text
    assert 'value: {{ default "role" .Values.oidc.roleClaim | quote }}' in deployment_text
    assert "- name: MYCELIS_OIDC_DEFAULT_ROLE" in deployment_text
    assert 'value: {{ default "viewer" .Values.oidc.defaultRole | quote }}' in deployment_text
    assert "{{- if .Values.oidc.existingSecret }}" in deployment_text
    assert "- name: MYCELIS_OIDC_HS256_SECRET" in deployment_text
    assert "name: {{ .Values.oidc.existingSecret }}" in deployment_text
    assert 'key: {{ default "mycelis-oidc-hs256-secret" .Values.oidc.hs256SecretKey }}' in deployment_text
