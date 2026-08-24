# Hound Media Server

<br />
<p align="center">
  <img src="https://github.com/Hound-Media-Server/hound/blob/main/web/public/hound-logo.png" alt="Logo" width="350">
</p>

<h3 align="center">
  <strong>The Modern Hybrid Media Server</strong>
</h3>

This branch is the desktop version of Hound. It uses Electron to wrap the web app and libmpv as the underlying players, and is meant to be the intended way to use hound on desktops. [electron-mpv-video](https://www.npmjs.com/package/electron-mpv-video) is the main dependency for video playback. We use a patched version to extend features such as subtitles.
