import logging
import base64
import datetime
from typing import List
from requests.sessions import Session
from bs4 import BeautifulSoup
from requests.exceptions import RequestException
from src.storage.storage import StorageUtils
from src.scraper.site.detect_technology import detect
from src.scraper.encoder.encoder import encode_url

logger = logging.getLogger(__name__)

class SiteScraper:
    def __init__(self, job_identifier: str, session: Session = None, storage_utils: StorageUtils = None):
        self.known_urls = []
        self.job_identifier = job_identifier
        self.storage_utils = storage_utils
        self.session = session or Session()

    def __call__(self, base_url: str):
        currrency_code = "ZAR"

        name, image, technology = self.get_site_info(base_url)
        site_id = encode_url(base_url)

        current_datetime = datetime.datetime.now()
        formatted_datetime = current_datetime.strftime("%Y-%m-%d|%H-%M-%S")

        path_prefix = f"site/{self.job_identifier}"
        file_name = f"{path_prefix}/{site_id}_{formatted_datetime}_site.csv"
        items = {
            "id": site_id,
            "name": name.strip() if name else "",
            "image": image.strip() if image else "",
            "currency": currrency_code,
            "technology": technology,
            "scraper_file": f"/{technology}/default.py"
        }
        if self.storage_utils:
            self.storage_utils.write_data_to_csv(file_name, [items])
            return {"file_name": file_name, "site_data": items}
        return items

    def get_site_info(self, url):
        # Send an HTTP GET request to the website and retrieve the HTML content
        response = self.session.get(url)
        html_content = response.content

        # Parse the HTML content using BeautifulSoup
        soup = BeautifulSoup(html_content, "html.parser")

        # Extract the name of the website
        name = soup.title.string
        
        # Extract the image of the website
        image = soup.find("meta", property="og:image")
        image = image["content"] if image else None

        technology = detect(soup)
        # Return the name of the website
        return name, image, technology


    def generate_symbol_map(self):
        # do currencies
        pass
        # import ccy

        # symbol_map = {}
        # currencies = ccy.currencydb()
        # for currency in currencies:
        #     symbol = currencies[currency].symbol
        #     symbol_map[symbol] = currency
        # return symbol_map



