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
PROVIDER_NAME = "torrentio"
# get command line arguments
args = get_args()
# add rd_url if key exists. This helps get torrent cache info
rd_url = "/realdebrid=" + args.rd_api_key if args.rd_api_key else ""

MOVIE_URL = "https://torrentio.strem.fun%s/stream/movie/%s.json" # imdb-id
TV_URL = "https://torrentio.strem.fun%s/stream/series/%s:%s:%s.json" # imdb-id:season:ep

# build the response dict
response = {"go_status":"error"}
try:
    if args.media_type == "movie":
        url = MOVIE_URL % (rd_url, args.imdb_id)
    elif args.media_type == "tvshow":
        url = TV_URL % (rd_url, args.imdb_id, args.season, args.episode)
except:
    exit(PROVIDER_NAME, "error parsing JSON in args")
# HTTP call
try:
    res = request_with_retry("GET", url)
    json_data = res.json()
    response["go_status"] = "success"
except requests.exceptions.RequestException as e:
    exit(PROVIDER_NAME, f"error during HTTP call - {e}")

if "streams" not in json_data or len(json_data["streams"]) <= 0:
    exit(PROVIDER_NAME, "unexpected response body")

torrents = []
for i in json_data["streams"]:
    # get cached status if rd api key is supplied
    cached = "unknown"
    infohash = i.get("infoHash", None)
    file_idx = i.get("fileIdx", -1)
    if rd_url:
        cached = "rd" if "[RD+]" in i.get("name", None) else "false"
        # infohash, fileIdx is returned differently by torrentio depending on whether api key is supplied
        # we want to standardize the output, so return infohash
        # infohash_parse = i.get("behaviorHints", {}).get("bingeGroup", None).split("|")
        url_parse = i.get("url", None).split("/")
        if url_parse:
            hash_idx = url_parse.index("realdebrid") + 2
            infohash = None if len(url_parse) < hash_idx + 1 else url_parse[hash_idx]
            file_idx = hash_idx + 2
            file_idx = -1 if len(url_parse) < file_idx + 1 else url_parse[file_idx]
            # some responses are not in the magnet format
            # probably not necessary since RTN checks magnet validity during ranking
            # if not is_valid_magnet_hash(infohash) or file_idx == -1:
            #     continue
    # get filename inside behaviorHints, remove bad items
    filename = i.get("behaviorHints", {}).get("filename", None)
    title = i.get("title", None)
    if filename is None or infohash is None or title is None:
        continue

    # get seeders and filesize for torrentio
    pattern = re.compile(r"👤\s*(\S+)\s*💾\s*([\d.]+\s*[A-Z]+)")
    match = pattern.search(title)
    seeders = 0
    if match:
        try:
            seeders = int(match.group(1))
        except:
            seeders = 0
        filesize = match.group(2)
    torrents.append({
        "filename": filename,
        "infohash": infohash,
        "file_idx": file_idx,
        "torrent_name": title,
        "correct_title": args.title, 
        "seeders": seeders,
        "cached": cached
    })
# use RTN to rank torrents. All providers should follow this response pattern
ranked = rank_and_parse_torrents(torrents, PROVIDER_NAME)
# sort list based on rank
ranked = sorted(list(ranked), key=lambda x: x["rank"], reverse=True)
response["provider"] = PROVIDER_NAME
response["streams"] = ranked

with open("output.json", "w", encoding="utf-8") as f:
    json.dump(response, f, indent=4)
