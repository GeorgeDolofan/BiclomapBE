from functools import lru_cache
from pydantic import BaseSettings

class Settings(BaseSettings):
    app_name: str = "Biclomap API"
    runtime_environment: str = "xxx"
    version: str = "0"

    class Config:
        env_file = ".env"

    def get_import_version_path(self):
        return 'app.api.v' + self.version + '.api'


@lru_cache()
def get_settings():
    return Settings()

