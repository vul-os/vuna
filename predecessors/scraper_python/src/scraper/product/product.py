from datetime import datetime
import hashlib
from typing import List
from urllib.parse import urlparse

from .loader import ScraperLoader
from src.storage.storage import StorageUtils


def scrape_product_data(product_url: str, scraper_loader: ScraperLoader, proxies: [] = None, storage_utils: StorageUtils = None) -> List[dict]:
    def get_site_url(url):
        parsed_url = urlparse(url)
        return f"{parsed_url.scheme}://{parsed_url.netloc}"

    print("URL:", product_url)
    scraper = scraper_loader()()
    product_data = scraper(product_url)
    print(product_data)
    # Get the current date and time
    current_datetime = datetime.now()
    # Format the date and time as a string
    formatted_datetime = current_datetime.strftime("%Y-%m-%d_%H-%M-%S")

    print(product_data)
    first_item = next(iter(product_data), None)

    # Validate product data
    if first_item.url is None or first_item.name is None:
        return []

    # Generate product & store IDs from their respective URLs
    product_id = hashlib.sha256(product_url.encode()).hexdigest()
    site_id = hashlib.sha256(get_site_url(product_url).encode()).hexdigest()

    # Create dictionaries for products, variations, and datapoints
    product_dict = {
        "id": product_id,
        "site_id": site_id,
        "url": first_item.url,
        "date_updated": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    }

    variation_dict = {}

    for p in product_data:
        variation_identifier = p.get("identifier", "default")

        variation_dict[variation_identifier] = {
            "id": hashlib.sha256(f"{site_id}:{product_id}:{variation_identifier}".encode()).hexdigest(),
            "identifier": variation_identifier,
            "product_id": product_id,
            "max_qty": p["max_qty"],
            "price": p["price"],
            "datetime": formatted_datetime,
        }
    
    if storage_utils is not None:
        base_string = f"{site_id}-{formatted_datetime}-{product_id}"
        # Write dictionaries to CSV files
        storage_utils.upload_csv_from_data(f"{base_string}-product.csv", [product_dict])
        storage_utils.upload_csv_from_data(f"{base_string}-variations.csv", variation_dict if len(variation_dict) > 0 else [variation_dict])

    return product_data
