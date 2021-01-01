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
    sitemap_rules = [('/product/', 'parse_product_data'), ('/products/', 'parse_product_data')] #, ('/product-category/*', 'parse_categories')]
    base_url = 'https://feedaseed.co.za/'
    sitemap_urls = [f'{base_url}/sitemap_index.xml', f'{base_url}/sitemap.xml']
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
        product_name_selector = '.product_title,.entry-title'
        if soup is None or not soup.select_one(product_name_selector):
            print("No Product Name")
            return None
        product_name = soup.select_one(product_name_selector).getText().strip()
        variations = soup.find('table', {'class': 'variations'})
        variations_soup = soup.find('form', {'class': 'variations_form cart'})
        if variations is not None:
            raw_json_string = str(variations_soup['data-product_variations'])
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
            variation_id_add_cart_button = soup.find('button', {'name': 'add-to-cart'})
            variation_id_input = soup.select_one('cwg-product-id')
            variation_id = variation_id_add_cart_button if variation_id_add_cart_button else variation_id_input
            variation_id = variation_id['value'] if variation_id is not None else None

            if not variation_id:
                return
            # if variation_id is None:
            #     # biltong & buddz
            #     variation_id = variations_soup['data-product_id']

            print("No Variations: ", product_name, product_price, product_stock, variation_id,
                  datetime.datetime.utcnow())

    @staticmethod
    def get_products_on_page(response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.products')
        if soup is None:
            return None
        soup = soup.select('.product')
        product_links = [str(links.find('a')['href']) for links in soup]
        return product_links

    # def parse_products_in_cat(self, response):
    #     max_pages = self.get_max_pages(response)
    #     if max_pages:
    #         urls = ["/".join([response.request.url, self.pagenation_str, str(i)]) for i in range(1, int(max_pages))]
    #         for url in urls:
    #             yield Request(url=url, callback=self.parse_products)
    #     else:
    #         products = self.get_products_on_page(response)
    #         if products:
    #             for product in products:
    #                 yield Request(url=product, callback=self.parse_product_data)
    # @staticmethod
    # def get_max_pages(response):
    #     soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    #     soup = soup.select_one('nav.woocommerce-pagination, nav.pagination')
    #     if soup:
    #         meta_text = soup.find('a', {'class': 'pagination-meta'})
    #         if meta_text:
    #             return int(meta_text.replace('Page 1 of ', ''))
    #         else:
    #             return max([int(s.getText().strip()) if s.getText().strip().isdigit() else 0
    #                         for s in soup.findAll('a', {'class': 'page-numbers'})])
    #     return None

    # def parse_categories(self, response):
    #     print("parsing categories")
    #     products = self.get_products_on_page(response)
    #     if products:
    #         for product in products:
    #             yield Request(url=product, callback=self.parse_product_data)
