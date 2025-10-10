import json
import re
import requests
from helpers.utils import get_args, rank_and_parse_torrents, exit, request_with_retry, is_valid_magnet_hash

'''
Parses 
Arguments
    --preset          -> preset of the search
    --imdb_id         -> tt12388
    --media_type      -> tvshow/movie
    --duration        -> in seconds, used to calculate bitrate
    --title           -> the clean title of the media
    --year(int)
    --season(int)
    --episode(int)
    --languages       -> comma seperated values
    --rd_api_key      -> rd api key, since torrentio has info on which torrents are cached if you provide an api key

Response
    {
        go_status: "success" or "error - error message"
        streams: [{}]
        provider: "torrentio"
    }
'''

PROVIDER_NAME = "prowlarr"
# get command line arguments
args = get_args()

PROWLARR_API_KEY = "bf55c78d6136444abeff369781a690f8"
PROWLARR_URL = "http://localhost:9696/api/v1/search?query=%s&type=%s&limit=30" # imdb-id

response = {"go_status":"error"}

# TODO REFACTOR FILENAMES AND RETEST

# build query string
try:
    year_str = ""
    if args.media_type == "movie":
        if args.year > 0:
            year_str = " (" + str(args.year) + ")"
        url = PROWLARR_URL % (args.title + year_str, "moviesearch")
    elif args.media_type == "tvshow":
        season_episode = " S" + str(args.season).zfill(2) + "E" + str(args.episode).zfill(2)
        url = PROWLARR_URL % (args.title + season_episode, "tvsearch") 
except:
    exit(PROVIDER_NAME, "error constructing query string. Check that the arguments are correct")

headers = {
    "Authorization": "Bearer " + PROWLARR_API_KEY
}
try:
    res = request_with_retry("GET", url, headers=headers)
    json_data = res.json()
    response["go_status"] = "success"
except requests.exceptions.RequestException as e:
    exit(PROVIDER_NAME, f"error during HTTP call - {e}")

torrents = []
for torrent in json_data:
    # only handle torrents for now
    # if not "protocol" not in torrent or torrent.protocol != "torrent":
    #     continue
    torrents.append({
        "torrent_name": torrent.get("title", None),
        "info_hash": torrent.get("infoHash", None),
        "seeders": torrent.get("seeders", 0),
        "leechers": torrent.get("leechers", -1),
        "clean_title": args.title,
        "cached": "unknown",
        "indexer": torrent.get("indexer", "")
    })

print(torrents)
ranked = rank_and_parse_torrents(torrents, PROVIDER_NAME)
# remove bad matches
filtered = []
for t in ranked:
    # skip bad matches
    if t.get("lev_ratio", 0) < 0.8:
        continue
    if "rank" not in t:
        continue
    # boost high seeder torrents
    seeders = t.get("seeders", 0)
    if seeders >= 100:
        t["rank"] += 900
    elif seeders >= 50:
        t["rank"] += 600
    elif seeders >= 25:
        t["rank"] += 300
    # penalize low seeders
    elif seeders < 10:
        t["rank"] -= 500
    elif seeders < 5:
        t["rank"] -= 1000
    filtered.append(t)
# sort list based on rank
filtered = sorted(list(filtered), key=lambda x: x["rank"], reverse=True)
response["addon"] = PROVIDER_NAME
response["streams"] = filtered

with open("prowlarr.json", "w", encoding="utf-8") as f:
    json.dump(response, f, indent=4)
