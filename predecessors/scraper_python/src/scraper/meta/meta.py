import logging
from typing import List

import requests
from bs4 import BeautifulSoup

from src.db.product import Product

logger = logging.getLogger(__name__)


class MetaScraper:
    def __init__(self):
        self.known_urls = []

    def __call__(self, base_url: str, site_id: uuid.UUID, session):
        products = self.scrape(base_url)

        # Save products to db
        for p in products:
            Product.merge(url=p.url, site_id=site_id, session=session)

    def scrape(self, base_url: str) -> List[str]:
        # Parse robots.txt
        robots_url = f'{base_url}/robots.txt'
        logger.info('Parsing robots.txt: %s', robots_url)
        robots_text = self._url_to_text(robots_url)
        sitemap_urls = self._parse_robots_txt(robots_text)

        # Add default sitemap URLs
        default_sitemap_urls = [
            f'{base_url}/sitemap.xml',
            f'{base_url}/sitemap_index.xml',
        ]
        sitemap_urls.extend(default_sitemap_urls)

        # Parse sitemaps for URLs
        for sitemap_url in sitemap_urls:
            logger.info('Parsing sitemap: %s', sitemap_url)
            sitemap_text = self._url_to_text(sitemap_url)
            urls = self._parse_sitemap(sitemap_text)
            self.known_urls.extend(urls)

        # Filter URLs for ecommerce product URLs
        product_urls = self._filter_product_urls(self.known_urls)

        return product_urls

    def _url_to_text(self, url: str) -> str:
        response = requests.get(url)
        response.raise_for_status()
        return response.text

    def _parse_robots_txt(self, text: str) -> List[str]:
        sitemap_urls = []
        for line in text.split('\n'):
            if line.lower().startswith('sitemap:'):
                url = line.split(':', 1)[1].strip()
                sitemap_urls.append(url)
        return sitemap_urls

    def _parse_sitemap(self, text: str) -> List[str]:
        soup = BeautifulSoup(text, 'html.parser')
        urls = [loc.text for loc in soup.find_all('loc')]
        return urls

    def _filter_product_urls(self, urls: List[str]) -> List[str]:
        product_urls = []
        for url in urls:
            path = url.split('//', 1)[-1].split('/', 1)[-1]
            if any(keyword in path for keyword in ['/product/', '/products/', '/shop/', '/buy/']):
                product_urls.append(url)
        return product_urls
