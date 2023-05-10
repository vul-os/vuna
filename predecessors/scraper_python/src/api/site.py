import uuid
from fastapi import APIRouter, HTTPException
from typing import List, Optional
from pydantic import BaseModel

from src.db.site import Site

router = APIRouter()


class SiteCreate(BaseModel):
    url: str
    name: str
    technology: str
    scraperfile: Optional[str]


class SiteUpdate(BaseModel):
    url: Optional[str] = None
    name: Optional[str] = None
    technology: Optional[str] = None
    scraperfile: Optional[str] = None


class SiteResponse(BaseModel):
    id: str
    url: str
    name: str
    technology: str
    date_added: str
    date_updated: str
    scraperfile: Optional[str] = None


@router.post("/", response_model=SiteResponse)
async def create_site(site: SiteCreate):
    new_site = Site.create(**site.dict())
    return SiteResponse(
        id=str(new_site.id),
        url=new_site.url,
        name=new_site.name,
        technology=new_site.technology,
        date_added=str(new_site.date_added),
        date_updated=str(new_site.date_updated),
        scraperfile=new_site.scraperfile,
    )


@router.get("/{site_id}", response_model=SiteResponse)
async def read_site(site_id: uuid.UUID):
    site = Site.get(site_id)
    if site is None:
        raise HTTPException(status_code=404, detail="Site not found")
    return SiteResponse(
        id=str(site.id),
        url=site.url,
        name=site.name,
        technology=site.technology,
        date_added=str(site.date_added),
        date_updated=str(site.date_updated),
        scraperfile=site.scraperfile,
    )


@router.get("/", response_model=List[SiteResponse])
async def read_sites():
    sites = Site.get_all()
    return [
        SiteResponse(
            id=str(site.id),
            url=site.url,
            name=site.name,
            technology=site.technology,
            date_added=str(site.date_added),
            date_updated=str(site.date_updated),
            scraperfile=site.scraperfile,
        )
        for site in sites
    ]


@router.put("/{site_id}", response_model=SiteResponse)
async def update_site(site_id: uuid.UUID, site: SiteUpdate):
    existing_site = Site.get(site_id)
    if existing_site is None:
        raise HTTPException(status_code=404, detail="Site not found")
    existing_site.update(**site.dict(exclude_unset=True))
    return SiteResponse(
        id=str(existing_site.id),
        url=existing_site.url,
        name=existing_site.name,
        technology=existing_site.technology,
        date_added=str(existing_site.date_added),
        date_updated=str(existing_site.date_updated),
        scraperfile=existing_site.scraperfile,
    )


@router.delete("/{site_id}")
async def delete_site(site_id: uuid.UUID):
    site = Site.get(site_id)
    if site is None:
        raise HTTPException(status_code=404, detail="Site not found")
    site.delete()
    return {"message": "Site deleted"}
