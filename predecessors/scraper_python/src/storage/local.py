import os
import csv
from typing import List

from src.storage.storage import StorageUtils


class StorageUtilsLocal(StorageUtils):
    def __init__(self, local_dir):
        super().__init__()
        self.local_dir = local_dir

    def write_data_to_csv(self, file_name: str, data: List[dict]):
        super().write_data_to_csv(file_name, data)

    def write_data_to_txt(self, file_name: str, data: List[dict]):
        super().write_data_to_txt(file_name, data)
