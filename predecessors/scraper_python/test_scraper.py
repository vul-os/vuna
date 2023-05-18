from src.scraper.product.examples.woocommerce import TheScraper
# Create an instance of TheScraper
scraper = TheScraper()

# Set the URL of the product to scrape
product_urls = [
    "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/",
    "https://www.biltongandbudz.co.za/product/red-jalapeno-seeds-30-seeds/",
    "https://abies.co.za/product/polly-cotton-denim-printed"
]
product_data_list = []
for product_url in product_urls:
    product_data_list.extend(scraper(product_url))

# Print the scraped product data
for product_data in product_data_list:
    print(product_data)