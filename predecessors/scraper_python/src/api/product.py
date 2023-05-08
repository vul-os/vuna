from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from uuid import UUID

from db.product import Product
from db.base import SessionLocal, engine

Base.metadata.create_all(bind=engine)
app = FastAPI()


class ProductCreate(BaseModel):
    url: str
    site_id: UUID


class ProductUpdate(BaseModel):
    url: str = None
    site_id: UUID = None


class ProductRead(BaseModel):
    id: UUID
    url: str
    site_id: UUID
    date_added: str
    date_updated: str


class ProductList(BaseModel):
    items: list[ProductRead]
    total: int


def get_db():
    try:
        db = SessionLocal()
        yield db
    finally:
        db.close()


@app.post("/products", response_model=ProductRead)
def create_product(product_create: ProductCreate, db=Depends(get_db)):
    product = Product.create(db=db, **product_create.dict())
    return product


@app.get("/products/{product_id}", response_model=ProductRead)
def get_product(product_id: UUID, db=Depends(get_db)):
    product = Product.get_by_id(db=db, product_id=product_id)
    if not product:
        raise HTTPException(status_code=404, detail="Product not found")
    return product


@app.put("/products/{product_id}", response_model=ProductRead)
def update_product(product_id: UUID, product_update: ProductUpdate, db=Depends(get_db)):
    product = Product.get_by_id(db=db, product_id=product_id)
    if not product:
        raise HTTPException(status_code=404, detail="Product not found")
    product.update(db=db, **product_update.dict(exclude_unset=True))
    return product


@app.delete("/products/{product_id}")
def delete_product(product_id: UUID, db=Depends(get_db)):
    product = Product.get_by_id(db=db, product_id=product_id)
    if not product:
        raise HTTPException(status_code=404, detail="Product not found")
    product.delete(db=db)
    return {"message": "Product deleted successfully"}


@app.get("/products", response_model=ProductList)
def get_products(db=Depends(get_db), skip: int = 0, limit: int = 100):
    products = Product.get_all(db=db, skip=skip, limit=limit)
    total = Product.get_count(db=db)
    return ProductList(items=products, total=total)