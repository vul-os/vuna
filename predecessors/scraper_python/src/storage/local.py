import os
import csv
from typing import List

from src.storage.storage import StorageUtils


class StorageUtilsLocal(StorageUtils):
    def __init__(self, local_dir):
        super().__init__()
        self.local_dir = local_dir

    def upload_image(self, image_url, site_id):
        pass

    def download_file(self, file_name):
        return os.path.join(self.local_dir, file_name)

    def _get_gcs_url(self, image_url, site_id):
        pass

    def upload_csv_from_dict(self, file_name: str, data: List[dict]):
        filepath = os.path.join(self.local_dir, file_name)
        self.write_dicts_to_csv(filepath, data)

    