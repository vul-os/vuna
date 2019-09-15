import scrapy
from ..items import NetramItem
import datetime
import html5lib
from bs4 import BeautifulSoup


class DiyelectronicsSpider(scrapy.Spider):
    name = "netram"

    def __init__(self, *a, **kw):
        super(DiyelectronicsSpider, self).__init__(*a, **kw)
        self.url = "https://www.netram.co.za"

    def get_categories(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"id": "cbp-hrmenu1"}).find('ul').findAll('li', recursive=False)
        links = []
        exclude = ['https://www.netram.co.za/931-products-of-the-week']
        for s in soup:
            link = s.find('a')['href']
            if link not in exclude:
                links.append(link)
        return links

    def get_all_page_url(self, response):
        url = response.request.url
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        id_cat = soup.find('input', {"name": "id_category"})
        n = soup.find('input', {"name": "n"})
        if id_cat is None or n is None:
            return None
        url = f"{url}?id_category={id_cat['value']}&n={n['value']}"
        return url

    def get_all_links_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find('div', {'id': 'center_column'})\
            .find('ul', {"class": "product_list grid row"})\
            .findAll('li')
        product_links = []
        for s in soup:
            data = s.find('a', {'class': 'product-name'})
            if data is not None:
                link = data['href'].strip()
                product_links.append(link)
        return product_links

    def get_all_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        product_name = soup.find("h1", {"itemprop": "name"}).getText().strip()
        product_price = soup.find("span", {"id": "our_price_display"})['content']
        product_stock = soup.find("span", {"id": "quantityAvailable"}).getText()

        item = NetramItem()
        item['name'] = product_name
        item['price'] = product_price
        item['stock'] = product_stock
        item['url'] = response.request.url
        item['date'] = datetime.datetime.utcnow()

        return item

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_categories(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.second_parse)

    def second_parse(self, response):
        _url = self.get_all_page_url(response) if self.get_all_page_url(response) is not None else response.request.url
        yield scrapy.Request(url=_url, callback=self.third_parse)

    def third_parse(self, response):
        links = self.get_all_links_page(response)
        for link in links:
            yield scrapy.Request(url=link, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_all_data(response)




