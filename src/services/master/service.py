"""Master service implementation."""
from typing import Dict, List, Any, Optional
from src.models.base import ServiceBase, GameState, Status
from src.utils.logging import logger

class MasterService:
    def __init__(self):
        self._services: Dict[str, ServiceBase] = {}
        self._game_state: Optional[GameState] = None
        
    async def register_service(self, service: ServiceBase) -> Dict[str, str]:
        """Register a new service."""
        if service.name in self._services:
            raise HTTPException(status_code=400, detail="Service already registered")
        self._services[service.name] = service
        logger.info(f"Service {service.name} registered at {service.url}")
        return {"message": "Service registered successfully"}

    async def deregister_service(self, service_name: str) -> Dict[str, str]:
        """Deregister an existing service."""
        if service_name not in self._services:
            raise HTTPException(status_code=404, detail="Service not found")
        del self._services[service_name]
        logger.info(f"Service {service_name} deregistered")
        return {"message": "Service deregistered successfully"}

    async def get_service_status(self) -> List[Dict[str, str]]:
        """Get status of all registered services."""
        return [{"serviceName": name, "status": svc.status} 
                for name, svc in self._services.items()]

    async def get_game_state(self) -> Dict[str, Any]:
        """Get current game state."""
        if not self._game_state:
            raise HTTPException(status_code=404, detail="No active game")
        return {"gameState": self._game_state.dict()}

    async def process_command(self, command: str) -> Dict[str, str]:
        """Process UI command."""
        # TODO: Implement command routing logic
        return {"message": "Command processed"}

    async def process_input(self, user_input: str, current_state: Dict[str, Any]) -> Dict[str, Any]:
        """Process user input via Claude."""
        # TODO: Implement Claude interaction
        return {"message": "Input processed", "state": current_state}
