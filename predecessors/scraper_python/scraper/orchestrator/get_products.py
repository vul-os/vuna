from sqlalchemy import create_engine

class Product:
    def __init__(self, ID, URL, Technology):
        self.ID = ID
        self.URL = URL
        self.Technology = Technology

def get_products(db_url, batch_size, offset):
    engine = create_engine(db_url)

    result = engine.execute(f"""
        WITH store_weights AS (
            SELECT site_id, 1.0 / COUNT(*) AS weight
            FROM products
            GROUP BY site_id
        ), ranked_products AS (
            SELECT p.site_id, p.product_id, p.url, ROW_NUMBER() OVER (PARTITION BY p.site_id ORDER BY p.product_id) - 1 AS row_num
            FROM products p
        ), site_technologies AS (
            SELECT s.site_id, s.technology
            FROM sites s
        )
        SELECT rp.product_id, rp.url, st.technology
        FROM (
            SELECT rp.product_id, rp.site_id, rp.url, rp.row_num, SUM(sw.weight) OVER (ORDER BY rp.row_num) AS cumulative_weight
            FROM ranked_products rp
            JOIN store_weights sw ON rp.site_id = sw.site_id
        ) ranked
        JOIN site_technologies st ON ranked.site_id = st.site_id
        ORDER BY ranked.cumulative_weight, ranked.row_num
        OFFSET {offset} ROWS
        FETCH NEXT {batch_size} ROWS ONLY;
    """)

    rows = result.fetchall()

    products = [Product(row.product_id, row.url, row.technology) for row in rows]

    return products