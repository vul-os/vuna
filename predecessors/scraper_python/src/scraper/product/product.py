from datetime import datetime
from typing import List

from sqlalchemy.orm import Session

from src.db.product import Product
from src.db.variation import Variation
from src.db.datapoint import DataPoint

from src.scraper.product.loader import ScraperLoader
from src.scraper.product.image.imageuploader import GCSUploader


class ProductScraper:
    def __init__(self, site_id: str, scraper: ScraperLoader, session: Session, proxies: List[str],
                 image_uploader: GCSUploader = None):
        self.site_id = site_id
        self.scraper = scraper
        self.session = session
        self.proxies = proxies
        self.image_uploader = image_uploader

    def scrape_product(self, product_url: str):
        # scraper_loader = ScraperLoader(site_id=self.site_id, proxies=self.proxies,
        #                                image_bucket_name=image_bucket_name)
        # scraper = scraper_loader()

        product_data = self.scraper(product_url, self.proxies)

        first_item = next(iter(product_data), None)

        # Validate product data
        if first_item.url is None or first_item.name is None:
            return

        # Merge product
        product = Product.merge(url=first_item["url"], site_id=self.site_id)

        # Save variation datapoints
        variation_identifier_to_id = {}
        for p in product_data:
            
            variation_identifier = p.get("sku")
            if not variation_identifier:
                variation_identifier = "default"

            variation_id = variation_identifier_to_id.get(variation_identifier)
            if not variation_id:
                variation = Variation.merge(
                    identifier=variation_identifier,
                    product_id=product.id,
                )
                variation_identifier_to_id[variation_identifier] = variation.id
                variation_id = variation.id

            datapoint = DataPoint.create(
                var_id=variation_id,
                max_qty=p["max_qty"],
                price=p["price"],
            )

            # Save variation images
            if self.image_uploader and "image_urls" in p and len(p["image_urls"]) > 0:
                for image_url in p["image_urls"]:
                    self.image_uploader.upload_image(image_url.strip(), self.site_id)

        # Update product metadata
        product.date_updated = datetime.now()

        self.session.commit()