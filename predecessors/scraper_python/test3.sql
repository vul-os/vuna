WITH diffs_the_third as (
    WITH diffs_the_second as (
        WITH diffs as (
            SELECT
                product,
                date_scraped,
                stock - lag(stock) over (partition BY product ORDER BY date_scraped) as difference
            FROM
                datapoints
        )
        SELECT
            product,
            date_scraped,
            -1 * sum( difference ) as units_sold
        FROM
            diffs
        WHERE difference < 0
        GROUP BY date_scraped, product
    ) SELECT tq.product, tq.units_sold as units_sold, tq.date_scraped, prods.price as price
    FROM diffs_the_second tq JOIN products prods
    ON tq.product = prods.id
) SELECT date_scraped, count(units_sold * price) AS total_rev
FROM diffs_the_third
WHERE date_scraped BETWEEN '{{ datetime.start }}' AND '{{ datetime.end }}'
GROUP BY date_scraped
ORDER BY date_scraped


