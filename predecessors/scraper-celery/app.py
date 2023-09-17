from celery import Celery
import requests

# Configuration for Celery
app = Celery('tasks', broker='redis://localhost:6379/0')

@app.task
def create_task(url, key, json_payload=None, scheduled_time=None):
    headers = {"Key": key}

    if scheduled_time:
        create_task.apply_async(args=[url, key, json_payload], eta=scheduled_time)
        return f"Task scheduled for {scheduled_time}"

    if json_payload:
        # Logic to create POST task with JSON
        headers["Content-Type"] = "application/json"
        response = requests.post(url, data=json_payload, headers=headers)
        return f"POST task: {response.status_code}, {response.text}"
    else:
        # Logic to create GET task
        response = requests.get(url, headers=headers)
        return f"GET task: {response.status_code}, {response.text}"
