import uuid
from typing import Optional
from datetime import datetime

from sqlalchemy import Column, String, DateTime, text
from sqlalchemy.orm import relationship

from src.db.base import Base, SessionLocal


class Site(Base):
    __tablename__ = "sites"

    id: str = Column(String(36), primary_key=True, default=str(uuid.uuid4()))
    url: str = Column(String(1000), nullable=False)
    name: str = Column(String(1000), nullable=False)
    technology: str = Column(String(100), nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("CURRENT_TIMESTAMP"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))

    scraperfile: Optional[str] = Column(String(100), nullable=True)

    products = relationship("Product", back_populates="site")

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
