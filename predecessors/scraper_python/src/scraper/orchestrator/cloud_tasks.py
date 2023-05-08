import base64
import datetime
import os

from google.auth import compute_engine
from google.cloud import tasks_v2

import requests


def create_task_for_product(client, queue_path, product, base_url):
    # credentials = compute_engine.Credentials()

    # client = tasks_v2.CloudTasksClient(credentials=credentials)

    # project_id = "<project-id>"
    # location = "<location>"
    # queue_id = "<queue-id>"
    # queue_path = client.queue_path(project_id, location, queue_id)

    technology = requests.utils.quote(product["Technology"])
    url_encoded = requests.utils.quote(product["URL"])
    url = f"{base_url}/?technology={technology}&url={url_encoded}"

    task = {
        "http_request": {
            "http_method": tasks_v2.HttpMethod.GET,
            "url": url,
            "headers": {
                "Content-Type": "application/json",
            },
        },
        "name": f"product-task-{product['ID']}",
        "schedule_time": tasks_v2.types.Timestamp().FromDatetime(
            datetime.datetime.utcnow() + datetime.timedelta(minutes=5)
        ),
    }

    response = client.create_task(request={"parent": queue_path, "task": task})

    return response

def create_tasks_for_products(products, batch_size):
    for i in range(0, len(products), batch_size):
        batch = products[i:i + batch_size]

        tasks = []
        for product in batch:
            task = create_task_for_product(product)
            tasks.append(task)

        for task in tasks:
            print(f"Created task: {task.name}")


if __name__ == "__main__":
    products = [
        {"ID": 1, "URL": "https://example.com/product1", "Technology": "Python"},
        {"ID": 2, "URL": "https://example.com/product2", "Technology": "Java"},
        {"ID": 3, "URL": "https://example.com/product3", "Technology": "JavaScript"},
    ]

    batch_size = 100

    # create_tasks_for_products(products, batch_size)