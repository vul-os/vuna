import scrapy
import datetime
import html5lib
import json
from bs4 import BeautifulSoup


class BotshopSpider(scrapy.Spider):
    name = "botshop"

    def __init__(self, *a, **kw):
        super(BotshopSpider, self).__init__(*a, **kw)
        self.url = "https://www.botshop.co.za/shop/"
        self.pagenation_str = "page/"

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

        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
        product_stock = soup.find("p", {"class": "stock in-stock"})
        product_stock = product_stock \
            .getText() \
            .replace("in", "") \
            .replace("stock", "") \
            .strip()\
        if product_stock is not None else 0
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

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('ul', {'class': 'page-numbers'}) \
            .findAll('a', {'class': 'page-numbers'})
        if len(soup) == 0:
            return 1
        return int(soup[-2].text.strip())

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        max_pages = self.get_max_pages(response)

        print(max_pages)
        for page in range(1, max_pages):
            yield scrapy.Request(url=f"{self.url}{self.pagenation_str}{page}", callback=self.second_parse)

    def second_parse(self, response):
        products = self.get_products_on_page(response)
        for prod in products:
            yield scrapy.Request(url=prod, callback=self.third_parse)

    def third_parse(self, response):
        self.get_product_data(response)