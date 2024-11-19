import matplotlib.pyplot as plt
import numpy as np



nnv = {
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

milvus = {
    1000: 3.46,
    5000: 16.94,
    10000: 33.70,
    50000: 170.54,
    100000: 357.51,
    300000: 1197.36,
    500000: 1998.60,
    700000: 2863.72,
    999999: 4225.28
}


datasets = [
    ("NNV", nnv),
    ("Milvus", milvus)
]


processed_datasets = []
for name, data in datasets:
    sorted_items = sorted(data.items())
    keys, times = zip(*sorted_items)
    processed_datasets.append((name, keys, times))


plt.figure(figsize=(14, 10))

for name, keys, times in processed_datasets:
    plt.plot(keys, times, marker='o', label=name)

plt.title('Compare Insert Latency', fontsize=16)
plt.xlabel('insert vector', fontsize=14)
plt.ylabel('latency (sec)', fontsize=14)
plt.xscale('log') 
plt.legend()
plt.grid(True, which="both", ls="--", linewidth=0.5)
plt.show()