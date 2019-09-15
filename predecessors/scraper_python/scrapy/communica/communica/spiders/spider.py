import scrapy
from ..items import CommunicaItem
import datetime
import time
import math
import json


class CommunicaSpider(scrapy.Spider):
    name = "communica"

    def __init__(self, *a, **kw):
        super(CommunicaSpider, self).__init__(*a, **kw)
        self.props = {
            "sort": "best-selling",
            "display": "grid",
            "product_available": "false",
            "variant_available": "false",
            "build_filter_tree": "false",
            "check_cache": "false",
            "sort_first": "available",
            "callback": "BCSfFilterCallback&event_type=page",
        }
        self.limit = 70
        self.base_url = "https://services.mybcapps.com/bc-sf-filter/filter"
        self.max_pages = 0
        self.cur_time = time.time()

    def get_product_data(self, product):
        var = product["variants"][0]
        name = var['sku']
        price = var['price']
        avail = var['available']
        stock = var['inventory_quantity']

        item = CommunicaItem()
        item['name'] = name
        item['price'] = price
        item['avail'] = avail
        item['stock'] = stock
        item['date'] = datetime.datetime.utcnow()

        return item

    def get_pages_url(self, url, cur_time, page, limit=70):
        _props = {**{
            "t": str(int(cur_time)),
            "shop": "communica-south-africa.myshopify.com",
            "page": str(page),
            "limit": str(limit)
        }, **self.props}
        _props_str = '&'.join([f'{key}={value}' for key, value in _props.items()])
        _url = f"{url}?{_props_str}"
        return _url

    def get_max_pages(self, page):
        num_products = int(page['total_product'])
        num_pages = math.ceil(num_products / self.limit)
        return num_pages

    def load_page(self, response):
        text = response.body
        text = text.decode("utf-8")

        str_list = [
            "/**/",
            "typeof",
            "BCSfFilterCallback === 'function' && BCSfFilterCallback("
        ]
        for st in str_list:
            text = text.replace(st, "")
        ret = text[:-2]
        if ret is not None and len(ret) > 0:
            return json.loads(ret)
        else:
            return None

    def start_requests(self):
        url = self.get_pages_url(self.base_url, self.cur_time, 1)
        yield scrapy.Request(url=url, callback=self.first_parse)

    def first_parse(self, response):
        first_page = self.load_page(response)
        max_pages = self.get_max_pages(first_page)
        pages = [self.get_pages_url(self.base_url, self.cur_time, pg) for pg in range(1, max_pages + 1)]
        for page in pages:
            yield scrapy.Request(url=page, callback=self.parse)

    def parse(self, response):
        page = self.load_page(response)
        for products in page['products']:
            yield self.get_product_data(products)
