import io
import os
import uuid
import hashlib
import requests
from urllib.parse import urlparse


class GCSUploader:
    def __init__(self, storage_client, bucket_name):
        self.storage_client = storage_client
        self.bucket_name = bucket_name

    def upload_image(self, image_url, site_id):
        gcs_url = self._get_gcs_url(image_url, site_id)

        bucket = self.client.bucket(self.bucket_name)
        blob = bucket.blob(gcs_url)

        if blob.exists():
            return gcs_url

        # Fetch the image from the URL
        response = requests.get(image_url)
        response.raise_for_status()
        image_bytes = io.BytesIO(response.content).read()

        blob.upload_from_string(image_bytes)

        return gcs_url

    def _get_gcs_url(self, image_url, site_id):
        # Calculate the hash of the image URL
        url_hash = hashlib.md5(image_url.encode()).hexdigest()

        file_ext = os.path.splitext(urlparse(image_url).path)[1]

        # Construct the object name with the hash and correct file extension
        return f"{site_id}/{url_hash}{file_ext}"