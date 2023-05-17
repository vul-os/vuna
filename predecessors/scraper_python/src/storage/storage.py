from abc import ABC, abstractmethod
from typing import List
import csv
import os


class StorageUtils(ABC):
    def __init__(self):
        self.local_dir = "/"
        pass
    
    # maybe seperate out these methods later
    def write_data_to_txt(self, file_name: str, data: List[dict]):
        print(f"localdir: ", self.local_dir)
        file_path = os.path.join(self.local_dir, file_name)
        with open(file_path, 'w') as file:
            for item in data:
                file.write(item + '\n')

    def write_data_to_csv(self, file_name: str, data: List[dict]):
        file_path = os.path.join(self.local_dir, file_name)
        # Implement this method to write the given data to a CSV file at the specified file path
        with open(file_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=data[0].keys())
            print(data[0].keys())
            writer.writeheader()
            writer.writerows(data)