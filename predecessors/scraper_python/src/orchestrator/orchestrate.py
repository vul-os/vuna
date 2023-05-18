
def get_urls_from_gcs_file(bucket_name, file_name):
    client = storage.Client()
    bucket = client.bucket(bucket_name)
    blob = bucket.blob(file_name)
    
    urls = []
    
    # Download the file as text
    content = blob.download_as_text()
    
    # Split the content into lines and add each line as a URL
    for line in content.splitlines():
        urls.append(line)
    
    return urls