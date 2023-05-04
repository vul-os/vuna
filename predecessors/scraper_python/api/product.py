from fastapi import APIRouter, Depends, HTTPException, status
from typing import List
import uuid

from db.site import Site
from db.product import Product
from db.base import SessionLocal

router = APIRouter()

class ProductCreate(BaseModel):
    url: str
    site_id: uuid.UUID

class ProductUpdate(BaseModel):
    url: str = None
    site_id: uuid.UUID = None

@router.post("/", response_model=Product)
def create_product(product_create: ProductCreate):
    site = Site.get(product_create.site_id)
    if site is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Site not found")
    product = Product.create(url=product_create.url, site_id=product_create.site_id)
    return product

@router.get("/{product_id}", response_model=Product)
def get_product(product_id: uuid.UUID):
    product = Product.get(product_id)
    if product is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Product not found")
    return product

@router.get("/", response_model=List[Product])
def get_all_products():
    return Product.get_all()

@router.put("/{product_id}", response_model=Product)
def update_product(product_id: uuid.UUID, product_update: ProductUpdate):
    product = Product.get(product_id)
    if product is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Product not found")
    product.update(url=product_update.url, site_id=product_update.site_id)
    return product

@router.delete("/{product_id}")
def delete_product(product_id: uuid.UUID):
    product = Product.get(product_id)
    if product is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Product not found")
    product.delete()
    return {"detail": "Product deleted"}
