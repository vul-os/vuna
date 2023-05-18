import requests
from flask import jsonify
from src.scraper import MetaScraper, ScraperLoader, scrape_product_data, SiteScraper
from src.storage.local import StorageUtilsLocal as StorageUtils

class OrchestratorAPI:
    def __init__(self):
        pass
    
    def meta(self, request):
        try:
            
        except Exception as exception:
            return str(exception), 500

    def site(self, request):
        try:

        except Exception as exception:
            return str(exception), 500

    def product(self, request):
        try:

        except Exception as ex:
            return str(ex), 500