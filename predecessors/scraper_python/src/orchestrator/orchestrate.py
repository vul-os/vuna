import requests
from google.cloud import storage

def get_urls_from_gcs_file(client: storage.Client, bucket_name: str, file_name: str):
    bucket = client.bucket(bucket_name)
    blob = bucket.blob(file_name)
    
    urls = []
    
    # Download the file as text
    content = blob.download_as_text()
    
    # Split the content into lines and add each line as a URL
    for line in content.splitlines():
        urls.append(line)
    
    return urls

def retrieve_proxys_from_file():
    file_url = 'https://spys.me/proxy.txt'
    # Send a GET request to retrieve the file contents
    response = requests.get(file_url)

    # Check if the request was successful
    if response.status_code == 200:
        # Get the text content of the response
        file_content = response.text

        # Split the file content by newlines
        lines = file_content.split('\n')

        # Extract the URLs from the lines
        urls = [line.split()[1] for line in lines if line.startswith('Http proxy')]

        return urls
    else:
        print('Failed to retrieve the file:', response.status_code)
        return []

def retrieve_recent_files(client: storage.Client, bucket_name, folder_prefix):
    # Get the bucket
    bucket = client.get_bucket(bucket_name)

    # Get the blobs in the specified folder
    blobs = bucket.list_blobs(prefix=folder_prefix)

    # Filter blobs with the desired file name format
    desired_files = [
        blob for blob in blobs if blob.name.endswith('_products.txt')
    ]

    # Sort the files by their job identifier (datetime)
    sorted_files = sorted(
        desired_files,
        key=lambda blob: int(blob.name.split('/')[1])
    )

    # Get the most recent file
    most_recent_file = sorted_files[-1]

    # Iterate through each file name
    for file_blob in sorted_files:
        file_name = file_blob.name.split('/')[2]

        # Extract encoded_site and formatted_datetime
        encoded_site, formatted_datetime, _ = file_name.split('_')

        # Perform operations with the file name here
        print('File name:', file_name)
        print('Encoded site:', encoded_site)
        print('Formatted datetime:', formatted_datetime)

    # Return the most recent file
    return most_recent_file