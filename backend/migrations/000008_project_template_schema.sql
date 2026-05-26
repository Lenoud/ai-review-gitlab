CREATE TABLE IF NOT EXISTS project_template (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(1024) NOT NULL DEFAULT '',
  extensions JSON NULL,
  review_prompt_template TEXT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_project_template_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
