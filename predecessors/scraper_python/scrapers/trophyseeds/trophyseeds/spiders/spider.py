import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup
import json

class TrophyseedsSpider(scrapy.Spider):
    name = "trophyseeds"

    def __init__(self, *a, **kw):
        super(TrophyseedsSpider, self).__init__(*a, **kw)
        self.url = "https://www.trophyseeds.com/shop/"
        self.pagenation_str = "page/"

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
            print(item)
        return None

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "summary entry-summary"})
        product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
        variations = soup.find('table', {'class': 'variations'})
        if variations is not None:
            json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
            print("Variations: ", self.get_variation_data(json_data, response.request.url, product_name))
            return self.get_variation_data(json_data, response.request.url, product_name)
        else:
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
            product_stock = soup.find("p", {"class": "stock in-stock"}).getText() \
                .replace("in", "") \
                .replace("stock", "") \
                .strip()
            product_stock = int(product_stock) if int(product_stock.isdigit()) else 0
            product_category = soup.find("span", {"class": "posted_in"}).findAll("a")
            if len(product_category) > 1:
                print("FUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUCK")

            print("No Variations: ", product_name, product_price, product_stock, datetime.datetime.utcnow())

            return None

    @staticmethod
    def get_max_pages(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"class": "woocommerce-pagination"})
        max_pages = soup.findAll("li")[-2].find('a').getText()
        return max_pages

    @staticmethod
    def get_products_on_page(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('main', {'id': 'main'})\
                   .find('ul', {'class': 'products columns-3'})\
                   .findAll('li')
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        max_pages = self.get_max_pages(response)
        urls = ["".join([self.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages) + 1)]
        for url in urls:
            yield scrapy.Request(url=url, callback=self.second_parse)

    def second_parse(self, response):
        products = self.get_products_on_page(response)
        for product in products:
            yield scrapy.Request(url=product, callback=self.third_parse)

    def third_parse(self, response):
        yield self.get_product_data(response)