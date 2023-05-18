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