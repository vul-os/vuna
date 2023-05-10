import os

os.environ['CONFIG_PATH'] = str("/workspace/scraper_python/config.env")

from fastapi import FastAPI

from src.db.base import Base, engine
# from src.api.scraper import router as scraper_router
# from src.api.site import router as site_router

import uvicorn

app = FastAPI()

# app.include_router(scraper_router, prefix="/scraper")
# app.include_router(site_router, prefix="/site")

@app.on_event("startup")
def startup():
    pass
    # Base.metadata.drop_all(engine)
    
@app.on_event("shutdown")
def shutdown():
    pass
    # engine.dispose()

@app.get("/")
async def root():
    return {"message": "Hello World"}


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)