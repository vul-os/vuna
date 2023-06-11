SELECT
  DateCreated,
  MaxQty,
  Price
FROM
  `scrapers.datapoint_raw`
WHERE
  ProductIdentifier = '{{ .product_identifier }}'
ORDER BY
  DateCreated;
