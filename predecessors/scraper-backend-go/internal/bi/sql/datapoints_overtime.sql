SELECT
  DateCreated,
  MaxQty,
  Price
FROM
  `scrapers.datapoint_partitioned`
WHERE
  ProductIdentifier = '{{ .ProductIdentifier }}'
ORDER BY
  DateCreated;
