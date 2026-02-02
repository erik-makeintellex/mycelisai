
from typer.testing import CliRunner
from cli.main import app

runner = CliRunner()

def test_help():
    result = runner.invoke(app, ["--help"])
    assert result.exit_code == 0
    assert "Synaptic Injector" in result.stdout

def test_status_help():
    result = runner.invoke(app, ["status", "--help"])
    assert result.exit_code == 0
    assert "health" in result.stdout
