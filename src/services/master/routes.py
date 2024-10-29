"""Master service API routes."""
from fastapi import APIRouter, HTTPException, Depends
from typing import List, Dict, Any
from src.models.base import ServiceBase, GameState
from .service import MasterService

router = APIRouter(prefix="/api/v1")
svc = MasterService()

@router.post("/services/register")
async def register_service(service: ServiceBase):
    return await svc.register_service(service)

@router.post("/services/deregister")
async def deregister_service(service_name: str):
    return await svc.deregister_service(service_name)

@router.get("/services/status")
async def get_service_status():
    return await svc.get_service_status()

@router.get("/game-state")
async def get_game_state():
    return await svc.get_game_state()

@router.post("/ui/command")
async def process_command(command: str):
    return await svc.process_command(command)

@router.post("/process-input")
async def process_input(user_input: str, current_state: Dict[str, Any]):
    return await svc.process_input(user_input, current_state)
