WITH 
distinct_products AS (
  SELECT DISTINCT
    ProductIdentifier
  FROM `scrapers.datapoint_partitioned`
),
period_1_sales AS (
  SELECT 
    ProductIdentifier, 
    SUM(-1 * difference * Price) AS revenue
  FROM 
    (SELECT
       DateCreated,
       ProductIdentifier,
       IFNULL(maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated), 0) AS difference,
       Price
     FROM `scrapers.datapoint_partitioned`
    ) t
  WHERE 
    difference < 0 AND DateCreated BETWEEN TIMESTAMP(DATE_SUB(PARSE_TIMESTAMP('%Y-%m-%d', '{{ .start_date }}'), INTERVAL 7 DAY)) AND TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '{{ .start_date }}'))
  GROUP BY ProductIdentifier
),
period_2_sales AS (
  SELECT 
    ProductIdentifier, 
    SUM(-1 * difference * Price) AS revenue
  FROM 
    (SELECT
       DateCreated,
       ProductIdentifier,
       IFNULL(maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated), 0) AS difference,
       Price
     FROM `scrapers.datapoint_partitioned`
    ) t
  WHERE 
    difference < 0 AND DateCreated BETWEEN TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '{{ .start_date }}')) AND TIMESTAMP(PARSE_TIMESTAMP('%Y-%m-%d', '{{ .end_date }}'))
  GROUP BY ProductIdentifier
),
period_1_rank AS (
  SELECT 
    ProductIdentifier, 
    RANK() OVER(ORDER BY revenue DESC) as rank
  FROM period_1_sales
),
period_2_rank AS (
  SELECT 
    ProductIdentifier, 
    RANK() OVER(ORDER BY revenue DESC) as rank,
    revenue AS current_period_revenue
  FROM period_2_sales
),
current_values AS (
  SELECT
    ProductIdentifier,
    Price AS current_price,
    maxqty AS current_maxqty
  FROM (
    SELECT 
      ProductIdentifier,
      Price,
      maxqty,
      ROW_NUMBER() OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated DESC) as rn
    FROM
      `scrapers.datapoint_partitioned`
  ) 
  WHERE rn = 1
)
SELECT 
  dp.ProductIdentifier, 
  IFNULL(p2.rank, 0) AS Rank,
  IFNULL(p1.rank, 0) - IFNULL(p2.rank, 0) AS RankChange,
  IFNULL(p2.current_period_revenue, 0) AS Revenue,
  c.current_price as Price,
  c.current_maxqty as MaxQty,
  c.current_price * c.current_maxqty as SalesValue,
  p.name AS ProductName,
  p.ImageUrls AS ImageUrls,
  p.URL as ProductUrl
FROM 
  distinct_products dp
LEFT JOIN 
  period_1_rank p1 ON dp.ProductIdentifier = p1.ProductIdentifier
LEFT JOIN 
  period_2_rank p2 ON dp.ProductIdentifier = p2.ProductIdentifier
LEFT JOIN
  current_values c ON dp.ProductIdentifier = c.ProductIdentifier
LEFT JOIN
  `scrapers.product_unique` p ON dp.ProductIdentifier = p.ProductIdentifier
WHERE c.current_maxqty > 0 -- Added condition
ORDER BY RankChange DESC
