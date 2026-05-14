from .k8s_standards import _shell_quote


def search_overrides(env):
    return {
        "provider": env.get("MYCELIS_K8S_SEARCH_PROVIDER", "").strip(),
        "searxng": env.get("MYCELIS_K8S_SEARXNG_ENDPOINT", "").strip(),
        "local_api": env.get("MYCELIS_K8S_SEARCH_LOCAL_API_ENDPOINT", "").strip(),
    }


def print_search_overrides(overrides):
    labels = {
        "provider": "Search provider",
        "searxng": "SearXNG endpoint",
        "local_api": "Search local API endpoint",
    }
    for key, label in labels.items():
        if overrides.get(key):
            print(f"   {label}: {overrides[key]}")


def append_search_overrides(cmd, overrides):
    if overrides.get("provider"):
        cmd += f"--set-string search.provider={_shell_quote(overrides['provider'])} "
    if overrides.get("searxng"):
        cmd += f"--set-string search.searxngEndpoint={_shell_quote(overrides['searxng'])} "
    if overrides.get("local_api"):
        cmd += f"--set-string search.localApiEndpoint={_shell_quote(overrides['local_api'])} "
    if overrides.get("searxng") or overrides.get("local_api"):
        cmd += "--set networkPolicy.searchEgress.enabled=true "
    return cmd
