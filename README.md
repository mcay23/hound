# Hound Media Server
 Watch and Track Movies and TV Shows. Self-hosted version of Plex/Stremio + Trakt, Simkl, etc. Hound aims to be a complete ecosystem of watching, tracking, downloading, and archiving media.
 **This project is under alpha development and not ready to self-host yet.**
 
# Features
- Current
  - Search Movies and TV Shows
  - Add watch data
  - Rewatch Shows from the beginning
  - View your watch history
  - Multi-user
  - Create collections (playlists)
  - Write reviews
- WIP
  - Stream and download media directly from p2p, http, or serve files directly from Hound
  - Manages and renames your downloads automatically
  - Download streams to device or server
  - Android Mobile and TV apps
- Future
  - Detailed watch statistics
  - Recommendations
  - Transcoding
  - Manually create your own movies/shows
  - Manually add your own media files
  - Data export
  - Third-party review score integration (eg. IMDB, Metacritic, RT)
  - View actor information (eg. movies they've played)
  - View public collections / other user's collections
  - Review individual seasons, episodes (TV Shows)
  - Add private notes for your media
 
# Screenshots
![home page](https://github.com/mcay23/hound/blob/main/screenshots/home.png)
![tv page](https://github.com/mcay23/hound/blob/main/screenshots/tvpage.png)
![tv page 2](https://github.com/mcay23/hound/blob/main/screenshots/tvpage2.png)

# API Keys
You need a [TMDB API key](https://developers.themoviedb.org/3/getting-started/introduction) to run Hound.

# Docker Compose
Set your API keys in the `compose.env` file. If you change the database username/password, make sure you change the DB connection string as well.
```
docker compose up
```

# Build from Source
Build both the frontend and backend separately. By default, the frontend runs on `http://localhost:3000` and the backend runs on `http://localhost:8080`. If you change the backend server host, adjust `server_host` in `/web/src/config.json` to point to the backend.
## Frontend
```bash
cd web

npm install

npm run start
```
## Backend  
Set your API keys in the `server/.env` file. If you change the database username/password, make sure you change the DB connection string as well.
```
cd server

go build
```

# Contributing
Project is still at a very early stage. Expect many bugs. Please report any you see. 
This project needs contributors. Feel free to contact me if you're interested.

