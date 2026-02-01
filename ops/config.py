from pathlib import Path
import platform

# -- Config --
CLUSTER_NAME = "mycelis-cluster"
NAMESPACE = "mycelis"
ROOT_DIR = Path(__file__).parent.parent.resolve() # tasks/../ -> root
CORE_DIR = ROOT_DIR / "core"
SDK_DIR = ROOT_DIR / "sdk/python"

def is_windows():
    return platform.system() == "Windows"
