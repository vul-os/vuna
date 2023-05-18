import requests
from bs4 import BeautifulSoup

def detect(soup: BeautifulSoup):
    try:
        # Check <script> and <link> tags
        scripts = soup.find_all('script')
        links = soup.find_all('link')

        for script in scripts:
            script_content = str(script)
            if 'Shopify.theme' in script_content:
                return "shopify"
            elif 'prestashop.com' in script_content:
                return "prestashop"
            elif 'cdn3.bigcommerce.com' in script_content:
                return "bigcommerce"
            elif 'varien/js.js' in script_content:  # This is a common Magento file
                return "magento"
            elif 'woocommerce' in script_content:
                return "woocommerce"

        for link in links:
            link_content = str(link)
            if 'cdn.shopify.com' in link_content:  # Shopify commonly uses this CDN
                return "shopify"
            elif 'themes/_prestashop' in link_content:  # Prestashop often has this in their theme URLs
                return "prestashop"
            elif 'stencil.bigcommerce.com' in link_content:  # BigCommerce often uses this in their URLs
                return "bigcommerce"
            elif 'skin/frontend' in link_content:  # Magento commonly has this in their URLs
                return "magento"
            elif 'woocommerce' in link_content:  
                return "woocommerce"

        return None

    except requests.exceptions.HTTPError as err:
        print(f"HTTP error occurred: {err}")
    except Exception as err:
        print(f"An error ocurred: {err}")


# # Test with your URL
# url = "https://www.biltongandbudz.co.za/"
# print(detect(url))
