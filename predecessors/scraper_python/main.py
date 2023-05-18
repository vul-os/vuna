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
# orchestrator_api = OrchestratorAPI(data_storage_utils)

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
def hello_http(path):
    if request.method == 'GET':
        if path == '':
            return scraper_api.root(request)
        elif path.startswith('site/'):
            return scraper_api.site(request, *path.split('/')[1:3])
        elif path.startswith('meta/'):
            return scraper_api.meta(request, *path.split('/')[1:3])
        # elif path.startswith('orchestrator/'):
        #     return scraper_api.meta(request, *path.split('/')[1:3])
    elif request.method == 'POST':
        if path.startswith('product/'):
            return scraper_api.product(request, *path.split('/')[1:])

    return 'Invalid request', 400

if os.getenv("GOOGLE_CLOUD_FUNCTION_TARGET"):
    from google.cloud import functions
    main = functions.wrap(main)
else:
    if __name__ == "__main__":
        app.run(host="0.0.0.0", port=8000)