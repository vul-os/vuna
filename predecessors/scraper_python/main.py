import os

os.environ['CONFIG_PATH'] = str("/workspace/scraper_python/config.env")

from fastapi import FastAPI

from src.db import base
from src.api.scraper import router as scraper_router
import uvicorn

app = FastAPI()

app.include_router(scraper_router, prefix="/scraper")


@app.get("/")
async def root():
    return {"message": "Hello World"}


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)