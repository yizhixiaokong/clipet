# Clipet 进化系统详解

## 进化树模型

进化系统是一个**有向无环图 (DAG)**，以 egg 阶段为根节点。

### 阶段 (Stage)

每个节点有三个属性：
- **id** — 唯一标识符，如 `baby_cat`
- **name** — 显示名称，如 `小猫咪`
- **phase** — 生命阶段，五选一：

| Phase   | 名称 | 建议最短持续 | 说明           |
|---------|------|-------------|----------------|
| egg     | 蛋期 | 1 小时      | 起始阶段       |
| baby    | 幼年 | 24 小时     | 单一路径       |
| child   | 少年 | 72 小时     | 首次分支       |
| adult   | 成年 | 720 小时    | 分支加深       |
| legend  | 传说 | —           | 终极形态，不再进化 |

### 进化边 (Evolution)

每条边定义 `from → to` 的转变，附带一组条件。

## 进化条件字段

| 字段                      | 类型             | 说明                          |
|--------------------------|------------------|-------------------------------|
| `min_age_hours`          | float64          | 宠物最低年龄（小时）           |
| `attr_bias`              | string           | 属性偏好方向                   |
| `min_dialogues`          | int              | 最低对话互动次数               |
| `min_adventures`         | int              | 最低冒险完成次数               |
| `min_feed_regularity`    | float64 (0-1)    | 喂食规律性比率                 |
| `night_interactions_bias`| bool             | 要求夜间互动>日间              |
| `day_interactions_bias`  | bool             | 要求日间互动>夜间              |
| `min_interactions`       | int              | 最低总互动次数                 |
| `min_attr`               | map[string]int   | 属性最低值要求 (如 happiness≥90) |

### attr_bias 取值

| 值         | 含义                     | 追踪字段         |
|-----------|--------------------------|------------------|
| happiness | 偏向快乐的养成           | AccHappiness     |
| health    | 偏向健康的养成           | AccHealth        |
| playful   | 偏向玩耍的养成           | AccPlayful       |

## 累积追踪机制

Pet 结构体中的累积字段在每次互动时自动更新：

| 操作      | 更新字段                                   |
|-----------|-------------------------------------------|
| `Feed()`  | FeedCount++, trackTimeOfDay()              |
| `Play()`  | AccPlayful++, trackTimeOfDay()             |
| `Talk()`  | AccHappiness++, DialogueCount++, trackTimeOfDay() |

`trackTimeOfDay()` 根据当前时间（6:00-18:00 为日间）更新 DayInteractions / NightInteractions。

## 喂食规律性

```
FeedRegularity = FeedCount / FeedExpectedCount
```

FeedExpectedCount 基于宠物年龄计算预期的合理喂食次数。

## 内置猫物种进化树

```
                    egg (神秘之蛋)
                        │
                 min_age ≥ 1h
                        │
                   baby_cat (小猫咪)
                  ╱     │      ╲
           happiness  health   playful
           +对话 ≥10  +冒险 ≥5  +喂食规律 ≥0.7
           min_age   min_age   min_age
            ≥24h      ≥24h      ≥24h
            ╱          │          ╲
    child_arcane  child_feral  child_mech
    (咒术小猫)    (战斗小猫)   (机甲小猫)
      ╱    ╲       ╱    ╲       ╱    ╲
    夜间    日间  hunger energy 对话≥30 冒险≥15
   偏好    偏好   偏好   偏好
    ╱        ╲    ╱      ╲      ╱        ╲
shadow_  crystal_ flame_  frost_ cyber_  chrome_
 mage    oracle   lion   panther lynx   jaguar
 (暗影    (水晶    (烈焰  (霜暴   (赛博  (合金
  魅猫)   预言猫)   狮)    豹)    猞猁)  猎豹)
   │        │       │      │      │       │
 互动≥500 互动≥500 互动≥500 互动≥500 互动≥500 互动≥500
 幸福≥90  幸福≥90  健康≥90  精力≥90  幸福≥85  健康≥85
 730h+    730h+    730h+   730h+   730h+   730h+
   │        │       │      │      │       │
  void_   astral_  immor.  cryo.  quantum stellar
  walker   sage    inferno deity  phantom predator
 (虚空    (星辰    (不灭  (极寒   (量子  (星际
  行者)   贤者)    炎帝)  霜神)   幽灵)  掠夺者)
```

## 进化判定流程（M2 实现）

```
每次游戏 tick / 互动：
  1. 获取当前 stageID 的所有可用进化边
  2. 对每条边检查所有条件：
     - pet.AgeHours() ≥ min_age_hours ?
     - 对应 Acc* 字段是否满足 attr_bias ?
     - DialogueCount ≥ min_dialogues ?
     - AdventuresCompleted ≥ min_adventures ?
     - FeedRegularity ≥ min_feed_regularity ?
     - Night/Day 互动比例是否满足 bias ?
     - TotalInteractions ≥ min_interactions ?
     - 各属性是否 ≥ min_attr ?
  3. 收集所有满足条件的进化路径
  4. 如果有多条满足 → 选择最佳匹配（或提示玩家选择）
  5. 执行进化 → 更新 Stage/StageID/Phase
  6. 保存存档
```
