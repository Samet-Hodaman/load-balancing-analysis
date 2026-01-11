import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from scipy.stats import gaussian_kde

def parse_latency_csv(path):
    df = pd.read_csv(path)

    def parse(v):
        v = str(v).strip()
        if v.endswith("ns"):
            return float(v.replace("ns","")) / 1e6
        elif v.endswith("µs") or v.endswith("us"):
            return float(v.replace("µs","").replace("us","")) / 1000
        elif v.endswith("ms"):
            return float(v.replace("ms",""))
        elif v.endswith("s"):
            return float(v.replace("s","")) * 1000
        else:
            return float(v)

    lat = df["latency_ms"].apply(parse).values
    # NaN ve Inf değerleri temizle
    lat = lat[np.isfinite(lat)]
    # Negatif veya sıfır değerleri filtrele (log scale için pozitif olmalı)
    lat = lat[lat > 0]
    return np.sort(lat)

def smooth_cdf(lat):
    # Minimum değeri güvenli hale getir (log scale için)
    min_val = lat.min()
    max_val = lat.max()
    
    # Eğer min çok küçükse, güvenli bir alt sınır belirle
    if min_val <= 0:
        min_val = max_val * 1e-6
    else:
        # Minimum değer çok küçükse, max'ın bir kısmını kullan
        min_val = max(min_val, max_val * 1e-6)
    
    # Log space hesapla
    log_min = np.log10(min_val)
    log_max = np.log10(max_val)
    x = np.logspace(log_min, log_max, 600)
    
    # KDE hesapla
    try:
        kde = gaussian_kde(lat)
        pdf = kde(x)
        # PDF'de NaN/Inf kontrolü
        pdf = np.nan_to_num(pdf, nan=0.0, posinf=0.0, neginf=0.0)
        cdf = np.cumsum(pdf)
        if cdf[-1] > 0:
            cdf /= cdf[-1]
        else:
            # Fallback: empirik CDF
            cdf = np.searchsorted(lat, x, side='right') / len(lat)
    except Exception as e:
        # KDE başarısız olursa empirik CDF kullan
        print(f"Warning: KDE failed ({e}), using empirical CDF")
        cdf = np.searchsorted(lat, x, side='right') / len(lat)
    
    return x, cdf

lat_lc = parse_latency_csv("latency_lc.csv")
lat_rr = parse_latency_csv("latency_rr.csv")

x_lc, cdf_lc = smooth_cdf(lat_lc)
x_rr, cdf_rr = smooth_cdf(lat_rr)

plt.figure(figsize=(10,6))

plt.plot(x_lc, cdf_lc, linewidth=2.5, label="Least Connections")
plt.plot(x_rr, cdf_rr, linewidth=2.5, label="Round Robin")

plt.xscale("log")
plt.xlabel("Latency (ms, log scale)")
plt.ylabel("CDF")
plt.title("Latency Distribution (CDF): LC vs RR", fontweight="bold")

plt.grid(True, which="both", alpha=0.3)
plt.legend()
plt.tight_layout()
plt.savefig("latency_cdf_lc_vs_rr.png", dpi=300)
plt.show()
