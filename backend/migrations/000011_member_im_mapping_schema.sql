CREATE TABLE IF NOT EXISTS member_im_mapping (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  git_username VARCHAR(128) NOT NULL,
  platform VARCHAR(32) NOT NULL,
  im_user_id VARCHAR(256) NOT NULL,
  display_name VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_member_im_mapping_git_platform (git_username, platform),
  KEY idx_member_im_mapping_git_username (git_username),
  KEY idx_member_im_mapping_platform (platform),
  KEY idx_member_im_mapping_display_name (display_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
