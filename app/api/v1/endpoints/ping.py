'''
API health check service
'''
from fastapi import APIRouter

router = APIRouter()

@router.get("/")
async def get_ping():
    return {"message": "OK"}

