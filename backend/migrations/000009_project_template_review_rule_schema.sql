CREATE TABLE IF NOT EXISTS project_template_review_rule (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  template_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  glob_patterns JSON NULL,
  content TEXT NOT NULL,
  priority INT NOT NULL DEFAULT 0,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_project_template_review_rule_template_id (template_id),
  KEY idx_project_template_review_rule_priority (priority),
  KEY idx_project_template_review_rule_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
