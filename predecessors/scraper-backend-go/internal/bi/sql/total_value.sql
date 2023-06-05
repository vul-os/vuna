SELECT SUM(PriceStock) AS TotalPriceStock
FROM (
  SELECT DISTINCT ProductIdentifier, MAX(DateCreated) AS LatestDate, SUM(Price * MaxQty) AS PriceStock
  FROM `scraping-is-hard.scrapers.datapoint_raw`
  WHERE MaxQty > 0
  GROUP BY ProductIdentifier
) AS subquery;
