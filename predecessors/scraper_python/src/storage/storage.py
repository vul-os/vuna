from abc import ABC, abstractmethod
import csv

class StorageUtils(ABC):
    @abstractmethod
    def write_data(self, data, file_type):
        pass
    
    @abstractmethod
    def read_data(self, file_type):
        pass

    @abstractmethod
    def get_latest_files(self, folder_prefix, text_in):
        pass

    @abstractmethod
    def get_latest_file(self, folder_prefix, text_in):
        pass


    @abstractmethod
    def create_dirs_for_file(self, full_path):
        pass