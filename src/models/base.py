"""Base models for game engine.
Defines data structures for services, game state, and metrics."""
from pydantic import BaseModel,Field
from typing import List,Dict,Any,Optional
from enum import Enum

class Status(str,Enum):
    """Service status enum"""
    ACTIVE="active"
    INACTIVE="inactive"
    ERROR="error"

class Service(BaseModel):
    """Service registration model"""
    name:str 
    url:str
    status:Status=Status.INACTIVE

class Direction(str,Enum):
    """Movement directions"""
    N="N";S="S";E="E";W="W"
    UP="up";DOWN="down"

class Object(BaseModel):
    """Game object model"""
    name:str
    tags:List[str]=Field(default_factory=list)
    material:Optional[str]=None
    properties:Dict[str,Any]=Field(default_factory=dict)

class Exit(BaseModel):
    """Scene exit model"""
    direction:Direction
    description:str
    target_scene:str

class Connector(BaseModel):
    """Scene connection model"""
    id:str
    exits:List[Exit]
    locked:bool=False
    key_required:Optional[str]=None

class Scene(BaseModel):
    """Game scene model"""
    id:str
    description:str
    objects:List[Object]=Field(default_factory=list)  
    connectors:List[Connector]=Field(default_factory=list)
    tags:List[str]=Field(default_factory=list)
    properties:Dict[str,Any]=Field(default_factory=dict)

class GameState(BaseModel):
    """Game state model"""
    scenes:List[Scene]=Field(default_factory=list)
    current_scene:str
    inventory:List[Object]=Field(default_factory=list)
    game_flags:Dict[str,Any]=Field(default_factory=dict)

class UiState(BaseModel):
    """UI state model"""
    current_view:str
    messages:List[str]=Field(default_factory=list) 
    input_history:List[str]=Field(default_factory=list)
    render_flags:Dict[str,bool]=Field(default_factory=dict)

class LogEntry(BaseModel):
    """Log entry model"""
    timestamp:str
    message:str
    level:str="INFO"
    service:str
    trace_id:Optional[str]=None

class Metrics(BaseModel):
    """Service metrics model"""
    latency:Dict[str,float]=Field(default_factory=dict)
    errors:Dict[str,int]=Field(default_factory=dict) 
    requests:Dict[str,int]=Field(default_factory=dict)
    memory:Dict[str,float]=Field(default_factory=dict)
