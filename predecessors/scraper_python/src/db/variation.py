import uuid
from typing import Optional
from datetime import datetime

from sqlalchemy import Column, String, DateTime, text, ForeignKey
from sqlalchemy.orm import relationship

from src.db.base import Base, SessionLocal


class Variation(Base):
    __tablename__ = "variations"
    __table_args__ = {'extend_existing': True}

    id: str = Column(String(32), primary_key=True, default=lambda: str(uuid.uuid4().hex))
    product_id: str = Column(String(32), ForeignKey("products.id"), nullable=False)
    identifier: str = Column(String(255), nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("now()"))

    product = relationship("Product", back_populates="variations")
    datapoints = relationship("DataPoint", back_populates="variation")
    
    def __repr__(self):
        return f"<Variation(id={self.id}, identifier={self.identifier})>"

    @classmethod
    def create(cls, product_id: uuid.UUID, identifier: str):
        variation = cls(product_id=product_id, identifier=identifier)
        with SessionLocal() as session:
            session.add(variation)
            session.commit()
            session.refresh(variation)
        return variation

    @classmethod
    def get(cls, variation_id: uuid.UUID):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == variation_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    @classmethod
    def merge(cls, product_id: uuid.UUID, identifier: str):
        with SessionLocal() as session:
            variation = session.query(cls).filter(cls.product_id == product_id, cls.identifier == identifier).one_or_none()
            if variation is None:
                variation = cls(product_id=product_id, identifier=identifier)
                session.add(variation)
            session.commit()
            return variation

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()
