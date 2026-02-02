from invoke import task, Collection

ns = Collection("interface")

@task
def dev(c):
    """Start Interface (Next.js) in Dev Mode."""
    c.run("npm run dev --prefix interface", pty=True)

ns.add_task(dev)

@task
def install(c):
    """
    Install Interface dependencies.
    """
    print("📦 Installing Interface Dependencies...")
    c.run("cd interface && npm install")

@task
def build(c):
    """
    Build the Interface for production.
    """
    print("🔨 Building Interface...")
    c.run("cd interface && npm run build")
