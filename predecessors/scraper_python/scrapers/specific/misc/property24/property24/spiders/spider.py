import scrapy
import datetime
import html5lib
import json
from bs4 import BeautifulSoup

class SacredSeedsSpider(scrapy.Spider):
    name = "property24"

    def __init__(self, *a, **kw):
        super(SacredSeedsSpider, self).__init__(*a, **kw)
        self.url = "https://www.property24.com/"
        # self.pagenation_str = "page/"

    def get_citys(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'col-xs-8'})
        soup = soup.findAll('a')
        urls = []
        for a in soup:
            if a.find('a', {'class': 'p24_bold'}):
                continue
            urls.append(a['href'])
        return urls

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('ul', {'class': 'pagination'})
        if soup is None:
            return 1
        soup = soup.findAll('li')
        if soup is None:
            return 1
        max_pages = soup[-1].find('a')['data-pagenumber']
        return max_pages

    def get_properties(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'class': 'col-xs-9'})
        soup = soup.findAll('a', {'class': ''})
        urls = []
        for a in soup:
            if a.find('span', {'class': 'js_listingTileImageHolder p24_image'}):
                urls.append(f"https://www.property24.com{a['href']}")
        return urls

    def get_prop_data(self, response):
        # print("yay")
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('a', {'class': 'js_displayMap p24_address'})
        if soup is not None:
            print(soup.text)

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        citys = self.get_citys(response)
        for city in citys:
            print(city)
            yield scrapy.Request(url=f"{self.url}{city}", callback=self.second_parse)

    def second_parse(self, response):
        print(response.request.url)
        max_pages = int(self.get_max_pages(response))
        for page in range(1, max_pages+1):
            yield scrapy.Request(url=f"{response.request.url}/p{page}", callback=self.third_parse)

    def third_parse(self, response):
        self.get_prop_data(response)


