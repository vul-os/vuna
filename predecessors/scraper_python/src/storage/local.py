import os
import csv
from typing import List

from src.storage.storage import StorageUtils


class StorageUtilsLocal(StorageUtils):
    def __init__(self, local_dir):
        super().__init__()
        self.local_dir = local_dir
