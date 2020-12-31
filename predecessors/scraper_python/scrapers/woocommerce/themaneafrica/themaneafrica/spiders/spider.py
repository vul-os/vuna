import scrapy
import datetime
import html5lib
import json
from bs4 import BeautifulSoup

class SacredSeedsSpider(scrapy.Spider):
    name = "themaneafrica"

    def __init__(self, *a, **kw):
        super(SacredSeedsSpider, self).__init__(*a, **kw)
        self.url = "https://thehighco.co.za/"
        # self.pagenation_str = "page/"

    def get_variation_data(self, json_data, product_name, url):
        for data in json_data:
            variation_id = data['variation_id']
            product_stock = data['max_qty']
            product_price = data['display_price']

            print(product_name, product_price, product_stock, variation_id, url)

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'summary entry-summary'})
        product_name = soup.find("h1", {"class": "product_title entry-title"}) \
            .getText().strip()
        variations = soup.find('table', {'class': 'variations'})
        if variations is not None:
            json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
            self.get_variation_data(json_data, response.request.url, product_name)
            # return self.get_variation_data(json_data, response.request.url, product_name)
        else:
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
            if product_price is None:
                return None
            product_price = product_price.getText().replace("R", "")
            product_stock = soup.find("div", {"class": "quantity"})
            if product_stock is None:
                product_stock = 0
            else:
                product_stock = product_stock.find("input", {"title": "Qty"})
                if product_stock is None:
                    product_stock = 0
                else:
                    if len(product_stock) > 0:
                        product_stock = product_stock['max']
                        product_stock = int(product_stock) if product_stock != '' else 0

            # product_short_disc = soup.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
            # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
            # product_category_link = product_category_['href']
            # product_category = product_category_.getText()

            print(product_name, product_price, product_stock)

        return None

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.products')
        soup = soup.select('.product')
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    def get_categories(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        category_links = [s['href'] for s in
                soup.find('ul', {'class': 'sub-menu'}).findAll('a', {'class': 'fusion-bar-highlight'})]
        return category_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_categories(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.second_parse)

    def second_parse(self, response):
        products = self.get_products_on_page(response)
        for prods in products:
            yield scrapy.Request(url=prods, callback=self.third_parse)

    def third_parse(self, response):
        self.get_product_data(response)
