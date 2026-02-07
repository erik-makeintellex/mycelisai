from invoke import task, Collection
from . import core, interface

@task
def all(c):
    """
    Run ALL Unit Tests (Core + Interface).
    """
    print("Executing Full Test Suite...")
    
    try:
        core.test(c)
        interface.test(c)
        print("All Tests Passed.")
    except Exception as e:
        print(f"Test Failure: {e}")
        exit(1)

ns = Collection("test")
ns.add_task(all)
