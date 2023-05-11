import uuid
import logging
import requests
from typing import List
from bs4 import BeautifulSoup

from requests.exceptions import RequestException
from src.storage.storage import StorageUtils

logger = logging.getLogger(__name__)

class MetaScraper:
    def __init__(self, storage_utils: StorageUtils = None):
        self.known_urls = []
        self.storage_utils = storage_utils

    def __call__(self, base_url: str):
        try:

            products = self.scrape_products(base_url)
            print(len(set(products)))
        except Exception as exception:
            print(exception)
            return 1

        # hashSiteId = hashStringFromUrl(base_url)
        # if len(products) > 0:
        #     site_info = self.get_site_info(base_url)
        #     if self.storage_utils:
        #         self.storage_utils.upload_csv_from_dict(f"{hashSiteId}-products.csv", products)
        #         self.storage_utils.upload_csv_from_dict(f"{hashSiteId}-site.csv", site_info)
        #     else:
        #         return site_info, products
        # return None

    @staticmethod
    def get_site_info(url):
        # Send an HTTP GET request to the website and retrieve the HTML content
        response = requests.get(url)
        html_content = response.content

        # Parse the HTML content using BeautifulSoup
        soup = BeautifulSoup(html_content, "html.parser")

        # Extract the name of the website
        name = soup.title.string
         # Extract the description of the website
        description = soup.find("meta", property="og:description")["content"]
        # Extract the image of the website
        image = soup.find("meta", property="og:image")["content"]
        # Return the name of the website
        return {"name": name, "description": description, "image": image}


    def parse_sitemaps(self, sitemap_urls):
        """
        Recursively parse sitemaps for URLs.

        Args:
            sitemap_urls (list): A list of sitemap URLs.

        Returns:
            list: A list of known URLs after parsing all sitemaps.
        """
        # Parse sitemaps for URLs
        for sitemap_url in sitemap_urls:
            logger.info('Parsing sitemap: %s', sitemap_url)
            sitemap_text = self._url_to_text(sitemap_url)
            urls = self._parse_sitemap(sitemap_text)
            # Check for nested sitemaps
            nested_sitemap_urls = [url for url in urls if ".xml" in url]
            if nested_sitemap_urls:
                self.parse_sitemaps(nested_sitemap_urls)
            else:
                u = [url for url in urls if 'jpg' not in url and 'cdn' not in url]
                self.known_urls.extend(u)

    def _url_to_text(self, url: str) -> str:
        try:
            response = requests.get(url)
            return response.text
        except RequestException as exception:
            print(exception)
            logger.exception(f"Error retrieving {url}: {exception}")
            return ""

    def _parse_robots_txt(self, text: str) -> List[str]:
        sitemap_urls = []
        for line in text.split('\n'):
            if line.lower().startswith('sitemap:'):
                url = line.split(':', 1)[1].strip()
                sitemap_urls.append(url)
        return sitemap_urls

    def _parse_sitemap(self, text) -> List[str]:
        # soup = BeautifulSoup(content, "lxml", features="xml")
        # soup = BeautifulSoup(html, features='html.parser')
        soup = BeautifulSoup(text, features="xml")

        urls = [loc.text for loc in soup.find_all('loc')]
        return urls

    def _filter_product_urls(self, urls: List[str]) -> List[str]:
        product_urls = []
        for url in urls:
            path = url.split('//', 1)[-1].split('/', 1)[-1]
            if any(keyword in path for keyword in ['product', 'products', 'collections', 'pages']) and not \
                any(keyword in path for keyword in ['product-tag', 'product-category']):
                # Add the base URL if it's not already included in the product URL
                base_url = url.split('/' + path, 1)[0]
                product_url = base_url + '/' + path
                product_urls.append(product_url)
        return product_urls

    def scrape_products(self, base_url: str) -> List[str]:
        # Parse robots.txt
        robots_url = f'{base_url}/robots.txt'
        logger.info('Parsing robots.txt: %s', robots_url)
        robots_text = self._url_to_text(robots_url)
        sitemap_urls = self._parse_robots_txt(robots_text)

        # Add default sitemap URLs
        default_sitemap_urls = [
            f'{base_url}/sitemap.xml',
            f'{base_url}/sitemap_index.xml',
        ]
        sitemap_urls.extend(default_sitemap_urls)
        # Parse sitemaps for URLs
        self.parse_sitemaps(sitemap_urls)
        # Filter URLs for ecommerce product URLs
        product_urls = self._filter_product_urls(self.known_urls)
        return product_urls