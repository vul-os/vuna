import os
from flask import Flask, request
# from google.cloud import functions

from src.api.api import ScraperAPI


app = Flask(__name__)
scraper_api = ScraperAPI()

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
def main(path):
    if request.method == "GET":
        if request.path == "/":
            return scraper_api.root(request)
        elif request.path.startswith("/meta/"):
            site_id, base_url = request.path.split("/")[2:4]
            print(site_id, base_url)
            return scraper_api.meta(request, site_id, base_url)
    elif request.method == "POST":
        if request.path.startswith("/product_scrape/"):
            site_id, product_url_encoded = request.path.split("/")[2:4]
            return scraper_api.product_scrape(request, site_id, product_url_encoded)

    return "Invalid request", 400

# if os.getenv("GOOGLE_CLOUD_FUNCTION_TARGET"):
#     main = functions.wrap(main)
# else:
if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000)
