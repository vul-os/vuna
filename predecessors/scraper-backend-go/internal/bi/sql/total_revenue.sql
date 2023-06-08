WITH diffs AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    maxqty - LAG(maxqty) OVER (PARTITION BY ProductIdentifier ORDER BY DateCreated) AS difference
  FROM
    `scrapers.datapoint_partitioned`
), 
the_query AS (
  SELECT
    DateCreated,
    ProductIdentifier,
    -1 * SUM(difference) AS Units_Sold
  FROM
    diffs
  WHERE difference < 0
  GROUP BY ProductIdentifier, DateCreated
), 
revenue_query AS (
  SELECT
    t.DateCreated,
    t.ProductIdentifier,
    t.Units_Sold * p.Price AS Revenue
  FROM
    the_query t
  JOIN
    `scrapers.datapoint_partitioned` p ON t.ProductIdentifier = p.ProductIdentifier AND t.DateCreated = p.DateCreated
  WHERE
      t.DateCreated BETWEEN TIMESTAMP('{{ .date_start }}') AND TIMESTAMP('{{ .date_end }}')
)
SELECT
  SUM(Revenue) as total_revenue
FROM
  revenue_query

