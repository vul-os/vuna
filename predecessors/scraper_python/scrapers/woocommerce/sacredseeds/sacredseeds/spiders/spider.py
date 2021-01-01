import scrapy
import datetime
import html5lib
from bs4 import BeautifulSoup
import json

class SacredSeedsSpider(scrapy.Spider):
    name = "sacredseeds"

    def __init__(self, *a, **kw):
        super(SacredSeedsSpider, self).__init__(*a, **kw)
        self.url = "https://sacredseeds.co.za/"
        # self.pagenation_str = "page/"

    def get_variation_data(self, json_data, product_name, url):
        for i, data in enumerate(json_data):
            variation_id = data['variation_id']
            product_price = data['display_price']
            product_stock_max_qty = data['max_qty']
            product_stock_max_qty = int(product_stock_max_qty) if str(product_stock_max_qty).isdigit() else 0
            product_stock_avail = data['availability_html'].replace('<p class="stock in-stock">', '') \
                .replace('</p>', '').strip().replace('in', '').replace('stock', '').strip()
            product_stock_avail = int(product_stock_avail) if product_stock_avail.isdigit() else 0
            product_stock = max(product_stock_avail, product_stock_max_qty)
            # pack_size = data['attributes']['attribute_pa_pack-size'].replace("-", "").replace("seeds", "").strip()

            print(f"Variations {i}: ", product_price, product_stock, variation_id, datetime.datetime.utcnow())

    def get_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.summary,.entry-summary,.product-summary,.product-info')
        product_name = soup.select_one('.product_title,.entry-title').getText().strip()
        variations = soup.find('table', {'class': 'variations'})
        if variations is not None:
            raw_json_string = str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations'])
            json_data = None
            try:
                json_data = json.loads(raw_json_string)
            except Exception as e:
                print(raw_json_string)
            if json_data:
                self.get_variation_data(json_data, product_name, response.request.url)
            # return self.get_variation_data(json_data, response.request.url, product_name)
        else:
            product_stock = soup.find("p", {"class": "stock in-stock"})
            product_stock = product_stock.getText().strip() if product_stock else 0
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
            product_price = product_price.getText().replace("R", "") if product_price else 0
            variation_id = soup.find('button', {'class': 'single_add_to_cart_button button alt'})
            if variation_id is None:
                variation_id = soup.find('input', {'class': 'cwg-product-id'})
            variation_id = variation_id['value'] if variation_id is not None else None

            print("No Variations: ", product_name, product_price, product_stock, variation_id,
                  datetime.datetime.utcnow())


    # def get_max_pages(self, response):
    #     soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    #     soup = soup.find("nav", {"class": "woocommerce-pagination"})
    #     if soup is None:
    #         return 1
    #     max_pages = soup.findAll("li")[-2].find('a').getText()
    #     return max_pages

    def get_products_on_page(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.products')
        soup = soup.select('.product')
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    def start_requests(self):
        yield scrapy.Request(url=self.url, callback=self.first_parse)

    def first_parse(self, response):
        categories = self.get_products_on_page(response)
        for cat in categories:
            yield scrapy.Request(url=cat, callback=self.third_parse)

    # def second_parse(self, response):
    #     max_pages = self.get_max_pages(response)
    #     urls = ["".join([response.request.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages) + 1)]
    #     for url in urls:
    #         yield scrapy.Request(url=url, callback=self.third_parse)

    def third_parse(self, response):
        products = self.get_products_on_page(response)
        for product in products:
            yield scrapy.Request(url=product, callback=self.fourth_parse)

    def fourth_parse(self, response):
        yield self.get_product_data(response)