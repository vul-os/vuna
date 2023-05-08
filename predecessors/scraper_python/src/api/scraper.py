from fastapi import FastAPI, HTTPException
from typing import List
import uuid
import requests
from cachetools import cached, TTLCache
from google.cloud import storage

from src.scraper import MetaScraper, ProductScraper, ScraperLoader
from src.db.site import Site
from src.db.base import SessionLocal
from src.scraper.product.image.imageuploader import GCSUploader

from src.config.setting import Settings  # import your Settings class

app = FastAPI()
settings = Settings()

# Define a cache for ScraperLoader instances
# The cache will store up to 100 instances for up to 10 minutes
scraper_cache = TTLCache(maxsize=100, ttl=600)

# Create a GCSUploader object
storage_client = storage.Client()
image_uploader = GCSUploader(storage_client, settings.gcs_bucket_name)

@app.get("/meta/{site_id}/{base_url}")
async def meta(site_id: uuid.UUID, base_url: str) -> List[str]:
    # Use the database settings to create a database session
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

        # Commit the changes to the database
        session.commit()

        return product_urls
    except Exception as e:
        # Roll back the changes to the database on error
        session.rollback()
        raise HTTPException(status_code=500, detail=str(e))


@cached(scraper_cache)
def get_scraper(script_gcs_bucket_name: str, scraper_filename: str) -> ScraperLoader:
    """
    Retrieve a cached ScraperLoader instance or create a new one if not available.
    """
    return ScraperLoader(script_gcs_bucket_name=script_gcs_bucket_name, scraper_filename=scraper_filename)


@app.post("/product_scrape/{site_id}/{product_url_encoded}")
async def product_scrape(site_id: str, product_url_encoded: str, proxies: List[str], 
                         script_gcs_bucket_name: str, scraper_filename: str, 
                         image_bucket_name: str) -> dict:
    try:
        product_url = requests.utils.unquote(product_url_encoded)

        # Use the database settings to create a database session
        session = SessionLocal(bind=settings.db_dsn)

        # Set up scraper and orchestrator
        scraper = get_scraper(script_gcs_bucket_name=script_gcs_bucket_name, scraper_filename=scraper_filename)()

        # Scrape product
        product_scraper = ProductScraper(site_id=site_id, scraper=scraper,
                                         session=session, proxies=proxies,
                                         image_uploader=image_uploader)
        product_data = product_scraper(product_url=product_url)

        # Commit changes to the database
        session.commit()

        return product_data
    except Exception as ex:
        session.rollback()
        raise HTTPException(status_code=500, detail=str(ex))
