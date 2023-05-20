import requests
from src.storage.gcs import StorageUtils

def create_proxy_list(file_path):
    # Send a GET request to retrieve the file contents
    response = requests.get('https://spys.me/proxy.txt')
    # Fetch the proxy list content from the URL
    proxies = response.text.split('\n')

    # Determine the number of header lines
    header_lines = 0
    for line in proxies:
        if line.startswith('IP address:Port'):
            break
        header_lines += 1

    # Remove the header lines
    proxies = proxies[header_lines:]

    proxy_list = []

    # Create proxy URLs for each proxy
    for proxy in proxies:
        proxy_parts = proxy.strip().split(' ')
        ip_port = proxy_parts[0]
        supports_ssl = 'S' in proxy_parts[1]

        # Create the proxy URL
        if not supports_ssl:
            proxy_url = f'http://{ip_port}'
            proxy_list.append(proxy_url)

    return proxy_list


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