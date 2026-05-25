# DTO Policy

DTOs define API request and response contracts. They are separate from GORM models.

In the API contract skeleton phase, handlers return `501 Not Implemented` and do not require full DTO coverage yet. When an endpoint is implemented, add its request/response DTOs in this package before writing service code.
