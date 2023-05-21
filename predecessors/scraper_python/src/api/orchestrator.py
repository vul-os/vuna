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
            products_files = self.storage_utils.get_latest_files('meta/', 'products.txt')
            for product_file in products_files:
                site_id = product_file.split('_')[0]
                site_info_file = self.storage_utils.get_latest_file('site/', site_id)
                site_info = self.storage_utils.read_data(site_info_file)
                if "scraper_file" in site_info.keys():
                    scraper_code_loc = f"scraper_code/{site_info["scraper_file"]}"
                    blob = self.storage_utils.bucket.blob(scraper_code_loc)
                    scraper_code = blob.download_as_text()
                    if site_info:
                        urls = self.storage_utils.read_data(product_file)
                        for url in urls:
                            self.task_creator.create_task_product(url, scraper_code, target_url)
            return "hopefully created", 200
        except Exception as exception:
            return str(exception), 500