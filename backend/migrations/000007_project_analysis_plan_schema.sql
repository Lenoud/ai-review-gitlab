CREATE TABLE IF NOT EXISTS project_analysis_plan (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  project_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL,
  prompt TEXT,
  cron_expression VARCHAR(128),
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  im_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  im_robot_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  html_report_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_project_analysis_plan_project_id (project_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS project_analysis_plan_execution_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plan_id BIGINT UNSIGNED NULL,
  project_id BIGINT UNSIGNED NULL,
  status VARCHAR(32) NOT NULL,
  started_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  result_content TEXT,
  result_actions TEXT,
  share_token VARCHAR(128),
  share_token_expires_at BIGINT NOT NULL DEFAULT 0,
  error_message TEXT,
  error_stack TEXT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_project_analysis_execution_plan_id (plan_id),
  KEY idx_project_analysis_execution_project_id (project_id),
  KEY idx_project_analysis_execution_status (status),
  KEY idx_project_analysis_execution_share_token (share_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
