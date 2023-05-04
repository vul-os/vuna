from datetime import datetime
import uuid

from sqlalchemy import Column, String, DateTime, text, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from db.base import Base, SessionLocal

class Product(Base):
    __tablename__ = "products"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    url = Column(String, nullable=False)
    site_id = Column(UUID(as_uuid=True), ForeignKey("sites.id"), nullable=False)

    date_added = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated = Column(DateTime, nullable=False, server_default=text("now()"))

    site = relationship("Site", back_populates="products")
    variations = relationship("Variation", back_populates="product")

    def __repr__(self):
        return f"<Product(id={self.id}, url={self.url})>"

    @classmethod
    def create(cls, url, site_id):
        product = cls(url=url, site_id=site_id)
        with SessionLocal() as session:
            session.add(product)
            session.commit()
            session.refresh(product)
        return product

    @classmethod
    def get(cls, product_id):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == product_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def update(self, url=None, site_id=None):
        with SessionLocal() as session:
            if url is not None:
                self.url = url
            if site_id is not None:
                self.site_id = site_id
            self.date_updated = datetime.now()
            session.commit()

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()