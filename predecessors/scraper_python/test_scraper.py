from src.scraper.product.examples.woocommerce import TheScraper
# Create an instance of TheScraper
scraper = TheScraper()

# Set the URL of the product to scrape
product_url = "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/"

# Scrape the product data
product_data_list = scraper.scrape_product(product_url)

# Print the scraped product data
for product_data in product_data_list:
    print(product_data)