from bs4 import BeautifulSoup


def get_max_pages(response):
    soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    soup = soup.find("nav", {"class": "woocommerce-pagination"})
    max_pages = soup.findAll("li")[-2].find('a').getText()
    return max_pages

def get_max_pages(self, response):
    soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    soup = soup.find('ul', {'class': 'page-numbers'}) \
        .findAll('a', {'class': 'page-numbers'})
    if len(soup) == 0:
        return 1
    return int(soup[-2].text.strip())

def get_max_pages(self, response):
    soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    soup = soup.find("nav", {"class": "woocommerce-pagination"}).find("ul", {
        "class": "page-numbers nav-pagination links text-center"})
    max_pages = soup.findAll("li")[-2].find('a').getText()
    # max_pages = 2
    return max_pages
