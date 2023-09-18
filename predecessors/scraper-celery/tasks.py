from celery import Celery, Task
import pytz
import requests
from datetime import datetime, timedelta

# Configuration for Celery
app = Celery(
    'tasks',
    broker='redis://localhost:6379/0',
    backend='redis://localhost:6379/1'
)


@app.task
def create_task(url, json_payload=None, scheduled_time=None):
    headers = {}
    print(json_payload)
    if scheduled_time:
        # Get the current time in UTC
        current_time_utc = datetime.now(pytz.utc)

        # Format it in ISO 8601 format
        scheduled_time_iso = current_time_utc.isoformat()
        
        # Use apply_async on the current task to schedule another task
        create_task.apply_async(args=[url, json_payload], eta=scheduled_time_iso)
        return f"Task scheduled for {scheduled_time}"        

    if json_payload:       
        # Logic to create POST task with JSON
        headers["Content-Type"] = "application/json"
        response = requests.post(url, json=json_payload, headers=headers)
        return f"POST task: {response.status_code}, {response.text}"
    else:
        # Logic to create GET task
        response = requests.get(url, headers=headers)
        return f"GET task: {response.status_code}, {response.text}"