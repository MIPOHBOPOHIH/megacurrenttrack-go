import requests

headers = {"Authorization": "OAuth y0_asd"}

response = requests.get(
    "http://localhost:8080/get_current_track_alpha", headers=headers
)

print(response.status_code)
print(response.text)
