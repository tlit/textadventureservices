"""JWT-based authentication service."""
from ..interfaces import AuthManager
from typing import List,Dict,Optional
from datetime import datetime,timedelta
import jwt
import bcrypt
from loguru import logger

class JWTAuthManager(AuthManager):
    """JWT authentication manager."""
    def __init__(self,secret_key:str,
               token_expiry:int=60): # mins
        self.secret=secret_key
        self.expiry=token_expiry
        # Demo users - replace with DB
        self._users={
            "admin":{"pwd":self._hash_pwd("admin"),
                    "perms":["admin"]},
            "player":{"pwd":self._hash_pwd("player"),
                     "perms":["play"]}
        }
        self._blacklist=set()

    def _hash_pwd(self,pwd:str)->bytes:
        """Hash password with bcrypt."""
        return bcrypt.hashpw(
            pwd.encode(),bcrypt.gensalt())

    def _check_pwd(self,pwd:str,hashed:bytes)->bool:
        """Verify password against hash."""
        return bcrypt.checkpw(
            pwd.encode(),hashed)

    def _gen_token(self,username:str,
                 perms:List[str])->str:
        """Generate JWT token."""
        now=datetime.utcnow()
        payload={
            "sub":username,
            "perms":perms,
            "iat":now,
            "exp":now+timedelta(minutes=self.expiry)
        }
        return jwt.encode(payload,self.secret,
                         algorithm="HS256")

    def _decode_token(self,token:str)->Optional[Dict]:
        """Decode and validate JWT token."""
        try:
            return jwt.decode(token,self.secret,
                            algorithms=["HS256"])
        except jwt.InvalidTokenError as e:
            logger.error(f"Invalid token: {e}")
            return None

    async def authenticate(self,
                       creds:Dict[str,str])->str:
        """Authenticate user and return token."""
        user=creds.get("username")
        pwd=creds.get("password")
        if not (user and pwd):
            raise ValueError("Missing credentials")

        if not (udata:=self._users.get(user)):
            raise ValueError("User not found")

        if not self._check_pwd(pwd,udata["pwd"]):
            raise ValueError("Invalid password")

        return self._gen_token(user,udata["perms"])

    async def validate_token(self,token:str)->bool:
        """Validate token is valid and not revoked."""
        if token in self._blacklist:
            return False
        return bool(self._decode_token(token))

    async def get_permissions(self,
                          token:str)->List[str]:
        """Get permissions from token."""
        if not (payload:=self._decode_token(token)):
            return []
        return payload.get("perms",[])

    async def revoke_token(self,token:str)->None:
        """Revoke token by adding to blacklist."""
        self._blacklist.add(token)