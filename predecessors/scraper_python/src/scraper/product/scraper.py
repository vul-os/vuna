from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional


class ProductData(dict):
    """
    Represents product data.

    Attributes:
        name (str): The name of the product.
        url (str): The URL of the product page.
        variants (List[VariantData]): A list of variant data.
    """
    name: str
    image_urls: List[str]
    attribute: str

    # these two identify product
    url: str
    product_id: str

    # these two identify variation
    variation_id: str
    sku: str

    # datapoint data
    price: float
    max_qty: int
    
    def __getattr__(self, name):
        return self.get(name)

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
    def __call__(self, site_url: str) -> List[ProductData]:
        """
        Scrape the site at the specified URL and return a list of ProductData objects.

        Args:
            site_url (str): The URL of the site to scrape.

        Returns:
            List[ProductData]: A list of ProductData objects representing the scraped product data.
        """