

WITH diffs as (
    SELECT
        product,
        stock - lag(stock) over (partition BY product ORDER BY date_scraped) as difference
    FROM
        datapoints
)
SELECT
    product,
    sum( difference ) as all_diffs
FROM
    diffs
GROUP BY product;




with the_query as (
    WITH diffs as (
        SELECT
            product,
            stock - lag(stock) over (partition BY product ORDER BY date_scraped) as difference
        FROM
            datapoints
    )
    SELECT
        product,
        -1 * sum( difference ) as "Units Sold"
    FROM
        diffs
    WHERE difference < 0
    GROUP BY product
) SELECT tq.product, tq."Units Sold",  prods.id, prods.name, prods.price, tq."Units Sold" * prods.price AS Revenue
FROM the_query tq JOIN products prods
ON tq.product = prods.id
ORDER BY tq."Units Sold" desc
