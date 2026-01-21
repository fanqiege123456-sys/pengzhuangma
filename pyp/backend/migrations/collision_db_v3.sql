/*
 碰撞交友 V3.0 完整数据库结构
 
 包含：
 - 原有表结构（保持不变）
 - V3.0 新增表（collision_results, user_contacts, hot_tags, email_logs, system_configs）
 
 执行方式：
 mysql -u root -p collision_db < collision_db_v3.sql
 
 Date: 2025-11-30
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户表 (原有)
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `open_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `union_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `nickname` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `wechat_no` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `gender` bigint NULL DEFAULT 0 COMMENT '0:未知 1:男 2:女',
  `age` bigint NULL DEFAULT 0 COMMENT '年龄',
  `coins` bigint NULL DEFAULT 0 COMMENT '碰撞币余额',
  `total_recharge` bigint NULL DEFAULT 0 COMMENT '累计充值',
  `country` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `province` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `city` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `district` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `location_visible` tinyint(1) NULL DEFAULT 1 COMMENT '地区是否可见',
  `allow_upper_level` tinyint(1) NULL DEFAULT 0,
  `allow_passive_add` tinyint(1) NULL DEFAULT 0,
  `allow_haidilao` tinyint(1) NULL DEFAULT 0,
  `allow_force_add` tinyint(1) NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_users_open_id`(`open_id`),
  INDEX `idx_users_deleted_at`(`deleted_at`),
  INDEX `idx_users_country`(`country`),
  INDEX `idx_users_province`(`province`),
  INDEX `idx_users_city`(`city`),
  INDEX `idx_users_district`(`district`),
  INDEX `idx_users_location_district`(`country`, `province`, `city`, `district`),
  INDEX `idx_users_location_city`(`country`, `province`, `city`, `allow_upper_level`),
  INDEX `idx_users_location_province`(`country`, `province`, `allow_upper_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ----------------------------
-- 2. 管理员表 (原有)
-- ----------------------------
DROP TABLE IF EXISTS `admins`;
CREATE TABLE `admins` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `nickname` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `role` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'admin',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'active',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_admins_username`(`username`),
  INDEX `idx_admins_deleted_at`(`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- ----------------------------
-- 3. 碰撞码表 (原有，核心匹配表)
-- ----------------------------
DROP TABLE IF EXISTS `collision_codes`;
CREATE TABLE `collision_codes` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `tag` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '碰撞词/兴趣标签',
  `country` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '搜索目标-国家',
  `province` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '搜索目标-省份',
  `city` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '搜索目标-城市',
  `district` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '搜索目标-区县',
  `gender` bigint NULL DEFAULT NULL COMMENT '搜索目标-性别',
  `age_min` bigint NULL DEFAULT 20 COMMENT '搜索目标-最小年龄',
  `age_max` bigint NULL DEFAULT 30 COMMENT '搜索目标-最大年龄',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'active' COMMENT 'active/expired/invalid',
  `expires_at` datetime(3) NULL DEFAULT NULL COMMENT '过期时间',
  `cost_coins` bigint NULL DEFAULT 0 COMMENT '消耗积分',
  `match_count` bigint NULL DEFAULT 0 COMMENT '匹配次数（支持多对多）',
  `is_matched` tinyint(1) NULL DEFAULT 0 COMMENT '是否已匹配（已废弃，保留兼容）',
  PRIMARY KEY (`id`),
  INDEX `idx_collision_codes_deleted_at`(`deleted_at`),
  INDEX `idx_collision_codes_user_id`(`user_id`),
  INDEX `idx_collision_codes_tag`(`tag`),
  INDEX `idx_collision_codes_country`(`country`),
  INDEX `idx_collision_codes_province`(`province`),
  INDEX `idx_collision_codes_city`(`city`),
  INDEX `idx_collision_codes_district`(`district`),
  INDEX `idx_collision_codes_gender`(`gender`),
  INDEX `idx_collision_codes_expires_at`(`expires_at`),
  INDEX `idx_collision_codes_status`(`status`),
  CONSTRAINT `fk_collision_codes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='碰撞码表（核心匹配）';

-- ----------------------------
-- 4. 碰撞记录表 (原有，匹配成功记录)
-- ----------------------------
DROP TABLE IF EXISTS `collision_records`;
CREATE TABLE `collision_records` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id1` bigint UNSIGNED NOT NULL COMMENT '用户1',
  `user_id2` bigint UNSIGNED NOT NULL COMMENT '用户2',
  `tag` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '匹配的碰撞词',
  `match_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '匹配级别: district/city/province/country',
  `match_country` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `match_province` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `match_city` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `match_district` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'matched' COMMENT 'matched/friend_added/missed',
  `add_friend_deadline` datetime(3) NULL DEFAULT NULL COMMENT '加好友截止时间',
  PRIMARY KEY (`id`),
  INDEX `idx_collision_records_deleted_at`(`deleted_at`),
  INDEX `idx_collision_records_user_id1`(`user_id1`),
  INDEX `idx_collision_records_user_id2`(`user_id2`),
  INDEX `idx_collision_records_tag`(`tag`),
  CONSTRAINT `fk_collision_records_user1` FOREIGN KEY (`user_id1`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_collision_records_user2` FOREIGN KEY (`user_id2`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='碰撞记录表';

-- ----------------------------
-- 5. 碰撞结果表 (V3.0 新增，用于前端展示)
-- ----------------------------
DROP TABLE IF EXISTS `collision_results`;
CREATE TABLE `collision_results` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '当前用户ID',
  `matched_user_id` bigint UNSIGNED NOT NULL COMMENT '匹配到的用户ID',
  `collision_list_id` bigint UNSIGNED NULL DEFAULT 0 COMMENT '关联的碰撞列表ID（可为0）',
  `keyword` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '碰撞词',
  `matched_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '' COMMENT '匹配用户的邮箱',
  `remark` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '' COMMENT '备注（最多10字）',
  `is_known` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已知',
  `email_sent` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已发送邮件通知',
  `email_sent_at` datetime NULL DEFAULT NULL COMMENT '邮件发送时间',
  `matched_at` datetime NOT NULL COMMENT '匹配时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_collision_results_user_id`(`user_id`),
  INDEX `idx_collision_results_matched_user_id`(`matched_user_id`),
  INDEX `idx_collision_results_keyword`(`keyword`),
  INDEX `idx_collision_results_matched_at`(`matched_at`),
  UNIQUE KEY `uk_user_matched_keyword`(`user_id`, `matched_user_id`, `keyword`) COMMENT '防止重复匹配'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='碰撞结果表（V3.0前端展示用）';

-- ----------------------------
-- 6. 用户联系方式表 (V3.0 新增)
-- ----------------------------
DROP TABLE IF EXISTS `user_contacts`;
CREATE TABLE `user_contacts` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '邮箱地址',
  `email_verified` tinyint(1) NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证',
  `email_verify_code` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '邮箱验证码',
  `email_verify_expire` datetime NULL DEFAULT NULL COMMENT '验证码过期时间',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '手机号',
  `phone_verified` tinyint(1) NOT NULL DEFAULT 0 COMMENT '手机是否已验证',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user_id`(`user_id`),
  INDEX `idx_user_contacts_email`(`email`),
  INDEX `idx_user_contacts_phone`(`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户联系方式表（V3.0）';

-- ----------------------------
-- 7. 热门标签表 (V3.0 新增)
-- ----------------------------
DROP TABLE IF EXISTS `hot_tags`;
CREATE TABLE `hot_tags` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `keyword` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关键词',
  `count_24h` int NOT NULL DEFAULT 0 COMMENT '24小时搜索次数',
  `count_total` int NOT NULL DEFAULT 0 COMMENT '总搜索次数',
  `last_search_at` datetime NULL DEFAULT NULL COMMENT '最后搜索时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_keyword`(`keyword`),
  INDEX `idx_hot_tags_count_24h`(`count_24h` DESC),
  INDEX `idx_hot_tags_count_total`(`count_total` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热门标签统计表（V3.0）';

-- ----------------------------
-- 8. 邮件发送记录表 (V3.0 新增)
-- ----------------------------
DROP TABLE IF EXISTS `email_logs`;
CREATE TABLE `email_logs` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `to_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '收件人邮箱',
  `subject` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮件主题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '邮件内容',
  `type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'system' COMMENT '类型: verify/collision/system',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态: pending/sent/failed',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '错误信息',
  `sent_at` datetime NULL DEFAULT NULL COMMENT '发送时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_email_logs_user_id`(`user_id`),
  INDEX `idx_email_logs_type`(`type`),
  INDEX `idx_email_logs_status`(`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮件发送记录表（V3.0）';

-- ----------------------------
-- 9. 系统配置表 (V3.0 新增)
-- ----------------------------
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `config_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置键',
  `config_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '配置值（JSON格式）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_config_key`(`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表（V3.0）';

-- ----------------------------
-- 10. 热门关键词表 (原有，兼容保留)
-- ----------------------------
DROP TABLE IF EXISTS `hot_keywords`;
CREATE TABLE `hot_keywords` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `keyword` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'show',
  `submit_count` bigint NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_hot_keywords_deleted_at`(`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热门关键词表（兼容保留）';

-- ----------------------------
-- 11. 好友表 (原有，V3.0已废弃但保留)
-- ----------------------------
DROP TABLE IF EXISTS `friends`;
CREATE TABLE `friends` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  `friend_id` bigint UNSIGNED NOT NULL,
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'pending',
  PRIMARY KEY (`id`),
  INDEX `idx_friends_deleted_at`(`deleted_at`),
  INDEX `fk_friends_friend`(`friend_id`),
  INDEX `fk_friends_user`(`user_id`),
  CONSTRAINT `fk_friends_friend` FOREIGN KEY (`friend_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_friends_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友表（V3.0已废弃）';

-- ----------------------------
-- 12. 好友条件表 (原有)
-- ----------------------------
DROP TABLE IF EXISTS `friend_conditions`;
CREATE TABLE `friend_conditions` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  `gender` bigint NULL DEFAULT NULL,
  `min_age` bigint NULL DEFAULT NULL,
  `max_age` bigint NULL DEFAULT NULL,
  `region` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `location` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `search_code` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_friend_conditions_deleted_at`(`deleted_at`),
  INDEX `fk_friend_conditions_user`(`user_id`),
  CONSTRAINT `fk_friend_conditions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友条件表';

-- ----------------------------
-- 13. 充值记录表 (原有)
-- ----------------------------
DROP TABLE IF EXISTS `recharge_records`;
CREATE TABLE `recharge_records` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  `amount` bigint NOT NULL COMMENT '充值金额（分）',
  `coins` bigint NOT NULL COMMENT '获得积分',
  `order_no` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'pending',
  `pay_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_recharge_records_order_no`(`order_no`),
  INDEX `idx_recharge_records_deleted_at`(`deleted_at`),
  INDEX `fk_recharge_records_user`(`user_id`),
  CONSTRAINT `fk_recharge_records_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值记录表';

-- ----------------------------
-- 14. 消费记录表 (原有)
-- ----------------------------
DROP TABLE IF EXISTS `consume_records`;
CREATE TABLE `consume_records` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  `coins` bigint NOT NULL COMMENT '消费积分',
  `type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消费类型',
  `reason` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '消费原因',
  `desc` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_consume_records_deleted_at`(`deleted_at`),
  INDEX `fk_consume_records_user`(`user_id`),
  CONSTRAINT `fk_consume_records_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消费记录表';

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 升级说明（从旧版本升级时执行）
-- ============================================
-- 如果是从旧版本升级，只需执行以下语句添加新表：
-- 
-- 1. collision_results - 碰撞结果展示表
-- 2. user_contacts - 用户联系方式表  
-- 3. hot_tags - 热门标签统计表
-- 4. email_logs - 邮件发送记录表
-- 5. system_configs - 系统配置表
--
-- 以及给 collision_codes 表添加索引：
-- ALTER TABLE `collision_codes` ADD INDEX `idx_collision_codes_status`(`status`);
-- ============================================
