import os
import sys
from flask import Flask, request

from src.api.scraper import ScraperAPI
from src.api.orchestrator import OrchestratorAPI

from src.storage.gcs import StorageUtilsGCS 
from src.storage.local import StorageUtilsLocal 

from src.orchestrator.tasks import TaskCreator

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

task_creator = TaskCreator(project_id="scraping-is-hard",
    location="us-central1", queue_id="scraper")

scraper_api = ScraperAPI(data_storage_utils)
orchestrator_api = OrchestratorAPI(task_creator, data_storage_utils)

def root():
    import sys
    print(sys.getrecursionlimit())
    return f"{sys.getrecursionlimit()}"

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
def hello_http(path):
    if request.method == "GET":
        if request.path == "/":
            return root()
    if 'scraper' in request.path:
        if request.path.startswith("/scraper/site/"):
            base_url = '/'.join(request.path.split("/")[3:4])
            return scraper_api.site(request, base_url)
        elif request.path.startswith("/scraper/meta/"):
            base_url = '/'.join(request.path.split("/")[3:4])
            return scraper_api.meta(request, base_url)
        elif request.path.startswith("/scraper/product/"):
            product_url = '/'.join(request.path.split("/")[3:])
            return scraper_api.product(request, product_url)

    elif 'orchestrator' in request.path:
        if request.path.startswith("/orchestrator/site/"):
            print("here")
            return orchestrator_api.site(request)
        elif request.path.startswith("/orchestrator/meta/"):
            return orchestrator_api.meta(request)

    return "Invalid request", 400

if os.getenv("GOOGLE_CLOUD_FUNCTION_TARGET"):
    from google.cloud import functions
    main = functions.wrap(main)
else:
    if __name__ == "__main__":
        app.run(host="0.0.0.0", port=8000)