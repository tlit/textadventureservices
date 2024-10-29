"""Natural language input processing using Ollama."""
from ..interfaces import InputProcessor
from ..models.base import GameState,Scene,Object
from typing import List,Dict,Any,Optional
import aiohttp,json,re
from loguru import logger

class OllamaInputProcessor(InputProcessor):
    """Process game commands using Ollama."""
    def __init__(self,base_url:str="http://localhost:11434",
                model:str="mistral"):
        self.base_url=base_url
        self.model=model
        self._cmd_prompts={
            "process":"""Process text adventure command: {cmd}
                    Current game state: 
                    - Scene: {scene}
                    - Inventory: {inv}
                    Return JSON with format:
                    {{"action":str,"targets":list,"valid":bool,"error":str?}}""",
            "suggest":"""Suggest completions for partial command: {cmd}
                    Current scene: {scene}
                    Return JSON list of possible completions.""",
            "validate":"""Validate if command is possible: {cmd}
                    Current scene: {scene}
                    Return JSON: {{"valid":bool,"reason":str}}"""
        }
        # Common command patterns
        self._cmd_patterns={
            r"^(go|move|walk)\s+(north|south|east|west|up|down|n|s|e|w)$":
                lambda m:{"action":"move","direction":m.group(2)},
            r"^(take|get|grab|pick\s+up)\s+(.+)$":
                lambda m:{"action":"take","target":m.group(2)},
            r"^(drop|put\s+down|discard)\s+(.+)$":
                lambda m:{"action":"drop","target":m.group(2)},
            r"^(look|l)(\s+at\s+(.+))?$":
                lambda m:{"action":"look","target":m.group(3)},
            r"^(inventory|inv|i)$":
                lambda _:{"action":"inventory"},
            r"^(use|activate)\s+(.+?)(\s+on\s+(.+))?$":
                lambda m:{"action":"use","item":m.group(2),"target":m.group(4)}
        }

    async def _query_ollama(self,prompt:str)->Dict:
        """Query Ollama with retry."""
        async with aiohttp.ClientSession() as ses:
            for _ in range(3):
                try:
                    async with ses.post(
                        f"{self.base_url}/api/generate",
                        json={"model":self.model,"prompt":prompt}
                    ) as r:
                        if r.status!=200:
                            continue
                        resp=await r.text()
                        # Parse last JSON block
                        json_start=resp.rindex("{")
                        return json.loads(resp[json_start:])
                except Exception as e:
                    logger.error(f"Ollama query failed: {e}")
                    continue
            raise RuntimeError("Failed to query Ollama")

    def _get_scene_desc(self,scene:Scene)->str:
        """Get compact scene description."""
        return f"{scene.description} Objects: {[o.name for o in scene.objects]}"

    def _get_inv_desc(self,state:GameState)->str:
        """Get inventory description."""
        return f"Items: {[o.name for o in state.inventory]}"

    def _pattern_match(self,cmd:str)->Optional[Dict[str,Any]]:
        """Try matching command against patterns."""
        for pat,handler in self._cmd_patterns.items():
            if m:=re.match(pat,cmd.lower()):
                return handler(m)
        return None

    async def process_command(self,cmd:str,
                          context:GameState)->Dict[str,Any]:
        """Process user command."""
        # Try pattern matching first
        if action:=self._pattern_match(cmd):
            return {**action,"valid":True}

        # Fall back to Ollama
        scene=context.scenes[context.current_scene]
        prompt=self._cmd_prompts["process"].format(
            cmd=cmd,
            scene=self._get_scene_desc(scene),
            inv=self._get_inv_desc(context)
        )
        resp=await self._query_ollama(prompt)
        return json.loads(resp["response"])

    async def get_suggestions(self,partial_cmd:str,
                          context:GameState)->List[str]:
        """Get command suggestions."""
        scene=context.scenes[context.current_scene]
        prompt=self._cmd_prompts["suggest"].format(
            cmd=partial_cmd,
            scene=self._get_scene_desc(scene)
        )
        resp=await self._query_ollama(prompt)
        return json.loads(resp["response"])

    async def validate_command(self,cmd:str,
                           context:GameState)->bool:
        """Validate command possibility."""
        # Pattern match is always valid
        if self._pattern_match(cmd):
            return True

        scene=context.scenes[context.current_scene]
        prompt=self._cmd_prompts["validate"].format(
            cmd=cmd,
            scene=self._get_scene_desc(scene)
        )
        resp=await self._query_ollama(prompt)
        result=json.loads(resp["response"])
        return result["valid"]