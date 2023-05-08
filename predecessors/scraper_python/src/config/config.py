import os
from pydantic import BaseSettings, PostgresDsn


class Config(BaseSettings):
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str
    db_password: str
    db_name: str
    gcs_bucket_name: str = None

    @property
    def db_url(self) -> PostgresDsn:
        return PostgresDsn.build(
            scheme="postgresql",
            user=self.db_user,
            password=self.db_password,
            host=self.db_host,
            port=str(self.db_port),
            path=f"/{self.db_name}",
        )

config_path = os.environ.get("CONFIG_PATH", ".env")
config = Config(_env_file=config_path, _env_file_encoding="utf-8")