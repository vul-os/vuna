from src.scraper.product.examples.woocommerce import TheScraper
import requests
from requests import Session
from src.orchestrator.proxies import create_proxy_list

proxies = create_proxy_list()
for proxy in proxies[0:10]:
    # print(proxy)
    session = Session()

    proxies = {'http': f'socks5://{proxy}'}
    session.proxies.update(proxies)

    # Create an instance of TheScraper
    scraper = TheScraper(session)


  
    # Set the URL of the product to scrape
    product_urls = [
        "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/",
        # "https://www.biltongandbudz.co.za/product/red-jalapeno-seeds-30-seeds/",
        # "https://abies.co.za/product/polly-cotton-denim-printed"
    ]
    product_data_list = []
    for product_url in product_urls:
        sc = scraper(product_url)
        if not sc:
            continue
        product_data_list.extend(sc)

    if len(product_data_list):
        print("success: ", product_urls[0])
    # # Print the scraped product data
    # for product_data in product_data_list:
    #     print(product_data)