from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional


class VariantData(dict):
    """
    Represents variant data.

    Attributes:
        identifier (str): The identifier of the variant.
        price (float): The price of the variant.
        max_qty (int): The maximum quantity of the variant available.
        attributes (Optional[List[Dict[str, Any]]]): A list of dictionaries representing the variant attributes, where each dictionary has keys 'name' and 'value'.
        image_urls (List[str]): A list of image URLs for the variant.
    """
    identifier: str
    image_urls: List[str]

    # datapoint data
    price: float
    max_qty: int

    attributes: Optional[List[Dict[str, Any]]]



class ProductData(dict):
    """
    Represents product data.

    Attributes:
        name (str): The name of the product.
        url (str): The URL of the product page.
        variants (List[VariantData]): A list of variant data.
    """
    name: str
    url: str
    variants: List[VariantData]


class Scraper(ABC):
    """
    Abstract base class for scrapers.

    Attributes:
        proxies (Optional[Dict[str, str]]): A dictionary of proxies to be used with the scraper.
    """
    def __init__(
        self,
        proxies: Optional[Dict[str, str]] = None,
    ):
        """
        Initialize the scraper with optional arguments.

        Args:
            proxies (dict, optional): A dictionary of proxies to be used with the scraper.
        """
        self.proxies = proxies

    @abstractmethod
    def __call__(self, site_url: str) -> ProductData:
        """
        Scrape the site at the specified URL and return a list of ProductData objects.

        Args:
            site_url (str): The URL of the site to scrape.

        Returns:
            List[ProductData]: A list of ProductData objects representing the scraped product data.
        """