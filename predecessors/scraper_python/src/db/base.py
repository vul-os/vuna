from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

from src.config.config import config  # import your Settings class

# format the connection string for PlanetScale
dsn = f"mysql://{config.db_user}:{config.db_password}@{config.db_host}/{config.db_name}"

# specify SSL parameters for encryption
ssl_args = {'ssl': {'ssl_mode': 'VERIFY_IDENTITY', 'ssl_ca': '/etc/ssl/certs/ca-certificates.crt'}}

# create the engine with the connection string and SSL parameters
engine = create_engine(dsn, connect_args=ssl_args, pool_pre_ping=True)

# create a session factory bound to the engine
SessionLocal = sessionmaker(bind=engine, autocommit=False, autoflush=False)
Base = declarative_base()
Base.metadata.drop_all(bind=engine)

from src.db.site import Site
from src.db.variation import Variation
from src.db.datapoint import DataPoint
from src.db.product import Product

with SessionLocal() as session:
    session.query(Product).delete()
    session.query(Site).delete()
    session.query(Variation).delete()
    session.query(DataPoint).delete()
    session.commit()
# Base.metadata.create_all(bind=engine)
# from src.db.variation import Variation

# with SessionLocal() as session:
#     session.query(Variation).delete()
#     session.commit()


# check if tables exist and create them if they don't
# inspector = inspect(Base)
# if not inspector.has_table(Site.__tablename__):
#     Site.__table__.create(bind=engine)
# if not inspector.has_table(Product.__tablename__):
#     Product.__table__.create(bind=engine)
# # if not inspector.has_table(Variation.__tablename__):
#     # Variation.__table__.create(bind=engine)
# if not inspector.has_table(DataPoint.__tablename__):
#     DataPoint.__table__.create(bind=engine)