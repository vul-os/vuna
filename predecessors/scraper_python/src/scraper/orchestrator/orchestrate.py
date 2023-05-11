import json
import requests
from typing import List
from sqlalchemy import Engine
from google.cloud import tasks_v2
from src.db.product import Product


class Orchestrator: 
    def __init__(self, engine: Engine, project_id: str, location: str, queue_id: str, 
        base_url: str, batch_size: int):
        self.engine = engine

        self.project_id = project_id
        self.location = location
        self.queue_id = queue_id
        self.client = tasks_v2.CloudTasksClient()
        self.queue_path = self.client.queue_path(self.project_id, self.location, self.queue_id)

        self.base_url = base_url
        self.batch_size = batch_size


    def create_task_for_product(self, product: Product):
        technology = requests.utils.quote(product.technology)
        url_encoded = requests.utils.quote(product.url)
        url = f"{self.base_url}/?technology={technology}&url={url_encoded}"

        task = {
            "http_request": {  # Specify the type of request.
                "http_method": tasks_v2.HttpMethod.GET,
                "url": url,  # The full url path that the task will be sent to.
                "headers": {
                    "Content-Type": "application/json",
                },
            },
            "name": f"product-task-{product.id}",
        }
        # Add payload to the request
        payload = json.dumps({
            "product_id": product.id,
            "url": product.url,
            "technology": product.technology
        })
        converted_payload = payload.encode()
        task["http_request"]["body"] = converted_payload

        response = self.client.create_task(request={"parent": self.queue_path, "task": task})
        return response

    def create_tasks_from_database(self):
        offset = 0
        while True:
            products = self.get_products(offset)

            if not products:
                break

            for product in products:
                task = self.create_task_for_product(product)
                print(f"Created task: {task.name}")

            offset += self.batch_size

    def get_products(self, offset: int) -> List[Product]:
        """
        Retrieve a batch of products from the database, sorted by site and evenly distributed.

        Args:
            engine (sqlalchemy.engine.Engine): The database engine to use for the query.
            batch_size (int): The size of the batch to retrieve.
            offset (int): The offset to use for the query.

        Returns:
            List[Product]: A list of `Product` objects representing the retrieved products.
        """
        result = self.engine.execute("""
            WITH store_weights AS (
                SELECT site_id, 
                    LEAST(CAST(COUNT(*) AS FLOAT) / CAST(:batch_size AS FLOAT) * 0.1, 1.0) AS weight
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
                WHERE sw.weight > 0
            ) ranked
            JOIN site_technologies st ON ranked.site_id = st.site_id
            ORDER BY ranked.cumulative_weight, ranked.row_num
            OFFSET :offset ROWS
            FETCH NEXT :batch_size ROWS ONLY;
        """, {"offset": offset, "batch_size": self.batch_size})

        rows = result.fetchall()

        products = [Product(row.product_id, row.url, row.technology) for row in rows]

        return products