#!/bin/bash
# Basit grafik oluşturma script'i
# Python ve gerekli paketlerin yüklü olduğunu varsayar
ALGORITHM=${1:-round_robin}

# Yerel Python kullanarak grafik oluştur
python3 load-balancer/plot_latency.py results/latency.csv "$ALGORITHM"

echo "Grafik oluşturuldu: results/latency_plot_${ALGORITHM}.png"
