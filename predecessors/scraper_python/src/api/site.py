from fastapi import FastAPI, HTTPException
from typing import List

from pydantic import UUID4, BaseModel
from uuid import UUID
from db.site import Site, create, get, get_all, update_site, delete_site

app = FastAPI()

class SiteCreate(BaseModel):
    url: str
    name: str
    technology: str
    scraperfile: str = None

class SiteUpdate(BaseModel):
    url: str = None
    name: str = None
    technology: str = None
    scraperfile: str = None

class SiteOut(BaseModel):
    id: UUID4
    url: str
    name: str
    technology: str
    scraperfile: str = None

class SiteListOut(BaseModel):
    sites: List[SiteOut]

@app.post('/sites', response_model=SiteOut)
def create_new_site(site_create: SiteCreate):
    site = create_site(**site_create.dict())
    return site

@app.get('/sites/{site_id}', response_model=SiteOut)
def get_site_by_id(site_id: UUID4):
    site = get_site(site_id)
    if not site:
        raise HTTPException(status_code=404, detail='Site not found')
    return site

@app.get('/sites', response_model=SiteListOut)
def get_all_sites():
    sites = get_all_sites()
    return SiteListOut(sites=[SiteOut.from_orm(site) for site in sites])

@app.put('/sites/{site_id}', response_model=SiteOut)
def update_site_by_id(site_id: UUID4, site_update: SiteUpdate):
    site = get_site(site_id)
    if not site:
        raise HTTPException(status_code=404, detail='Site not found')
    updated_site = update_site(site, **site_update.dict(exclude_unset=True))
    return updated_site

@app.delete('/sites/{site_id}')
def delete_site_by_id(site_id: UUID4):
    site = get_site(site_id)
    if not site:
        raise HTTPException(status_code=404, detail='Site not found')
    delete_site(site)
    return {'message': 'Site deleted successfully'}