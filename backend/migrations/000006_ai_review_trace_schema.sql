CREATE TABLE IF NOT EXISTS ai_review_trace (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  review_event_type VARCHAR(32) NOT NULL,
  review_event_id BIGINT UNSIGNED NOT NULL,
  prompt LONGTEXT NULL,
  response LONGTEXT NULL,
  provider VARCHAR(64) NOT NULL,
  model_code VARCHAR(128) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_ai_review_trace_event (review_event_type, review_event_id),
  KEY idx_ai_review_trace_review_event_type (review_event_type),
  KEY idx_ai_review_trace_review_event_id (review_event_id),
  KEY idx_ai_review_trace_provider (provider),
  KEY idx_ai_review_trace_model_code (model_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
