from datetime import datetime
from typing import List

from urllib.parse import urlparse

from .loader import ScraperLoader
from src.scraper.product.scraper import ProductData
from src.storage.storage import StorageUtils
from src.scraper.encoder.encoder import encode_url

def scrape_product_data(product_url: str, scraper_loader: ScraperLoader, job_identifier: str, proxies: [] = None, storage_utils: StorageUtils = None):
    def get_site_url(url):
        parsed_url = urlparse(url)
        return f"{parsed_url.scheme}://{parsed_url.netloc}"

    print("URL:", product_url)
    scraper = scraper_loader()()
    product_data = scraper(product_url)
    # Get the current date and time
    current_datetime = datetime.now()
    # Format the date and time as a string
    formatted_datetime = current_datetime.strftime("%Y-%m-%d|%H-%M-%S")

    first_item = next(iter(product_data), None)
    first_product = ProductData(**first_item)
    # Validate product data
    if first_product.url is None or first_product.name is None:
        return None

    # Generate product & store IDs from their respective URLs
    product_id = encode_url(product_url)
    site_id = encode_url(get_site_url(product_url))

    # Create dictionaries for products, variations, and datapoints
    products_to_save = []
    for p in product_data:
        product_dict = {
            "id": product_id,
            "date_updated": formatted_datetime,
        }
        product_dict.update(p)
        products_to_save.append(product_dict)
        
    if storage_utils is not None:
        path_prefix = f"/product/{job_identifier}"
        file_name = f"{path_prefix}/{site_id}_{formatted_datetime}_product.csv"
        # Write dictionaries to CSV files
        storage_utils.write_data_to_csv(file_name, products_to_save)

    return product_data
