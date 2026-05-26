# DTO Policy

DTOs define API request and response contracts. They are separate from GORM models.

Implemented endpoints should keep request/response DTOs explicit at the handler boundary before mapping into service inputs or model-backed response shapes.
