WITH latest_datapoint AS (
  SELECT
    ProductIdentifier, 
    DateCreated, 
    Price, 
    MaxQty,
    ROW_NUMBER() OVER(PARTITION BY ProductIdentifier ORDER BY DateCreated DESC) AS row_num
  FROM
    `scraping-is-hard.scrapers.datapoint_raw`
  WHERE 
    MaxQty > 0
), latest_product_info AS (
  SELECT
    ProductIdentifier, 
    DateCreated AS LatestDate, 
    Price, 
    MaxQty
  FROM
    latest_datapoint
  WHERE
    row_num = 1
)
SELECT
  SUM(Price * MaxQty) AS total_sum
FROM
  latest_product_info
