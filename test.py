import json

# JSON dosyalarını yükle
with open("stations_full.json", "r", encoding="utf-8") as stations_file:
    stations = json.load(stations_file)

with open("pairs.json", "r", encoding="utf-8") as pairs_file:
    pairs = json.load(pairs_file)

with open("cities.json", "r", encoding="utf-8") as cities_file:
    cities = json.load(cities_file)

# Yeni JSON'u oluştur
updated_stations = []

for station in stations:
    if station.get("showOnQuery", False):
        matching_pairs = [pair["pairs"] for pair in pairs if pair["id"] == station["id"]]
        city_name = next((city["name"] for city in cities if city["id"] == station.get("cityId")), None)
        updated_station = {
            "id": station["id"],
            "name": station["name"],
            "pairs": matching_pairs[0] if matching_pairs else [],
            "cityName": city_name
        }
        updated_stations.append(updated_station)

# Yeni JSON dosyasını kaydet
with open("updated_stations.json", "w", encoding="utf-8") as output_file:
    json.dump(updated_stations, output_file, ensure_ascii=False, indent=4)

print("İşlem tamamlandı! updated_stations.json dosyası oluşturuldu.")
