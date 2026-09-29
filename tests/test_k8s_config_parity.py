from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]

BOOT_BUNDLES = ("mycelis-runtime-core.yaml", "mycelis-dev-swarm-optional.yaml")
RETIRED_BUNDLE = "v8-migration-standing-team-bridge"

EXACT_CONFIG_PAIRS = (
    ("cognitive/config/engine.yaml", "charts/mycelis-core/config/engine.yaml"),
    ("core/config/policy.yaml", "charts/mycelis-core/config/policy.yaml"),
    ("core/config/cognitive.yaml", "charts/mycelis-core/config/cognitive.yaml"),
    ("core/config/homepage.yaml", "charts/mycelis-core/config/homepage.yaml"),
    *(
        (f"core/config/templates/{name}", f"charts/mycelis-core/config/templates/{name}")
        for name in BOOT_BUNDLES
    ),
    ("core/config/teams/admin.yaml", "charts/mycelis-core/config/teams/admin.yaml"),
    ("core/config/teams/agui-design-architect.yaml", "charts/mycelis-core/config/teams/agui-design-architect.yaml"),
    ("core/config/teams/council.yaml", "charts/mycelis-core/config/teams/council.yaml"),
    ("core/config/teams/genesis.yaml", "charts/mycelis-core/config/teams/genesis.yaml"),
    ("core/config/teams/prime-architect.yaml", "charts/mycelis-core/config/teams/prime-architect.yaml"),
    ("core/config/teams/prime-development.yaml", "charts/mycelis-core/config/teams/prime-development.yaml"),
    ("core/config/teams/telemetry.yaml", "charts/mycelis-core/config/teams/telemetry.yaml"),
)


def test_chart_config_copies_match_canonical_sources():
    chart_root = ROOT / "charts" / "mycelis-core" / "config"
    mapped_chart_files = {
        Path(chart_path).relative_to("charts/mycelis-core/config").as_posix()
        for _, chart_path in EXACT_CONFIG_PAIRS
    }
    discovered_chart_files = {
        path.relative_to(chart_root).as_posix() for path in chart_root.rglob("*") if path.is_file()
    }
    assert mapped_chart_files == discovered_chart_files

    for source_path, chart_path in EXACT_CONFIG_PAIRS:
        assert (ROOT / chart_path).read_bytes() == (ROOT / source_path).read_bytes()


def test_boot_bundles_are_the_only_shipped_templates():
    for templates in ("core/config/templates", "charts/mycelis-core/config/templates"):
        shipped = sorted(path.name for path in (ROOT / templates).iterdir() if path.is_file())
        assert shipped == sorted(BOOT_BUNDLES), templates


def test_compose_and_chart_default_to_runtime_core_bundle():
    compose = (ROOT / "docker-compose.yml").read_text(encoding="utf-8")
    compose_example = (ROOT / ".env.compose.example").read_text(encoding="utf-8")
    configmap = (ROOT / "charts/mycelis-core/templates/configmap-config.yaml").read_text(encoding="utf-8")
    deployment = (ROOT / "charts/mycelis-core/templates/deployment.yaml").read_text(encoding="utf-8")

    assert "MYCELIS_BOOTSTRAP_TEMPLATE_ID: ${MYCELIS_BOOTSTRAP_TEMPLATE_ID:-mycelis-runtime-core}" in compose
    assert "MYCELIS_BOOTSTRAP_TEMPLATE_ID=mycelis-runtime-core\n" in compose_example
    for name in BOOT_BUNDLES:
        assert f'.Files.Get "config/templates/{name}"' in configmap
        assert f"path: templates/{name}" in deployment
    # The chart sets no bootstrap id, so Core's built-in default selects mycelis-runtime-core.
    assert "MYCELIS_BOOTSTRAP_TEMPLATE_ID" not in deployment
    for text in (compose, compose_example, configmap, deployment):
        assert RETIRED_BUNDLE not in text
