with the_query as (
    WITH diffs as (
        SELECT
            date_scraped,
            product,
            stock - lag(stock) over (partition BY product ORDER BY date_scraped) as difference
        FROM
            datapoints
        WHERE date_scraped BETWEEN  '{{ datetime.start }}' AND '{{ datetime.end }}'
    )
    SELECT
        date_scraped,
        product,
        -1 * sum( difference ) as "Units Sold"
    FROM
        diffs
    WHERE difference < 0
    GROUP BY product, date_scraped
) SELECT tq."Units Sold" * prods.price AS Revenue, prods.name, tq."Units Sold", prods.price, tq.product, prods.id, tq.date_scraped as date_scraped
FROM the_query tq JOIN products prods
                       ON tq.product = prods.id
ORDER BY Revenue desc
