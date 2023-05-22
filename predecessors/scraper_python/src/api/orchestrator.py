from google.protobuf.timestamp_pb2 import Timestamp
from datetime import datetime, timedelta

from src.storage.local import StorageUtilsLocal as StorageUtils
from src.storage.gcs import StorageUtils
from src.orchestrator.tasks import TaskCreator


class OrchestratorAPI:
    def __init__(self, task_creator: TaskCreator, storage_utils: StorageUtils):
        self.task_creator = task_creator
        self.storage_utils = storage_utils # gcs only
        self.target_url = "https://function-1-gizrqdvcaq-uc.a.run.app"

    def meta(self, request):
        try:
            latest_file = self.storage_utils.get_latest_file('root/', "sites.txt")
            print(latest_file)
            urls = self.storage_utils.read_data(latest_file)
            print(urls)
            for url in urls:
                task = self.task_creator.create_task_meta(url, self.target_url)
            return "hopefully created meta task"
        except Exception as exception:
            return str(exception)

    def site(self, request):
        try:
            latest_file = self.storage_utils.get_latest_file('root/', "sites.txt")
            print(latest_file)
            urls = self.storage_utils.read_data(latest_file)
            print(urls)
            for url in urls:
                self.task_creator.create_task_site(url, self.target_url)
            return "hopefully created site task"
        except Exception as exception:
            return str(exception)


    def product(self, request):
        try:
            start_time = datetime.utcnow()

            products_files_per_site = self.storage_utils.get_latest_files('meta/', 'products.txt')
            for products_file_per_site in products_files_per_site:
                site_id = products_file_per_site.split('_')[0]
                site_id = site_id.replace("meta/", "")
                site_info_file = self.storage_utils.get_latest_file('site/', site_id.strip())
                site_info = self.storage_utils.read_data(site_info_file)
                rate_limit = 1  # Number of requests per second

                if len(site_info) and len(site_info[0]):
                    scraper_code_loc = f"scraper_code{site_info[0][-1]}"
                    print(scraper_code_loc)
                    blob = self.storage_utils.bucket.blob(scraper_code_loc)
                    scraper_code = blob.download_as_text()
                    if site_info:
                        urls = self.storage_utils.read_data(products_file_per_site)
                        for url in urls:
                            url = url.replace("https://", "")
                            scheduled_time = start_time + timedelta(seconds=rate_limit)
                            scheduled_timestamp = Timestamp()
                            scheduled_timestamp.FromDatetime(scheduled_time)
                            self.task_creator.create_task_product(url, scraper_code, self.target_url, scheduled_timestamp)
            return "hopefully created", 200
        except Exception as exception:
            return str(exception), 500