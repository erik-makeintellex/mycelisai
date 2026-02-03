
import psycopg2
import os
import uuid

# Configuration
DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DB_USER = os.getenv("POSTGRES_USER", "postgres")
DB_PASS = os.getenv("POSTGRES_PASSWORD", "postgres")
DB_NAME = os.getenv("POSTGRES_DB", "mycelis")

def get_connection():
    return psycopg2.connect(
        host=DB_HOST,
        user=DB_USER,
        password=DB_PASS,
        dbname=DB_NAME
    )

def run_verification():
    print("🔌 Connecting to Database...")
    try:
        conn = get_connection()
        conn.autocommit = True
        cursor = conn.cursor()
    except Exception as e:
        print(f"❌ Connection Failed: {e}")
        return

    admin_id = None
    try:
        cursor.execute("SELECT id FROM users LIMIT 1")
        result = cursor.fetchone()
        if result:
            admin_id = result[0]
        else:
            print("⚠️ No users found. Inserting default admin placeholder...")
            admin_id = str(uuid.uuid4())
            cursor.execute("INSERT INTO users (id, email, password_hash, role) VALUES (%s, 'admin@test', 'hash', 'admin')", (admin_id,))
    except Exception as e:
         print(f"⚠️ Could not fetch users (Table might not exist yet?): {e}")
         return

    print(f"👤 Using User ID: {admin_id}")

    # UUIDs for testing active constraint linkage
    mission_id = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
    root_team_id = "b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"
    sub_team_id = str(uuid.uuid4())

    # Test 1: Create Mission
    print("\n🧪 Test 1: Create Mission 'Project Absolute'")
    try:
        cursor.execute("""
            INSERT INTO missions (id, owner_id, name, directive)
            VALUES (%s, %s, 'Project Absolute', 'Symbiotic Swarm Intelligence')
            ON CONFLICT (id) DO NOTHING
        """, (mission_id, admin_id))
        print("✅ Success: Mission created.")
    except Exception as e:
        print(f"❌ Failed: {e}")

    # Test 2: Create Root Team
    print("\n🧪 Test 2: Create Root Team (Attached to Mission)")
    try:
        cursor.execute("""
            INSERT INTO teams (id, mission_id, name, path)
            VALUES (%s, %s, 'Command Root', %s)
            ON CONFLICT (id) DO NOTHING
        """, (root_team_id, mission_id, root_team_id))
        print("✅ Success: Root Team created.")
    except Exception as e:
        print(f"❌ Failed: {e}")

    # Test 3: Create Sub-Team
    print("\n🧪 Test 3: Create Sub-Team (Recursive Link)")
    try:
        cursor.execute("""
            INSERT INTO teams (id, mission_id, parent_id, name, path)
            VALUES (%s, %s, %s, 'Aerial Division', %s)
            ON CONFLICT (id) DO NOTHING
        """, (sub_team_id, mission_id, root_team_id, f"{root_team_id}.{sub_team_id}"))
        print("✅ Success: Sub-Team created.")
    except Exception as e:
        print(f"❌ Failed: {e}")

    # Test 4: Orphan Check
    print("\n🧪 Test 4: Orphan Check (Expect Constraint Failure)")
    try:
        cursor.execute("""
            INSERT INTO teams (name, mission_id) 
            VALUES ('Rogue Team', '00000000-0000-0000-0000-000000000000')
        """)
        print("❌ Failed: Orphan insertion succeded (Constraints broken).")
    except psycopg2.errors.ForeignKeyViolation:
         print("✅ Success: Constraint caught the orphan.")
    except Exception as e:
         print(f"⚠️ Unexpected Error: {e}")

    conn.close()

if __name__ == "__main__":
    run_verification()
