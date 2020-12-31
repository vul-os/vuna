from scrapy.spiders import SitemapSpider

class MySpider(SitemapSpider):
    sitemap_urls = ['https://www.biltongandbudz.co.za/sitemap_index.xml']

    def parse(self, response):
        pass # ... scrape item here ...