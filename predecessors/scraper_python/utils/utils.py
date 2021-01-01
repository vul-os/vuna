import psycopg2
import datetime


def connect():
    return psycopg2.connect(
        database="scrapers_2", user="scrapers", password="scrapers", host="38.17.53.117", port=17435)


def upsert_product(connection,
                   cursor,
                   store_id,
                   categories: dict,
                   product_name: str,
                   url: str,
                   product_price: float,
                   product_stock: int,
                   scrape_date: datetime):





    # Insert Products
    query = f"""
      INSERT INTO products (name, store, url, price, date_added)
      VALUES('{product_name}', {store_id}, '{url}', {product_price}, '{datetime.datetime.now()}')
      ON CONFLICT (url) DO UPDATE SET
            name = '{product_name}',
            store = {store_id},
            url = '{url}',
            price = {product_price},
            date_added = '{datetime.datetime.now()}'
        RETURNING id;
      """
    cursor.execute(query)
    connection.commit()

    # Get Inserted Product ID (hopefully) todo: be sure
    product_id = cursor.fetchone()
    product_id = product_id if not len(product_id) > 0 else product_id[0]
    if product_id:
        # insert data point
        print(product_id, product_stock, scrape_date, datetime.datetime.now())
        query = f"""
          INSERT INTO datapoints (product, stock, date_scraped, date_added)
          VALUES({product_id}, {product_stock}, '{scrape_date}', '{datetime.datetime.now()}')
          RETURNING id;
          """
        cursor.execute(query)
        connection.commit()

        for cat_name, data in categories.items():
            print(cat_name, data['url'], data['subcat'])
            query = f"""
              INSERT INTO categories (product, stock, date_scraped, date_added)
              VALUES({product_id}, {product_stock}, '{scrape_date}', '{datetime.datetime.now()}')
              RETURNING id;
              """
            cursor.execute(query)
            connection.commit()


