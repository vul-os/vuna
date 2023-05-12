import os
import csv
from typing import List

from src.storage.storage import StorageUtils


class StorageUtilsLocal(StorageUtils):
    def __init__(self, local_dir):
        self.local_dir = local_dir

    def upload_image(self, image_url, site_id):
        # Implement this method to upload the image file to the storage backend
        pass

    def download_file(self, file_name):
        # Implement this method to download the file with the given name from the storage backend
        pass

    def _get_gcs_url(self, image_url, site_id):
        # Implement this method to get the GCS URL for the given image URL and site ID
        pass

    def write_data_to_csv(self, file_path: str, data: List[dict]):
        # Implement this method to write the given data to a CSV file at the specified file path
        with open(file_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=data[0].keys())
            writer.writeheader()
            writer.writerows(data)

    def upload_csv_from_data(self, file_name: str, data: List[dict]):
        # Implement this method to upload the given data as a CSV file with the specified name
        file_path = f"{self.local_dir}/{file_name}"
        self.write_data_to_csv(file_path, data)
        # Implement the upload to the storage backend here
        pass