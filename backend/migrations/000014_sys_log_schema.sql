CREATE TABLE IF NOT EXISTS sys_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    level VARCHAR(32),
    module VARCHAR(64),
    action VARCHAR(128),
    message VARCHAR(255),
    detail TEXT,
    error_stack MEDIUMTEXT,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    INDEX idx_sys_log_level (level),
    INDEX idx_sys_log_module (module),
    INDEX idx_sys_log_action (action),
    INDEX idx_sys_log_created_at (created_at)
);
