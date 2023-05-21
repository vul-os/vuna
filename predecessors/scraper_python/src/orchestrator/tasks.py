from google.cloud import tasks
import json
import time


class TaskCreator:
    def __init__(self, project_id, location, queue_id):
        self.client = tasks.CloudTasksClient()
        self.parent = self.client.queue_path(project_id, location, queue_id)

    def create_task(self, task: dict):
        response = self.client.create_task(request={"parent": self.parent, "task": task})
        print(f"Task created: {response.name}")

    def create_task_site(self, url: str, target_url: str):
        current_time = int(time.time())
        job_id = str(current_time)

        return {
            "http_request": {
                "http_method": tasks.HttpMethod.POST,
                "url": f"{target_url}/scraper/site/{job_id}/{url}"
            }
        }

    def create_task_meta(self, url: str, target_url: str):
        current_time = int(time.time())
        job_id = str(current_time)

        return {
            "http_request": {
                "http_method": tasks.HttpMethod.POST,
                "url": f"{target_url}/scraper/meta/{job_id}/{url}"
            }
        }

    def create_task_product(self, url: str, scraper_code: str, target_url: str):
        current_time = int(time.time())
        job_id = str(current_time)

        task_url = f"{target_url}/scraper/product/{job_id}/{url}"

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
