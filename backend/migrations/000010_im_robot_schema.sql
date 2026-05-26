CREATE TABLE IF NOT EXISTS im_robot (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  platform VARCHAR(32) NOT NULL,
  name VARCHAR(128) NOT NULL,
  webhook_url VARCHAR(1024) NOT NULL,
  secret VARCHAR(512) NOT NULL DEFAULT '',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_im_robot_platform (platform),
  KEY idx_im_robot_name (name),
  KEY idx_im_robot_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
