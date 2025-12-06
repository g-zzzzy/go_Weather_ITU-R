import matplotlib.pyplot as plt
import re
from matplotlib.patches import Rectangle

# ---------------------- 配置参数（严格对齐参考脚本） ----------------------
# 输入文件路径
FILES = {
    "hilbert": "../coordinator/hilbert_terminal_groups.txt",  # hilbert分区文件
    "id": "../coordinator/id_terminal_groups.txt"             # id分区文件
}
# 颜色列表（和参考脚本完全一致）
COLORS = ["red", "blue", "green", "orange", "purple", "brown", "cyan", "magenta"]
# 图表保存路径
SAVE_PATH_HILBERT = "hilbert_sin_distribution.png"  # 参考脚本的命名风格
SAVE_PATH_ID = "id_distribution.png"
# 矩形框配置（不影响基础样式）
RECT_ALPHA = 0.2  # 矩形填充透明度
RECT_LINE_WIDTH = 2  # 矩形边框宽度

# ---------------------- 解析函数 ----------------------
def parse_terminal_file(file_path):
    """
    解析终端分组文件，返回：
    {节点ID: {
        "lons": 经度列表,
        "lats": 纬度列表,
        "lon_range": (min_lon, max_lon),
        "lat_range": (min_lat, max_lat)
    }}
    """
    node_data = {}
    current_node = None
    current_lons = []
    current_lats = []
    current_lon_range = (None, None)
    current_lat_range = (None, None)
    
    # 正则匹配规则
    node_pattern = re.compile(r"--- 节点(\d+) 终端分组 ---")
    coord_pattern = re.compile(r"终端\d+: 经度=([-+]?\d+\.\d+), 纬度=([-+]?\d+\.\d+)")
    range_pattern = re.compile(r"经纬度范围: 经度\[([-+]?\d+\.\d+), ([-+]?\d+\.\d+)\], 纬度\[([-+]?\d+\.\d+), ([-+]?\d+\.\d+)\]")

    with open(file_path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            # 匹配节点ID
            node_match = node_pattern.match(line)
            if node_match:
                if current_node is not None:
                    node_data[current_node] = {
                        "lons": current_lons,
                        "lats": current_lats,
                        "lon_range": current_lon_range,
                        "lat_range": current_lat_range
                    }
                current_node = int(node_match.group(1))
                current_lons = []
                current_lats = []
                current_lon_range = (None, None)
                current_lat_range = (None, None)
                continue
            
            # 匹配单个终端经纬度
            coord_match = coord_pattern.match(line)
            if coord_match:
                lon = float(coord_match.group(1))
                lat = float(coord_match.group(2))
                current_lons.append(lon)
                current_lats.append(lat)
                continue
            
            # 匹配经纬度范围
            range_match = range_pattern.match(line)
            if range_match:
                min_lon = float(range_match.group(1))
                max_lon = float(range_match.group(2))
                min_lat = float(range_match.group(3))
                max_lat = float(range_match.group(4))
                current_lon_range = (min_lon, max_lon)
                current_lat_range = (min_lat, max_lat)
                continue
        
        # 保存最后一个节点
        if current_node is not None:
            node_data[current_node] = {
                "lons": current_lons,
                "lats": current_lats,
                "lon_range": current_lon_range,
                "lat_range": current_lat_range
            }
    
    return node_data

# ---------------------- 绘图函数（严格对齐参考脚本） ----------------------
def plot_terminal_distribution(node_data, title, save_path):
    """
    严格按照参考脚本规格绘图
    """
    # 1. 画布尺寸（和参考脚本完全一致：10,5）
    plt.figure(figsize=(10, 5))
    ax = plt.gca()
    
    # 2. 遍历节点绘图（按参考脚本逻辑）
    for idx, (node_id, data) in enumerate(node_data.items()):
        lons = data["lons"]
        lats = data["lats"]
        lon_range = data["lon_range"]
        lat_range = data["lat_range"]
        color = COLORS[idx % len(COLORS)]  # 用参考脚本的颜色列表
        
        # 绘制散点（参数和参考脚本完全一致）
        plt.scatter(
            lons, lats, 
            c=color, 
            s=10,  # 点大小和参考一致
            label=f"Group {idx}",  # 图例格式和参考一致（Group X）
            alpha=0.7  # 透明度和参考一致
        )
        
        # 绘制经纬度范围矩形框（可选）
        if lon_range[0] is not None and lat_range[0] is not None:
            min_lon, max_lon = lon_range
            min_lat, max_lat = lat_range
            width = max_lon - min_lon
            height = max_lat - min_lat
            rect = Rectangle(
                (min_lon, min_lat),
                width,
                height,
                linewidth=RECT_LINE_WIDTH,
                edgecolor=color,
                facecolor=color,
                alpha=RECT_ALPHA,
                zorder=0  # 矩形框置于散点下方，避免遮挡
            )
            ax.add_patch(rect)
    
    # 3. 图表样式（和参考脚本完全一致）
    plt.xlabel("Longitude")  # x轴标签（无单位，和参考一致）
    plt.ylabel("Latitude")  # y轴标签（无单位，和参考一致）
    plt.title(title)  # 标题传参（匹配参考的命名风格）
    plt.legend()  # 图例（默认位置，和参考一致）
    plt.grid(True)  # 网格（无额外参数，和参考一致）
    
    # 4. 保存参数（和参考脚本完全一致）
    plt.savefig(save_path, dpi=300, bbox_inches="tight")
    plt.close()
    print(f"图表已保存至: {save_path}")

# ---------------------- 主执行逻辑 ----------------------
if __name__ == "__main__":
    # 1. 绘制 Hilbert 分区（标题匹配参考脚本风格）
    hilbert_data = parse_terminal_file(FILES["hilbert"])
    plot_terminal_distribution(
        hilbert_data,
        title="Terminal Groups Distribution (Hilbert Partition)",  # 参考风格标题
        save_path=SAVE_PATH_HILBERT
    )
    
    # 2. 绘制 ID 分区
    id_data = parse_terminal_file(FILES["id"])
    plot_terminal_distribution(
        id_data,
        title="Terminal Groups Distribution (ID Partition)",
        save_path=SAVE_PATH_ID
    )
    
    print("所有图表绘制完成！")