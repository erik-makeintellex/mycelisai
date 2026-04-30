from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def _read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


def test_search_runtime_env_projects_through_compose_and_helm():
    surfaces = {
        "docker-compose.yml": [
            "MYCELIS_SEARCH_PROVIDER: ${MYCELIS_SEARCH_PROVIDER:-searxng}",
            "MYCELIS_SEARXNG_ENDPOINT: ${MYCELIS_SEARXNG_ENDPOINT:-http://searxng:8080}",
            "MYCELIS_SEARCH_LOCAL_API_ENDPOINT: ${MYCELIS_SEARCH_LOCAL_API_ENDPOINT:-}",
            "MYCELIS_SEARCH_MAX_RESULTS: ${MYCELIS_SEARCH_MAX_RESULTS:-8}",
        ],
        ".env.compose.example": [
            "MYCELIS_SEARCH_PROVIDER=searxng",
            "MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080",
            "MYCELIS_SEARCH_LOCAL_API_ENDPOINT=",
            "MYCELIS_SEARCH_MAX_RESULTS=8",
        ],
        "ops/compose_env.py": [
            '"MYCELIS_SEARCH_PROVIDER"',
            '"MYCELIS_SEARXNG_ENDPOINT"',
            '"MYCELIS_SEARCH_LOCAL_API_ENDPOINT"',
            '"MYCELIS_SEARCH_MAX_RESULTS"',
        ],
        "charts/mycelis-core/values.yaml": [
            "search:",
            "provider: disabled",
            'searxngEndpoint: ""',
            'localApiEndpoint: ""',
            "maxResults: 8",
        ],
        "charts/mycelis-core/templates/deployment.yaml": [
            "MYCELIS_SEARCH_PROVIDER",
            "MYCELIS_SEARXNG_ENDPOINT",
            "MYCELIS_SEARCH_LOCAL_API_ENDPOINT",
            "MYCELIS_SEARCH_MAX_RESULTS",
            'value: {{ default "disabled" .provider | quote }}',
            'value: {{ default "" .searxngEndpoint | quote }}',
            'value: {{ default "" .localApiEndpoint | quote }}',
            "value: {{ default 8 .maxResults | quote }}",
        ],
    }

    missing: list[str] = []
    for relative, snippets in surfaces.items():
        text = _read(relative)
        missing.extend(f"{relative} missing `{snippet}`" for snippet in snippets if snippet not in text)

    assert not missing, "Search runtime env projection is incomplete:\n" + "\n".join(missing)
