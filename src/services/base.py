"""Abstract base service implementations."""
from typing import Dict,List,Any,Optional
from ..interfaces import *
from ..models.base import *
import aiohttp,asyncio
from datetime import datetime,timezone

class BaseServiceRegistry(ServiceRegistry):
    """Basic service registry impl."""
    def __init__(self):
        self._services:Dict[str,Service]={}
        self._health_checks:Dict[str,asyncio.Task]={}

    async def register(self,svc:Service)->bool:
        self._services[svc.name]=svc
        self._health_checks[svc.name]=asyncio.create_task(
            self._check_health(svc))
        return True

    async def deregister(self,name:str)->bool:
        if task:=self._health_checks.get(name):
            task.cancel()
        self._services.pop(name,None)
        return True

    async def get_service(self,name:str)->Optional[Service]:
        return self._services.get(name)

    async def list_services(self)->List[Service]:
        return list(self._services.values())

    async def health_check(self)->Dict[str,bool]:
        return {n:s.status==Status.ACTIVE 
               for n,s in self._services.items()}

    async def _check_health(self,svc:Service)->None:
        while True:
            try:
                async with aiohttp.ClientSession() as ses:
                    async with ses.get(f"{svc.url}/health") as r:
                        svc.status=(Status.ACTIVE if r.status==200 
                                  else Status.ERROR)
            except: svc.status=Status.ERROR
            await asyncio.sleep(30)

class BaseGameStateManager(GameStateManager):
    """Basic game state management."""
    def __init__(self):
        self._states:Dict[str,GameState]={}

    async def load_state(self,game_id:str)->GameState:
        if not (state:=self._states.get(game_id)):
            raise KeyError(f"No state for game {game_id}")
        return state

    async def save_state(self,game_id:str,state:GameState)->bool:
        self._states[game_id]=state
        return True

    async def update_state(self,game_id:str,
                        updates:Dict[str,Any])->GameState:
        state=await self.load_state(game_id)
        for k,v in updates.items():
            setattr(state,k,v)
        return state

    async def delete_state(self,game_id:str)->bool:
        self._states.pop(game_id,None)
        return True

class BaseUIManager(UIManager):
    """Basic UI management."""
    def __init__(self):
        self._state=UiState(current_view="main")

    async def get_state(self)->UiState:
        return self._state

    async def update_state(self,updates:Dict[str,Any])->UiState:
        for k,v in updates.items():
            setattr(self._state,k,v)
        return self._state

    async def add_message(self,msg:str)->None:
        self._state.messages.append(msg)

    async def clear_messages(self)->None:
        self._state.messages.clear()

    async def render_scene(self,scene:Scene)->str:
        return (f"{scene.description}\n\n"
                f"Objects: {', '.join(o.name for o in scene.objects)}\n"
                f"Exits: {', '.join(f'{e.direction}:{e.description}' for c in scene.connectors for e in c.exits)}")

class BaseEventLogger(EventLogger):
    """Basic event logging."""
    def __init__(self):
        self._logs:List[LogEntry]=[]

    async def log(self,entry:LogEntry)->None:
        entry.timestamp=datetime.now(timezone.utc).isoformat()
        self._logs.append(entry)

    async def get_logs(self,filters:Dict[str,Any])->List[LogEntry]:
        logs=self._logs
        for k,v in filters.items():
            logs=[l for l in logs if getattr(l,k)==v]
        return logs

    async def clear_logs(self,older_than:str)->None:
        cutoff=datetime.fromisoformat(older_than)
        self._logs=[l for l in self._logs 
                   if datetime.fromisoformat(l.timestamp)>=cutoff]

class BaseMetricsCollector(MetricsCollector):
    """Basic metrics collection."""
    def __init__(self):
        self._metrics=Metrics()

    async def record_metric(self,name:str,value:float)->None:
        cat,metric=name.split(".",1)
        getattr(self._metrics,cat)[metric]=value

    async def get_metrics(self)->Metrics:
        return self._metrics

    async def clear_metrics(self)->None:
        self._metrics=Metrics()