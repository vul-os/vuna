import os
import importlib
from ..utils import StorageUtils


class ScraperLoader:
    def __init__(self, storage_utils: StorageUtils, scraper_filename: str):
        self.storage_utils = storage_utils
        self.scraper_filename = scraper_filename
        
    def __call__(self):
        # Check cache directory for scraper file
        scraper_file = self.storage_utils.download_file(self.scraper_filename)

        # Execute scraper file dynamically
        spec = importlib.util.spec_from_file_location("scraper_module", scraper_file)
        scraper_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(scraper_module)

        # Get the scraper class from the module
        scraper_class = getattr(scraper_module, "Scraper")

        # Instantiate the scraper and return it
        return scraper_class()
