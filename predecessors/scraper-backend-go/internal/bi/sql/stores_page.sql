WITH
product_sites AS (
  SELECT DISTINCT
    ProductIdentifier,
    SiteIdentifier
  FROM `scrapers.product_unique`
),
distinct_products AS (
  SELECT DISTINCT
    ps.SiteIdentifier,
    dp.ProductIdentifier
  FROM `scrapers.datapoint_partitioned` dp
  JOIN product_sites ps ON dp.ProductIdentifier = ps.ProductIdentifier
),
period_1_sales AS (
  SELECT 
    ps.SiteIdentifier,
    dp.ProductIdentifier, 
    SUM(-1 * dp.difference * dp.Price) AS revenue
  FROM 
    (SELECT
       DateCreated,
       ProductIdentifier,
       IFNULL(maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated), 0) AS difference,
       Price
     FROM `scrapers.datapoint_partitioned`
    ) dp
  JOIN product_sites ps ON dp.ProductIdentifier = ps.ProductIdentifier
  WHERE 
    dp.difference < 0 AND dp.DateCreated BETWEEN TIMESTAMP(DATE_SUB(PARSE_TIMESTAMP('%Y-%m-%d', '2023-06-01'), INTERVAL 7 DAY)) AND TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '2023-06-01'))
  GROUP BY ps.SiteIdentifier, dp.ProductIdentifier
),
period_2_sales AS (
  SELECT 
    ps.SiteIdentifier,
    dp.ProductIdentifier, 
    SUM(-1 * dp.difference * dp.Price) AS revenue
  FROM 
    (SELECT
       DateCreated,
       ProductIdentifier,
       IFNULL(maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated), 0) AS difference,
       Price
     FROM `scrapers.datapoint_partitioned`
    ) dp
  JOIN product_sites ps ON dp.ProductIdentifier = ps.ProductIdentifier
  WHERE 
    dp.difference < 0 AND dp.DateCreated BETWEEN TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '2023-06-01')) AND TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '2023-06-08'))
  GROUP BY ps.SiteIdentifier, dp.ProductIdentifier
),
period_1_rank AS (
  SELECT 
    SiteIdentifier,
    ProductIdentifier, 
    RANK() OVER(PARTITION BY SiteIdentifier ORDER BY revenue DESC) as rank
  FROM period_1_sales
),
period_2_rank AS (
  SELECT 
    SiteIdentifier,
    ProductIdentifier, 
    RANK() OVER(PARTITION BY SiteIdentifier ORDER BY revenue DESC) as rank,
    revenue AS current_period_revenue
  FROM period_2_sales
),
current_values AS (
  SELECT
    ps.SiteIdentifier,
    dp.ProductIdentifier,
    dp.Price AS current_price,
    dp.maxqty AS current_maxqty
  FROM (
    SELECT 
      ProductIdentifier,
      Price,
      maxqty,
      ROW_NUMBER() OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated DESC) as rn
    FROM
      `scrapers.datapoint_partitioned`
  ) dp
  JOIN product_sites ps ON dp.ProductIdentifier = ps.ProductIdentifier
  WHERE dp.rn = 1
),
site_values AS (
  SELECT 
    SiteIdentifier,
    SUM(current_price * current_maxqty) as TotalValue
  FROM current_values
  GROUP BY SiteIdentifier
)
SELECT 
  su.SiteIdentifier, 
  su.Name AS SiteName,
  su.Url AS SiteUrl,
  su.Image AS SiteImage,
  sv.TotalValue,
  IFNULL(p2.rank, 0) AS Rank,
  IFNULL(p1.rank, 0) - IFNULL(p2.rank, 0) AS RankChange
FROM 
  `scrapers.site_unique` su
JOIN 
  site_values sv ON su.SiteIdentifier = sv.SiteIdentifier
LEFT JOIN 
  period_1_rank p1 ON su.SiteIdentifier = p1.SiteIdentifier
LEFT JOIN 
  period_2_rank p2 ON su.SiteIdentifier = p2.SiteIdentifier
ORDER BY RankChange DESC
