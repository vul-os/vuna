import scrapy
from ..items import PishopItem
import datetime
import json
import html5lib
import re
from bs4 import BeautifulSoup


class PishopSpider(scrapy.Spider):
    name = "pishop"

    def __init__(self, *a, **kw):
        super(PishopSpider, self).__init__(*a, **kw)
        self.url = "https://www.pishop.co.za/store/all-products"
        self.pagenation_str = "&page="

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')

        product_name = soup.find("h1", {"class": "productname"}).find('span').getText()
        product_stock = soup.find("div", {"class": "productprice"})\
            .find("label").getText().replace("Availability: ", "")

        soup_ = soup.find("div", {"class": "productfilneprice"}).getText()
        product_price = float(re.search('[0-9,.]+', soup_).group().replace(",", ""))
        product_stock = 0 if 'Out' in product_stock else int(product_stock)

        item = PishopItem()
        item['name'] = product_name
        item['price'] = product_price
        item['stock'] = product_stock
        item['url'] = response.request.url
        item['date'] = datetime.datetime.utcnow()
        yield item

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "contentpanel"})
        soup = soup.find('div', {'class': 'thumbnails grid row list-inline'})\
                   .findAll('div', {"class": "col-md-3 col-sm-6 col-xs-12"})
        if len(soup) is 0:
            return None

        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("ul", {"class": "pagination"})
        soup = str(soup.findAll("li")[-1].find('a')['href'])
        soup = soup.split("&")
        max_page = 1
        for s in soup:
            if 'page=' in s:
                max_page = s.replace("page=", "")
        return int(max_page)

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

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





