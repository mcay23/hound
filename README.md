# Hound
 Track TV Shows, Movies, etc.

# API Keys
You need a [TMDB API key](https://developers.themoviedb.org/3/getting-started/introduction) and [IGDB Client ID and secret](https://api-docs.igdb.com/) to run Hound.


# Docker Compose
Set your API keys in the `compose.env` file. If you change the database username/password, make sure you change the DB connection string as well.
```
docker compose up
```

# Build from Source
- Frontend
```bash
cd web

npm install

npm run start
```
- Backend  
Set your API keys in the `server/.env` file. If you change the database username/password, make sure you change the DB connection string as well.
```
go build
```
