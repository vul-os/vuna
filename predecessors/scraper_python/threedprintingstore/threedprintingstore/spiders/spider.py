import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup
from bson.objectid import ObjectId
import pymongo


class ThreeDPrintingStore(scrapy.Spider):
    name = "threedprintingstore"

    def __init__(self, *a, **kw):
        super(ThreeDPrintingStore, self).__init__(*a, **kw)
        self.url = "http://www.3dprintingstore.co.za/sitemap/categories"
        self.pagenation_str = "?page="
        self.store_id = '5fe676a0039b3ee228e3b324'
        self.client = pymongo.MongoClient(
            "mongodb+srv://scraperama:scraperama@cluster0.i0xw4.mongodb.net/scrapers?retryWrites=true&w=majority")
        self.scrape_date = datetime.datetime.now()

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

        result_upsert = self.client.scrapers.products.update_one({
          'name': product_name,
          'store': ObjectId(self.store_id),
          'url': url
        },{
          '$set': {
            'price': product_price,
            'store': ObjectId(self.store_id),
            'url': url,
            'date': datetime.datetime.now()
          }
        }, upsert=True)

        product_id = self.client.scrapers.products.find_one({
            'name': product_name,
            'store': ObjectId(self.store_id),
            'url': url
        })

        result_insert = self.client.scrapers.datapoint.insert_one({
          'product': ObjectId(product_id['_id']),
          'stock': product_stock,
          'date': datetime.datetime.now(),
          'scrape_date': self.scrape_date
        })

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
