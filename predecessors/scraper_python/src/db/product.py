from datetime import datetime
import uuid

from sqlalchemy import Column, String, DateTime, text, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from db.base import Base, SessionLocal

class Product(Base):
    __tablename__ = "products"

    id: uuid.UUID = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    url: str = Column(String, nullable=False)
    site_id: uuid.UUID = Column(UUID(as_uuid=True), ForeignKey("sites.id"), nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("now()"))

    site = relationship("Site", back_populates="products")
    variations = relationship("Variation", back_populates="product")

    def __repr__(self):
        return f"<Product(id={self.id}, url={self.url})>"

    @classmethod
    def get(cls, product_id: uuid.UUID):
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
    def merge(cls, url: str, site_id: uuid.UUID):
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
    def create(cls, url: str, site_id: uuid.UUID):
        product = cls(url=url, site_id=site_id)
        with SessionLocal() as session:
            session.add(product)
            session.commit()
            session.refresh(product)
        return product

