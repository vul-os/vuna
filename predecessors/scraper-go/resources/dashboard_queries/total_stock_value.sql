with themainq as (
    with that_query as (
        SELECT variation_id as varid, MAX(date_added) AS MaxDateTime, stock, price
        FROM datapoints
        WHERE stock > 0 and price > 0
        GROUP BY varid, stock, price
        ORDER BY stock DESC
    ) select p.product_name, p.store, tq.stock, tq.price
    from product_variations_stores p inner join that_query tq on tq.varid = p.id
    WHERE store_name = any(' {{ stores }} '::text[])
) SELECT 'R' || ' ' ||  to_char(ROUND(CAST(sum(price * stock) AS decimal), 2), 'FM999,999,999,999') FROM themainq