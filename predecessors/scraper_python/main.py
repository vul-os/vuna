import os
import sys
from flask import Flask, request

from src.api.scraper import ScraperAPI

from src.storage.gcs import StorageUtilsGCS 
from src.storage.local import StorageUtilsLocal 


app = Flask(__name__)

data_storage_utils = None
if '--local' in sys.argv:
    script_dir = os.path.dirname(os.path.abspath(__file__))
    data_storage_utils = StorageUtilsLocal(f"{script_dir}/test_data")
else:
    from google.cloud import storage
    storage_client = storage.Client()
    bucket_name = "exolution-scraper-data"
    data_storage_utils = StorageUtilsGCS(storage_client, bucket_name)

scraper_api = ScraperAPI(data_storage_utils)

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
def hello_http(path):
    if request.method == "GET":
        if request.path == "/":
            return scraper_api.root(request)
        elif request.path.startswith("/site/"):
            base_url = '/'.join(request.path.split("/")[2:3])
            return scraper_api.site(request, base_url)
        elif request.path.startswith("/meta/"):
            base_url = '/'.join(request.path.split("/")[2:3])
            return scraper_api.meta(request, base_url)
    elif request.method == "POST":
        if request.path.startswith("/product/"):
            product_url = '/'.join(request.path.split("/")[2:3])
            return scraper_api.product(request, product_url)

    return "Invalid request", 400

if os.getenv("GOOGLE_CLOUD_FUNCTION_TARGET"):
    # from google.cloud import functions
    # main = functions.wrap(main)
else:
    if __name__ == "__main__":
        app.run(host="0.0.0.0", port=8000)