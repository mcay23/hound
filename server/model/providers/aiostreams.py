import json
import requests
from helpers.utils import get_args, rank_and_parse_torrents, exit, request_with_retry, is_valid_magnet_hash

'''
Parses 
Arguments
    --preset                 -> preset of the search
    --imdb_id                -> tt12388
    --media_type             -> tvshow/movie
    --duration               -> in seconds, used to calculate bitrate
    --title                  -> the clean title of the media
    --year(int)
    --season(int)
    --episode(int)
    --languages              -> comma seperated values
    --rd_api_key             -> rd api key, since torrentio has info on which torrents are cached if you provide an api key
    --connection_string      -> host;uuid;password string: eg. http://abc.com:5000;uuid-str-here;my-password

Response
    {
        go_status: "success" or "error - error message"
        streams: [{}]
        provider: "aiostreams"
    }
'''
PROVIDER_NAME = "aiostreams"
# get command line arguments
args = get_args()
try:
    (host, uuid, password) = args.connection_string.split("|")
except:
    exit(PROVIDER_NAME, "Invalid or missing --connection_string")

MOVIE_URL = host + "/stremio/%s/%s/stream/movie/%s.json" # imdb-id
TV_URL = host + "/stremio/%s/%s/stream/series/%s:%s:%s.json" # imdb-id:season:ep
USER_URL = host + "/api/v1/user?uuid=%s&password=%s" % (uuid, password)

# build the response dict
response = {"go_status":"error"}

# get user
try:
    res = request_with_retry("GET", USER_URL)
    user_data = res.json()
    if not user_data["success"]:
        exit(PROVIDER_NAME, str(user_data["error"]))
    # get encrypted password for next calls
    encryptedPassword = user_data["data"]["encryptedPassword"]
    # check if debrid services are used
    useDebrid = False
    for provider in user_data["data"]["userData"]["services"]:
        if provider["enabled"]:
            useDebrid = True 
            break # at least one debrid provider
except Exception as e:
    exit(PROVIDER_NAME, "Error getting user data, check the credentials are correct and aiostreams is running" + str(e))

try:
    if args.media_type == "movie":
        url = MOVIE_URL % (uuid, encryptedPassword, args.imdb_id)
    elif args.media_type == "tvshow":
        url = TV_URL % (uuid, encryptedPassword, args.imdb_id, args.season, args.episode)
except:
    exit(PROVIDER_NAME, "error parsing JSON in args")

# HTTP call
try:
    res = request_with_retry("GET", url)
    json_data = res.json()
    response["go_status"] = "success"
except requests.exceptions.RequestException as e:
    exit(PROVIDER_NAME, f"error during HTTP call - {e}")

if "streams" not in json_data:
    exit(PROVIDER_NAME, "unexpected response body" + json.dumps(json_data))

# parse the descriptions from the response into dicts
# descriptions have the form 
# "addon:Torrentio\ntitle:My Show\nyear:2022\nseason:1\nepisode:4\nservice:RD\ncached:true\np2p:debrid\nresolution:2160p\nquality:WEB-DL" ... etc
torrents = []
for t in json_data["streams"]:
    if "description" not in t:
        continue
    torrent = {}
    for line in t["description"].split("\n"):
        if ":" in line:
            key, value = line.split(":", 1)
            torrent[key] = value if value else ""
            # convert to int
            try:
                torrent["seeders"] = int(torrent.get("seeders", -1))
            except:
                torrent["seeders"] = -1
            try:
                torrent["file_size"] = int(torrent.get("file_size", -1))
            except:
                torrent["file_size"] = -1
    # when service such as debrid is note used, no url is returned, but file index is
    torrent["url"] = t.get("url", "")
    # get p2p trackers, not usually returned for debrid searches
    torrent["sources"] = t.get("sources", [])
    torrent["file_idx"] = t.get("fileIdx", -1)
    if torrent["file_name"] == "":
        continue
    # generate a string for RTN to parse
    # we append folder and file name since sometimes filenames don't have video details
    torrent["parse_title"] = torrent["file_name"] + " " + torrent["folder_name"]
    torrents.append(torrent)

ranked = rank_and_parse_torrents(torrents, PROVIDER_NAME, useDebrid=useDebrid)
        

# sort list based on rank
ranked = sorted(list(ranked), key=lambda x: x["rank"], reverse=True)

response["provider"] = PROVIDER_NAME
response["streams"] = ranked
# with open("output.json", "w", encoding="utf-8") as f:
#     json.dump(response, f, indent=4)
print(json.dumps(response))

"""
Below is the required template in the formatter for aiostreams for this script to parse
{stream.title::exists["{stream.title::title}"||""]}{stream.year::exists[" ({stream.year})"||""]}

addon:{addon.name}
title:{stream.title::exists["{stream.title::title}"||""]}
year:{stream.year::exists["{stream.year}"||""]}
season:{stream.season::exists["{stream.season}"||""]}
episode:{stream.episode::exists["{stream.episode}"||""]}
info_hash:{stream.infoHash::exists["{stream.infoHash}"||""]}
service:{service.id::exists["{service.shortName}"||""]}
cached:{service.cached::istrue["true"||"false"]}
p2p:{stream.type::exists["{stream.type}"||"Unknown"]}
proxied:{stream.proxied::istrue["true"||"false"]}
resolution:{stream.resolution::exists["{stream.resolution}"||""]}
quality:{stream.quality::exists["{stream.quality}"||""]}
encode:{stream.encode::exists["{stream.encode}"||""]}
scene_group:{stream.releaseGroup::exists["{stream.releaseGroup}"||""]}
visual_tags:{stream.visualTags::exists["{stream.visualTags::join(',')}"||""]}
audio_tags:{stream.audioTags::exists["{stream.audioTags::join(',')}"||""]}
audio_channel:{stream.audioChannels::exists["{stream.audioChannels::join(',')}"||""]}
file_size:{stream.size::>0["{stream.size}"||""]}
folder_size:{stream.folderSize::>0["{stream.folderSize}"||""]}
file_size_string:{stream.size::>0["{stream.size::bytes}"||""]}
folder_size_string:{stream.folderSize::>0["{stream.folderSize::bytes}"||""]}
duration:{stream.duration::>0["{stream.duration}"||""]}
duration_string:{stream.duration::>0["{stream.duration::time}"||""]}
seeders:{stream.seeders::>0["{stream.seeders}"||"0"]}
age:{stream.age::exists["{stream.age}"||""]}
indexer:{stream.indexer::exists["{stream.indexer}"||""]}
languages:{stream.languages::exists["{stream.languages::join(',')}"||""]}
languages_codes:{stream.languageCodes::exists["{stream.languageCodes::join(',')::lower}"||""]}
file_name:{stream.filename::exists["{stream.filename}"||""]}
folder_name:{stream.folderName::exists["{stream.folderName}"||""]}
message:{stream.message::exists["{stream.message}"||""]}
"""
