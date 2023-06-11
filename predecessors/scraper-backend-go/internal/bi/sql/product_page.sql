WITH diffs AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    maxqty - IFNULL(LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated), 0) AS difference
  FROM
    `scrapers.datapoint_partitioned`
), the_query AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    -1 * SUM(difference) AS Units_Sold
  FROM
    diffs
  WHERE difference < 0
    -- AND ProductIdentifier = 'Your_Product_Id' -- replace with your actual product id
  GROUP BY ProductIdentifier, DateCreated
), revenue_query AS (
  SELECT
    t.ProductIdentifier,
    SUM(t.Units_Sold * p.Price) AS Total_Revenue
  FROM
    the_query t
  JOIN
    `scrapers.datapoint_partitioned` p ON t.ProductIdentifier = p.ProductIdentifier AND t.DateCreated = p.DateCreated
  GROUP BY t.ProductIdentifier
), product_data AS (
  SELECT
    p.ProductIdentifier,
    p.Name,
    p.ImageURLs,
    p.Url AS producturl,
    s.Url AS siteurl,
    s.Image,
    d.Price AS latest_price,
    d.MaxQty AS latest_maxqty
  FROM
    `scrapers.product_unique` p
  JOIN
    `scrapers.site_unique` s ON p.SiteIdentifier = s.SiteIdentifier
  JOIN
    (
      SELECT
        ProductIdentifier,
        Price,
        MaxQty,
        ROW_NUMBER() OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated DESC) AS row_num
      FROM
        `scrapers.datapoint_raw`
    ) d ON p.ProductIdentifier = d.ProductIdentifier AND d.row_num = 1
)
SELECT 
  p.ProductIdentifier,
  p.Name,
  SUM(d.price * r.Total_Revenue) AS Total_Revenue,
  p.producturl,
  p.siteurl,
  p.imageurls,
  p.image,
  p.latest_price as price,
  p.latest_maxqty as max_qty
FROM 
  revenue_query r
JOIN
  `scrapers.datapoint_raw` d ON r.ProductIdentifier = d.ProductIdentifier
JOIN
  product_data p ON r.ProductIdentifier = p.ProductIdentifier
GROUP BY p.ProductIdentifier, p.Name, p.producturl, p.siteurl, p.imageurls, p.image, p.latest_price, p.latest_maxqty
ORDER BY Total_Revenue DESC;
