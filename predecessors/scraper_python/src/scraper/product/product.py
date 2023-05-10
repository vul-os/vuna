from datetime import datetime
import csv
import hashlib
from typing import List

from .loader import ScraperLoader
from ..utils import StorageUtils, hashStringFromUrl


class ProductScraper:
    def __init__(self, site_id: str, scraper: ScraperLoader, proxies: List[str],
                 storage_utils: StorageUtils = None):
        self.site_id = site_id
        self.scraper = scraper
        self.proxies = proxies
        self.storage_utils = storage_utils

    def __call__(self, product_url: str):

        product_data = self.scraper(product_url, self.proxies)

        first_item = next(iter(product_data), None)

        # Validate product data
        if first_item.url is None or first_item.name is None:
            return

        # Generate product ID from product URL
        product_id = hashlib.sha256(first_item.url.encode()).hexdigest()

        # Create dictionaries for products, variations, and datapoints
        product_dict = {
            "id": product_id,
            "url": first_item.url,
            "site_id": self.site_id,
            "date_updated": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        }

        variation_dict = {}
        datapoint_dict = {}

        for p in product_data:
            variation_identifier = p.get("identifier")
            if not variation_identifier:
                variation_identifier = "default"

            variation_dict[variation_identifier] = {
                "id": hashlib.sha256(f"{product_id}:{variation_identifier}".encode()).hexdigest(),
                "identifier": variation_identifier,
                "product_id": product_id,
            }
            variation_id = variation_dict[variation_identifier]["id"]

            datapoint_dict[f"{variation_id}:{p['max_qty']}:{p['price']}"] = {
                "var_id": variation_id,
                "max_qty": p["max_qty"],
                "price": p["price"],
            }

            # # Save variation images
            # if self.gcs_utils and "image_urls" in p and len(p["image_urls"]) > 0:
            #     for image_url in p["image_urls"]:
            #         self.gcs_utils.upload_image(image_url.strip(), self.site_id)

        # Write dictionaries to CSV files
        self.storage_utils.upload_csv_from_dict(f"{product_id}-products.csv", [product_dict])
        self.storage_utils.upload_csv_from_dict(f"{product_id}-variations.csv", variation_dict.values())
        self.storage_utils.upload_csv_from_dict(f"{product_id}-datapoints.csv", datapoint_dict.values())
        self.storage_utils.upload_csv_from_dict(f"{product_id}-images.csv", [{"images": p["image_urls"]}])

        return product_data


