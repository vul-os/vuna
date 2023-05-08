from datetime import datetime
import uuid
from typing import Optional

from sqlalchemy import Column, String, DateTime, text
from sqlalchemy.dialects.postgresql import UUID

from src.db.base import Base, SessionLocal

class Site(Base):
    __tablename__ = "sites"

    id: uuid.UUID = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    url: str = Column(String, nullable=False)
    name: str = Column(String, nullable=False)
    technology: str = Column(String, nullable=False)

    technology: str = Column(String, nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("now()"))

    scraperfile: Optional[str] = Column(String, nullable=False)

    def __repr__(self):
        return f"<Site(id={self.id}, name={self.name}, url={self.url})>"

    @classmethod
    def create(cls, url: str, name: str, technology: str, scraperfile: Optional[str] = None):
        site = cls(url=url, name=name, technology=technology, scraperfile=scraperfile)
        with SessionLocal() as session:
            session.add(site)
            session.commit()
            session.refresh(site)
        return site

    @classmethod
    def get(cls, site_id: uuid.UUID):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == site_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def update(self, url: Optional[str] = None, name: Optional[str] = None, technology: Optional[str] = None, scraperfile: Optional[str] = None):
        with SessionLocal() as session:
            if url is not None:
                self.url = url
            if name is not None:
                self.name = name
            if technology is not None:
                self.technology = technology
            if scraperfile is not None:
                self.scraperfile = scraperfile
            self.date_updated = datetime.now()
            session.commit()
    
    @classmethod
    def merge(cls, url: str, name: str, technology: str, scraperfile: Optional[str] = None):
        with SessionLocal() as session:
            site = session.query(cls).filter(cls.url == url).one_or_none()
            if site is None:
                site = cls(url=url, name=name, technology=technology, scraperfile=scraperfile)
                session.add(site)
            else:
                site.name = name
                site.technology = technology
                if scraperfile is not None:
                    site.scraperfile = scraperfile
            session.commit()
            return site

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()

