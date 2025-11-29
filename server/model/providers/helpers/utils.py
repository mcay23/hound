import argparse
import re
import sys
from RTN.parser import parse
from RTN import RTN
from RTN.models import SettingsModel, CustomRank, CustomRanksConfig, DefaultRanking, ResolutionConfig, QualityRankModel, BaseRankingModel
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

RES_4K = ('2160', '216o', '.4k', 'ultrahd', 'ultra.hd', '.uhd.')
RES_1080 = ('1080', '1o8o', '108o', '1o80', '.fhd.')
RES_720 = ('720', '72o')
RES_SCR = ('dvdscr', 'screener', '.scr.', '.r5', '.r6')
RES_CAM = ('1xbet', 'betwin', '.cam.', 'camrip', 'cam.rip', 'dvdcam', 'dvd.cam', 'dvdts', 'hdcam', '.hd.cam', '.hctc', '.hc.tc', '.hdtc',
				'.hd.tc', 'hdts', '.hd.ts', 'hqcam', '.hg.cam', '.tc.', '.tc1', '.tc7', '.ts.', '.ts1', '.ts7', 'tsrip', 'telecine', 'telesync', 'tele.sync')

def get_resolution(term):
	if any(i in term for i in RES_SCR): return 'SCR'
	elif any(i in term for i in RES_CAM): return 'CAM'
	elif any(i in term for i in RES_720): return '720p'
	elif any(i in term for i in RES_1080): return '1080p'
	elif any(i in term for i in RES_4K): return '4K'
	elif '.hd.' in term: return '720p'
	else: return 'SD'
	
def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--imdb_id", type=str, help="(string) imdb_id")
    parser.add_argument("--preset", type=str, help="(string) preset")
    # "browser": suitable for browser playback, try to avoid transcoding, good network
    # "player": suitable for playback in players with wider codec support, still aim for good network
    # "quality": prioritize quality above all, bitrate not an issue
    parser.add_argument("--media_type", type=str, help="(string) movie/tvshow") 
    parser.add_argument("--duration", type=int, help="(int) duration of the media in seconds") # to calculate bitrate, since we can't quickly parse duration from torrents manually
    parser.add_argument("--title", type=str, help="(string) clean title of the media")
    parser.add_argument("--year", type=int, help="(int) publishing year of the media")
    parser.add_argument("--season", type=int, help="(int) tv show season (if applicable)")
    parser.add_argument("--episode", type=int, help="(int) tv show episode (if applicable)")
    parser.add_argument("--languages", type=str, help="(string) comma separated ISO 639-1 language codes")
    parser.add_argument("--rd_api_key", type=str, help="(string) Real debrid api key, if available")
    parser.add_argument("--connection_string", type=str, help="(string) For passing connection info such as host, username, pw, etc.")
    return parser.parse_args()

# handle exiting with go_status
def exit(provider, reason):
    print({"provider": provider, "go_status":"error - " + reason})
    sys.exit(1)
    
def request_with_retry(
    method, 
    url,
    headers=None, 
    data=None, # For POST/PUT requests
    json=None, # For POST/PUT requests with JSON payload
    retries=3, 
    backoff_factor=1, 
    status_forcelist=(500, 502, 503, 504)
):
    """
    Performs an HTTP request (GET, POST, etc.) with a specified number of retries 
    on certain HTTP status codes and connection errors.

    :param method: The HTTP method (e.g., 'GET', 'POST'). Case-insensitive.
    :param url: The URL to perform the request on.
    :param data: (optional) Dictionary, bytes, or file-like object to send in the body.
    :param json: (optional) JSON data to send in the body.
    :param retries: The number of times to retry the request.
    :param backoff_factor: A factor to calculate the sleep time between retries.
    :param status_forcelist: A set of HTTP status codes that trigger a retry.
    :return: A requests.Response object.
    :raises requests.exceptions.RequestException: If the request fails after all retries.
    """
    retry_strategy = Retry(
        total=retries, 
        read=retries,
        connect=retries, 
        backoff_factor=backoff_factor,
        status_forcelist=status_forcelist,
        allowed_methods={"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"} 
    )

    adapter = HTTPAdapter(max_retries=retry_strategy)
    http = requests.Session()
    http.mount("https://", adapter)
    http.mount("http://", adapter)
    if not headers:
        headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                        "AppleWebKit/537.36 (KHTML, like Gecko) "
                        "Chrome/124.0.0.0 Safari/537.36",
            "Accept": "application/json, text/plain, */*"
        }
    response = http.request(
        method=method,
        url=url,
        headers=headers,
        data=data,
        json=json
    )
    
    # Raise an exception for persistent bad status codes
    response.raise_for_status() 
    
    return response

def is_valid_magnet_hash(h: str) -> bool:
    hex_pattern = re.compile(r"^[0-9a-fA-F]{40}$")
    base32_pattern = re.compile(r"^[A-Z2-7]{32}$")

    return bool(hex_pattern.match(h) or base32_pattern.match(h))

# settings for browsers, try to avoid transcoding
browser_ranking = DefaultRanking(
    # prefer avc over hevc for performance
    avc=450,
    hevc=400,
    webdl=0,
    webmux=0,
    remux=0,
    xvid=-150,
    # audio codecs, penalize ac3 for browsers
    atmos=-300, 
    dolby_digital=-300,
    dolby_digital_plus=-300,
    dts_lossless=-300,
    dts_lossy=-500,
    # others
    bit_10=0,
    bdrip=0,
    brrip=0,
    hdrip=0,
    uhdrip=0,
    webrip=0,
    webdlrip=0,
    size=0,
    # hdr
    bit10=0,
    dolby_vision=0,
    hdr=10,
    # extras
    dubbed=-100,
    network=0,
    edition=0,
    proper=0,
    repack=0,
    scene=0,
    site=0,
    upscaled=0,
)

# Rank and parse torrent languages, codecs, etc.
def rank_and_parse_torrents(torrents, provider, useDebrid=False):
    rtn = RTN(settings=SettingsModel(), ranking_model=browser_ranking)
    torrentsData = []
    for t in torrents:
        try:
            # use torrent name since quality and codec info is not always in each file
            if provider == "torrentio" or provider == "prowlarr":
                data = rtn.rank(t["torrent_name"], t["info_hash"], correct_title=t["clean_title"])
            elif provider == "aiostreams":
                # parse name is concatenation of file and folder names
                data = rtn.rank(t["file_name"], t["info_hash"], correct_title=t["title"])
                # try parsing folder name if no resolution, since for some packs the data is only in the folder name
                try:
                    t_dict = data.model_dump()
                    if t_dict["data"]["resolution"] == "unknown" and t["folder_name"] != "":
                        data = rtn.rank(t["folder_name"], t["info_hash"], correct_title=t["title"])
                except:
                    pass
            else:
                # might need to do some episode / season parsing to validate from other sources such as jackett, etc.
                data = rtn.rank(t["file_name"], t["info_hash"])
        except Exception:
            continue
    
        t_dict = data.model_dump()
        t_dict["addon"] = t.get("addon", "")
        t_dict["indexer"] = t.get("indexer", "")
        t_dict["url"] = t.get("url", "")
        t_dict["seeders"] = t.get("seeders", 0)
        t_dict["leechers"] = t.get("leechers", -1)
        t_dict["file_idx"] = t.get("file_idx", -1)
        t_dict["file_size"] = t.get("file_size", -1)
        t_dict["file_size_string"] = t.get("file_size_string", -1)
        t_dict["duration"] = t.get("duration", -1)
        t_dict["duration_string"] = t.get("duration_string", -1)
        t_dict["service"] = t.get("service", "")
        t_dict["resolution"] = t.get("resolution", "")
        t_dict["sources"] = t.get("sources", [])

        # don't always know if torrent is cached
        t_dict["cached"] = t.get("cached", "unknown")

        t_dict["file_name"] = t.get("file_name", "")
        t_dict["folder_name"] = t.get("folder_name", "")
        t_dict["p2p"] = t.get("p2p", "")

        # rename raw_title to filename
        # check if torrentio/prowlarr needs this
        # t_dict["raw_title"] = t.get("file_name", "")
        t_dict.pop('raw_title', None)

        # penalize unknown resolutions
        if t_dict["resolution"] == "":
            t_dict["rank"] -= 2000
        # penalize unknown audio codecs, as this may disrupt browser playback 
        if len(t_dict["data"]["audio"]) == 0:
            t_dict["rank"] -= 300
        # boost cached torrents for debrid
        if useDebrid:
            if t_dict["cached"] == "true":
                t_dict["rank"] += 10000
        else:
            # if torrent streaming, boost ones with more seeders
            seeders = t_dict["seeders"]
            if seeders >= 50:
                t_dict["rank"] += 500
            elif seeders >= 25:
                t_dict["rank"] += 250
            elif seeders < 10:
                t_dict["rank"] -= 200
        # penalize files below 50MB
        if t_dict["file_size"] < 50000000:
            t_dict["rank"] -= 2000
        # REVIEW: penalize italian, russian streams since browser playback can't change
        # audio language
        if "it" in t.get("languages", []) or "ru" in t.get("languages", []):
            t_dict["rank"] -= 50
        torrentsData.append(t_dict)
    return torrentsData

