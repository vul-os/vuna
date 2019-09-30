import scrapy
from ..items import FlyingrobotItem
import datetime
import html5lib
from bs4 import BeautifulSoup


class FlyingRobotSpider(scrapy.Spider):
    name = "flyingrobot"

    def __init__(self, *a, **kw):
        super(FlyingRobotSpider, self).__init__(*a, **kw)
        self.url = "https://flyingrobot.co/"
        self.pagenation_str = "?page="

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        product_name = soup.find("h1", {"class": "product-single__title"})
        product_name = product_name.getText()
        product_price = soup.find("span", {"class": "firstPriceValue"}).getText().replace("R", "")
        product_stock_str = soup.find("span", {"id": "ProductStock-product-template"}).getText()\
                                                                                      .replace("available", "")\
                                                                                      .strip()
        product_stock = int(product_stock_str) if int(product_stock_str.isdigit()) else 0

        item = FlyingrobotItem()
        item['name'] = product_name
        item['price'] = product_price
        item['stock'] = product_stock
        item['url'] = response.request.url
        item['date'] = datetime.datetime.utcnow()

        return item

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "pagination"})
        if soup is None:
            return 1
        max_pages = int(soup.findAll("span")[-2].find('a').getText())
        return max_pages

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.findAll('a', {'class': 'product-card'})

        if len(soup) is 0:
            return None
        product_links = (str(links['href']) for links in soup)
        return product_links

    def get_categories(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("ul", {"class": "drawer__nav"}).findAll("li", {'class': 'drawer__nav-item'})
        links = []
        if len(soup) > 0:
            for s in soup:
                link = s.find('a')['href']
                links.append("".join([self.url, link]))
        return links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_categories(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.second_parse)

    def second_parse(self, response):
        max_pages = self.get_max_pages(response)
        for page in range(1, max_pages + 1):
            yield scrapy.Request(url="".join([response.request.url, self.pagenation_str, str(page)]), callback=self.third_parse)

    def third_parse(self, response):
        products = self.get_products_on_page(response)
        for product in products:
            yield scrapy.Request(url="".join([self.url, product]), callback=self.third_parse)

    def fourth_parse(self, response):
        yield self.get_product_data(response)



