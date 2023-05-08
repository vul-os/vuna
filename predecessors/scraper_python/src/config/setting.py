from pydantic import BaseSettings, PostgresDsn


class Settings(BaseSettings):
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str
    db_password: str
    db_name: str
    gcs_bucket_name: str

    @property
    def postgre_dsn(self) -> PostgresDsn:
        return PostgresDsn.build(
            scheme="postgresql",
            user=self.db_user,
            password=self.db_password,
            host=self.db_host,
            port=str(self.db_port),
            path=f"/{self.db_name}",
        )

    class Config:
        env_prefix = ""
        case_sensitive = False

settings = Settings()