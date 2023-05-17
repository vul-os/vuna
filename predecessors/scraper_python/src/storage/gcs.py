import os
import io
import csv
import uuid
import hashlib
import tempfile
import requests
from typing import List
from urllib.parse import urlparse
from google.cloud import storage

from src.storage.storage import StorageUtils


class StorageUtilsGCS(StorageUtils):
    def __init__(self, storage_client, bucket_name):
        # unnecesary
        super().__init__()

        self.storage_client = storage_client
        self.bucket_name = bucket_name
        self.local_dir = tempfile.TemporaryDirectory().name

    def upload_image(self, image_url, site_id):
        gcs_url = self._get_gcs_url(image_url, site_id)

        bucket = self.storage_client.bucket(self.bucket_name)
        blob = bucket.blob(gcs_url)

        if blob.exists():
            return gcs_url

        # Fetch the image from the URL
        response = requests.get(image_url)
        response.raise_for_status()
        image_bytes = io.BytesIO(response.content).read()

        blob.upload_from_string(image_bytes)

        return gcs_url

    def upload_file(self, file_name):
        file_path = os.path.join(self.local_dir, file_name)
        # Create a blob object with the destination path and name
        blob = self.storage_client.bucket(self.bucket_name).blob(f"meta/{file_name}")
        # Upload the file to GCS
        blob.upload_from_filename(file_path)

    def download_file(self, file_name):
        scraper_file = os.path.join(self.local_dir, f"{file_name}")

        if not os.path.isfile(scraper_file):
            # Download scraper file from GCS
            bucket = self.storage_client.get_bucket(self.bucket_name)
            blob = bucket.blob(f"{file_name}")
            scraper_code = blob.download_as_string().decode('utf-8')

            # Save scraper file to cache directory
            with open(scraper_file, "w") as f:
                f.write(scraper_code)
        return scraper_file

    def _get_gcs_url(self, image_url, site_id):
        # Calculate the hash of the image URL
        url_hash = hashlib.md5(image_url.encode()).hexdigest()

        file_ext = os.path.splitext(urlparse(image_url).path)[1]

        # Construct the object name with the hash and correct file extension
        return f"{site_id}/{url_hash}{file_ext}"

    def write_data_to_txt(self, file_name: str, data: List[dict]):
        super().write_data_to_txt(file_name, data)
        print(self.local_dir, file_name, data)
        print(os.path.exists(os.path.join(self.local_dir, file_name)))
        self.upload_file(file_name)


    def write_data_to_csv(self, file_name: str, data: List[dict]):
        super().write_data_to_csv(file_name, data)
        self.upload_file(file_name)
