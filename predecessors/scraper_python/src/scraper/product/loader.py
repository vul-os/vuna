import os
import importlib


class ScraperLoader:
    """Loader for dynamically loading a scraper class from a code string."""

    def __init__(self, scraper_code_string: str):
        """
        Initialize the ScraperLoader.

        Args:
            scraper_code_string (str): The code string representing the scraper class.
        """
        self.scraper_code_string = scraper_code_string

    def __call__(self):
        """
        Load and instantiate the scraper class.

        Returns:
            object: An instance of the scraper class.
        """
        # Create a module name
        module_name = "scraper_module"
        # Create a module spec from the module name and code string
        module_spec = importlib.util.spec_from_loader(module_name, loader=None)
        # Create an empty module
        module = importlib.util.module_from_spec(module_spec)
        # Execute the code string in the module's namespace
        exec(self.scraper_code_string, module.__dict__)
        # Get the scraper class from the module
        scraper_class = getattr(module, "TheScraper")
        return scraper_class
