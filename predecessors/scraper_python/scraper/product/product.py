import uuid
from datetime import datetime
from typing import Optional, List, Dict, Any

from sqlalchemy.orm import Session

from db.models.product import Product
from db.models.variation import Variation
from db.models.datapoint import DataPoint
from db.models.attribute import Attribute

from scraper_loader import ScraperLoader


def scrape_product(site_id: str, scraper: ScraperLoader,
                   product_url: str, session: Session, proxies: Optional[dict] = None,
                   gcs_uploader: GCSUploader = None):
    # scraper_loader = ScraperLoader(site_id=site_id, proxies=proxies,
    #                                image_bucket_name=image_bucket_name)
    # scraper = scraper_loader()

    product_data = scraper(product_url)

    first_item = next(iter(product_data), None)

    # Validate product data
    if not scraper.validate_data(first_item):
        return

    # Merge product
    product = Product.merge(session=session, url=first_item["url"], site_id=site_id)

    # Save variation datapoints
    variation_sku_to_id = {}
    for p in product_data:
        
        variation_sku = p.get("sku")
        if not variation_sku:
            variation_sku = "default"

        variation_id = variation_sku_to_id.get(variation_sku)
        if not variation_id:
            variation = Variation.merge(
                session=session,
                sku=variation_sku,
                product_id=product.id,
            )
            variation_sku_to_id[variation_sku] = variation.id
            variation_id = variation.id

        datapoint = DataPoint.create(
            session=session,
            var_id=variation_id,
            max_qty=p["max_qty"],
            price=p["price"],
        )

        # Save variation attributes
        attributes_data = p.get("attributes", [])
        for attribute_data in attributes_data:
            attribute = Attribute.merge(
                session=session,
                name=attribute_data["name"],
                value=attribute_data["value"],
            )
            variation.attributes.append(attribute)

        # Save variation images
        if image_bucket_name and "image_urls" in p and len(p["image_urls"]) > 0:
            for image_url in p["image_urls"]:
                gcs_uploader.upload_image(image_url.strip(), site_id)

    # Update product metadata
    product.date_updated = datetime.now()

    session.commit()