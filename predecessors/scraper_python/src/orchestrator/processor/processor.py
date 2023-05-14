from abc import ABC, abstractmethod

class BatchProcessor(ABC):
    def __init__(self, folder_path, chunk_size=1024, ):
        self.folder_path = folder_path
        self.chunk_size = chunk_size
        self.rate_limits = {}

    @abstractmethod
    def process_files(self):
        pass

