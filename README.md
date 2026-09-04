# Hound Media Server

<br />
<p align="center">
  <img src="https://github.com/Hound-Media-Server/hound/blob/main/web/public/hound-logo.png" alt="Logo" width="350">
</p>

<h3 align="center">
  <strong>The Modern Hybrid Media Server</strong>
</h3>

Watch and Track Movies and TV Shows. Self-hosted version of Plex/Stremio + Trakt, Simkl, etc. Hound aims to be a complete ecosystem of watching, tracking, downloading, and archiving media.

Hound is a fully-featured media server, like Jellyfin or Plex, but with the additional ability to stream content through P2P (torrent), HTTP/Debrid sources and Usenet. With Hound, you get the benefits of fully controlling your media like Jellyfin, but can also stream instantly like Stremio. You can also watch Live TV using IPTV sources (xtream and m3u8). Everything comes together seamlessly, into a single, intuitive, platform.

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

The desktop clients for Windows, MacOS (arm) are available in the releases section. You can download the Android and Android TV apps from the [App Repo](https://github.com/Hound-Media-Server/hound-app) in the releases page. You'll need to sideload the .APKs. iOS and tvOS share the same codebase, but are not available yet since they have more requirements to publish, for now you can only run them on XCode. Stay tuned.

# Installation

Docker compose is the preferred method for installing Hound. Read the installation docs [here](https://hound-media-server.github.io/hound-site/installation.html).

# Features

### Current

- Stream and download your own content from your drives, or stream content directly from P2P (torrent), HTTP/Debrid sources, and Usenet through Stremio addons (AIOStreams)
- Connect and watch your IPTV sources (Xtream and M3U8 playlists)
- Trakt-like features, all your watches are automatically tracked and easily browsable
- Customize your home layout with built-in catalogs or MDBList lists
- Create custom collections/lists
- Add reviews and comments to your media
- Watch calendar and statistics
- Android and Android TV clients (iOS and tvOS coming soon)
- Focus on UI/UX, and Admin experience
- Really fast to setup, zero to watching content in <10 mins, few dependencies

### Planned

- Transcoding
- Manually create your own movies/shows
- Data export
- Third-party review score integration (eg. IMDB, Metacritic, RT)
- View actor information (eg. movies they've played)
- Review individual seasons, episodes (TV Shows)
- Add private notes for your movies, episodes

# Development

Make sure postgres is running on your machine. Copy `server/dev.env.example` to `server/dev.env`, then customize for your dev environment — at minimum, set `HOUND_SECRET` to a unique random value (e.g. `openssl rand -base64 48`); the server panics on startup if it's empty. Build and run both the frontend and backend separately. By default, the frontend runs on `http://localhost:3000` and the backend runs on `http://localhost:2323`.

Hound uses vite and electron for the web apps and desktop clients. Devs should develop frontend features for both at the same time. The desktop version uses [electron-mpv-video](https://www.npmjs.com/package/electron-mpv-video) for media playback. In order to build for electron, follow the instructions [here](https://www.npmjs.com/package/electron-mpv-video), in the 'Native build requirements' section.

### Backend

```
cd server
go run main.go
```

### Frontend

```bash
cd web/electron
npm install
npm run dev
```

# Screenshots

![home page](https://github.com/Hound-Media-Server/hound/blob/main/assets/home.png)
![tv page](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage.png)
![tv page 2](https://github.com/Hound-Media-Server/hound/blob/main/assets/tvpage2.png)
