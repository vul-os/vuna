import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup


class MarajuanasaSpider(scrapy.Spider):
    name = "marajuanasa"

    def __init__(self, *a, **kw):
        super(MarajuanasaSpider, self).__init__(*a, **kw)
        self.url = "https://marijuanasa.co.za/shop/"
        self.pagenation_str = "page/"
        self.count = 0

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "summary entry-summary"})

        product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
        product_stock_str = soup.find("p", {"class": "stock in-stock"})
        product_stock_ = product_stock_str.getText() \
                             .replace("in", "") \
                             .replace("stock", "") \
                             .replace("(can be backordered)", "") \
                             .strip() if product_stock_str is not None else 0

        product_stock = int(product_stock_) if product_stock_ != '' else 0

        # self.count = self.count + 1
        print(product_name, product_price, product_stock, response.request.url, datetime.datetime.utcnow())
        return None

    # @staticmethod
    # def get_max_pages(response):
    #     soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    #     soup = soup.find("nav", {"class": "woocommerce-pagination"})
    #     max_pages = soup.findAll("li")[-2].find('a').getText()
    #     return max_pages

    @staticmethod
    def get_products_on_page(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('main', {'id': 'main'})\
                   .find('ul', {'class': 'products columns-4'})\
                   .findAll('li')
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    # def first_parse(self, response):
    #     max_pages = self.get_max_pages(response)
    #     urls = ["".join([self.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages) + 1)]
    #     for url in urls:
    #         yield scrapy.Request(url=url, callback=self.second_parse)

    def first_parse(self, response):
        products = self.get_products_on_page(response)
        for product in products:
            yield scrapy.Request(url=product, callback=self.third_parse)

    def third_parse(self, response):
        yield self.get_product_data(response)

    # def closed(self, reason):
    #     print(self.count)