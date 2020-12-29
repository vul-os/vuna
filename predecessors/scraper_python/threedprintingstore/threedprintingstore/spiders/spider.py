import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup
from bson.objectid import ObjectId
import pymongo
import psycopg2


class ThreeDPrintingStore(scrapy.Spider):
    name = "threedprintingstore"

    def __init__(self, *a, **kw):
        super(ThreeDPrintingStore, self).__init__(*a, **kw)
        self.url = "http://www.3dprintingstore.co.za/sitemap/categories"
        self.pagenation_str = "?page="
        self.scrape_date = datetime.datetime.now()
        self.store_id = 1
        self.connection = psycopg2.connect(
            database="scrapers", user="scrapers", password="scrapers", host="38.17.53.117", port=17435)
        self.cursor = self.connection.cursor()

    def spider_closed(self, spider):
        self.cursor.close()
        self.connection.close()
        spider.logger.info('Spider closed: %s', spider.name)

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "ContentArea"})

        try:
            product_name = soup.find("h1", {"class": "title"}).getText().strip()
            product_price = soup.find("span", {"class": "ProductDetailsPriceIncTax"}).getText()\
                .replace("R", "").replace("(inc VAT)", "").replace(",", "").strip()
            product_stock = soup.find("div", {"class": "DetailRow InventoryLevel"}).\
                find('div', {'class': 'Value'}).getText().replace("Currently out of stock", "").strip()
            product_stock = int(product_stock) if product_stock != '' else 0
        except Exception:
            print("Error")
            return None

        url = response.request.url

        query = f"""
          INSERT INTO products (name, store, url, price, date_added)
          VALUES('{product_name}', {self.store_id}, '{url}', {product_price}, '{datetime.datetime.now()}')
          ON CONFLICT (url) DO UPDATE SET
                name = '{product_name}',
                store = {self.store_id},
                url = '{url}',
                price = {product_price},
                date_added = '{datetime.datetime.now()}'
            RETURNING id;
          """

        self.cursor.execute(query)
        self.connection.commit()

        product_id = self.cursor.fetchone()
        product_id = product_id if not len(product_id) > 0 else product_id[0]
        if product_id:
            print(product_id, product_stock, self.scrape_date, datetime.datetime.now())
            query = f"""
              INSERT INTO datapoints (product, stock, date_scraped, date_added)
              VALUES({product_id}, {product_stock}, '{self.scrape_date}', '{datetime.datetime.now()}')
              RETURNING id;
              """
            self.cursor.execute(query)
            self.connection.commit()

        # result_insert = self.client.scrapers.datapoint.insert_one({
        #   'product': ObjectId(product_id['_id']),
        #   'stock': product_stock,
        #   'date': datetime.datetime.now(),
        #   'scrape_date': self.scrape_date
        # })

        return None

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
            yield scrapy.Request("/".join([response.request.url, self.pagenation_str, str(page)]),
                                 callback=self.third_parse)

    def third_parse(self, response):
        items = self.get_products_on_page(response)
        if items is not None:
            for item in items:
                yield scrapy.Request(item, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_product_data(response)
