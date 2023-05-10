import os
from pydantic import BaseSettings


class Config(BaseSettings):
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str
    db_password: str
    db_name: str
    db_organization: str
    
    gcs_bucket_name: str = None

config_path = os.environ.get("CONFIG_PATH", ".env")
config = Config(_env_file=config_path, _env_file_encoding="utf-8")