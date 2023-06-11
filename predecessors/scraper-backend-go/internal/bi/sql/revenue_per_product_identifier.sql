WITH diffs AS (
  SELECT
    ProductIdentifier,
    maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated) AS difference
  FROM
    `scrapers.datapoint_raw`
), filtered_diffs AS (
  SELECT
    ProductIdentifier,
    CASE
      WHEN difference > 0 THEN 0
      ELSE -difference
    END AS positive_difference
  FROM diffs
  WHERE difference IS NOT NULL
), sales_data AS (
  SELECT
    ProductIdentifier,
    SUM(positive_difference) AS total_difference
  FROM filtered_diffs
  GROUP BY ProductIdentifier
)
SELECT 
  s.ProductIdentifier,
  p.Name AS ProductName,
  SUM(d.price * s.total_difference) AS Total_Revenue
FROM 
  sales_data s
JOIN
  `scrapers.datapoint_raw` d ON s.ProductIdentifier = d.ProductIdentifier
JOIN
  `scrapers.product_unique` p ON s.ProductIdentifier = p.ProductIdentifier
WHERE s.total_difference > 0
GROUP BY s.ProductIdentifier, p.Name, s.total_difference
ORDER BY Total_Revenue DESC
LIMIT 100