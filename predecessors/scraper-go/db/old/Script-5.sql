

with the_query as (
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
        -1 * sum( difference ) as "Units Sold"
    FROM
        diffs
    WHERE difference < 0
    GROUP BY product, date_scraped
) select * from the_query