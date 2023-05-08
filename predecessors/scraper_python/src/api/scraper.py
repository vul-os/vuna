from fastapi import APIRouter, HTTPException
from typing import List, Dict, Union
import uuid
import requests
from cachetools import cached, TTLCache
from google.cloud import storage

from src.scraper import MetaScraper, ProductScraper, ScraperLoader
from src.db.site import Site
from src.db.base import SessionLocal
from src.scraper.product.image.imageuploader import GCSUploader

from src.config.config import config  # import your Settings class

router = APIRouter()
# Define a cache for ScraperLoader instances
# The cache will store up to 100 instances for up to 10 minutes
scraper_cache = TTLCache(maxsize=100, ttl=600)

# Create a GCSUploader object
storage_client = storage.Client() if config.gcs_bucket_name else None
image_uploader = GCSUploader(storage_client, config.gcs_bucket_name) if storage_client else None

@router.get("/meta/{site_id}/{base_url}")
async def meta(site_id: str, base_url: str) -> Dict[str, Union[List[str], int]]:
    # Use the database settings to create a database session
    try:
        url = f"https://{requests.utils.unquote(base_url)}"
        
        scraper = MetaScraper()
        product_urls = scraper.scrape(base_url=url)

        return {"product_urls": product_urls, "len": len(product_urls)}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@cached(scraper_cache)
def get_scraper(script_gcs_bucket_name: str, scraper_filename: str) -> ScraperLoader:
    """
    Retrieve a cached ScraperLoader instance or create a new one if not available.
    """
    return ScraperLoader(script_gcs_bucket_name=script_gcs_bucket_name, scraper_filename=scraper_filename)


@router.post("/product_scrape/{site_id}/{product_url_encoded}")
async def product_scrape(site_id: str, product_url_encoded: str, proxies: List[str], 
                         script_gcs_bucket_name: str, scraper_filename: str, 
                         image_bucket_name: str) -> dict:
    try:
        product_url = f"https://{requests.utils.unquote(product_url_encoded)}"

        # Set up scraper and orchestrator
        scraper = get_scraper(script_gcs_bucket_name=script_gcs_bucket_name, scraper_filename=scraper_filename)()

        # Scrape product
        product_scraper = ProductScraper(site_id=site_id, scraper=scraper, proxies=proxies,
                                         image_uploader=image_uploader)
        product_data = product_scraper(product_url=product_url)

        return product_data
    except Exception as ex:
        raise HTTPException(status_code=500, detail=str(ex))
