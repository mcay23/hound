# Hound Media Server

<br />
<p align="center">
  <img src="https://github.com/Hound-Media-Server/hound/blob/main/web/public/hound-logo.png" alt="Logo" width="350">
</p>

<h3 align="center">
  <strong>The Modern Hybrid Media Server</strong>
</h3>

Watch and Track Movies and TV Shows. Self-hosted version of Plex/Stremio + Trakt, Simkl, etc. Hound aims to be a complete ecosystem of watching, tracking, downloading, and archiving media.

Hound is a fully-featured media server, like Jellyfin or Plex, but with the additional ability to stream content through P2P (torrent) or HTTP/Debrid sources. With Hound, you get the benefits of fully controlling your media like Jellyfin, but can also stream instantly like Stremio. It's the best of both worlds.

> [!CAUTION]
> Hound is still under heavy development and may contain bugs. Please backup your data periodically.

> [!NOTE]
> Full documentation and installation guide can be found [here](https://hound-media-server.github.io/hound-site/).

# Links

- [Documentation](https://hound-media-server.github.io/hound-site/)
- [Subreddit](https://www.reddit.com/r/HoundMediaServer/) <- Follow updates
- [Installation](https://hound-media-server.github.io/hound-site/installation.html)
- [App Repo (Android, iOS)](https://github.com/Hound-Media-Server/hound-app) <- Download the clients here
- [API Docs](https://hound-media-server.github.io/hound-site/operations/authentication.html)

# Demo
Access the demo [here](https://hound-demo.yuwono.xyz)
```
username: github
password: password
```

# Platforms
You can download the Android and Android TV apps from the [App Repo](https://github.com/Hound-Media-Server/hound-app) in the releases page. You'll need to sideload the .APKs. iOS and tvOS share the same codebase, but are not available yet since they have more requirements to publish, for now you can only run them on XCode. Stay tuned. 

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

# Installation

Docker compose is the preferred method for installing Hound. Read the installation docs [here](https://hound-media-server.github.io/hound-site/installation.html).

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

# Screenshots

![home page](https://github.com/Hound-Media-Server/hound/blob/main/assets/home.png)
![tv page](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage.png)
![tv page 2](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage2.png)
