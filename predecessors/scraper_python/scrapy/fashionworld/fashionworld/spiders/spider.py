import scrapy
from ..items import FashionworldItem
import datetime
import json
import html5lib
from bs4 import BeautifulSoup


class FashionworldSpider(scrapy.Spider):
    name = "fashionworld"

    def __init__(self, *a, **kw):
        super(FashionworldSpider, self).__init__(*a, **kw)
        self.url = "https://www.fashionworld.co.za/products"
        self.pagenation_str = "?page="

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')

        product_data = soup.find("div", {"data-component": "productDetails"})['data-product-variations']
        product_name_ = soup.find("div", {"class": "small-12 columns title-price no-padding"}).find('h1')
        product_name = product_name_.getText().strip() if product_name_ is not None else None
        if product_name_ is None or product_data is None:
            return None

        for data in json.loads(product_data):
            item = FashionworldItem()
            item['name'] = product_name
            item['price'] = data['price']
            item['salePrice'] = data['salePrice']
            item['sizeId'] = data['sizeId']
            item['sku'] = data['sku']
            item['stock'] = data['stock']
            item['url'] = response.request.url
            item['date'] = datetime.datetime.utcnow()
            yield item

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'large-9 columns'})\
                   .findAll('div', {'class': 'columns block product'})
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("ul", {"class": "pagination"})
        soup = soup.find("li", {"class": "current"}).getText().replace("You're on page", "").strip()
        max_page = int(soup)
        return max_page

    def start_requests(self):
        new_url = "".join([self.url, self.pagenation_str, str(9999)])
        yield scrapy.Request(url=new_url, callback=self.first_parse)

    def first_parse(self, response):
        max_pages = self.get_max_pages(response)
        for page in range(1, max_pages + 1):
            new_url = "".join([self.url, self.pagenation_str, str(page)])
            yield scrapy.Request(url=new_url, callback=self.second_parse)

    def second_parse(self, response):
        products = self.get_products_on_page(response)
        if products is not None:
            for product in products:
                yield scrapy.Request(url=product, callback=self.third_parse)

    def third_parse(self, response):
        return self.get_product_data(response)





