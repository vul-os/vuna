import os
import psycopg2
import datetime
from urllib.parse import urlparse


def connect():
    return psycopg2.connect(
        database="scrapers_ecom", user="scrapers", password="scrapers", host="38.17.53.117", port=17435)


def upsert_product(cursor,
                   connection,
                   sku: str,
                   store_url: str,
                   url: str,
                   variation_id: str,
                   product_name: str,
                   categories: list,
                   tags: list,
                   attributes: dict,
                   product_price: float,
                   product_stock: int,
                   scrape_date: datetime):
    # find store id from store url
    query = f"""
      SELECT id FROM stores WHERE '{store_url.strip('/')}' LIKE '%' || url || '%';
      """
    cursor.execute(query)
    connection.commit()

    store_id = cursor.fetchone()
    if not store_id:
        print("No Store found")
        return None
    store_id = store_id if not len(store_id) > 0 else store_id[0]

    # find categories from store id
    #   if not found add categories
    if categories:
        for cat in categories:
            query = f"""
              INSERT INTO categories (category_name, url, store, date_added, date_updated)
              VALUES('{cat['name']}', '{cat['url']}', '{store_id}', '{datetime.datetime.now()}', '{datetime.datetime.now()}')
              ON CONFLICT (url) DO UPDATE SET
                    category_name = '{cat['name']}',
                    url = '{cat['url']}',
                    store = {store_id},
                    date_updated = '{datetime.datetime.now()}'
                RETURNING id;
              """
            cursor.execute(query)
            connection.commit()

            cat_id = cursor.fetchone()  # Get Inserted Product ID (hopefully) todo: be sure
            cat_id = cat_id if not len(cat_id) > 0 else cat_id[0]
            print(f"Category Upserted -> ID: {cat_id}, Name: {cat['name']}, URL: {cat['url']}")

    # find tags from store url
    #   if not found add tags
    if tags:
        for tag in tags:
            query = f"""
              INSERT INTO tags (tag_name, url, store, date_added, date_updated)
              VALUES('{tag['name']}', '{tag['url']}', '{store_id}', '{datetime.datetime.now()}', '{datetime.datetime.now()}')
              ON CONFLICT (url) DO UPDATE SET
                    tag_name = '{tag['name']}',
                    url = '{tag['url']}',
                    store = {store_id},
                    date_updated = '{datetime.datetime.now()}'
                RETURNING id;
              """
            cursor.execute(query)
            connection.commit()

            cat_id = cursor.fetchone()
            cat_id = cat_id if not len(cat_id) > 0 else cat_id[0]
            print(f"Tag Upserted -> ID: {cat_id}, Name: {tag['name']}, URL: {tag['url']}")

    # Insert Products -> WHERE '{url}' LIKE '%' || url || '%'
    url = urlparse(url).path.strip('/')
    query = f"""
      INSERT INTO products (product_name, url, store, date_added, date_updated)
      VALUES('{product_name}', '{url}', '{store_id}', '{datetime.datetime.now()}', '{datetime.datetime.now()}')
      ON CONFLICT (url) DO UPDATE SET
            product_name = '{product_name}',
            url = '{url}',
            store = {store_id},
            date_updated = '{datetime.datetime.now()}'
        RETURNING id;
      """
    cursor.execute(query)
    connection.commit()

    product_id = cursor.fetchone()  # Get Inserted Product ID (hopefully) todo: be sure
    product_id = product_id if not len(product_id) > 0 else product_id[0]

    if attributes:
        query = f"""
          INSERT INTO attributes (attribute_name, url, store, date_added, date_updated)
          VALUES('{attributes['name']}', '{attributes['url']}', '{store_id}', 
            '{datetime.datetime.now()}', '{datetime.datetime.now()}')
          ON CONFLICT (url) DO UPDATE SET
                attribute_name = '{attributes['name']}',
                url = '{attributes['url']}',
                store = {store_id},
                date_updated = '{datetime.datetime.now()}'
            RETURNING id;
          """
        cursor.execute(query)
        connection.commit()

        atrr_id = cursor.fetchone()
        atrr_id = atrr_id if not len(atrr_id) > 0 else atrr_id[0]
        print(f"Attribute Upserted -> ID: {atrr_id}, Name: {attributes['name']}, URL: {attributes['url']}")

        query = f"""
          INSERT INTO product_attributes (product_id, attribute_id, date_added, date_updated)
          VALUES('{product_id}', '{atrr_id}', '{datetime.datetime.now()}', '{datetime.datetime.now()}')
          ON CONFLICT (url) DO UPDATE SET
                store = {store_id},
                date_updated = '{datetime.datetime.now()}'
            RETURNING id;
          """
        cursor.execute(query)
        connection.commit()

        product_attributes_id = cursor.fetchone()
        product_attributes_id = product_attributes_id \
            if not len(product_attributes_id) > 0 else product_attributes_id[0]

        # variations
        query = f"""
          INSERT INTO variations (product_id, attribute_id, date_added, date_updated)
          VALUES('{product_id}', '{atrr_id}', '{datetime.datetime.now()}', '{datetime.datetime.now()}')
          ON CONFLICT (url) DO UPDATE SET
                store = {store_id},
                date_updated = '{datetime.datetime.now()}'
            RETURNING id;
          """
        cursor.execute(query)
        connection.commit()

    # find variation from url or (variation id + store id)
    #   if not found add variation



    # find attributes from store url
    #   if not found add attributes

    # insert DataPoints

    # # Insert Products
    # query = f"""
    #   INSERT INTO products (name, store, url, price, date_added)
    #   VALUES('{product_name}', {store_id}, '{url}', {product_price}, '{datetime.datetime.now()}')
    #   ON CONFLICT (url) DO UPDATE SET
    #         name = '{product_name}',
    #         store = {store_id},
    #         url = '{url}',
    #         price = {product_price},
    #         date_added = '{datetime.datetime.now()}'
    #     RETURNING id;
    #   """
    # cursor.execute(query)
    # connection.commit()
    #
    # # Get Inserted Product ID (hopefully) todo: be sure
    # product_id = cursor.fetchone()
    # product_id = product_id if not len(product_id) > 0 else product_id[0]
    # if product_id:
    #     # insert data point
    #     print(product_id, product_stock, scrape_date, datetime.datetime.now())
    #     query = f"""
    #       INSERT INTO datapoints (product, stock, date_scraped, date_added)
    #       VALUES({product_id}, {product_stock}, '{scrape_date}', '{datetime.datetime.now()}')
    #       RETURNING id;
    #       """
    #     cursor.execute(query)
    #     connection.commit()
    #
    #     for cat_name, data in categories.items():
    #         print(cat_name, data['url'], data['subcat'])
    #         query = f"""
    #           INSERT INTO categories (product, stock, date_scraped, date_added)
    #           VALUES({product_id}, {product_stock}, '{scrape_date}', '{datetime.datetime.now()}')
    #           RETURNING id;
    #           """
    #         cursor.execute(query)
    #         connection.commit()


# if __name__ == '__main__':
#     connection = connect()
#     cursor = connection.cursor()
#
#     # stores = {
#     #     "Biltong & Budz": "biltongandbudz.co.za",
#     #     "Trophy Seeds": "trophyseeds.com",
#     #     "Bot Shop": "botshop.co.za",
#     # }
#     # for name, url in stores.items():
#     #     cursor.execute(f"""
#     #         INSERT INTO stores (store_name, url, date_added)
#     #         VALUES ('{name}', '{url}', '{datetime.datetime.now()}')
#     #     """)
#     # connection.commit()
#
#     upsert_product(cursor, connection, sku='29419',
#                    store_url='https://www.trophyseeds.com',
#                    url='https://www.trophyseeds.com/product/garden-of-green-phantom-cookies-domina-5-pack-3-free/',
#                    variation_id='29419',
#                    product_name='Garden of Green – Phantom Cookies Domina (5-Pack) + 3 Free',
#                    categories=[None],
#                    tags=[None],
#                    attributes={},
#                    product_price=0,
#                    product_stock=0,
#                    scrape_date=datetime.datetime.now())
#
#     cursor.close()
#     connection.close()