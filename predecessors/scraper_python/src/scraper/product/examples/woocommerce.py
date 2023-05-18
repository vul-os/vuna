import json
from bs4 import BeautifulSoup
from requests.sessions import Session
from src.scraper.product.scraper import Scraper, ProductData
from src.scraper.product.utils import price_to_float, max_qty_to_int


class TheScraper(Scraper):
    """
    A web scraper for WooCommerce stores.

    Args:
        proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        bucket_name (str, optional): The name of the Google Cloud Storage bucket to upload scraped images to.

    Attributes:
        proxies (dict): A dictionary of proxies to be used with the scraper.
        session (requests.Session): A persistent HTTP session to be used for making requests.
    """

    def __init__(self, session: Session = None):
        """
        Initialize a new instance of the WooCommerceScraper class.

        Args:
            proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        """
        super().__init__()
        self.session = session or Session()

            
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
        product_id_tag = soup.find('input', {'name': 'add-to-cart'}) or soup.find('button', {'name': 'add-to-cart'})
        product_id = product_id_tag['value'] if product_id_tag else None

   
        product_data_list = self.scrape_product_with_variations(product_url, product_id, product_name, soup) \
            if soup.find('form', {'class': 'variations_form'}) \
                else [self.scrape_product_without_variations(product_url, product_id, product_name, soup)]
        
        # Convert dictionaries to ProductData instances
        return [ProductData(**product_data) for product_data in product_data_list]


    def scrape_product_without_variations(self, product_url, product_id, product_name, soup):
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
        summary_div = soup.find("div", {"class": "summary"})
        price = summary_div.find('span', {'class': 'woocommerce-Price-amount amount'}).text.strip()
        sku = summary_div.find('span', {'class': 'sku'}).text.strip() if soup.find('span', {'class': 'sku'}) else ''
        max_qty = summary_div.find('p', {'class': 'stock'}).text.strip() if soup.find('p', {'class': 'stock'}) else ''
        
        price = price_to_float(price)
        max_qty = max_qty_to_int(max_qty)
        
        image_url = soup.find('div', {'class': 'woocommerce-product-gallery__image'}).find('img')['src']
        return {
            'name': product_name,
            'url': product_url,
            'image_url': image_url,
            'sku': sku,
            'product_id': product_id,
            'variation_id': None,
            'price': price,
            'max_qty': max_qty,
        }

    def scrape_product_with_variations(self, product_url, product_id, product_name, soup):
        product_variations = soup.find('form', class_='variations_form')['data-product_variations']
        variations_data = json.loads(product_variations)

        product_info = []

        for variation in variations_data:
            availability_html = variation['availability_html']
            display_price = variation['display_price']
            image_url = variation['image']['src']
            sku = variation['sku']
            variation_id = variation['variation_id']

            price = price_to_float(display_price)
            max_qty = max_qty_to_int(availability_html)

            attributes = variation['attributes']
            first_key = next(iter(attributes))
            first_value = attributes[first_key]

            product_info.append({
                'name': first_value if first_value else product_name,
                'url': product_url,
                'image_url': image_url,
                'sku': sku,
                'product_id': product_id,
                'variation_id': variation_id,
                'price': price,
                'max_qty': max_qty,
            })

        return product_info