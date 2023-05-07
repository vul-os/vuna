from datetime import datetime
import uuid
from typing import Optional
from sqlalchemy import Column, Integer, Float, String, DateTime, text, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
from db.base import Base, SessionLocal

class DataPoint(Base):
    __tablename__ = "datapoints"

    id: str = Column(String, primary_key=True)
    var_id: uuid.UUID = Column(UUID(as_uuid=True), ForeignKey("variations.id"), nullable=False)

    max_qty: int = Column(Integer, nullable=False)
    price: float = Column(Float, nullable=False)

    date_added: datetime = Column(DateTime, nullable=False, server_default=text("now()"))
    date_updated: datetime = Column(DateTime, nullable=False, server_default=text("now()"))

    variation = relationship("Variation", back_populates="datapoints")

    def __repr__(self):
        return f"<DataPoint(id={self.id}, max_qty={self.max_qty}, price={self.price})>"

    @classmethod
    def create(cls, var_id: uuid.UUID, max_qty: int, price: float):
        datapoint = cls(id=str(uuid.uuid4()), var_id=var_id, max_qty=max_qty, price=price)
        with SessionLocal() as session:
            session.add(datapoint)
            session.commit()
            session.refresh(datapoint)
        return datapoint

    @classmethod
    def get(cls, datapoint_id: str):
        with SessionLocal() as session:
            return session.query(cls).filter(cls.id == datapoint_id).one_or_none()

    @classmethod
    def get_all(cls):
        with SessionLocal() as session:
            return session.query(cls).all()

    def delete(self):
        with SessionLocal() as session:
            session.delete(self)
            session.commit()

    @classmethod
    def merge(cls, var_id: uuid.UUID, max_qty: int, price: float):
        with SessionLocal() as session:
            datapoint = session.query(cls).filter(cls.var_id == var_id).one_or_none()
            if datapoint is None:
                datapoint = cls(id=str(uuid.uuid4()), var_id=var_id, max_qty=max_qty, price=price)
                session.add(datapoint)
            else:
                datapoint.max_qty = max_qty
                datapoint.price = price
                datapoint.date_updated = datetime.now()
            session.commit()
            return datapoint