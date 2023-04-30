create or replace view product_variations_stores as (
    select p.product_name , p.url, p.store, v.variation_id, v.variation_id_raw, v.sku, s.store_name
    from products p
             inner join product_variations pv on pv.product_id = p.id
             inner join variations v on v.id = pv.variation_id
             inner join stores s on p.store = s.id
)