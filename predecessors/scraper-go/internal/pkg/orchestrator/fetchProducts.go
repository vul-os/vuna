package orchestrator

import (
	"gorm.io/gorm"
)

type Product struct {
    ID        uint
    URL       string
    Technology string
}

func FetchProducts(db *gorm.DB, batchSize int) ([]Product, error) {
    var products []Product
    var offset = 0

    for {
        rows, err := db.Raw(`
            WITH store_weights AS (
                SELECT site_id, 1.0 / COUNT(*) AS weight
                FROM products
                GROUP BY site_id
            ), ranked_products AS (
                SELECT p.site_id, p.product_id, p.url, ROW_NUMBER() OVER (PARTITION BY p.site_id ORDER BY p.product_id) - 1 AS row_num
                FROM products p
            ), site_technologies AS (
                SELECT s.site_id, s.technology
                FROM sites s
            )
            SELECT rp.product_id, rp.url, st.technology
            FROM (
                SELECT rp.product_id, rp.site_id, rp.url, rp.row_num, SUM(sw.weight) OVER (ORDER BY rp.row_num) AS cumulative_weight
                FROM ranked_products rp
                JOIN store_weights sw ON rp.site_id = sw.site_id
            ) ranked
            JOIN site_technologies st ON ranked.site_id = st.site_id
            ORDER BY ranked.cumulative_weight, ranked.row_num
            OFFSET ? ROWS
            FETCH NEXT ? ROWS ONLY;
        `, offset, batchSize).Rows()

        if err != nil {
            return nil, err
        }

        if rows == nil {
            break // no more rows
        }

        for rows.Next() {
            var product Product
            if err := db.ScanRows(rows, &product); err != nil {
                return nil, err
            }
            products = append(products, product)
        }

        offset += batchSize
    }

    return products, nil
}
