sc_token = " " # https://now.es3n1n.eu/sc/
import requests
sc__info = requests.get("http://localhost:8080/get_current_track_soundcloud", headers = {
    "sc-token": sc_token,
})
print(sc__info.text)

#ya_token = " " # https://github.com/MarshalX/yandex-music-api/discussions/513
#ya__info = requests.get("http://localhost:8080/get_current_track_beta", headers = {
#     "ya-token": ya_token,
#})
#print(ya__info.text)   
