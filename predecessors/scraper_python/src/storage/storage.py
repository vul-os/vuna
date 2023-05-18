from abc import ABC
from typing import List
import csv
import os


class StorageUtils(ABC):
    def __init__(self):
        self.local_dir = "/"
    
    def get_filename(self, file_path):
        path_parts = file_path.rsplit("/", 1)
        filename = path_parts[1]
        return filename

    # maybe seperate out these methods later
    def write_data_to_txt(self, file_path: str, data: List[dict]):
        filename = self.get_filename(file_path)
        full_file_path = os.path.join(self.local_dir, filename)
        with open(full_file_path, 'w') as file:
            for item in data:
                file.write(item + '\n')

    def write_data_to_csv(self, file_path: str, data: List[dict]):
        filename = self.get_filename(file_path)
        full_file_path = os.path.join(self.local_dir, filename)
        with open(full_file_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=data[0].keys())
            print(data[0].keys())
            writer.writeheader()
            writer.writerows(data)