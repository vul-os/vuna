from bs4 import BeautifulSoup
import datetime
import json


def get_variation_data(self, json_data, product_name, url):
    for data in json_data:
        variation_id = data['variation_id']
        product_price = data['display_price']
        product_stock = data['max_qty']
        if not product_stock:
            product_stock = data['availability_html'].replace('<p class="stock in-stock">', '') \
                .replace('</p>', '').strip().replace('in', '').replace('stock', '').strip()
        # pack_size = data['attributes']['attribute_pa_pack-size'].replace("-", "").replace("seeds", "").strip()

        print("Variations: ", product_name, product_price, product_stock, variation_id, url,
              datetime.datetime.utcnow())


def get_product_data(self, response):
    soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    soup = soup.select_one('.summary,.entry-summary,.product-summary,.product-info')
    product_name = soup.select_one('.product_title,.entry-title').getText().strip()
    print(product_name)
    variations = soup.find('table', {'class': 'variations'})
    if variations is not None:
        raw_json_string = str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations'])
        try:
            json_data = json.loads(raw_json_string)
            self.get_variation_data(json_data, response.request.url, product_name)
        except Exception as e:
            print(raw_json_string)

        # return self.get_variation_data(json_data, response.request.url, product_name)
    else:
        product_stock = soup.find("p", {"class": "stock in-stock"})
        product_stock = product_stock.getText().strip() if product_stock else 0
        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
        product_price = product_price.getText().replace("R", "") if product_price else 0
        variation_id = soup.find('button', {'class': 'single_add_to_cart_button button alt'})
        if variation_id is None:
            variation_id = soup.find('input', {'class': 'cwg-product-id'})
        variation_id = variation_id['value'] if variation_id is not None else None

        print("No Variations: ", product_name, product_price, product_stock, variation_id,
              datetime.datetime.utcnow())
# def get_product_data(self, response):
#     soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#     soup = soup.find('div', {'class': 'summary entry-summary'})
#     product_name = soup.find("h1", {"class": "product_title entry-title"}) \
#         .getText().strip()
#
#     product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
#     product_stock = soup.find("p", {"class": "stock in-stock"})
#     product_stock = product_stock \
#         .getText() \
#         .replace("in", "") \
#         .replace("stock", "") \
#         .strip() \
#         if product_stock is not None else 0
#     # product_short_disc = soup.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
#     # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
#     # product_category_link = product_category_['href']
#     # product_category = product_category_.getText()
#
#     print(product_name, product_price, product_stock)
#
#     return None
#
# ##
#
#
#
#
#
# ##
#
#     def get_product_data(self, response):
#         soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#         soup = soup.find("div", {"class": "summary entry-summary"})
#
#         product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
#         product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
#         product_stock_str = soup.find("p", {"class": "stock in-stock"})
#         product_stock_ = product_stock_str.getText() \
#                              .replace("in", "") \
#                              .replace("stock", "") \
#                              .replace("(can be backordered)", "") \
#                              .strip() if product_stock_str is not None else 0
#
#         product_stock = int(product_stock_) if product_stock_ != '' else 0
#
#         # self.count = self.count + 1
#         print(product_name, product_price, product_stock, response.request.url, datetime.datetime.utcnow())
#         return None
#
# #
#     def get_product_data(self, response):
#         soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#         soup = soup.find("div", {"class": "summary entry-summary"})
#
#         product_name = soup.find("h1", {"class": "product_title entry-title"}).getText()
#         product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
#         product_stock = soup.find("div", {"class": "quantity"})
#         product_stock = product_stock.find("input", {"title": "Qty"}) if product_stock is not None else 0
#         product_stock = product_stock['max'] if product_stock is not None else 0
#         product_stock = int(product_stock) if product_stock != '' else 0
#
#
#         print(product_name, product_price, product_stock)
#
#
#         return None
#
#
# #
#
#
#
#     def get_product_data(self, response):
#         soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#         soup = soup.find('div', {'class': 'product-info summary col-fit col entry-summary product-summary'})
#         product_name = soup.find("h1", {"class": "product-title product_title entry-title"}).getText().strip()
#         variations = soup.find('table', {'class': 'variations'})
#         if variations is not None:
#             json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
#             print("Variations: ", self.get_variation_data(json_data, response.request.url, product_name))
#             return self.get_variation_data(json_data, response.request.url, product_name)
#         else:
#             product_stock = soup.find("p", {"class": "stock in-stock"})
#             product_stock = product_stock.getText().strip() if product_stock else 0
#             product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
#             product_price = product_price.getText().replace("R", "") if product_price else 0
#             variation_id = soup.find('button', {'class': 'single_add_to_cart_button button alt'})
#             if variation_id is None:
#                 variation_id = soup.find('input', {'class': 'cwg-product-id'})
#             variation_id = variation_id['value'] if variation_id is not None else None
#
#             print("No Variations: ", product_name, product_price, product_stock, variation_id,
#                   datetime.datetime.utcnow())
#
#             yield None
#
# #
#
#
#
#
#     def get_product_data(self, response):
#         soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#         soup = soup.find('div', {'class': 'summary entry-summary'})
#         product_name = soup.find("h1", {"class": "product_title entry-title"}) \
#             .getText().strip()
#         variations = soup.find('table', {'class': 'variations'})
#         if variations is not None:
#             json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
#             self.get_variation_data(json_data, response.request.url, product_name)
#             # return self.get_variation_data(json_data, response.request.url, product_name)
#         else:
#             product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
#             product_stock = soup.find("div", {"class": "quantity"})
#             product_stock = product_stock.find("input", {"title": "Qty"}) if product_stock is not None else 0
#             product_stock = product_stock['max'] if product_stock is not 0 else 0
#             product_stock = int(product_stock) if product_stock != '' else 0
#             # product_short_disc = soup.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
#             # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
#             # product_category_link = product_category_['href']
#             # product_category = product_category_.getText()
#
#             print(product_name, product_price, product_stock)
#
#         return None
#
# #
#
#
#     def get_product_data(self, response):
#         soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
#         soup = soup.find('div', {'class': 'summary entry-summary'})
#         product_name = soup.find("h1", {"class": "product_title entry-title"}) \
#             .getText().strip()
#         variations = soup.find('table', {'class': 'variations'})
#         if variations is not None:
#             json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
#             self.get_variation_data(json_data, response.request.url, product_name)
#             # return self.get_variation_data(json_data, response.request.url, product_name)
#         else:
#             product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"})
#             if product_price is None:
#                 return None
#             product_price = product_price.getText().replace("R", "")
#             product_stock = soup.find("div", {"class": "quantity"})
#             if product_stock is None:
#                 product_stock = 0
#             else:
#                 product_stock = product_stock.find("input", {"title": "Qty"})
#                 if product_stock is None:
#                     product_stock = 0
#                 else:
#                     if len(product_stock) > 0:
#                         product_stock = product_stock['max']
#                         product_stock = int(product_stock) if product_stock != '' else 0
#
#             # product_short_disc = soup.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
#             # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
#             # product_category_link = product_category_['href']
#             # product_category = product_category_.getText()
#
#             print(product_name, product_price, product_stock)
#
#         return None
#
# #
#

    #
    # def get_product_data(self, response):
    #     soup = BeautifulSoup(response.body.decode('utf-8'), 'html5lib')
    #     soup = soup.find("div", {"class": "summary entry-summary"})
    #     product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
    #     variations = soup.find('table', {'class': 'variations'})
    #     if variations is not None:
    #         json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))
    #         print("Variations: ", self.get_variation_data(json_data, response.request.url, product_name))
    #         return self.get_variation_data(json_data, response.request.url, product_name)
    #     else:
    #         product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
    #         product_stock = soup.find("p", {"class": "stock in-stock"}).getText() \
    #             .replace("in", "") \
    #             .replace("stock", "") \
    #             .strip()
    #         product_stock = int(product_stock) if int(product_stock.isdigit()) else 0
    #         product_category = soup.find("span", {"class": "posted_in"}).findAll("a")
    #         if len(product_category) > 1:
    #             print("FUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUCK")
    #
    #         print("No Variations: ", product_name, product_price, product_stock, datetime.datetime.utcnow())
    #
    #         return None