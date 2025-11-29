# GCP Tech Conf 2025 Website

A 1-day technical conference informational site built with Flask.

## Prerequisites

- Python 3.6+
- pip

## Setup and Run

1.  **Run the Application:**
    Navigate to the project directory:
    ```bash
    cd d:/MakeIntellex/Projects/mycelisai/scratch/gcp_conf_site
    ```
    Run the app using `uv` (this will automatically handle the Flask dependency):
    ```bash
    uv run --with flask app.py
    ```

2.  **Access the Website:**
    Open your web browser and go to `http://127.0.0.1:5000/`.

## Features

-   **Schedule:** View the 1-day event schedule with 8 talks and a lunch break.
-   **Search:** Filter talks by title, speaker, or category using the search bar.
-   **Speaker Details:** View speaker names and link to their LinkedIn profiles.
-   **Responsive Design:** Works on desktop and mobile.

## Project Structure

-   `app.py`: Flask application with dummy data and routes.
-   `templates/index.html`: Main HTML template.
-   `static/style.css`: CSS styles.
-   `static/script.js`: JavaScript for search functionality.
