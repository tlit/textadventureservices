"""World generation service using Ollama."""
from ..interfaces import WorldGenerator
from ..models.base import Scene,Object,Exit,Connector,Direction
from typing import List,Dict,Any,Optional
import aiohttp,json,uuid
from loguru import logger

class OllamaWorldGenerator(WorldGenerator):
    """World generator using Ollama for scene generation."""
    def __init__(self,base_url:str="http://localhost:11434",
                model:str="mistral"):
        self.base_url=base_url
        self.model=model
        self._gen_prompts={
            "world":"""Generate a text adventure game world based on: {prompt}
                   Return a JSON list of connected scenes with format:
                   [{{"id":str,"description":str,"tags":list,"objects":[{{"name":str,"tags":list,"material":str}}],"exits":[{{"direction":str,"description":str,"target":str}}]}}]
                   Keep descriptions vivid but concise. Use valid JSON.""",
            "extend":"""Extend existing game world with new scenes.
                    Current scenes: {scenes}
                    Extension prompt: {prompt}
                    Return only new scenes in same JSON format.""",
            "modify":"""Modify scene based on changes: {changes}
                    Current scene: {scene}
                    Return modified scene in same JSON format."""
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
                        # Parse last JSON block from response
                        json_start=resp.rindex("{")
                        json_str=resp[json_start:]
                        return json.loads(json_str)
                except Exception as e:
                    logger.error(f"Ollama query failed: {e}")
                    continue
            raise RuntimeError("Failed to query Ollama")

    def _parse_scene(self,data:Dict)->Scene:
        """Parse scene data into Scene model."""
        return Scene(
            id=data.get("id",str(uuid.uuid4())),
            description=data["description"],
            tags=data.get("tags",[]),
            objects=[
                Object(
                    name=o["name"],
                    tags=o.get("tags",[]),
                    material=o.get("material")
                ) for o in data.get("objects",[])
            ],
            connectors=[
                Connector(
                    id=str(uuid.uuid4()),
                    exits=[
                        Exit(
                            direction=Direction(e["direction"]),
                            description=e["description"],
                            target_scene=e["target"]
                        ) for e in data.get("exits",[])
                    ]
                )
            ]
        )

    async def generate_world(self,prompt:str,
                         constraints:Dict[str,Any])->List[Scene]:
        """Generate initial world scenes."""
        prompt=self._gen_prompts["world"].format(
            prompt=f"{prompt}. Constraints: {constraints}")
        resp=await self._query_ollama(prompt)
        scenes_data=json.loads(resp["response"])
        return [self._parse_scene(s) for s in scenes_data]

    async def extend_world(self,scenes:List[Scene],
                        prompt:str)->List[Scene]:
        """Generate additional scenes."""
        scenes_json=json.dumps([{
            "id":s.id,
            "description":s.description,
            "exits":[{
                "direction":e.direction,
                "target":e.target_scene
            } for c in s.connectors for e in c.exits]
        } for s in scenes])
        prompt=self._gen_prompts["extend"].format(
            scenes=scenes_json,prompt=prompt)
        resp=await self._query_ollama(prompt)
        new_scenes=json.loads(resp["response"])
        return [self._parse_scene(s) for s in new_scenes]

    async def modify_scene(self,scene:Scene,
                        changes:Dict[str,Any])->Scene:
        """Modify existing scene."""
        prompt=self._gen_prompts["modify"].format(
            scene=scene.json(),changes=changes)
        resp=await self._query_ollama(prompt)
        modified=json.loads(resp["response"])
        return self._parse_scene(modified)