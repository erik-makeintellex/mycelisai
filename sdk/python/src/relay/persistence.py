
import sqlite3
import time
import os
from typing import Generator, Tuple

DB_PATH = "agent_buffer.db"

def init_db(db_path: str = DB_PATH):
    """Ensure the buffer table exists."""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS buffer (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            topic TEXT NOT NULL,
            payload BLOB NOT NULL,
            timestamp TEXT NOT NULL
        )
    """)
    conn.commit()
    conn.close()

def save_impulse(topic: str, payload: bytes, db_path: str = DB_PATH):
    """Commit impulse to disk."""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    ts = time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
    cursor.execute(
        "INSERT INTO buffer (topic, payload, timestamp) VALUES (?, ?, ?)",
        (topic, payload, ts)
    )
    conn.commit()
    conn.close()
    # print(f"⚠️ [Buffer] Saved impulse to {topic} ({len(payload)} bytes)")

def drain_buffer(db_path: str = DB_PATH) -> Generator[Tuple[int, str, bytes, str], None, None]:
    """
    Yields rows and deletes them one by one.
    This ensures that if the process crashes during drain, 
    unprocessed items remain in the DB.
    """
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Fetch all IDs first to avoid locking issues during iteration?
    # Or just fetch one by one.
    # Simple approach: Select all, iterate, delete each.
    
    try:
        cursor.execute("SELECT id, topic, payload, timestamp FROM buffer ORDER BY id ASC")
        rows = cursor.fetchall()
        
        for row in rows:
            row_id, topic, payload, ts = row
            yield row_id, topic, payload, ts
            
            # Delete confirmed row
            cursor.execute("DELETE FROM buffer WHERE id = ?", (row_id,))
            conn.commit()
            
    finally:
        conn.close()
