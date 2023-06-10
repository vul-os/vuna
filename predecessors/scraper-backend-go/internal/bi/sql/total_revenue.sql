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
), revenue_data AS (
  SELECT 
    s.ProductIdentifier,
    MAX(d.price) * s.total_difference AS Total_Revenue
  FROM 
    sales_data s
  JOIN
    `scrapers.datapoint_partitioned` d ON s.ProductIdentifier = d.ProductIdentifier
  GROUP BY s.ProductIdentifier, s.total_difference
)
SELECT SUM(Total_Revenue) as total_revenue
FROM revenue_data;
