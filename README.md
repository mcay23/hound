# Hound
 Track TV Shows, Movies, etc.
 
# Demo
Access the demo [here](http://107.174.11.52/)
```
username: demo
password: demodemo
```

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

# Contributing
Project is still at a very early stage. Expect many bugs. Please report any you see. 
This project needs contributors. Feel free to contact me if you're interested.

# Screenshots
![home page](https://github.com/mcay23/hound/blob/main/screenshots/home.png)
![tv page](https://github.com/mcay23/hound/blob/main/screenshots/tvpage.png)
![tv page 2](https://github.com/mcay23/hound/blob/main/screenshots/tvpage2.png)

