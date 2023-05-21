from src.storage.local import StorageUtilsLocal as StorageUtils
from src.storage.gcs import StorageUtils
from src.orchestrator.tasks import TaskCreator

class OrchestratorAPI:
    def __init__(self, task_creator: TaskCreator, storage_utils: StorageUtils):
        self.task_creator = task_creator
        self.storage_utils = storage_utils # gcs only

    def meta(self, request):
        try:
            latest_file = self.storage_utils.get_latest_file(None, "sites.txt")
            print(latest_file)
            urls = self.storage_utils.read_data('csv', latest_file)
            print(urls)
            for url in urls:
                self.task_creator.create_task_meta(url, "https://function-1-gizrqdvcaq-uc.a.run.app")
            return "hopefully created meta task"
        except Exception as exception:
            return str(exception)

    def site(self, request):
        try:
            latest_file = self.storage_utils.get_latest_file(None, "sites.txt")
            print(latest_file)
            urls = self.storage_utils.read_data('csv', latest_file)
            print(urls)
            for url in urls:
                self.task_creator.create_task_site(url, "https://function-1-gizrqdvcaq-uc.a.run.app")
            return "hopefully created site task"
        except Exception as exception:
            return str(exception)


    # def product(self, request):
    #     try:
    #         products_files = self.storage_utils.get_latest_files('products', 'products.txt')

    #         for url in urls:
    #             self.task_creator.create_task_site(url)
    #         return "hopefully created", 200
    #     except Exception as exception:
    #         return str(exception), 500