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
    `scrapers.datapoint_raw` d ON s.ProductIdentifier = d.ProductIdentifier
  GROUP BY s.ProductIdentifier, s.total_difference
)
SELECT 
  p.SiteIdentifier,
  si.Url,
  SUM(r.Total_Revenue) as total_revenue_per_site,
  (SUM(r.Total_Revenue) / (SELECT SUM(Total_Revenue) FROM revenue_data)) * 100 AS percentage_of_total
FROM 
  revenue_data r
JOIN
  `scrapers.product_raw` p ON r.ProductIdentifier = p.ProductIdentifier
JOIN
  `scrapers.site_raw` si ON p.SiteIdentifier = si.site_identifier
GROUP BY p.SiteIdentifier, si.Url;
