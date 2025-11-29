-- ============================================
-- 案例1：社区论坛系统 - 数据库表设计
-- 数据库：MySQL 8.0+
-- 字符集：utf8mb4（支持 emoji）
-- ============================================

-- ==================== 1. 用户系统 ====================

-- 用户表
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名（唯一）',
  `email` VARCHAR(100) NOT NULL COMMENT '邮箱（唯一）',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称',
  `avatar_url` VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
  `bio` VARCHAR(500) DEFAULT NULL COMMENT '个人简介',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-正常 2-禁用 3-待激活',
  `is_deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除：0-否 1-是（软删除）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ==================== 2. 社交关系 ====================

-- 用户关注关系表（自关联多对多）
CREATE TABLE `user_follows` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关系ID',
  `follower_id` BIGINT UNSIGNED NOT NULL COMMENT '关注者ID（粉丝）',
  `following_id` BIGINT UNSIGNED NOT NULL COMMENT '被关注者ID（博主）',
  `is_mutual` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否互相关注：0-否 1-是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follower_following` (`follower_id`, `following_id`),
  KEY `idx_follower` (`follower_id`),
  KEY `idx_following` (`following_id`),
  CONSTRAINT `fk_follower` FOREIGN KEY (`follower_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_following` FOREIGN KEY (`following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注关系表';

-- ==================== 3. 内容管理 ====================

-- 版块表（分区/分类）
CREATE TABLE `sections` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '版块ID',
  `name` VARCHAR(50) NOT NULL COMMENT '版块名称',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '版块描述',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序（数字越小越靠前）',
  `is_active` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用：0-否 1-是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sort` (`sort_order`, `is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='版块表';

-- 标签表
CREATE TABLE `tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `name` VARCHAR(30) NOT NULL COMMENT '标签名称（唯一）',
  `description` VARCHAR(200) DEFAULT NULL COMMENT '标签描述',
  `usage_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '使用次数',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_usage` (`usage_count`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

-- 帖子表
CREATE TABLE `posts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '作者ID',
  `section_id` BIGINT UNSIGNED NOT NULL COMMENT '版块ID',
  `title` VARCHAR(200) NOT NULL COMMENT '标题',
  `content` TEXT NOT NULL COMMENT '内容（支持 Markdown）',
  `content_html` TEXT DEFAULT NULL COMMENT '渲染后的HTML（可选，提升查询性能）',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-草稿 2-待审核 3-已发布 4-已删除',
  `view_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
  `like_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数（冗余字段，提升查询性能）',
  `comment_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '评论数（冗余字段）',
  `is_top` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否置顶：0-否 1-是',
  `is_deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除（软删除）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_section` (`section_id`),
  KEY `idx_status` (`status`, `is_deleted`),
  KEY `idx_created` (`created_at`),
  KEY `idx_top` (`is_top`, `created_at`),
  CONSTRAINT `fk_post_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_post_section` FOREIGN KEY (`section_id`) REFERENCES `sections` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子表';

-- 帖子-标签关联表（多对多）
CREATE TABLE `post_tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `post_id` BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
  `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_tag` (`post_id`, `tag_id`),
  KEY `idx_tag` (`tag_id`),
  CONSTRAINT `fk_pt_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pt_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子标签关联表';

-- 评论表（支持多级回复）
CREATE TABLE `comments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `post_id` BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '评论者ID',
  `parent_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论ID（NULL表示一级评论）',
  `content` TEXT NOT NULL COMMENT '评论内容',
  `like_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数（冗余字段）',
  `is_deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除（软删除）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_post` (`post_id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_parent` (`parent_id`),
  KEY `idx_created` (`created_at`),
  CONSTRAINT `fk_comment_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comment_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_comment_parent` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- 点赞表（通用设计：可点赞帖子/评论）
CREATE TABLE `likes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '点赞用户ID',
  `target_type` TINYINT NOT NULL COMMENT '目标类型：1-帖子 2-评论',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '目标ID（帖子ID或评论ID）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_target` (`user_id`, `target_type`, `target_id`),
  KEY `idx_target` (`target_type`, `target_id`),
  CONSTRAINT `fk_like_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表';

-- ==================== 4. 权限管理（RBAC模型） ====================

-- 角色表
CREATE TABLE `roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `name` VARCHAR(50) NOT NULL COMMENT '角色名称（唯一）',
  `code` VARCHAR(30) NOT NULL COMMENT '角色代码（唯一，如：admin/moderator/user）',
  `description` VARCHAR(200) DEFAULT NULL COMMENT '角色描述',
  `is_system` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否系统角色：0-否 1-是（系统角色不可删除）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 权限表
CREATE TABLE `permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `name` VARCHAR(50) NOT NULL COMMENT '权限名称',
  `code` VARCHAR(50) NOT NULL COMMENT '权限代码（唯一，如：post.create/post.delete）',
  `resource` VARCHAR(50) NOT NULL COMMENT '资源（如：post/comment/user）',
  `action` VARCHAR(20) NOT NULL COMMENT '操作（如：create/read/update/delete）',
  `description` VARCHAR(200) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_resource` (`resource`, `action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 角色-权限关联表（多对多）
CREATE TABLE `role_permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
  KEY `idx_permission` (`permission_id`),
  CONSTRAINT `fk_rp_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_rp_permission` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- 用户-角色关联表（多对多，支持一个用户多个角色）
CREATE TABLE `user_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
  KEY `idx_role` (`role_id`),
  CONSTRAINT `fk_ur_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ur_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- ==================== 5. 扩展表（可选） ====================

-- 版主表（版块-用户关联，记录谁管理哪个版块）
CREATE TABLE `section_moderators` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `section_id` BIGINT UNSIGNED NOT NULL COMMENT '版块ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '版主用户ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_section_user` (`section_id`, `user_id`),
  CONSTRAINT `fk_sm_section` FOREIGN KEY (`section_id`) REFERENCES `sections` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_sm_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='版主表';

