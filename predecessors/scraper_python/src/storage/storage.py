from abc import ABC, abstractmethod
from typing import List
import csv


class StorageUtils(ABC):
    def __init__(self):
        pass
    
    @abstractmethod
    def upload_image(self, image_url, site_id):
        pass

    @abstractmethod
    def download_file(self, file_name):
        pass

    @abstractmethod
    def _get_gcs_url(self, image_url, site_id):
        pass

    @abstractmethod
    def write_data_to_csv(self, file_path: str, data: List[dict]):
        pass

    @abstractmethod
    def upload_csv_from_data(self, file_name: str, data: List[dict]):
        pass