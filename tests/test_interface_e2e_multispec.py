import shlex
from ops import interface_runtime as interface


def test_build_playwright_command_supports_multiple_space_separated_specs():
    # A quoted, space-separated --spec value becomes extra positional args
    # Playwright treats as additional test files (used by CI's mvp-browser job).
    command = interface._build_playwright_command(project="chromium", spec="a.spec.ts b.spec.ts", workers="1")
    assert shlex.split(command) == ["npx", "playwright", "test", "--reporter=dot", "--project=chromium", "a.spec.ts", "b.spec.ts", "--workers=1"]
