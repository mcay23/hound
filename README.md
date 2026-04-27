# Hound Media Server

<p align="center">
  <img src="https://github.com/Hound-Media-Server/hound/blob/main/web/public/hound-logo.png" alt="Logo" width="200">
</p>

<p align="center">
  <strong>The Modern Hybrid Media Server</strong>
</p>

Watch and Track Movies and TV Shows. Self-hosted version of Plex/Stremio + Trakt, Simkl, etc. Hound aims to be a complete ecosystem of watching, tracking, downloading, and archiving media.

Hound is a fully-featured media server, like Jellyfin or Plex, but with the additional ability to stream content through P2P (torrent) or HTTP/Debrid sources. With Hound, you get the benefits of fully controlling your media like Jellyfin, but can also stream instantly like Stremio. It's the best of both worlds.

# Links

- [Documentation](https://hound-media-server.github.io/hound-site/)
- [App Repo (Android, iOS)](https://github.com/Hound-Media-Server/hound-app) <- Download the clients here

# Features

### Current

- Stream and download your own content from your drives, or stream content directly from P2P (torrent) and HTTP/ Debrid sources through Stremio addons
- Trakt-like features, all your watches are automatically tracked and easily browsable
- Create custom collections/lists
- Add reviews and comments to your media
- Android and Android TV clients (iOS and tvOS coming soon)
- Focus on UI/UX, and Admin experience
- Really fast to setup, zero to watching content in <10 mins, few dependencies

### Planned

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
- Add private notes for your movies, episodes

# Screenshots

![home page](https://github.com/Hound-Media-Server/hound/blob/main/assets/home.png)
![tv page](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage.png)
![tv page 2](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage2.png)

# Docker Compose

Docker is the preferred method for installing Hound.
Change:

- `POSTGRES_PASSWORD` in both hound-postgres and hound-server to the same, strong password
- `HOUND_SECRET` to a random, strong key

```yaml
services:
  hound-postgres:
    container_name: hound-postgres
    image: postgres:18
    environment:
      POSTGRES_DB: hound_db
      POSTGRES_USER: hound
      POSTGRES_PASSWORD: super-strong-password
    volumes:
      - postgres_data:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hound -d hound_db"]
      interval: 5s
      timeout: 5s
      retries: 5

  hound-server:
    container_name: hound-server
    image: houndmediaserver/hound:latest
    depends_on:
      hound-postgres:
        condition: service_healthy
    ports:
      - "2323:2323"
    environment:
      - POSTGRES_DB=hound_db
      - POSTGRES_USER=hound
      - POSTGRES_PASSWORD=super-strong-password
      - HOUND_SECRET=super-strong-secret
    volumes:
      - ./Hound Data:/app/Hound Data
      # (Optional) attach your media library
      # IMPORTANT: Please read the docs before doing this
      # - /path/to/movies:/app/External Library/Movies
      # - /path/to/shows:/app/External Library/TV Shows

volumes:
  postgres_data:
```

then `docker compose up -d`

Next, you'll want to [set up a provider](https://hound-media-server.github.io/hound-site/provider.html) to start watching content.

# Development

Make sure postgres is running on your machine. Modify `/server/dev.env` to suit your dev environment. Build and run both the frontend and backend separately. By default, the frontend runs on `http://localhost:3000` and the backend runs on `http://localhost:2323`.

### Backend

```
cd server
go run main.go
```

### Frontend

```bash
cd web
npm install
npm run start
```
