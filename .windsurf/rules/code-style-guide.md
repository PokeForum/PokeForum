---
trigger: always_on
---

## Task Guidelines

- Project dependencies: Go + Gin + EntORM
- Project middleware dependencies: PostgreSQL + Redis
- Parameter validation implemented using github.com/go-playground/validator/v10 library
- Request/response body code must be created under the Schema directory
- Interfaces require comments compliant with gin-swagger specifications for API documentation generation
- API interfaces must adhere to REST API style
- Strictly prohibit foreign key associations in database table design; all related logic must be implemented via logical queries at the application layer
- Code comment structure rule: {English} | {Chinese}
- The tracing.WithTraceIDField method must be placed as the second argument in s.logger.\*() methods. Placement elsewhere or as the last argument is not permitted. [To standardize log formats]
