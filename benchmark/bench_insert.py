import matplotlib.pyplot as plt
import numpy as np


# Dataset 1: Cache
cache = {
    1000: 0.65,
    5000: 3.42,
    10000: 6.81,
    50000: 31.64,
    100000: 61.25,
    300000: 174.69,
    500000: 290.88,
    700000: 402.35,
    999999: 583.83
}

# Dataset 2: Disk
disk = {
    1000: 0.75,
    5000: 3.52,
    10000: 6.77,
    50000: 32.53,
    100000: 64.12,
    300000: 200.83,
    500000: 344.25,
    700000: 486.86,
    999999: 704.01
}

# Dataset 3: update commit logic
dataset3 = {
    1000: 0.69,
    5000: 3.32,
    10000: 6.77,
    50000: 36.59,
    100000: 75.66,
    300000: 223.49,
    500000: 370.56,
    700000: 522.80,
    999999: 716.43
}

# Dataset 4: map => concurrentmap
dataset4 = {
    1000: 0.74,
    5000: 3.43,
    10000: 6.72,
    50000: 32.45,
    100000: 66.61,
    300000: 195.11,
    500000: 315.05,
    700000: 432.60,
    999999: 612.41
}


datasets = [
    ("Cache", cache),
    ("Disk", disk),
    ("DiskUpdate1", dataset3),
    ("DiskFinalUpdate", dataset4)
]

processed_datasets = []
for name, data in datasets:
    sorted_items = sorted(data.items())
    keys, times = zip(*sorted_items)
    processed_datasets.append((name, keys, times))


plt.figure(figsize=(14, 10))

for name, keys, times in processed_datasets:
    plt.plot(keys, times, marker='o', label=name)

plt.title('CompareNNV History', fontsize=16)
plt.xlabel('insert vector', fontsize=14)
plt.ylabel('latency (sec)', fontsize=14)
plt.xscale('log') 
plt.legend()
plt.grid(True, which="both", ls="--", linewidth=0.5)
plt.show()