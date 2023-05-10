from fastapi import FastAPI, HTTPException
from typing import List, Dict, Union
import uuid
import requests
from cachetools import cached, TTLCache
from google.cloud import storage

from src.scraper import MetaScraper, ProductScraper, ScraperLoader
from src.scraper.utils import StorageUtils

app = FastAPI()
# Define a cache for ScraperLoader instances
# The cache will store up to 100 instances for up to 10 minutes
scraper_cache = TTLCache(maxsize=100, ttl=600)
gcs_bucket_name = "asdasd"
# Create a GCSUploader object
storage_client = storage.Client() if gcs_bucket_name else None
data_storage_utils = StorageUtils(storage_client=storage_client, bucket_name="mybucket")
scraper_storage_utils = StorageUtils(storage_client=storage_client, bucket_name="mybucket")

@app.get("/meta/{site_id}/{base_url}")
async def meta(site_id: str, base_url: str) -> Dict[str, Union[List[str], int]]:
    # Use the database settings to create a database session
    try:
        url = f"https://{requests.utils.unquote(base_url)}"
        
        scraper = MetaScraper(data_storage_utils)
        meta_data = scraper(base_url=url)

        return {"meta_data": meta_data, "len": len(meta_data)}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@cached(scraper_cache)
def get_product_scraper(site_id: str, scraper_filename: str) -> ProductScraper:
    """
    Retrieve a cached ScraperLoader instance or create a new one if not available.
    """
    scraper = ScraperLoader(storage_utils=scraper_storage_utils, scraper_filename=scraper_filename)
    product_scraper = ProductScraper(site_id=site_id, scraper=scraper, proxies=proxies,
                                        storage_utils=data_storage_utils)
    return product_scraper


@app.post("/product_scrape/{site_id}/{product_url_encoded}")
async def product_scrape(site_id: str, product_url_encoded: str, proxies: List[str], 
                         script_gcs_bucket_name: str, scraper_filename: str, 
                         image_bucket_name: str) -> dict:
    try:
        product_url = f"https://{requests.utils.unquote(product_url_encoded)}"

        # Set up scraper and orchestrator
        scraper = get_product_scraper(scraper_filename=scraper_filename)

        product_data = product_scraper(product_url=product_url)

        return product_data
    except Exception as ex:
        raise HTTPException(status_code=500, detail=str(ex))



if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)