import scrapy
from ..items import ThreeDPrintingStoreItem
import datetime
import html5lib
from bs4 import BeautifulSoup


class ThreeDPrintingStore(scrapy.Spider):
    name = "threedprintingstore"

    def __init__(self, *a, **kw):
        super(ThreeDPrintingStore, self).__init__(*a, **kw)
        self.url = "http://www.3dprintingstore.co.za/sitemap/categories"
        self.pagenation_str = "?page="

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "ContentArea"})

        product_name = soup.find("h1", {"class": "title"}).getText().strip()
        product_price = soup.find("span", {"class": "ProductDetailsPriceIncTax"}).getText().replace("R", "").replace("(inc VAT)", "").replace(",", "").strip()
        product_stock_str = soup.find("div", {"class": "DetailRow InventoryLevel"}).find('div', {'class': 'Value'}).getText().replace("Currently out of stock", "").strip()
        product_stock_str = int(product_stock_str) if product_stock_str != '' else 0

        item = ThreeDPrintingStoreItem()
        item['name'] = product_name
        item['price'] = float(product_price)
        item['stock'] = int(product_stock_str)
        item['url'] = response.request.url
        item['date'] = datetime.datetime.utcnow()

        return item

    def get_categories(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "SitemapCategories"}).find('ul', {'class': ''}).findAll('li', recursive=False)
        links = []
        exclude = ['http://www.3dprintingstore.co.za/categories/training.html']
        for s in soup:
            link = s.find('a')['href']
            if link not in exclude:
                links.append(link)
        return links

    def get_max_pages(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("ul", {"class": "PagingList"})
        if soup is not None:
            max_pages = int(soup.findAll("li")[-1].find('a').getText())
            return max_pages
        return 1

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("ul", {"class": "ProductList"}).findAll('li', recursive=False)
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_categories(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.second_parse)

    def second_parse(self, response):
        max_pages = self.get_max_pages(response)
        for page in range(1, max_pages + 1):
            yield scrapy.Request("/".join([response.request.url, self.pagenation_str, str(page)]), callback=self.third_parse)

    def third_parse(self, response):
        items = self.get_products_on_page(response)
        if items is not None:
            for item in items:
                yield scrapy.Request(item, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_product_data(response)







