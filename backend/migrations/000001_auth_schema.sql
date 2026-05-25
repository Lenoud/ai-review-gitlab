CREATE TABLE IF NOT EXISTS sys_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  avatar VARCHAR(512) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_sys_user_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sys_role (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_sys_role_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sys_permission (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  category VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_sys_permission_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sys_user_role (
  user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NULL,
  PRIMARY KEY (user_id, role_id),
  KEY idx_sys_user_role_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sys_role_permission (
  role_id BIGINT UNSIGNED NOT NULL,
  permission_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NULL,
  PRIMARY KEY (role_id, permission_id),
  KEY idx_sys_role_permission_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

