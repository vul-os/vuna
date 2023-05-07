import requests
from fake_useragent import UserAgent

class Scraper:
    def __init__(self, url):
        self.url = url
        self.json_url = f"{self.url}.json"
        self.headers = {'User-Agent': UserAgent().random}

    def scrape(self):
        response = requests.get(self.json_url, headers=self.headers)
        if response.ok:
            products = response.json()
            return self.scrape_product(products['product'])
        else:
            return []

    def scrape_product(self, product):
        # Extract product information
        product_id = product['id']
        product_name = product['title']
        product_price = product['variants'][0]['price']
        product_quantity = product['variants'][0]['inventory_quantity']

        # Extract variant information
        variant_info = []
        for variant in product['variants']:
            variant_price = variant['price']
            variant_quantity = variant['inventory_quantity']
            variant_info.append({
                'price': variant_price,
                'quantity': variant_quantity,
            })

        # Extract image information for product (not variant-specific)
        image_urls = [image['src'] for image in product['images']]

        # Return a dictionary of scraped data
        return {
            'product_id': product_id,
            'product_name': product_name,
            'product_price': product_price,
            'product_quantity': product_quantity,
            'variant_info': variant_info,
            'image_urls': image_urls
        }
