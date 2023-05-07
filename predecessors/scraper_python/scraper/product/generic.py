import importlib.util
import os

from db.models import Site
from gcs_uploader import GCSUploader


class Scraper:
    def __init__(self, script_cache_dir, scraper_filename, image_bucket_name=None):
        self.script_cache_dir = script_cache_dir
        self.scraper_filename = scraper_filename
        self.image_uploader = GCSUploader(image_bucket_name) if image_bucket_name else None

    def __call__(self, url):
        # Check cache directory for scraper file
        scraper_file = os.path.join(self.script_cache_dir, f"{self.scraper_filename}")

        if not os.path.isfile(scraper_file):
            # Download scraper file from GCS
            gcs_path = f"{self.site_id}/{self.scraper_filename}"
            scraper_code = self.image_uploader.download_file(gcs_path)

            # Save scraper file to cache directory
            with open(scraper_file, "w") as f:
                f.write(scraper_code)

        # Execute scraper file dynamically
        spec = importlib.util.spec_from_file_location(scraper_module_name, scraper_file)
        scraper_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(scraper_module)

        # Get the scraper class from the module
        scraper_class = getattr(scraper_module, scraper_class_name)

        # Instantiate the scraper
        scraper = scraper_class()

        # Run the scraper
        data = scraper(site_url)

        # Save data to db
        a = 
        # Save the images
        if self.image_uploader is not None and "image_urls" in data and len(data["image_urls"]) > 0:
            for image_url in data["image_urls"]:
                self.image_uploader.upload_image(image_url, self.site_id)


if __name__ == "__main__":
    # Example usage
    site_id = "example"
    proxies = {"http": "http://proxy.example.com:1234", "https": "http://proxy.example.com:1234"}
    image_bucket_name = "example-bucket"
    script_cache_dir = "/path/to/cache/dir"
    scraper_filename = "example_scraper.py"
    url = "https://example.com/product/1234"

    s = Scraper(script_cache_dir=script_cache_dir, scraper_filename=scraper_filename, image_bucket_name=image_bucket_name)
    s.site_id = site_id
    s.proxies = proxies
    s(url)