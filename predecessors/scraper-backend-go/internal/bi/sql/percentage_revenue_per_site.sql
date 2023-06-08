WITH diffs AS (
  SELECT
    d.DateCreated,
    d.ProductIdentifier,
    pr.SiteIdentifier,
    d.MaxQty - LAG(d.MaxQty) OVER (PARTITION BY d.ProductIdentifier ORDER BY d.DateCreated) AS difference
  FROM
    `scrapers.datapoint_partitioned` d
  JOIN
    `scrapers.product_unique` pr ON d.ProductIdentifier = pr.ProductIdentifier
  WHERE
    d.DateCreated BETWEEN '{{ .date_start }}' AND '{{ .date_end }}'
), 
the_query AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    SiteIdentifier,
    -1 * SUM(difference) AS Units_Sold
  FROM
    diffs
  WHERE difference < 0
  GROUP BY ProductIdentifier, SiteIdentifier, DateCreated
), 
revenue_query AS (
  SELECT
    t.DateCreated,
    t.ProductIdentifier,
    t.SiteIdentifier,
    t.Units_Sold * p.Price AS Revenue
  FROM
    the_query t
  JOIN
    `scrapers.datapoint_partitioned` p ON t.ProductIdentifier = p.ProductIdentifier AND t.DateCreated = p.DateCreated
),
site_revenue AS (
  SELECT
    s.Name,
    SUM(r.Revenue) as total_revenue
  FROM
    revenue_query r
  JOIN
    `scrapers.site_unique` s ON r.SiteIdentifier = s.Site_Identifier
  GROUP BY
    s.Name
),
total_revenue AS (
  SELECT
    SUM(total_revenue) as total_revenue
  FROM
    site_revenue
)
SELECT
  s.Name,
  (s.total_revenue / t.total_revenue) * 100 as revenue_percentage
FROM
  site_revenue s, total_revenue t
