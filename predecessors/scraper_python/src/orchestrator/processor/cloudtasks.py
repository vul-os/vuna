import time
from google.cloud import storage, tasks


def read_urls_from_file(bucket_name, file_name):
    client = storage.Client()
    bucket = client.bucket(bucket_name)
    blob = bucket.blob(file_name)
    content = blob.download_as_text()
    urls = content.splitlines()
    return urls

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


if __name__ == "__main__":
    
    # Define the GCS bucket and text file information
    bucket_name = "your-bucket-name"
    file_name = "path/to/your-file.txt"

    # Define the Cloud Tasks information
    project_id = "your-project-id"
    queue_id = "your-queue-id"
    location = "your-location"  # e.g., "us-central1"

    host = input("Enter the host: ")
    urls = read_urls_from_file(bucket_name, file_name)
    for url in urls:
        task = create_task_meta(host, url)
        create_task(task, project_id, location, queue_id)
