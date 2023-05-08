from fastapi import FastAPI
from db import base
from db.site import Site
from db.product import Product
from db.variation import Variation
from db.datapoint import DataPoint
from api.site import site_router
from api.product import product_router
from api.variation import variation_router
from api.datapoint import datapoint_router

app = FastAPI()

app.include_router(site_router)
app.include_router(product_router)
app.include_router(variation_router)
app.include_router(datapoint_router)


@app.on_event("startup")
async def startup():
    await base.Base().connect()
    base.Base().metadata.create_all(bind=base.engine)


@app.on_event("shutdown")
async def shutdown():
    await base.Base().disconnect()


@app.get("/")
async def root():
    return {"message": "Hello World"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)