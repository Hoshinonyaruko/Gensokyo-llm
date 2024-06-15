#### promptMarks 配置部分

1. **promptMarks**:
   - 类型：数组（包含 `BranchConfig` 结构体）
   - 功能：定义对话过程中的标记和跳转规则。

2. **`BranchConfig` 结构体**：
   - `BranchName`：分支标识，表示跳转的目标分支名称。
   - `Keywords`：关键字列表，包含与该分支关联的关键词。

#### 配置示例

1. **promptMarkType == 0**：
   - 格式：
     ```yaml
     promptMarks:
       - BranchName: "去逛街路上"
         Keywords: []
       - BranchName: "在家准备"
         Keywords: []
     ```
   - 功能：当对话达到 `promptMarksLength`（即标记数组的长度）时，会随机选择一个分支进行跳转。
   - 示例解释：
     - 当对话达到指定长度时，将在 `去逛街路上` 和 `在家准备` 之间随机选择一个分支进行跳转。

2. **promptMarkType == 1**：
   - 格式：
     ```yaml
     promptMarks:
       - BranchName: "去逛街路上"
         Keywords: ["坐车", "走路", "触发"]
       - BranchName: "在家准备"
         Keywords: ["等一下", "慢慢", "准备"]
     ```
   - 功能：在对话过程中，判断 Q（问题）和 A（回答）是否包含指定关键词，包含则跳转到相应的分支。
   - 示例解释：
     - 如果对话中包含关键词 "坐车", "走路", 或 "触发"，则跳转到 `去逛街路上` 分支。
     - 如果对话中包含关键词 "等一下", "慢慢", 或 "准备"，则跳转到 `在家准备` 分支。

### promptMarkType 解释

- `promptMarkType = 0`:
  - 代表按 `promptMarksLength` 来切换提示词文件。
  - `promptMarksLength` 代表本提示词文件维持的上下文长度。
  - 当 `promptMarksLength` 小于 0 时，会从 `promptMarks` 中读取之后的分支，并从中随机一个切换。

- `promptMarkType = 1`:
  - 代表按条件触发，当 `promptMarksLength` 达到时也会触发。
  - 条件格式：
    ```yaml
    promptMarks:
      - BranchName: "aaaa"
        Keywords: ["xxx", "xxx", "xxxx", "xxx"]
    ```
  - 示例解释：
    - `aaaa` 是 `promptMarks` 中的分支标识。
    - `["xxx", "xxx", "xxxx", "xxx"]` 是关键字列表，当 Q 或 A 包含列表中的任意一个关键字时，会触发跳转到 `aaaa` 分支。

