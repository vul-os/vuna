from bs4 import BeautifulSoup


def get_products_on_page(self, response):
    soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    soup = soup.find('ul', {'class': 'products'}, partial=True) \
               .find('li', {'class': 'product'}, partial=True) \
               .findAll('a')
    if len(soup) == 0:
        return None
    product_links = [s['href'] for s in soup]

    product_links = [s['href'] for s in soup]
    product_links = [s['href'] for s in soup]
    product_links = [s['href'] for s in soup]
    product_links = [s['href'] for s in soup]

    product_links = (str(links.find('a')['href']) for links in soup)
    product_links = (str(links.find('a')['href']) for links in soup)

    product_links = (str(links.find('a')['href']) for links in soup)
    product_links = [str(links.find('a')['href']) for links in soup]
    print(product_links)

    soup = soup.find('ul', {'class': 'products columns-4'})\
               .findAll('li')


    soup = soup.find('ul', {'class': 'products columns-3'})\
               .findAll('li')

    soup = soup.find('ul', {'class': 'products clearfix products-3'}) \
               .findAll('a', {'class': 'product-images'})


    soup = soup.find('ul', {'class': 'products columns-3'})\
               .findAll('li')

    soup = soup.find('ul', {'class': 'products clearfix products-3'}) \
        .findAll('a', {'class': 'product-images'})

    soup = soup.find('ul', {'class': 'products columns-3'}) \
        .findAll('a', {'class': 'woocommerce-LoopProduct-link woocommerce-loop-product__link'})


    soup = soup.find('div', {'class': 'shop-container'}) \
        .findAll('div', {'class': 'image-fade_in_back'})

    return product_links