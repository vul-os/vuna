import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup


class DiyelectronicsSpider(scrapy.Spider):
    name = "diyelectronics"

    def __init__(self, *a, **kw):
        super(DiyelectronicsSpider, self).__init__(*a, **kw)
        self.url = "https://www.diyelectronics.co.za/store/"

    def get_categories(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"id": "cbp-hrmenu"}).find('ul').findAll('li', recursive=False)
        links = []
        exclude = ['https://www.diyelectronics.co.za/store/', 'https://www.diyelectronics.co.za/store/209-featured']
        for s in soup:
            link = s.find('a')['href']
            if link not in exclude:
                links.append(link)
        return links

    def get_all_page_url(self, response):
        url = response.request.url
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        id_cat = soup.find('input', {"name": "id_category"})['value']
        n = soup.find('input', {"name": "n"})['value']
        url = f'{url}?id_category={id_cat}&n={n}'
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

    #2090

    def get_all_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        product_name = soup.find("div", {"class": "product-title"}).find("h1").getText().strip()
        product_price = soup.find("span", {"id": "our_price_display"}).getText().strip().replace("R", "").replace(",", "").strip()
        product_stock = soup.find("span", {"id": "quantityAvailable"}).getText()

        print(product_name, product_price, product_stock)

        return None


    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_categories(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.second_parse)

    def second_parse(self, response):
        yield scrapy.Request(url=self.get_all_page_url(response), callback=self.third_parse)

    def third_parse(self, response):
        links = self.get_all_links_page(response)
        for link in links:
            yield scrapy.Request(url=link, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_all_data(response)



