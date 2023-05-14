import requests
from bs4 import BeautifulSoup

from src.scraper.product.scraper import Scraper, ProductData


class WooCommerceScraper(Scraper):
    """
    A web scraper for WooCommerce stores.

    Args:
        proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        bucket_name (str, optional): The name of the Google Cloud Storage bucket to upload scraped images to.

    Attributes:
        proxies (dict): A dictionary of proxies to be used with the scraper.
        session (requests.Session): A persistent HTTP session to be used for making requests.
    """

    def __init__(self, proxies=None):
        """
        Initialize a new instance of the WooCommerceScraper class.

        Args:
            proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        """
        self.proxies = proxies or {}
        self.session = requests.Session()

            
    def __call__(self, url):
        return self.scrape_product(url)

    def scrape_product(self, product_url):
        """
        Scrape product information from a WooCommerce website.

        Args:
            product_url (str): The URL of the product to be scraped.

        Returns:
            List[ProductData]: A list of ProductData objects representing the scraped product data.
        """
        response = self.session.get(product_url)
        soup = BeautifulSoup(response.text, 'html.parser')
        product_name = soup.find('h1', {'class': 'product_title'}).text.strip()
        product_data_list = self.scrape_product_with_variations(product_url, soup, product_name) \
            if soup.find('form', {'class': 'variations_form'}) \
                else [self.scrape_product_without_variations(product_url, soup, product_name)]
        
        # Convert dictionaries to ProductData instances
        return [ProductData(**product_data) for product_data in product_data_list]


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
