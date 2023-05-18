import time
import json
from google.cloud import tasks


def create_task(task: dict, project_id, location, queue_id):
    client = tasks.CloudTasksClient()
    parent = client.queue_path(project_id, location, queue_id)
    response = client.create_task(request={"parent": parent, "task": task})
    print(f"Task created: {response.name}")

def create_task_site(host, url):
    current_time = int(time.time())  # Get current timestamp as job_id
    job_id = str(current_time)

    return {
        "http_request": {
            "http_method": tasks.HttpMethod.POST,
            "url": f"{host}/site/{job_id}/{url}"
        }
    }


def create_task_meta(host, url):
    current_time = int(time.time())  # Get current timestamp as job_id
    job_id = str(current_time)

    return {
        "http_request": {
            "http_method": tasks.HttpMethod.POST,
            "url": f"{host}/meta/{job_id}/{url}"
        }
    }

def create_task_product(host, url, scraper_code):
    current_time = int(time.time())  # Get current timestamp as job_id
    job_id = str(current_time)

    task_url = f"{host}/product/{job_id}/{url}"

    payload = {
        "scraper_code": scraper_code,
    }
    json_payload = json.dumps(payload)

    return {
        "http_request": {
            "http_method": tasks.HttpMethod.POST,
            "url": task_url,
            "headers": {"Content-Type": "application/json"},
            "body": json_payload.encode()
        }
    }