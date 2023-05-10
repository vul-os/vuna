import uuid
from typing import Optional
from datetime import datetime

from sqlalchemy import Column, Integer, Float, String, DateTime, text, ForeignKey
from sqlalchemy.orm import relationship

from src.db.base import Base, SessionLocal

class Product(Base):
    __tablename__ = "products"

    id: str = Column(String(36), primary_key=True, default=str(uuid.uuid4()), unique=True, index=True)
    url: str = Column(String(1000), nullable=False)
    site_id: uuid.UUID = Column(String(36), ForeignKey("sites.id"), nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("CURRENT_TIMESTAMP"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))

    site = relationship("Site", back_populates="products")
    variations = relationship("Variation", back_populates="product")

    def __repr__(self):
        return f"<Product(id={self.id}, url={self.url})>"

    @classmethod
    def get(cls, product_id: str):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == product_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()
    
    @classmethod
    def merge(cls, url: str, site_id: str):
        with SessionLocal() as session:
            product = session.query(cls).filter(cls.url == url).one_or_none()
            if product is None:
                product = cls(url=url, site_id=site_id)
                session.add(product)
            else:
                product.url = url
                product.site_id = site_id
                product.date_updated = datetime.now()
            session.commit()
            return product

    @classmethod
    def create(cls, url: str, site_id: str):
        product = cls(url=url, site_id=site_id)
        with SessionLocal() as session:
            session.add(product)
            session.commit()
            session.refresh(product)
        return product
