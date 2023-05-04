import importlib.util
import os
import tempfile

from db.models import Site
from gcs_uploader import GCSUploader


class Scraper:
    def __init__(self, script_cache_dir, gcs_bucket_name):
        self.script_cache_dir = script_cache_dir
        self.scraper_filename = scraper_filename
        self.gcs_bucket_name = gcs_bucket_name
        self.gcs_uploader = GCSUploader(gcs_bucket_name)

    def __call__(self, url):
        # Check cache directory for scraper file
        scraper_file = os.path.join(self.script_cache_dir, f"{self.scraper_filename}")

        if not os.path.isfile(scraper_file):
            # Download scraper file from GCS
            gcs_path = f"{self.scraper_filename}"
            scraper_code = self.gcs_uploader.download_file(gcs_path)

            # Save scraper file to cache directory
            with open(scraper_file, "w") as f:
                f.write(scraper_code)

        # Execute scraper file dynamically
        spec = importlib.util.spec_from_file_location("scraper_module", scraper_file)
        scraper_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(scraper_module)

        # Run the `scrape` function in the scraper file
        data = scraper_module.Scraper(self.site_id)(url)

        # Save data to db

        # Save the images
        if "image_urls" in data and data["image_urls"].len() > 0:
            for image_url in data["image_urls"]:
                self.upload_image(image_url)

    def upload_image(self, image_url):
        return self.gcs_uploader.upload_image(image_url, self.site_id)


if __name__ == "__main__":
    s = Scraper("")()