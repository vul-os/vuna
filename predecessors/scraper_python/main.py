from fastapi import FastAPI

from src.db import base
import uvicorn

app = FastAPI()

app.include_router(site_router)


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
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)