import os
import base64
import re
import datetime
from google.cloud.storage import Client
from src.storage.storage import StorageUtils


class StorageUtilsGCS(StorageUtils):
    def __init__(self, storage_client: Client, bucket_name: str):
        self.bucket_name = bucket_name
        self.client = storage_client
        self.bucket = self.client.get_bucket(self.bucket_name)
    
    def write_data(self, data, file_path):
        self.create_dirs_for_file(file_path)
        file_type = self.get_file_extension(file_path)
        
        if file_type == 'csv':
            self.write_csv_data(data, file_path)
        elif file_type == 'txt':
            self.write_txt_data(data, file_path)
        else:
            print("Invalid file type.")
    
    def read_data(self, file_path):
        file_type = self.get_file_extension(file_path)

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

    @staticmethod
    def get_file_extension(file_path):
        _, file_extension = os.path.splitext(file_path)
        return file_extension.lstrip(".")
    
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
    
    @staticmethod
    def flatten_list(lst):
        flattened_list = []
        for item in lst:
            if isinstance(item, list):
                flattened_list.extend(item)
            else:
                flattened_list.append(item)
        return flattened_list

    def read_txt_data(self, file_path):
        blob = self.bucket.blob(file_path)
        content = blob.download_as_text()
        data = content.split('\n')
        data = self.flatten_list(data)
        return data

    def create_dirs_for_file(self, file_path):
        directory = os.path.dirname(file_path)
        if directory:
            # Create empty placeholder object to emulate directory
            blob = self.bucket.blob(directory + '/')
            blob.upload_from_string('')
            print("Directories created in GCS: " + directory)

    @staticmethod
    def extract_datetime(string):
        regex_pattern = r"\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}"
        datetime_matches = re.findall(regex_pattern, string)
        if datetime_matches:
            datetime_objects = [datetime.datetime.strptime(dt_str, "%Y-%m-%d-%H-%M-%S") for dt_str in datetime_matches]
            return datetime_objects[0]  # Get the first datetime object
        else:
            return None
            
    def get_latest_files(self, folder_prefix, text_in):
        # Dictionary to store the latest file for each encoded_site
        blobs = self.bucket.list_blobs(prefix=folder_prefix, delimiter="/")
        latest_files = {}
        for blob in blobs:
            blob_name = blob.name
            if text_in in blob_name:
                datetime_obj = self.extract_datetime(blob_name)
                if not datetime_obj:
                    continue
                encoded_site, formatted_datetime, _ = blob_name.split("_")
                # Check if this blob is the latest for the encoded_site
                if encoded_site not in latest_files or datetime_obj > latest_files[encoded_site]["datetime"]:
                    latest_files[encoded_site] = {"file": blob_name, "datetime": datetime_obj}

        return [value['file'] for value in latest_files.values() if value]

    def get_latest_file(self, folder_prefix, text_in):
        latest_filename = None
        latest_datetime = None
        # Dictionary to store the latest file for each encoded_site
        blobs = self.bucket.list_blobs(prefix=folder_prefix, delimiter="/")
        for blob in blobs:
            blob_name = blob.name
            if text_in in blob_name:
                datetime_obj = self.extract_datetime(blob_name)
                if not datetime_obj:
                    continue
                # Check if this blob is the latest based on the datetime
                if latest_datetime is None or datetime_obj > latest_datetime:
                    latest_filename = blob_name
                    latest_datetime = datetime_obj

        return latest_filename