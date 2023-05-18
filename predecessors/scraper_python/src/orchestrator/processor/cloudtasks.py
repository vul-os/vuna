from google.cloud import storage, tasks


def read_urls_from_file(bucket_name, file_name):
    client = storage.Client()
    bucket = client.bucket(bucket_name)
    blob = bucket.blob(file_name)
    content = blob.download_as_text()
    urls = content.splitlines()
    return urls


def create_task(project_id, location, queue_id, url):
    client = tasks.CloudTasksClient()
    parent = client.queue_path(project_id, location, queue_id)

    task = {
        "http_request": {
            "http_method": tasks.HttpMethod.POST,
            "url": url
        }
    }

    response = client.create_task(request={"parent": parent, "task": task})
    print(f"Task created: {response.name}")


if __name__ == "__main__":
    # Define the GCS bucket and text file information
    bucket_name = "your-bucket-name"
    file_name = "path/to/your-file.txt"

    # Define the Cloud Tasks information
    project_id = "your-project-id"
    queue_id = "your-queue-id"
    location = "your-location"  # e.g., "us-central1"

    urls = read_urls_from_file(bucket_name, file_name)
    for url in urls:
        create_task(project_id, location, queue_id, url)