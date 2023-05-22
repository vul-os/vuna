import requests
from requests import Session
from flask import jsonify
from src.scraper import MetaScraper, ScraperLoader, scrape_product_data, SiteScraper
from src.storage.local import StorageUtilsLocal as StorageUtils

class ScraperAPI:
    def __init__(self, data_storage_utils: StorageUtils):
        self.data_storage_utils = data_storage_utils
        self.scraper_cache = {}
        self.proxies = [""]

    def meta(self, request, base_url):
        try:
            url = f"https://{requests.utils.unquote(base_url)}"
        
            scraper = MetaScraper(storage_utils=self.data_storage_utils)
            meta_data = scraper(base_url=url)
        
            return jsonify(meta_data), 200
            # return {"meta_data": meta_data, "len": len(meta_data)}

        except Exception as exception:
            return str(exception), 500

    def site(self, request, base_url):
        try:
            url = f"https://{requests.utils.unquote(base_url)}"
        
            scraper = SiteScraper(storage_utils=self.data_storage_utils)
            site_data = scraper(base_url=url)
        
            return jsonify(site_data), 200
            # return {"meta_data": meta_data, "len": len(meta_data)}

        except Exception as exception:
            return str(exception), 500

    def product(self, request, product_url):
        try:
            proxies = request.json.get("proxies", None)
            scraper_code = request.json.get("scraper_code", None)

            session = Session()    
            if proxies:
                session.proxies.update(proxies)

            product_url = f"https://{requests.utils.unquote(product_url)}"
            scraper_loader = ScraperLoader(scraper_code)
            product_data = scrape_product_data(scraper_loader=scraper_loader,
                                        session=session,
                                        storage_utils=self.data_storage_utils,
                                        product_url=product_url)
            if product_data:
                return jsonify(product_data), 200
            return "no product data"

        except Exception as ex:
            return str(ex), 500