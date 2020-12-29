SELECT dp.product, dp.stock, prods.id, prods.name
FROM datapoints dp JOIN products prods
ON dp.product = prods.id
INNER JOIN
    (SELECT product as prod, MAX(date_scraped) AS MaxDateTime
    FROM datapoints
    GROUP BY prod) date_grouped_latest
ON dp.product = date_grouped_latest.prod
AND date_scraped = date_grouped_latest.MaxDateTime

SELECT dp.product, dp.stock, dp.date_scraped, prods.id, prods.name
FROM datapoints dp JOIN products prods
ON dp.product = prods.id
INNER JOIN
    (SELECT product as prod, MAX(date_scraped) AS MaxDateTime
    FROM datapoints
    GROUP BY prod) date_grouped_latest
ON dp.product = date_grouped_latest.prod
AND date_scraped = date_grouped_latest.MaxDateTime
ORDER BY dp.stock DESC




SELECT dp.product, dp.stock, dp.date_scraped, prods.id, prods.name
FROM datapoints dp JOIN products prods
ON dp.product = prods.id
INNER JOIN
    (SELECT product as prod, MAX(date_scraped) AS MaxDateTime
    FROM datapoints
    GROUP BY prod) date_grouped_latest
ON dp.product = date_grouped_latest.prod
AND date_scraped = date_grouped_latest.MaxDateTime
ORDER BY dp.stock DESC

