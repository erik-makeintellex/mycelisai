from invoke import task, Collection
import time

@task
def up(c, clean=False):
    """
    Launch the Full Mycelis Organism (Local Dev).
    Order: Cluster -> Dependencies -> Core -> Bridge.
    """
    print("🚀 Stimulating the Organism...")
    
    # 1. Check Infrastructure (Body)
    res = c.run("kubectl get nodes", warn=True, hide=True)
    if res.ok:
         print("✅ Body (Cluster) is Awake.")
    else:
         print("💤 Body is Asleep. Awakening...")
         c.run("inv k8s.init")
    
    # 2. Deploy Brain (Core + NATS + PG)
    # This handles building and helm install
    print("🧠 Enacting the Brain...")
    c.run("inv k8s.deploy", pty=True)
    
    # 3. Operations (Bridge)
    # We run this in background or just inform user? 
    # User said "launched together". 
    # But Bridge blocks. So we should spawn it or tell user to run it?
    # "inv k8s.bridge" runs kubectl port-forward which blocks.
    # It's better to instruct user or run in new terminal.
    # For now, we verify status.
    print("🩺 Checking Vitals...")
    c.run("inv k8s.status", pty=True)
    
    print("\n✨ The Organism is Conscious.")
    print("👉 Run 'inv k8s.bridge' in a separate terminal to connect Nerves.")
    print("👉 Run 'inv interface.dev' for the Visual Cortex.")

@task
def down(c):
    """
    Hibernate the Organism (Teardown).
    """
    print("💤 Inducing Hibernation...")
    c.run("kind delete cluster --name mycelis-cluster")
    print("✅ Hibernation Complete.")

dev_ns = Collection("dev")
dev_ns.add_task(up)
dev_ns.add_task(down)
