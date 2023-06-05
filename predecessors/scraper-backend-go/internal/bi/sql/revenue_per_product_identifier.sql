WITH diffs AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated) AS difference
  FROM
    `scrapers.datapoint_raw`
), the_query AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    -1 * SUM(difference) AS Units_Sold
  FROM
    diffs
  WHERE difference < 0
  GROUP BY ProductIdentifier, DateCreated
), revenue_query AS (
  SELECT
    t.ProductIdentifier,
    SUM(t.Units_Sold * p.Price) AS Total_Revenue
  FROM
    the_query t
  JOIN
    `scrapers.datapoint_raw` p ON t.ProductIdentifier = p.ProductIdentifier AND t.DateCreated = p.DateCreated
  WHERE
    t.DateCreated  BETWEEN '{{ .date_start }}' AND '{{ .date_end }}'
  GROUP BY t.ProductIdentifier
)
SELECT DISTINCT
  r.ProductIdentifier,
  p.Name AS ProductName,
  r.Total_Revenue
FROM
  revenue_query r
JOIN
  `scrapers.product_raw` p ON r.ProductIdentifier = p.ProductIdentifier
ORDER BY r.Total_Revenue DESC, r.ProductIdentifier
LIMIT 250