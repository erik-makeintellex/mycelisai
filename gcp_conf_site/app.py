from flask import Flask, render_template
from datetime import datetime

app = Flask(__name__)

# Dummy Data
SPEAKERS = {
    1: {"first_name": "Alice", "last_name": "Chen", "linkedin": "https://www.linkedin.com/in/alicechen-dummy"},
    2: {"first_name": "Bob", "last_name": "Smith", "linkedin": "https://www.linkedin.com/in/bobsmith-dummy"},
    3: {"first_name": "Carol", "last_name": "Davis", "linkedin": "https://www.linkedin.com/in/caroldavis-dummy"},
    4: {"first_name": "David", "last_name": "Wilson", "linkedin": "https://www.linkedin.com/in/davidwilson-dummy"},
    5: {"first_name": "Eve", "last_name": "Johnson", "linkedin": "https://www.linkedin.com/in/evejohnson-dummy"},
    6: {"first_name": "Frank", "last_name": "Miller", "linkedin": "https://www.linkedin.com/in/frankmiller-dummy"},
    7: {"first_name": "Grace", "last_name": "Lee", "linkedin": "https://www.linkedin.com/in/gracelee-dummy"},
    8: {"first_name": "Heidi", "last_name": "Klum", "linkedin": "https://www.linkedin.com/in/heidiklum-dummy"},
    9: {"first_name": "Ivan", "last_name": "Petrov", "linkedin": "https://www.linkedin.com/in/ivanpetrov-dummy"},
    10: {"first_name": "Judy", "last_name": "Garland", "linkedin": "https://www.linkedin.com/in/judygarland-dummy"},
}

TALKS = [
    {
        "id": 1,
        "title": "Keynote: The Future of Cloud Computing",
        "speakers": [1],
        "category": "Keynote",
        "description": "An overview of where Google Cloud is heading in the next 5 years.",
        "time": "09:00 - 10:00"
    },
    {
        "id": 2,
        "title": "Serverless Architectures with Cloud Run",
        "speakers": [2, 3],
        "category": "Serverless",
        "description": "Deep dive into building scalable applications using Cloud Run.",
        "time": "10:15 - 11:00"
    },
    {
        "id": 3,
        "title": "BigQuery for Data Warehousing",
        "speakers": [4],
        "category": "Data",
        "description": "Best practices for managing petabytes of data with BigQuery.",
        "time": "11:15 - 12:00"
    },
    {
        "id": 4,
        "title": "Lunch Break",
        "speakers": [],
        "category": "Break",
        "description": "Enjoy a catered lunch and networking.",
        "time": "12:00 - 13:00"
    },
    {
        "id": 5,
        "title": "Kubernetes 101: Managing Containers",
        "speakers": [5, 6],
        "category": "Containers",
        "description": "Getting started with GKE and container orchestration.",
        "time": "13:00 - 13:45"
    },
    {
        "id": 6,
        "title": "AI & ML on GCP: Vertex AI",
        "speakers": [7],
        "category": "AI/ML",
        "description": "Building and deploying machine learning models with Vertex AI.",
        "time": "14:00 - 14:45"
    },
    {
        "id": 7,
        "title": "Security Best Practices in the Cloud",
        "speakers": [8],
        "category": "Security",
        "description": "Securing your cloud infrastructure and data.",
        "time": "15:00 - 15:45"
    },
    {
        "id": 8,
        "title": "Networking Deep Dive",
        "speakers": [9, 10],
        "category": "Networking",
        "description": "Understanding VPCs, Load Balancers, and Hybrid Connectivity.",
        "time": "16:00 - 16:45"
    },
    {
        "id": 9,
        "title": "Closing Remarks",
        "speakers": [1],
        "category": "Keynote",
        "description": "Wrap up of the day and Q&A.",
        "time": "17:00 - 17:30"
    }
]

def get_speaker_details(speaker_ids):
    return [SPEAKERS[sid] for sid in speaker_ids]

@app.route('/')
def index():
    # Enrich talks with speaker details
    enriched_talks = []
    for talk in TALKS:
        new_talk = talk.copy()
        new_talk['speaker_details'] = get_speaker_details(talk['speakers'])
        enriched_talks.append(new_talk)
    
    current_date = datetime.now().strftime("%B %d, %Y")
    return render_template('index.html', talks=enriched_talks, date=current_date)

if __name__ == '__main__':
    app.run(debug=True)
