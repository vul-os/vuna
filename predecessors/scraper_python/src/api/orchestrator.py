import requests
from flask import jsonify
from src.scraper import MetaScraper, ScraperLoader, scrape_product_data, SiteScraper
from src.storage.local import StorageUtilsLocal as StorageUtils
from src.orchestrator.utils import get_urls_from_gcs_file
from src.storage.gcs import StorageUtils
from src.orchestrator.tasks import TaskCreator

class OrchestratorAPI:
    def __init__(self, task_creator: TaskCreator, storage_utils: StorageUtils, site_gcs_file_path: str):
        self.task_creator = task_creator
        self.storage_utils = storage_utils # gcs only
        self.site_gcs_file_path = site_gcs_file_path

    def meta(self, request):
        try:
            urls = self.storage_utils.get_file(self.site_gcs_file_path)
            for url in urls:
                self.task_creator.create_task_meta(url)
            return "hopefully created", 200
        except Exception as exception:
            return str(exception), 500

    def site(self, request):
        try:
            urls = self.storage_utils.get_file(self.site_gcs_file_path)
            for url in urls:
                self.task_creator.create_task_site(url)
            return "hopefully created", 200
        except Exception as exception:
            return str(exception), 500


    def product(self, request):
        try:

        except Exception as ex:
            return str(ex), 500