# 案例1：社区论坛系统 - 设计说明文档

## 一、整体架构

### 1.1 模块划分
- **用户系统**：`users` 表
- **社交关系**：`user_follows` 表
- **内容管理**：`posts`、`comments`、`tags`、`post_tags`、`likes` 表
- **权限管理**：`roles`、`permissions`、`role_permissions`、`user_roles` 表（RBAC模型）
- **扩展功能**：`sections`、`section_moderators` 表

### 1.2 设计原则
1. **软删除**：核心表（users、posts、comments）使用 `is_deleted` + `deleted_at`，便于数据恢复和审计
2. **冗余字段**：`posts.like_count`、`posts.comment_count` 等，用空间换查询性能
3. **索引优化**：所有外键、常用查询字段都建索引
4. **字符集**：`utf8mb4` 支持 emoji 和特殊字符

---

## 二、核心表设计详解

### 2.1 用户表（users）

**关键字段：**
- `username`、`email` 唯一索引，防止重复注册
- `status`：1-正常 2-禁用 3-待激活（支持邮箱激活流程）
- `is_deleted` + `deleted_at`：软删除，保留历史数据

**设计考虑：**
- 密码存储 `password_hash`（实际用 bcrypt/argon2）
- `avatar_url` 支持 CDN 地址
- `bio` 限制 500 字符，防止过长

---

### 2.2 关注关系表（user_follows）

**关键设计：**
- `follower_id`（粉丝）→ `following_id`（被关注者）
- `is_mutual`：标记互相关注（可通过触发器或应用层维护）
- 唯一索引 `(follower_id, following_id)` 防止重复关注

**查询场景：**
- 某用户的粉丝列表：`WHERE following_id = ?`
- 某用户关注的人：`WHERE follower_id = ?`
- 互相关注：`WHERE is_mutual = 1 AND (follower_id = ? OR following_id = ?)`

**性能优化：**
- 两个方向的索引：`idx_follower`、`idx_following`
- 外键 `ON DELETE CASCADE`：用户删除时自动清理关注关系

---

### 2.3 帖子表（posts）

**关键字段：**
- `status`：1-草稿 2-待审核 3-已发布 4-已删除（支持审核流程）
- `content_html`：可选字段，存储渲染后的 HTML（提升列表页性能）
- `view_count`、`like_count`、`comment_count`：冗余字段，避免 JOIN 查询

**索引策略：**
- `idx_status`：按状态筛选（审核后台）
- `idx_created`：按时间排序（首页列表）
- `idx_top`：置顶帖优先显示

**设计考虑：**
- 外键 `ON DELETE RESTRICT`：用户/版块删除时不允许删除帖子（需先处理帖子）
- `section_id`：支持多版块分类

---

### 2.4 评论表（comments）

**多级回复设计：**
- `parent_id`：NULL 表示一级评论，非 NULL 表示回复某条评论
- 支持无限层级（但前端通常只显示 2-3 层）

**查询场景：**
- 某帖子的所有评论：`WHERE post_id = ? AND parent_id IS NULL`
- 某评论的所有回复：`WHERE parent_id = ?`
- 递归查询子评论（需要应用层或存储过程）

**性能优化：**
- `idx_post`、`idx_parent`：快速定位评论树
- `like_count` 冗余字段

---

### 2.5 点赞表（likes）

**通用设计（Polymorphic）：**
- `target_type`：1-帖子 2-评论（未来可扩展：3-用户动态等）
- `target_id`：对应目标的 ID
- 唯一索引 `(user_id, target_type, target_id)` 防止重复点赞

**设计优势：**
- 一张表支持多种点赞场景
- 查询某用户所有点赞：`WHERE user_id = ?`
- 查询某帖子/评论的点赞列表：`WHERE target_type = ? AND target_id = ?`

**注意：**
- 点赞/取消点赞时，需要同步更新 `posts.like_count` 或 `comments.like_count`（可用触发器或应用层事务）

---

### 2.6 标签系统（tags + post_tags）

**多对多关系：**
- `tags`：标签主表，`usage_count` 记录使用次数（热门标签排序）
- `post_tags`：关联表，一个帖子可以有多个标签

**查询场景：**
- 某帖子的所有标签：`JOIN post_tags ON tags.id = post_tags.tag_id WHERE post_tags.post_id = ?`
- 某标签的所有帖子：`JOIN post_tags ON posts.id = post_tags.post_id WHERE post_tags.tag_id = ?`
- 热门标签：`ORDER BY usage_count DESC`

---

### 2.7 权限管理（RBAC模型）

**三层结构：**
1. **用户（users）** ←→ **角色（roles）** ←→ **权限（permissions）**
2. `user_roles`：用户-角色关联（一个用户可以有多个角色）
3. `role_permissions`：角色-权限关联（一个角色可以有多个权限）

**权限表设计：**
- `resource` + `action`：如 `post.create`、`comment.delete`
- `code` 唯一索引，便于程序判断：`IF user.hasPermission('post.delete')`

**查询用户权限：**
```sql
SELECT DISTINCT p.code
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = ?
```

**设计优势：**
- 灵活：新增角色/权限无需改表结构
- 可扩展：支持权限继承、权限组等高级特性

---

## 三、设计亮点与优化

### 3.1 软删除 vs 硬删除
- **软删除**：users、posts、comments（重要数据，需要恢复/审计）
- **硬删除**：likes、user_follows（关系数据，删除即清理）

### 3.2 冗余字段策略
- `posts.like_count`、`comment_count`：避免 `COUNT(*)` 查询
- `tags.usage_count`：热门标签排序
- **维护方式**：应用层事务保证一致性，或使用触发器

### 3.3 索引优化
- **唯一索引**：防止重复数据（username、email、关注关系等）
- **复合索引**：`idx_status`、`idx_top` 支持多条件查询
- **外键索引**：所有外键自动建索引（MySQL InnoDB）

### 3.4 扩展性考虑
- **版块表（sections）**：支持多分区，`sort_order` 灵活排序
- **版主表（section_moderators）**：版块-用户多对多，一个版块多个版主
- **通用点赞表**：`target_type` 设计，未来可扩展点赞其他资源

---

## 四、常见查询示例

### 4.1 获取某用户的粉丝列表
```sql
SELECT u.*, uf.created_at AS follow_time
FROM user_follows uf
JOIN users u ON uf.follower_id = u.id
WHERE uf.following_id = ?
ORDER BY uf.created_at DESC;
```

### 4.2 获取某版块的热门帖子（按点赞数+时间）
```sql
SELECT * FROM posts
WHERE section_id = ? AND status = 3 AND is_deleted = 0
ORDER BY like_count DESC, created_at DESC
LIMIT 20;
```

### 4.3 获取某帖子的评论树（一级评论+回复数）
```sql
-- 一级评论
SELECT c.*, u.nickname, u.avatar_url,
  (SELECT COUNT(*) FROM comments WHERE parent_id = c.id) AS reply_count
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.post_id = ? AND c.parent_id IS NULL AND c.is_deleted = 0
ORDER BY c.created_at ASC;
```

### 4.4 判断用户是否有某权限
```sql
SELECT COUNT(*) > 0 AS has_permission
FROM user_roles ur
JOIN role_permissions rp ON ur.role_id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE ur.user_id = ? AND p.code = 'post.delete';
```

---

## 五、潜在问题与改进建议

### 5.1 性能优化
1. **分表策略**：帖子表按时间或版块分表（数据量大时）
2. **缓存层**：Redis 缓存热门帖子、用户信息
3. **读写分离**：主从复制，读操作走从库

### 5.2 数据一致性
1. **冗余字段维护**：`like_count` 等字段需要事务保证一致性
2. **互相关注标记**：`is_mutual` 可通过触发器或应用层维护

### 5.3 扩展功能
1. **消息通知**：新增 `notifications` 表（点赞通知、评论通知等）
2. **内容审核**：`posts.audit_status`、`audit_log` 表
3. **数据统计**：`user_stats` 表（发帖数、获赞数等）

---

## 六、总结

本设计覆盖了：
- ✅ 社交关系（关注/粉丝）
- ✅ 内容管理（帖子/评论/标签）
- ✅ 权限管理（RBAC）
- ✅ 软删除、审计字段
- ✅ 性能优化（冗余字段、索引）
- ✅ 扩展性（通用设计、多对多关系）

**适合场景**：中小型社区论坛、技术博客、问答平台等。

