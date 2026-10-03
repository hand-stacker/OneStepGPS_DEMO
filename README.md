## The purpose of this app is as a test project that uses Go as the backend and Vue.js as a frontend.
## The app will use the OneStepGPS API to collect sample gps tracker data and store user prefrences data in an sqlite database.

## The app is available at [ip_address] or you can deploy it locally through the instructions below

## TESTING IN A LOCAL ENV
to test backend only run this command in the backend folder
"""
    # load .env into this terminal session
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]*)=(.*)$') {
            Set-Item "env:$($matches[1].Trim())" $matches[2].Trim()
        }
    }
    go run .
"""
and the endpoints root is
"127.0.0.1:ENV_VAR_PORT/"

to test front end make sure the backend is running on the same device, and run this command in the frontend folder
"""
    npm run dev
"""
and the root url is 
"http://localhost:5173"

## BACKEND

## DATABASE MODELS

## FRONTEND
