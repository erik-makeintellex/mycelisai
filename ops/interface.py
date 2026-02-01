from invoke import task

ns = None # Will be populated if we used a Class, but for now just function tasks

@task
def dev(c):
    """
    Start the Next.js development server.
    """
    print("🚀 Starting Interface Dev Server...")
    c.run("cd interface && npm run dev")

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
