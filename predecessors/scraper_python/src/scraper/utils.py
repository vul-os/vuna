import os
import csv
import uuid
import hashlib
import requests
from typing import List
from urllib.parse import urlparse
from google.cloud import storage
import io
import tempfile


def hashStringFromUrl(url: str): 
    return hashlib.sha256(url.encode()).hexdigest()


class StorageUtils:
    def __init__(self, storage_client, bucket_name):
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

    def write_dicts_to_csv(self, filepath: str, data: List[dict]):
        with open(filepath, 'a', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=data[0].keys())
            if f.tell() == 0:
                writer.writeheader()
            writer.writerows(data)

    def upload_csv_from_dict(self, filename: str, data: List[dict]):
        # Create a blob object with the destination path and name
        blob = self.storage_client.bucket(self.bucket_name).blob(filename)

        filepath = os.path.join(self.local_dir, filename)
        self.write_dicts_to_csv(filepath, data)

        # Upload the file to GCS
        blob.upload_from_filename(filepath)

        print(f"File {filename} uploaded to gs://{self.bucket_name}/{filename}")
