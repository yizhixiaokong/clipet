# Set 命令扩展说明

## 扩展内容

`clipet-dev set` 命令现在支持修改所有宠物数据，包括：
- 核心属性（4个）
- 基本信息（4个）
- **统计数据（5个，新增）**
- **进化累积（5个，新增）**
- **生命周期状态（1个，新增）**

## 支持的字段

### 核心属性 (0-100)
| 键名 | 显示名 | 类型 | 说明 |
|-----|--------|------|------|
| `hunger` | 饱腹 | int | 饱食度 |
| `happiness` | 快乐 | int | 快乐度 |
| `health` | 健康 | int | 健康度 |
| `energy` | 精力 | int | 精力值 |

### 基本信息
| 键名 | 显示名 | 类型 | 说明 |
|-----|--------|------|------|
| `name` | 名字 | string | 宠物名字 |
| `species` | 物种 | string | 物种ID（如 "cat"）|
| `stage_id` | 阶段ID | string | 进化阶段（如 "baby", "adult_arcane_shadow"）|
| `alive` | 存活 | bool | 是否存活 |

### 统计数据 🆕
| 键名 | 显示名 | 类型 | 说明 |
|-----|--------|------|------|
| `interactions` | 总互动 | int | 总互动次数 |
| `feeds` | 喂食次数 | int | 喂食次数 |
| `dialogues` | 对话次数 | int | 对话次数 |
| `adventures` | 冒险次数 | int | 完成的冒险次数 |
| `wins` | 游戏胜利 | int | 游戏胜利次数 |

### 进化累积器 🆕
| 键名 | 显示名 | 类型 | 说明 |
|-----|--------|------|------|
| `acc_happiness` | 快乐累积 | int | 快乐累积点数 |
| `acc_health` | 健康累积 | int | 健康累积点数 |
| `acc_playful` | 玩耍累积 | int | 玩耍累积点数 |
| `night` | 夜间互动 | int | 夜间互动次数 |
| `day` | 日间互动 | int | 日间互动次数 |

### 生命周期 (M7) 🆕
| 键名 | 显示名 | 类型 | 说明 |
|-----|--------|------|------|
| `lifecycle_warning` | 生命预警 | bool | 是否显示过生命预警 |

## 使用方式

### 1. 交互式界面（推荐）

```bash
./clipet-dev set
```

显示所有字段，用方向键选择，回车编辑，输入新值。

**界面示例**：
```
 ╭─ 小猫咪 [baby] ╮
 │                │
 ▸ 饱腹      65  ████████░░░░
   快乐      70  ███████░░░░░
   健康      80  ████████░░░░
   精力      75  ███████░░░░░
   总互动    25  ░░░░░░░░░░░░  ← 新增
   喂食次数  10  ░░░░░░░░░░░░  ← 新增
   对话次数  5   ░░░░░░░░░░░░  ← 新增
   冒险次数  3   ░░░░░░░░░░░░  ← 新增
   ...
```

### 2. 直接命令行

```bash
# 设置核心属性
./clipet-dev set hunger 100

# 设置统计数据
./clipet-dev set interactions 500
./clipet-dev set adventures 30
./clipet-dev set feeds 50

# 设置进化累积
./clipet-dev set night 100
./clipet-dev set acc_happiness 200

# 设置生命周期
./clipet-dev set lifecycle_warning true
```

## 实际应用场景

### 1. 测试进化条件

```bash
# 快速达成"冒险一生"终局条件
./clipet-dev set adventures 30

# 快速达成"快乐累积"进化条件
./clipet-dev set acc_happiness 100
./clipet-dev set interactions 500

# 测试夜间偏好进化
./clipet-dev set night 50
```

### 2. 测试生命周期

```bash
# 触发生命预警（需要配合 timeskip）
./clipet-dev set lifecycle_warning false  # 重置预警
./clipet-dev timeskip --hours 180          # 跳到7.5天触发预警

# 查看预警状态
./clipet status
```

### 3. 修正统计数据

```bash
# 如果统计数据异常，可以手动修正
./clipet-dev set feeds 20
./clipet-dev set dialogues 15
```

### 4. 强制进化

```bash
# 设置满足进化条件的属性
./clipet-dev set happiness 90
./clipet-dev set interactions 500
./clipet-dev set acc_happiness 100
# 然后执行任何互动操作会自动触发进化检查
```

## 与旧版本的区别

| 项目 | 旧版本 | 新版本 |
|-----|--------|--------|
| 支持字段 | 8个 | 19个 |
| 统计数据 | ❌ 不支持 | ✅ 完整支持 |
| 进化累积 | ❌ 不支持 | ✅ 完整支持 |
| 生命周期 | ❌ 不存在 | ✅ 支持M7字段 |
| 交互界面 | 仅核心属性 | 显示所有字段 |

## 技术实现

### 修改的文件

1. **internal/game/pet.go**
   - 扩展 `SetField()` 方法支持所有新字段
   - 添加字段别名（如 `interactions` = `total_interactions`）
   - 优化错误提示，分类显示有效字段

2. **cmd/clipet-dev/set.go**
   - 扩展 `settableFields` 列表
   - 更新 `getCurrentPetValue()` 读取所有新字段
   - 更新帮助文档，分类说明字段类型

### 兼容性

- ✅ 向后兼容：所有旧字段名仍然有效
- ✅ 别名支持：`interactions` = `total_interactions`, `feeds` = `feed_count`
- ✅ 类型安全：int/string/bool 类型自动转换
- ✅ 范围限制：核心属性自动 clamp 到 0-100

## 示例输出

```bash
$ ./clipet-dev set interactions 500
set 总互动: 25 -> 500

$ ./clipet-dev set adventures 30
set 冒险次数: 3 -> 30

$ ./clipet-dev set
# 交互式界面显示所有字段...
```

## 后续改进建议

1. **添加 feed_regularity 支持**：目前只能通过修改 feed_count 间接影响
2. **批量修改**：支持一次修改多个字段
3. **预设模板**：一键设置到某个进化分支的理想状态
4. **实时预览**：在TUI中显示修改后会达成的进化条件
