from google.cloud import storage
import os
from src.storage.storage import StorageUtils


class StorageUtilsGCS(StorageUtils):
    def __init__(self, bucket_name):
        self.bucket_name = bucket_name
        self.client = storage.Client()
        self.bucket = self.client.get_bucket(bucket_name)
    
    def write_data(self, data, file_type, file_path):
        self.create_dirs_for_file(file_path)
        
        if file_type == 'csv':
            self.write_csv_data(data, file_path)
        elif file_type == 'txt':
            self.write_txt_data(data, file_path)
        else:
            print("Invalid file type.")
    
    def read_data(self, file_type, file_path):
        if file_type == 'csv':
            return self.read_csv_data(file_path)
        elif file_type == 'txt':
            return self.read_txt_data(file_path)
        
        print("Invalid file type.")
        return []
    
    def write_csv_data(self, data, file_path):
        blob = self.bucket.blob(file_path)
        content = '\n'.join(','.join(str(value) for value in row.values()) for row in data)
        blob.upload_from_string(content)
        print(f"Data written to CSV file '{file_path}' in GCS.")
    
    def read_csv_data(self, file_path):
        blob = self.bucket.blob(file_path)
        content = blob.download_as_text()
        data = []
        for line in content.split('\n'):
            if line:
                values = line.split(',')
                data.append(values)
        return data
    
    def write_txt_data(self, data, file_path):
        blob = self.bucket.blob(file_path)
        content = '\n'.join(data)
        blob.upload_from_string(content)
        print(f"Data written to TXT file '{file_path}' in GCS.")
    
    def read_txt_data(self, file_path):
        blob = self.bucket.blob(file_path)
        content = blob.download_as_text()
        data = content.split('\n')
        return data

    def create_dirs_for_file(self, file_path):
        directory = os.path.dirname(file_path)
        if directory:
            # Create empty placeholder object to emulate directory
            blob = self.bucket.blob(directory + '/')
            blob.upload_from_string('')
            print("Directories created in GCS: " + directory)

    
    def retrieve_last_file(self, folder_prefix, ends_with):
        # Get the blobs in the specified folder
        blobs = self.bucket.list_blobs(prefix=folder_prefix)

        # Filter blobs with the desired file name format
        desired_files = [
            blob for blob in blobs if blob.name.endswith(ends_with)
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