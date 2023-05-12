import requests
from bs4 import BeautifulSoup
import json
from fake_useragent import UserAgent

class Scraper:
    """
    A web scraper for WooCommerce stores.

    Args:
        proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        bucket_name (str, optional): The name of the Google Cloud Storage bucket to upload scraped images to.

    Attributes:
        proxies (dict): A dictionary of proxies to be used with the scraper.
        session (requests.Session): A persistent HTTP session to be used for making requests.
        headers (dict): HTTP headers to be included with requests made by the scraper.
        bucket_name (str): The name of the Google Cloud Storage bucket to upload scraped images to.
        gcs_uploader (GCSUploader): An instance of the GCSUploader class for uploading scraped images 
        to Google Cloud Storage.

    """

    def __init__(self, proxies=None, bucket_name=None):
        """
        Initialize a new instance of the WooCommerceScraper class.

        Args:
            proxies (dict, optional): A dictionary of proxies to be used with the scraper.
            bucket_name (str, optional): The name of the Google Cloud Storage bucket to upload scraped images to.

        """
        self.proxies = proxies or {}
        self.session = requests.Session()
        self.session.proxies = self.proxies
        self.headers = {
            'User-Agent': UserAgent().random
        }
        self.bucket_name = bucket_name
        if self.bucket_name:
            self.gcs_uploader = GCSUploader(bucket_name)
            
    def __call__(self, url):
        return self.scrape_product(url)

    def scrape_product(self, product_url):
        """
        Scrape product information from a WooCommerce website.

        Args:
            product_url (str): The URL of the product to be scraped.

        Returns:
            str: A JSON string containing the scraped product information.

        """
        response = self.session.get(url, headers=self.headers)
        soup = BeautifulSoup(response.text, 'html.parser')
        product_name = soup.find('h1', {'class': 'product_title'}).text.strip()
        product_data = self.scrape_product_with_variations(url, soup, product_name) \
            if soup.find('form', {'class': 'variations_form'}) \
                else self.scrape_product_without_variations(product_url, soup, product_name)
        return json.dumps(product_data)

    def scrape_product_without_variations(self, product_url, soup, product_name):
        """
        Scrape a product without variations.

        Args:
            product_url (str): The URL of the product to be scraped.
            soup (bs4.BeautifulSoup): A BeautifulSoup object representing the HTML of the product page.
            product_name (str): The name of the product being scraped.

        Returns:
            dict: A dictionary representing the product with keys 'name', 'url', 'sku', 'price', 
            'max_qty', 'attributes', and 'image_url'.

        """
        price = soup.find('span', {'class': 'woocommerce-Price-amount amount'}).text.strip()
        if '-' in price:
            price_range = [float(x.strip().replace('$', '').replace(',', '')) for x in price.split('-')]
            price = sum(price_range) / len(price_range)
        else:
            price = float(price.replace('$', '').replace(',', ''))
        sku = soup.find('span', {'class': 'sku'}).text.strip() if soup.find('span', {'class': 'sku'}) else ''
        max_qty = soup.find('p', {'class': 'stock'}).text.strip() if soup.find('p', {'class': 'stock'}) else ''
        if isinstance(max_qty, str):
            max_qty = max_qty.replace('in', '').replace('stock', '').strip()
        max_qty = int(max_qty) if max_qty.isdigit() else None
        image_url = soup.find('div', {'class': 'woocommerce-product-gallery__image'}).find('img')['src']
        return {
            'name': product_name,
            'url': product_url,
            'sku': sku,
            'price': price,
            'max_qty': max_qty,
            'attributes': [],
            'image_url': image_url
        }

    def scrape_product_with_variations(self, product_url, soup, product_name):
        """
        Scrape a product with multiple variants and return a list of product variants.

        Args:
            product_url (str): The URL of the product to be scraped.
            soup (bs4.BeautifulSoup): A BeautifulSoup object representing the HTML of the product page.
            product_name (str): The name of the product being scraped.

        Returns:
            list: A list of dictionaries, where each dictionary represents a product variant and has 
            the keys 'name', 'url', 'sku', 'price', 'max_qty', 'attributes', and 'image_url'.

        """
        price = soup.find('span', {'class': 'woocommerce-Price-amount amount'}).text.strip()
        if '-' in price:
            price_range = [float(x.strip().replace('$', '').replace(',', '')) for x in price.split('-')]
            price = sum(price_range) / len(price_range)
        else:
            price = float(price.replace('$', '').replace(',', ''))
        sku = soup.find('span', {'class': 'sku'}).text.strip() if soup.find('span', {'class': 'sku'}) else ''
        max_qty = soup.find('p', {'class': 'stock'}).text.strip() if soup.find('p', {'class': 'stock'}) else ''
        if isinstance(max_qty, str):
            max_qty = max_qty.replace('in', '').replace('stock', '').strip()
        max_qty = int(max_qty) if max_qty.isdigit() else None
        image_url = soup.find('div', {'class': 'woocommerce-product-gallery__image'}).find('img')['src']
        return [{
            'name': product_name,
            'url': product_url,
            'sku': sku,
            'price': price,
            'max_qty': max_qty,
            'attributes': [],
            'image_url': image_url
        }]
