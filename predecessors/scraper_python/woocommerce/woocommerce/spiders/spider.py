import datetime
import html5lib
from bs4 import BeautifulSoup
import json
from scrapy.linkextractors import LinkExtractor
from scrapy.spiders import SitemapSpider, CrawlSpider, Rule
from scrapy import Request


class WoocommerceSpider(SitemapSpider, CrawlSpider):
    name = "woocommerce"
    rules = ( Rule(LinkExtractor(allow=('', )), callback='parse_item', follow=True), )
    sitemap_rules = [ ('/', 'parse_products'), ]
    sitemap_rules = [('/product/', 'parse_product_data')]
    base_url = 'https://www.smokinggunseeds.co.za'
    sitemap_urls = [f'{base_url}/sitemap_index.xml']
    start_urls = [base_url]
    pagenation_str = 'page'

    def parse_variation_data(self, json_data, product_name, url):
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

    def parse_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.summary,.entry-summary,.product-summary,.product-info')
        if soup is None:
            return
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
                self.parse_variation_data(json_data, product_name, response.request.url)
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

    @staticmethod
    def get_max_pages(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('nav.woocommerce-pagination, nav.pagination')
        if soup:
            meta_text = soup.find('a', {'class': 'pagination-meta'})
            if meta_text:
                return int(meta_text.replace('Page 1 of ', ''))
            else:
                return max([int(s.getText().strip()) if s.getText().strip().isdigit() else 0
                            for s in soup.findAll('a', {'class': 'page-numbers'})])
        return None

    @staticmethod
    def get_products_on_page(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.products')
        if soup is None:
            return None
        soup = soup.select('.product')
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    def parse_products_in_cat(self, response):
        max_pages = self.get_max_pages(response)
        if max_pages:
            urls = ["/".join([response.request.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages))]
            for url in urls:
                yield Request(url=response.request.url, callback=self.parse_products_per_page)
        products = self.get_products_on_page(response)
        if products:
            for product in products:
                yield Request(url=product, callback=self.parse_products_per_page)

    def parse_products_per_page(self, response):
        products = self.get_products_on_page(response)
        if products:
            for product in products:
                yield Request(url=product, callback=self.parse_product_data)
