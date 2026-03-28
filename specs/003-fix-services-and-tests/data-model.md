# Data Model

*(No structural data model changes introduced in this testing/fix phase. Existing models for Order and Product apply.)*

## Testing Entities Context
Tests will interact with the following aggregates:
- **Order Aggregate**: State transitions during the "order creation journey".
- **Product Aggregate**: Stock level management in the inventory service.