import requests
from google.cloud import storage
from src.scraper import MetaScraper, ProductScraper, ScraperLoader
from src.storage.local import StorageUtils

class ScraperAPI:
    def __init__(self):
        # self.data_storage_utils = StorageUtils(storage_client=self.storage_client, bucket_name="mybucket")
        # self.scraper_storage_utils = StorageUtils(storage_client=self.storage_client, bucket_name="mybucket")
        # self.storage_client = storage.Client()

        self.data_storage_utils = StorageUtils(local_dir="/workspace/scraper_python/src/scraper/product/examples")

        self.scraper_cache = {}
        self.proxies = [""]

    def root(self, request):
        return {"message": "Hello, World"}

    def meta(self, request, base_url):
        try:
            url = f"https://{requests.utils.unquote(base_url)}"
        
            scraper = MetaScraper(self.data_storage_utils)
            meta_data = scraper(base_url=url)
        
            return "Hello"
            # return {"meta_data": meta_data, "len": len(meta_data)}

        except Exception as exception:
            return str(exception), 500

    def product_scrape(self, request, product_url, scraper_code):
        try:
            proxies = request.json.get("proxies", [])
            scraper_filename = request.json.get("scraper_filename")
       
            product_url = f"https://{requests.utils.unquote(product_url)}"
            scraper = ScraperLoader(scraper_code)
            product_scraper = ProductScraper(scraper=scraper, proxies=self.proxies,
                                        storage_utils=self.data_storage_utils)
            product_data = product_scraper(product_url=product_url)
        
            return product_data

        except Exception as ex:
            return str(ex), 500