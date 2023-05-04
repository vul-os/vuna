from datetime import datetime
import uuid

from sqlalchemy import Column, String, DateTime, text
from sqlalchemy.dialects.postgresql import UUID

from db.base import Base, SessionLocal

class Site(Base):
    __tablename__ = "sites"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    url = Column(String, nullable=False)
    name = Column(String, nullable=False)
    technology = Column(String, nullable=False)

    date_added = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated = Column(DateTime, nullable=False, server_default=text("now()"))

    def __repr__(self):
        return f"<Site(id={self.id}, name={self.name})>"

    @classmethod
    def create(cls, url, name, technology):
        site = cls(url=url, name=name, technology=technology)
        with SessionLocal() as session:
            session.add(site)
            session.commit()
            session.refresh(site)
        return site

    @classmethod
    def get(cls, site_id):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == site_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def update(self, url=None, name=None, technology=None):
        with SessionLocal() as session:
            if url is not None:
                self.url = url
            if name is not None:
                self.name = name
            if technology is not None:
                self.technology = technology
            self.date_updated = datetime.now()
            session.commit()

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()