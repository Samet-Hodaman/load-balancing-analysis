#!/usr/bin/env python3
import sys
import pandas as pd
import numpy as np
import matplotlib
matplotlib.use('Agg')  # Non-interactive backend
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
    # Clean NaN and Inf values
    lat = lat[np.isfinite(lat)]
    # Filter negative or zero values (must be positive for log scale)
    lat = lat[lat > 0]
    return np.sort(lat)

def smooth_cdf(lat):
    # Make minimum value safe (for log scale)
    min_val = lat.min()
    max_val = lat.max()
    
    if min_val <= 0:
        min_val = max_val * 1e-6
    else:
        min_val = max(min_val, max_val * 1e-6)
    
    # Calculate in log space
    log_min = np.log10(min_val)
    log_max = np.log10(max_val)
    x = np.logspace(log_min, log_max, 600)
    
    # Calculate KDE
    try:
        kde = gaussian_kde(lat)
        pdf = kde(x)
        pdf = np.nan_to_num(pdf, nan=0.0, posinf=0.0, neginf=0.0)
        cdf = np.cumsum(pdf)
        if cdf[-1] > 0:
            cdf /= cdf[-1]
        else:
            cdf = np.searchsorted(lat, x, side='right') / len(lat)
    except Exception as e:
        print(f"Warning: KDE failed ({e}), using empirical CDF")
        cdf = np.searchsorted(lat, x, side='right') / len(lat)
    
    return x, cdf

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: plot_latency.py <csv_file> [algorithm] [output_file]")
        sys.exit(1)
    
    csv_file = sys.argv[1]
    algorithm = sys.argv[2] if len(sys.argv) > 2 else "unknown"
    
    # Generate output filename with algorithm name in snake_case
    if len(sys.argv) > 3:
        output_file = sys.argv[3]
    else:
        # Default: results/latency_plot_<algorithm>.png
        output_file = f"results/latency_plot_{algorithm}.png"
    
    try:
        lat = parse_latency_csv(csv_file)
        
        if len(lat) == 0:
            print("No valid latency data found")
            sys.exit(1)
        
        x, cdf = smooth_cdf(lat)
        
        plt.figure(figsize=(10,6))
        # Format algorithm name for display (replace underscores with spaces, capitalize)
        algorithm_display = algorithm.replace("_", " ").title()
        plt.plot(x, cdf, linewidth=2.5, label=f"Latency Distribution ({algorithm_display})")
        
        plt.xscale("log")
        plt.xlabel("Latency (ms, log scale)")
        plt.ylabel("CDF")
        plt.title(f"Latency Distribution (CDF) - {algorithm_display}", fontweight="bold")
        
        plt.grid(True, which="both", alpha=0.3)
        plt.legend()
        plt.tight_layout()
        plt.savefig(output_file, dpi=300)
        print(f"Plot saved to {output_file}")
        plt.close()
    except Exception as e:
        print(f"Error generating plot: {e}")
        sys.exit(1)
