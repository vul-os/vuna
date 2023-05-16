import os
from flask import Flask, request

from src.api.api import ScraperAPI
from src.storage.gcs import StorageUtilsGCS 
from src.storage.local import StorageUtilsLocal 


app = Flask(__name__)

# from google.cloud import storage
# storage_client = storage.Client()
# data_storage_utils = StorageUtilsGCS(client, )

data_storage_utils = StorageUtilsLocal("/workspace/scraper_python/src/scraper/product/examples")


scraper_api = ScraperAPI(data_storage_utils)

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
def hello_http(path):
    if request.method == "GET":
        if request.path == "/":
            return scraper_api.root(request)
        elif request.path.startswith("/meta/"):
            base_url = '/'.join(request.path.split("/")[2:3])
            return scraper_api.meta(request, base_url)
    elif request.method == "POST":
        if request.path.startswith("/product_scrape/"):
            product_url = '/'.join(request.path.split("/")[2:3])
            return scraper_api.product_scrape(request, product_url)

    return "Invalid request", 400

if os.getenv("GOOGLE_CLOUD_FUNCTION_TARGET"):
    from google.cloud import functions
    main = functions.wrap(main)
else:
    if __name__ == "__main__":
        app.run(host="0.0.0.0", port=8000)