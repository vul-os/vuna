# Bring your packages onto the path
import sys, os
import scrapy
import datetime
from pathlib import Path
from bs4 import BeautifulSoup

sys.path.append(str(Path(__file__).parent.parent.parent.parent.parent / Path('utils')))
from utils import connect, upsert_product


class ThreeDPrintingStore(scrapy.Spider):
    name = "threedprintingstore"

    def __init__(self, *a, **kw):
        super(ThreeDPrintingStore, self).__init__(*a, **kw)
        self.url = "http://www.3dprintingstore.co.za/sitemap/categories"
        self.pagenation_str = "?page="
        self.scrape_date = datetime.datetime.now()
        self.store_id = 1
        self.connection = connect()
        self.cursor = self.connection.cursor()

    def spider_closed(self, spider):
        self.cursor.close()
        self.connection.close()
        spider.logger.info('Spider closed: %s', spider.name)

    def get_product_categories(self, soup):
        breadcrumbs = soup.find('div', {'id': 'ProductBreadcrumb'})
        cats = {}
        for cat in breadcrumbs.findAll('ul'):
            crumbs = [cat.findAll('li')]
            if len(crumbs) == 0:
                continue
            cats[crumbs[0][1].find('a').text] = {
                'url': crumbs[0][1].find('a')['href'],
                'subcat': {
                    k.find('a').text: k.find('a')['href'] for i, k in enumerate(crumbs[0][2:-1])
                } if len(crumbs[0]) > 2 else None
            }
        return cats

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        content_area = soup.find("div", {"class": "ContentArea"})

        try:
            product_name = content_area.find("h1", {"class": "title"}).getText().strip()
            product_price = content_area.find("span", {"class": "ProductDetailsPriceIncTax"}).getText() \
                .replace("R", "").replace("(inc VAT)", "").replace(",", "").strip()
            product_stock = content_area.find("div", {"class": "DetailRow InventoryLevel"}). \
                find('div', {'class': 'Value'}).getText().replace("Currently out of stock", "").strip()
            product_stock = int(product_stock) if product_stock != '' else 0
            product_categories = self.get_product_categories(soup)
        except Exception as e:
            print(f"Error Parsing: {e}")
            return None

        result = upsert_product(
            self.connection,
            self.cursor,
            self.store_id,
            product_categories,
            str(product_name),
            str(response.request.url),
            float(product_price),
            int(product_stock),
            self.scrape_date
        )
        return result

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
        if len(soup) == 0:
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
            yield scrapy.Request("/".join([response.request.url, self.pagenation_str, str(page)]),
                                 callback=self.third_parse)

    def third_parse(self, response):
        items = self.get_products_on_page(response)
        if items is not None:
            for item in items:
                yield scrapy.Request(item, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_product_data(response)
