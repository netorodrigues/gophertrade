# Data Model: Backend Order and Stock Management

## Overview

Entities are split between two bounded contexts: `Inventory` (Product) and `Order`.

## Inventory Context

### Product Entity

**Table**: `products`

| Field | Type | Description | Constraints |
|---|---|---|---|
| id | UUID | Primary Key | Not Null |
| name | String | Product Name | Not Null |
| price | Integer | Price in Cents | Not Null, >= 0 |
| stock_quantity | Integer | Available inventory | Not Null, >= 0 |
| version | Integer | Optimistic Locking | Not Null, Default 1 |
| created_at | Timestamp | Creation Time | Not Null |
| updated_at | Timestamp | Last Update | Not Null |

**Repository Interfaces**:
- `GetByID(ctx, id) -> Product`
- `UpdateStock(ctx, id, delta, version) -> Error` (Fails if version mismatch or result < 0)

## Order Context

### Order Entity

**Table**: `orders`

| Field | Type | Description | Constraints |
|---|---|---|---|
| id | UUID | Primary Key | Not Null |
| created_at | Timestamp | Order Date | Not Null |
| total_price | Integer | Sum of items (cents) | Not Null, >= 0 |
| status | Enum | Order Status | CREATED, CANCELLED |

### OrderItem Entity

**Table**: `order_items`

| Field | Type | Description | Constraints |
|---|---|---|---|
| order_id | UUID | Foreign Key (Order) | Not Null |
| product_id | UUID | Reference to Product | Not Null |
| quantity | Integer | Quantity ordered | Not Null, > 0 |
| unit_price | Integer | Snapshot price (cents) | Not Null, >= 0 |

**Repository Interfaces**:
- `Create(ctx, Order, []OrderItem) -> Error`
- `GetByID(ctx, id) -> (Order, []OrderItem)`
- `List(ctx, filter) -> []Order`

## Read Models (NoSQL)

### Firestore Structure

**Collection**: `products_view`
- Document ID: `product_uuid`
- Fields: `name`, `price`, `stock_quantity`, `last_updated`

**Collection**: `orders_view`
- Document ID: `order_uuid`
- Fields: `total_price`, `status`, `created_at`, `items: [{product_id, quantity, unit_price}]`

### ElasticSearch Documents

**Index**: `products`
- Fields: `name` (text, searchable), `description` (text), `price` (keyword)

**Index**: `orders`
- Fields: `status` (keyword), `created_at` (date), `total_price` (integer)
