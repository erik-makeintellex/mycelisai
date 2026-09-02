from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]

EXACT_CONFIG_PAIRS = (
    ("cognitive/config/engine.yaml", "charts/mycelis-core/config/engine.yaml"),
    ("core/config/policy.yaml", "charts/mycelis-core/config/policy.yaml"),
    ("core/config/cognitive.yaml", "charts/mycelis-core/config/cognitive.yaml"),
    ("core/config/homepage.yaml", "charts/mycelis-core/config/homepage.yaml"),
    (
        "core/config/templates/v8-migration-standing-team-bridge.yaml",
        "charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml",
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
