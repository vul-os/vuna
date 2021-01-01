import datetime
import html5lib
from bs4 import BeautifulSoup
import json
from scrapy.linkextractors import LinkExtractor
from scrapy.spiders import SitemapSpider, CrawlSpider, Rule
from scrapy import Request


class CarscozaSpider(SitemapSpider, CrawlSpider):
    name = "carscoza"
    # rules = ( Rule(LinkExtractor(allow=('', )), callback='parse_item', follow=True), )
    # sitemap_rules = [('/for-sale/', 'parse_car_data')]
    base_url = 'https://cars.co.za'
    sitemap_urls = [f'{base_url}/sitemap.xml']
    start_urls = [base_url]

    def parse_item(self, response):
        if 'for-sale' in response.request.url:
            print( response.request.url)
    #
    # def parse_car_data(self, response):
    #     print("here")
        # soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        # car_name = soup.find('h1', {'class': 'heading heading_size_xl'})
        # car_price = soup.find('div', {'class': 'price price_view vehicle-view__price'})
        #
        # print(car_name, car_price)