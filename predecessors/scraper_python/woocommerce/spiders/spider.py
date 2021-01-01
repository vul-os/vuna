import datetime
import html5lib
from bs4 import BeautifulSoup
import json
from scrapy.linkextractors import LinkExtractor
from scrapy.spiders import SitemapSpider, CrawlSpider, Rule
from scrapy import Request
from pprint import pprint


def upsert_product(sku: str,
                   store_url: str,
                   variation_id: str,
                   url: str,
                   product_name: str,
                   categories: list,
                   tags: list,
                   attributes: dict,
                   product_price: float,
                   product_stock: int,
                   scrape_date: datetime):
    print(
        "Variation --> \n" if variation_id is not None else "No Variation --> \n",
        f"name: {product_name}, store_url: {store_url}, sku: {sku}, var_id: {variation_id} \n",
        f"stock: {product_stock}, price: {product_price}",
        f"cats: {categories} \n",
        f"tag: {tags} \n",
        f"attributes: {attributes} \n",
        f"url: {url}, date: {scrape_date}"
    )


class WoocommerceSpider(SitemapSpider, CrawlSpider):
    name = "woocommerce"
    rules = (Rule(LinkExtractor(allow=('',)), callback='parse_item', follow=True),)
    sitemap_rules = [('/product/', 'parse_product_data'), ('/products/', 'parse_product_data')]

    def __init__(self, base_url, *args, **kwargs):
        super(WoocommerceSpider, self).__init__(*args, **kwargs)
        self.base_url = base_url
        self.sitemap_urls = [f'{base_url}/sitemap_index.xml', f'{base_url}/sitemap.xml']
        self.start_urls = [base_url]

    def parse_product_data(self, response):
        soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
        soup = soup.select_one('.summary,.entry-summary,.product-summary,.product-info')
        product_name_selector = '.product_title,.entry-title'
        if soup is None or not soup.select_one(product_name_selector):
            print("No Product Name")
            return None
        product_name = soup.select_one(product_name_selector).getText().strip()
        product_meta = soup.find('div', {'class': 'product_meta'})
        tags = product_meta.select('.tagged_as > a')
        tags = [{'url': t['href'], 'name': t.text} for t in tags]
        categories = product_meta.select('.posted_in > a')
        categories = [{'url': c['href'], 'name': c.text} for c in categories]
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
                for i, data in enumerate(json_data):
                    sku = data['sku']
                    variation_id = data['variation_id']
                    attributes = data['attributes']
                    # attributes = {k.replace('attribute_', ''): v for k, v in attributes.items()} \
                    #     if len(attributes) > 0 else None
                    product_price = data['display_price']
                    product_stock_max_qty = data['max_qty']
                    product_stock_max_qty = int(product_stock_max_qty) if str(product_stock_max_qty).isdigit() else 0
                    product_stock_avail = data['availability_html'].replace('<p class="stock in-stock">', '') \
                        .replace('</p>', '').strip().replace('in', '').replace('stock', '').strip()
                    product_stock_avail = int(product_stock_avail) if product_stock_avail.isdigit() else 0
                    product_stock = max(product_stock_avail, product_stock_max_qty)
                    upsert_product(sku=sku, store_url=self.base_url, variation_id=variation_id, categories=categories,
                                   tags=tags, attributes=attributes,
                                   product_name=product_name, product_price=product_price, product_stock=product_stock,
                                   url=response.request.url, scrape_date=datetime.datetime.now())

        else:
            product_stock = soup.find("p", {"class": "stock in-stock"})
            product_stock = product_stock.getText().strip() if product_stock else 0
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
            product_price = product_price.getText().replace("R", "") if product_price else 0
            variation_id_add_cart_button = soup.find('button', {'name': 'add-to-cart'})
            variation_id_input = soup.select_one('cwg-product-id')
            variation_id = variation_id_add_cart_button if variation_id_add_cart_button else variation_id_input
            variation_id = variation_id['value'] if variation_id is not None else None
            sku = product_meta.select_one('.sku').text if product_meta.select_one('.sku') else None

            upsert_product(sku=sku, store_url=self.base_url, variation_id=variation_id, categories=categories,
                           tags=tags, attributes={},
                           product_name=product_name, product_price=product_price, product_stock=product_stock,
                           url=response.request.url, scrape_date=datetime.datetime.now())
