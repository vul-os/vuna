from datetime import datetime
import uuid

from sqlalchemy import Column, String, DateTime, text, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from db.base import Base, SessionLocal

class Variation(Base):
    __tablename__ = "variations"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    product_id = Column(UUID(as_uuid=True), ForeignKey("products.id"), nullable=False)
    identifier = Column(String, nullable=False)

    date_added = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated = Column(DateTime, nullable=False, server_default=text("now()"))

    product = relationship("Product", back_populates="variations")
    datapoints = relationship("DataPoint", back_populates="variation")

    def __repr__(self):
        return f"<Variation(id={self.id}, identifier={self.identifier})>"

    @classmethod
    def create(cls, product_id, identifier):
        variation = cls(product_id=product_id, identifier=identifier)
        with SessionLocal() as session:
            session.add(variation)
            session.commit()
            session.refresh(variation)
        return variation

    @classmethod
    def get(cls, variation_id):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == variation_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def update(self, product_id=None, identifier=None):
        with Session
