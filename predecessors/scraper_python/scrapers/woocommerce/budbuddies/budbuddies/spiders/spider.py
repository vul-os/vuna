import scrapy
import datetime
import html5lib
import json
from bs4 import BeautifulSoup
from itertools import islice


class BudbuddiesSpider(scrapy.Spider):
    name = "budbuddies"

    def __init__(self, *a, **kw):
        super(BudbuddiesSpider, self).__init__(*a, **kw)
        self.url = "https://budbuddies.co.za/shop/?count=10000"
        self.partition_size = 10

    def get_variation_data(self, json_data, product_name, url):
        for data in json_data:
            variation_id = data['variation_id']
            pack_size = data['attributes']['attribute_pa_pack-size'].replace("-", "").replace("seeds", "").strip()
            product_stock = data['availability_html'].replace('<p class="stock in-stock">', '')\
                .replace('</p>', '').strip().replace('in', '').replace('stock', '').strip()
            product_price = data['display_price']

            item = {}
            item['name'] = product_name
            item['varId'] = variation_id
            item['price'] = product_price
            item['stock'] = product_stock
            item['url'] = url
            item['date'] = datetime.datetime.utcnow()

            yield item

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'summary entry-summary'})
        product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
        print(product_name)
        variations = soup.find('table', {'class': 'variations'})
        if variations is not None:
            raw_json_string = str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations'])
            try:
                json_data = json.loads(raw_json_string)
                print("Variations: ", self.get_variation_data(json_data, response.request.url, product_name))
            except Exception as e:
                print(raw_json_string)

            # return self.get_variation_data(json_data, response.request.url, product_name)
        else:
            product_stock = soup.find("p", {"class": "stock in-stock"})
            product_stock = product_stock.getText().strip() if product_stock else 0
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
            product_price = product_price.getText().replace("R", "") if product_price else 0
            variation_id = soup.find('button', {'class': 'single_add_to_cart_button button alt'})
            if variation_id is None:
                variation_id = soup.find('input', {'class': 'cwg-product-id'})
            variation_id = variation_id['value'] if variation_id is not None else None

            print("No Variations: ", product_name, product_price, product_stock, variation_id,
                  datetime.datetime.utcnow())


    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.products')
        soup = soup.select('.product')
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        products = self.get_products_on_page(response)
        for prod in products:
            yield scrapy.Request(url=prod, callback=self.second_parse)

    def second_parse(self, response):
        self.get_product_data(response)