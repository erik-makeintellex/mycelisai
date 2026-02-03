
import psycopg2
import os
import glob
from dotenv import load_dotenv

load_dotenv()

# Configuration
DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DB_USER = os.getenv("POSTGRES_USER", "postgres")
DB_PASS = os.getenv("POSTGRES_PASSWORD", "postgres")
DB_NAME = os.getenv("POSTGRES_DB", "mycelis")

MIGRATION_DIR = r"core/migrations"

def get_connection():
    return psycopg2.connect(
        host=DB_HOST,
        user=DB_USER,
        password=DB_PASS,
        dbname=DB_NAME
    )

def deploy_schema():
    print(f"🚀 Deploying Scema from: {MIGRATION_DIR}")
    
    # Find all SQL files (both .sql and .up.sql)
    files = glob.glob(os.path.join(MIGRATION_DIR, "*.sql"))
    files.sort() # Ensure numerical order (001, 002, 003)

    if not files:
        print("❌ No migration files found.")
        return

    try:
        conn = get_connection()
        conn.autocommit = True
        cursor = conn.cursor()
    except Exception as e:
        print(f"❌ Connection Failed: {e}")
        return

    for file_path in files:
        filename = os.path.basename(file_path)
        print(f"📜 Applying: {filename}...")
        
        with open(file_path, 'r') as f:
            sql_content = f.read()

        try:
            cursor.execute(sql_content)
            print(f"   ✅ Success.")
        except psycopg2.errors.DuplicateTable:
            print(f"   ⚠️  Table already exists (Skipping).")
        except psycopg2.errors.DuplicateObject:
            print(f"   ⚠️  Object already exists (Skipping).")
        except Exception as e:
            # We continue? Or Halt? For now, print error but continue (some scripts might be idempotent)
            print(f"   ❌ Failed: {e}")
    
    conn.close()
    print("✨ Schema Deployment Complete.")

if __name__ == "__main__":
    deploy_schema()
