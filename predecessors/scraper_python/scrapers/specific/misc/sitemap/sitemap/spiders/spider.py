from scrapy.spiders import SitemapSpider

class MySpider(SitemapSpider):
    name = 'sitemap'
    sitemap_urls = ['https://www.botshop.co.za/product_cat-sitemap.xml']
    start_urls = ['https://www.biltongandbudz.co.za/']

    def parse(self, response):
        print(response)
        pass # ... scrape item here ...