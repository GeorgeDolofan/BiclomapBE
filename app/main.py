from fastapi import FastAPI
from importlib import import_module

from . import config
settings = config.get_settings()

# here use a computed path to take into account API version
api = import_module(settings.get_import_version_path())
from mangum import Mangum

app = FastAPI(
        title=settings.app_name,
        root_path='/' + settings.runtime_environment
        )

@app.get("/")
async def root():
    return {"message": "Biclomap Rules!"}

app.include_router(api.router, prefix="/api/v1")

handler = Mangum(app)

