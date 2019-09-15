import scrapy
from ..items import BiltongandbudzItem
import datetime
import html5lib
from bs4 import BeautifulSoup


class BiltongandbudzSpider(scrapy.Spider):
    name = "biltongandbudz"

    def __init__(self, *a, **kw):
        super(BiltongandbudzSpider, self).__init__(*a, **kw)
        self.url = "https://www.biltongandbudz.co.za/shop/"
        self.pagenation_str = "page/"

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        product_name = soup.find("h1", {"class": "product-title product_title entry-title"}).getText().strip()
        product_stock = soup.find("p", {"class": "stock in-stock"}).getText().strip()
        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")

        item = BiltongandbudzItem()
        item['name'] = product_name
        item['price'] = product_price
        item['stock'] = product_stock
        item['url'] = response.request.url
        item['date'] = datetime.datetime.utcnow()

        return item


    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"class": "woocommerce-pagination"}).find("ul", {
            "class": "page-numbers nav-pagination links text-center"})
        max_pages = soup.findAll("li")[-2].find('a').getText()
        # max_pages = 2
        return max_pages

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'shop-container'}) \
            .findAll('div', {'class': 'image-fade_in_back'})
        if len(soup) is 0:
            return None
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, cookies={'age_gate': 21, 'tk_ai': 'woo:dBlQRi3iIybLcJZIZaok+wyL'}, callback=self.first_parse)

    def first_parse(self, response):
        max_pages = self.get_max_pages(response)
        urls = ["".join([self.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages) + 1)]
        for url in urls:
            yield scrapy.Request(url=url, cookies={'age_gate': 21, 'tk_ai': 'woo:dBlQRi3iIybLcJZIZaok+wyL'}, callback=self.second_parse)

    def second_parse(self, response):
        products = self.get_products_on_page(response)
        for product in products:
            yield scrapy.Request(url=product, cookies={'age_gate': 21, 'tk_ai': 'woo:dBlQRi3iIybLcJZIZaok+wyL'}, headers={"forever": "true"}, callback=self.third_parse)

    def third_parse(self, response):
        yield self.get_product_data(response)