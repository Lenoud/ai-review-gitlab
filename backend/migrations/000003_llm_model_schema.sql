CREATE TABLE IF NOT EXISTS llm_model (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  provider VARCHAR(64) NOT NULL,
  model_code VARCHAR(128) NOT NULL,
  api_base_url VARCHAR(512) NOT NULL,
  api_key VARCHAR(512) NOT NULL,
  max_tokens INT NOT NULL DEFAULT 4096,
  is_default TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_llm_model_provider (provider),
  KEY idx_llm_model_model_code (model_code),
  KEY idx_llm_model_is_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

