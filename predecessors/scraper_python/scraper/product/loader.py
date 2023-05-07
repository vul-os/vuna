from google.cloud import storage

class ScraperLoader:
    def __init__(self, site_id, script_gsc_bucket_name, script_cache_dir, scraper_filename):
        self.site_id = site_id
        self.script_gsc_bucket_name = script_gsc_bucket_name
        self.script_cache_dir = script_cache_dir
        self.scraper_filename = scraper_filename
        
    def get_scraper(self):
        # Check cache directory for scraper file
        scraper_file = os.path.join(self.script_cache_dir, f"{self.scraper_filename}")

        if not os.path.isfile(scraper_file):
            # Download scraper file from GCS
            storage_client = storage.Client()
            bucket = storage_client.get_bucket(self.script_gsc_bucket_name)
            blob = bucket.blob(f"{self.site_id}/{self.scraper_filename}")
            scraper_code = blob.download_as_string().decode('utf-8')

            # Save scraper file to cache directory
            with open(scraper_file, "w") as f:
                f.write(scraper_code)

        # Execute scraper file dynamically
        spec = importlib.util.spec_from_file_location("scraper_module", scraper_file)
        scraper_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(scraper_module)

        # Get the scraper class from the module
        scraper_class = getattr(scraper_module, "Scraper")

        # Instantiate the scraper and return it
        return scraper_class()