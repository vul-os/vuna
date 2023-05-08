from fastapi import APIRouter, HTTPException
from typing import List
import uuid

from src.scraper import MetaScraper, ProductScraper, Orchestrator
from src.db.site import Site


router = APIRouter()

@router.get("/meta/{site_id}/{base_url}")
async def meta(site_id: uuid.UUID, base_url: str) -> List[str]:
    """
    Scrape sitemaps for product URLs and return a list of filtered product URLs.

    Args:
        site_id (uuid.UUID): The ID of the site to associate with the products.
        base_url (str): The base URL of the site to scrape.

    Returns:
        List[str]: A list of filtered product URLs.
    """
    try:
        scraper = MetaScraper()
        product_urls = scraper.scrape(base_url=base_url)

        # Get the site from the database
        site = Site.get(site_id)
        if site is None:
            raise HTTPException(status_code=404, detail="Site not found")

        # Add the products to the database
        for url in product_urls:
            site.add_product(url)

        return product_urls
    except Exception as ex:
        raise HTTPException(status_code=500, detail=str(ex))