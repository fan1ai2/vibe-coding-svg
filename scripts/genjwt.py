#!/usr/bin/env python3
"""Generate a dev JWT token for API testing (bypasses OAuth flow)."""
import jwt
import time
import sys
import uuid

SECRET = "dev-secret-change-in-prod"
USER_ID = sys.argv[1] if len(sys.argv) > 1 else str(uuid.uuid4())

payload = {
    "sub": USER_ID,
    "iat": int(time.time()),
    "exp": int(time.time() + 7 * 24 * 3600),
}

token = jwt.encode(payload, SECRET, algorithm="HS256")
print(token)
