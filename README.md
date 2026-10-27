# Auth
Handles user authentication and session tokens. Does not handle user registration.

## Registration
Users register with user service. Frontend provides a username and password, user service checks to make sure user name isn't already taken and maybe password length verification. If the username and password are valid, user service adds them to the database. If not, user service returns a problem that frontend will handle. Auth is not involved.

## Logging in
Logging in means getting a session token. Session tokens are handled by this repo. When a user wants to login, frontend provides a username and password to auth. Auth checks with user service to make sure the creds are valid. If they are, user service returns the uuid of the user. If not, user service returns a problem of type UserNotFound or InvalidPassword depending on the issue which auth will forward to the frontend.

If the creds are valid, auth generates a long random string to be a session token. Auth places the session token and user uuid into redis, with the session token as the key and the user uuid as the value. A TTL is placed on the entry, which will force users to log in again for security. Auth then returns the session token to frontend. Frontend saves the session token, and includes the token in all requests inside the gRPC metadata.

## Request Authentication
Except for user registration and logging in, all requests will be checked against auth. Session tokens will be passed to backend inside gRPC metadata. Details about this are TBD. Auth will read the session token, and check if it exists in redis. If it does, the uuid is added to the request inside 

## TODO: finish
