from create_task import create_task_for_product
from get_products import get_products

def orchestrate():
    # this is more to do with limitations of python cloud run instance, 
    # can send all requests at once if need be
    batch_size = 100
    offset = 0

    while True:
        products = get_products(db_url, batch_size, offset)

        if not products:
            break

        create_task_for_product(products)

        offset += batch_size

if __name__ == "__main__":
    orchestrate()