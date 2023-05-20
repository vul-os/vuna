


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