import json
import numpy as np
import matplotlib.pyplot as plt
import sys

# JSON formatını kontrol et
# Önce tek bir JSON objesi olarak okumayı dene (özet format)
latencies_list = []
is_summary_format = False
data = None

with open("./results.json") as f:
    content = f.read().strip()
    if not content:
        print("HATA: Dosya boş.")
        sys.exit(1)
    
    # Önce tek bir JSON objesi olarak parse etmeyi dene
    try:
        data = json.loads(content)
        # Eğer "latencies" anahtarı varsa ve dict ise, özet formatı
        if isinstance(data.get("latencies"), dict):
            is_summary_format = True
    except json.JSONDecodeError:
        # Tek bir JSON objesi değilse, json-stream formatı olabilir
        # Dosyayı satır satır oku
        f.seek(0)
        for line in f:
            if line.strip():
                try:
                    req_data = json.loads(line)
                    if "latency" in req_data:
                        latencies_list.append(req_data["latency"])
                except json.JSONDecodeError:
                    continue

if is_summary_format:
    print("UYARI: Bu JSON dosyası özet istatistikler içeriyor, ham latency verileri yok.")
    print("Ham veriler için Vegeta'yı şu şekilde çalıştırın:")
    print("  vegeta attack -format=json-stream -output=results.json < targets.txt")
    print("\nMevcut özet istatistikleri kullanarak yaklaşık bir grafik oluşturuluyor...")
    
    # Özet istatistiklerden yaklaşık bir dağılım oluştur
    lat_stats = data["latencies"]
    requests = data.get("requests", 1000)
    
    # Min, mean, max ve percentile değerlerini kullanarak yaklaşık dağılım oluştur
    min_lat = lat_stats["min"] / 1e6  # ns'den ms'ye
    mean_lat = lat_stats["mean"] / 1e6
    max_lat = lat_stats["max"] / 1e6
    p50 = lat_stats["50th"] / 1e6
    p90 = lat_stats["90th"] / 1e6
    p95 = lat_stats["95th"] / 1e6
    p99 = lat_stats["99th"] / 1e6
    
    # Normal dağılım kullanarak yaklaşık latency değerleri oluştur
    # (Bu tam olarak doğru değil ama görselleştirme için yeterli)
    std = (p95 - mean_lat) / 1.645  # p95 için yaklaşık std hesapla
    latencies = np.random.normal(mean_lat, std, min(requests, 10000))
    latencies = np.clip(latencies, min_lat, max_lat)  # Min-max aralığında tut
    
    # Bilinen percentile değerlerini ekle
    known_percentiles = [min_lat, p50, p90, p95, p99, max_lat]
    latencies = np.concatenate([latencies, known_percentiles])
    
elif latencies_list:
    # json-stream formatından latency verileri alındı
    latencies = np.array(latencies_list) / 1e6  # ns'den ms'ye çevir
else:
    print("HATA: JSON dosyasından latency verileri okunamadı.")
    sys.exit(1)

# Sırala ve CDF hesapla
latencies_sorted = np.sort(latencies)
cdf = np.arange(1, len(latencies_sorted) + 1) / len(latencies_sorted)

# Percentile'lar
p95 = np.percentile(latencies_sorted, 95)
p99 = np.percentile(latencies_sorted, 99)

# Plot
plt.figure()
plt.plot(latencies_sorted, cdf)
plt.axvline(p95, linestyle='--', label=f"p95 = {p95:.2f} ms")
plt.axvline(p99, linestyle='--', label=f"p99 = {p99:.2f} ms")
plt.xlabel("Latency (ms)")
plt.ylabel("CDF")
plt.legend()
plt.grid(True)
plt.show()
