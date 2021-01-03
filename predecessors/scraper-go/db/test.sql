INSERT INTO product_variations (product_id, variation_id, sku, date_added, date_updated)
VALUES({product_id}, {variation_id_db}, '{sku}', {datetime.datetime.now()}, {datetime.datetime.now()})
ON CONFLICT (product_id, variation_id) DO UPDATE SET
    variation_id = {variation_id_db},
    product_id = {product_id},
    sku = '{sku}',
    date_updated = {datetime.datetime.now()}
RETURNING id;