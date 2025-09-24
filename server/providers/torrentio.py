import urllib.request
import urllib.parse
import json

movie_url = "https://torrentio.strem.fun/stream/movie/%s.json" # imdb-id
tv_url = "https://torrentio.strem.fun/stream/series/%s:%s:%s.json" # imdb-id:season:ep
params = {"userId": 1}

url = movie_url % "tt0068646"
data = {"go_status":False}
try:
    with urllib.request.urlopen(url) as response:
        res = response.read().decode("utf-8")
        json_data = json.loads(res)
        data.update(json_data)
        data["go_status"] = True
except urllib.error.HTTPError as e:
    print("HTTP Error:", e.code, e.reason)
except urllib.error.URLError as e:
    print("URL Error:", e.reason)
except Exception as e:
    print("Unexpected Error:", e)

print(json.dumps(data))