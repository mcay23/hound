# Hound
 Track TV Shows, Movies, etc. Self-hosted version of Trakt, Simkl, etc.
 
# Demo
Access the demo [here](http://107.174.11.52/)
```
username: demo
password: demodemo
```
# Features
- Current
  - Search Movies, TV Shows, and Games
  - Multiple users
  - Create collections (playlists)
  - Add watch data
  - Write reviews
- Future
  - Integrate more media types (books, manga, etc.)
  - Third-party integrations (plex, jellyfin)
  - Add more views/information to each page (eg. DLCs for games)
  - Recommendations
  - Add your own metadata, manually add media
  - Data export
  - Third-party review score integration (eg. IMDB, Metacritic, RT)
  - View actor information (eg. movies they've played)
  - View public collections / other user's collections
  - Review individual seasons, episodes (TV Shows)
  - Add private notes for your media

# API Keys
You need a [TMDB API key](https://developers.themoviedb.org/3/getting-started/introduction) and [IGDB Client ID and secret](https://api-docs.igdb.com/) to run Hound.


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

# Screenshots
![home page](https://github.com/mcay23/hound/blob/main/screenshots/home.png)
![tv page](https://github.com/mcay23/hound/blob/main/screenshots/tvpage.png)
![tv page 2](https://github.com/mcay23/hound/blob/main/screenshots/tvpage2.png)

