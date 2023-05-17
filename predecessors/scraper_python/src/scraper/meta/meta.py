import logging
import requests
import base64
import datetime
from typing import List
from bs4 import BeautifulSoup
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
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

        encoded_site = base64.b64encode(base_url.encode()).decode()
        if len(products) > 0:
            name, image = self.get_site_info(base_url)
            current_datetime = datetime.datetime.now()
            formatted_datetime = current_datetime.strftime("%Y-%m-%d_%H-%M-%S")
            file_name = f"{encoded_site}-{formatted_datetime}-products.csv"
            if self.storage_utils:
                first_items = [
                    f"name: {name.strip()}",
                    f"image: {image.strip()}",
                    f"num_product_urls: {len(products)}"
                ]
                first_items.extend(products)
                self.storage_utils.write_data_to_txt(file_name, first_items)
            else:
                return products
        return None

    @staticmethod
    def get_site_info(url):
        # Send an HTTP GET request to the website and retrieve the HTML content
        response = requests.get(url)
        html_content = response.content

        # Parse the HTML content using BeautifulSoup
        soup = BeautifulSoup(html_content, "html.parser")

        # Extract the name of the website
        name = soup.title.string
        
        # Extract the image of the website
        image = soup.find("meta", property="og:image")
        image = image["content"] if image else None
        # Return the name of the website
        return name, image


    def parse_sitemaps(self, sitemap_urls):
        """
        Parse sitemaps for URLs.

        Args:
            sitemap_urls (list): A list of sitemap URLs.

        Returns:
            list: A list of known URLs after parsing all sitemaps.
        """
        stack = list(sitemap_urls)
        while stack:
            sitemap_url = stack.pop()
            logger.info('Parsing sitemap: %s', sitemap_url)
            sitemap_text = self._url_to_text(sitemap_url)
            urls = self._parse_sitemap(sitemap_text)
            nested_sitemap_urls = [url for url in urls if ".xml" in url]
            if nested_sitemap_urls:
                stack.extend(nested_sitemap_urls)
            else:
                u = [url for url in urls if 'jpg' not in url and 'cdn' not in url]
                self.known_urls.extend(u)
                
    def _url_to_text(self, url: str) -> str:
        try:
            # Create a session
            session = requests.Session()

            # Define the retry behavior
            retry = Retry(total=1, backoff_factor=0, status_forcelist=(),
            raise_on_redirect=True, raise_on_status=True)

            # Mount it for both http and https usage
            adapter = HTTPAdapter(max_retries=retry)
            session.mount('http://', adapter)
            session.mount('https://', adapter)
            response = session.get(url)
            
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

        # Add default sitemap URLs
        sitemap_urls = [
            f'{base_url}/sitemap.xml',
            f'{base_url}/sitemap_index.xml',
        ]
        if robots_text != "":
            sitemap_urls.extend(self._parse_robots_txt(robots_text))

        # Parse sitemaps for URLs
        self.parse_sitemaps(sitemap_urls)
        # Filter URLs for ecommerce product URLs
        product_urls = self._filter_product_urls(self.known_urls)
        return product_urls