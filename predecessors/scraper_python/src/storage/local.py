import os
import csv
from typing import List

from src.storage.storage import StorageUtils


class StorageUtilsLocal(StorageUtils):

    def __init__(self, local_dir):
        self.local_dir = local_dir

    def get_latest_files(self, folder_prefix, text_in):
        print("yaa")
        return "hello"

    def get_latest_file(self, folder_prefix, text_in):
        print("yaa")
        return "hello"

    def write_data(self, data, file_type, file_path):
        file_path = os.path.join(self.local_dir, file_path)
        self.create_dirs_for_file(file_path)

        if file_type == 'csv':
            self.write_csv_data(data, file_path)
        elif file_type == 'txt':
            self.write_txt_data(data, file_path)
        else:
            print("Invalid file type.")
    
    def read_data(self, file_type, file_path):
        file_path = os.path.join(self.local_dir, file_path)
        if file_type == 'csv':
            return self.read_csv_data(file_path)
        elif file_type == 'txt':
            return self.read_txt_data(file_path)
        else:
            print("Invalid file type.")
            return []
    
    def write_csv_data(self, data, file_path):
        fieldnames = data[0].keys() if len(data) > 0 else []
        with open(file_path, 'w', newline='') as file:
            writer = csv.DictWriter(file, fieldnames=fieldnames)
            writer.writeheader()
            writer.writerows(data)
        print("Data written to CSV file: " + file_path)
    
    def read_csv_data(self, file_path):
        data = []
        with open(file_path, 'r') as file:
            reader = csv.DictReader(file)
            for row in reader:
                data.append(row)
        return data
    
    def write_txt_data(self, data, file_path):
        with open(file_path, 'w') as file:
            for item in data:
                file.write(str(item) + '\n')
        print("Data written to TXT file: " + file_path)
    
    def read_txt_data(self, file_path):
        data = []
        with open(file_path, 'r') as file:
            for line in file:
                data.append(line.strip())
        return data

    def create_dirs_for_file(self, file_path):
        directory = os.path.join(self.local_dir, os.path.dirname(file_path))
        os.makedirs(directory, exist_ok=True)
        print("Directories created for file: " + file_path)