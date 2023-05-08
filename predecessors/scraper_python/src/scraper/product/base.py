from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional, Union


class ProductData(dict):
    name: Optional[str]
    url: Optional[str]
    sku: Optional[str]
    price: Optional[float]
    max_qty: Optional[int]
    attributes: Optional[List[Dict[str, Any]]]
    image_url: Optional[str]


class Scraper(ABC):
    def __init__(
        self,
        proxies: Optional[Dict[str, str]] = None,
    ):
        """
        Initialize the scraper with optional arguments.

        Args:
            proxies (dict, optional): A dictionary of proxies to be used with the scraper.
            bucket_name (str, optional): The name of the Google Cloud Storage bucket to upload scraped images to.
        """
        self.proxies = proxies

    @abstractmethod
    def __call__(self, site_url: str) -> List[ProductData]:
        """
        Scrape the site at the specified URL and return a dictionary with the following keys:

        - 'name': the name of the product (str)
        - 'url': the URL of the product page (str)
        - 'sku': the product SKU (str)
        - 'price': the price of the product (float)
        - 'max_qty': the maximum quantity of the product available (int)
        - 'attributes': a list of dictionaries representing the product attributes, where each dictionary has keys
          'name' and 'value' (List[Dict[str, Any]])
        - 'image_url': the URL of the product image (str)
        """
        pass