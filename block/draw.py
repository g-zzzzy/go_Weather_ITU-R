import matplotlib.pyplot as plt

# 文件列表
files = [
    "../data/terminals_group_0.txt",
    "../data/terminals_group_1.txt",
    "../data/terminals_group_2.txt",
    "../data/terminals_group_3.txt",
    "../data/terminals_group_4.txt",
    "../data/terminals_group_5.txt",
    "../data/terminals_group_6.txt",
    "../data/terminals_group_7.txt",
]

colors = ["red", "blue", "green", "orange", "purple", "brown", "cyan", "magenta"]
# colors = ["red", "blue", "green", "orange", "purple", "brown"]
# colors = ["red", "blue", "green", "orange"]
plt.figure(figsize=(10, 5))

for idx, fname in enumerate(files):
    lats, lons = [], []
    with open(fname, "r") as f:
        for line in f:
            parts = line.strip().split()
            if len(parts) != 2:
                continue
            lat, lon = map(float, parts)
            lats.append(lat)
            lons.append(lon)
    plt.scatter(lons, lats, c=colors[idx], s=10, label=f"Group {idx}", alpha=0.7)

plt.xlabel("Longitude")
plt.ylabel("Latitude")
plt.title("Terminal Groups Distribution (HilbertKMeans Partition)")
plt.legend()
plt.grid(True)
# 保存图表到文件，支持png、pdf、svg等多种格式
plt.savefig("hilbertk_2000_8.png", dpi=300, bbox_inches="tight")
