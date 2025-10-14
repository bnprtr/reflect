# Comprehensive Test Data

This directory contains a comprehensive set of protobuf files designed to test all aspects of the Reflect documentation system.

## Structure

```
comprehensive/
├── common/          # Shared types used across all services
├── users/           # User management service
├── products/        # Product catalog and inventory service
├── orders/          # Order management and fulfillment service
└── notifications/   # Notification delivery service
```

## Features Tested

### Proto Features
- ✅ Multiple packages with cross-package imports
- ✅ Well-known types (google.protobuf.Timestamp, FieldMask)
- ✅ Nested messages and enums
- ✅ Maps (string -> string, etc.)
- ✅ Repeated fields
- ✅ Oneofs
- ✅ Optional fields
- ✅ Deprecated fields
- ✅ Enums with detailed documentation

### gRPC Service Patterns
- ✅ Unary RPC (standard request-response)
- ✅ Server streaming RPC (stream responses)
- ✅ Client streaming RPC (stream requests)
- ✅ Bidirectional streaming RPC

### Documentation Features
- ✅ Comprehensive leading comments on all types
- ✅ Field-level documentation
- ✅ Enum value documentation
- ✅ Multi-paragraph descriptions
- ✅ Usage notes and examples in comments
- ✅ Deprecation warnings

### Real-World Patterns
- ✅ Pagination with page tokens
- ✅ Field masks for partial updates
- ✅ Metadata and versioning patterns
- ✅ Error handling with detailed error types
- ✅ Money and currency handling
- ✅ Address and geolocation types
- ✅ Status and state machines
- ✅ Batch operations
- ✅ Search and filtering
- ✅ Real-time subscriptions

## Services

### 1. Common Types (`common/types.proto`)
Shared foundational types used across all services:
- **Money**: Currency handling with ISO 4217 codes
- **Address**: Physical addresses with geo-coordinates
- **Pagination**: Standard pagination patterns
- **ErrorInfo**: Comprehensive error handling
- **Metadata**: Common resource metadata (ID, timestamps, versions, labels)
- **Status**: Generic status enum

### 2. User Service (`users/users.proto`)
Complete user management with 11 RPC methods:
- CRUD operations (Create, Read, Update, Delete)
- List and search with pagination
- Batch operations
- Server streaming (StreamUsers)
- Subscription to real-time updates (SubscribeToUserUpdates)
- Client streaming bulk updates (BulkUpdateUsers)
- Bidirectional streaming sync (SyncUsers)

**Key features:**
- User profiles with rich metadata
- Role-based access control
- Notification preferences
- Privacy settings
- Social links
- Verification status tracking

### 3. Product Service (`products/products.proto`)
Product catalog and inventory management with 9 RPC methods:
- CRUD operations for products
- Advanced search with price range filtering
- Inventory management with reason tracking
- Real-time inventory updates streaming

**Key features:**
- SKU-based product identification
- Category hierarchy
- Pricing with sales and tax information
- Physical dimensions and weight
- Multiple product images
- Attribute maps (color, size, etc.)
- Stock status tracking
- Warehouse location management

### 4. Order Service (`orders/orders.proto`)
Complete order lifecycle management with 9 RPC methods:
- Order creation and management
- Payment processing with multiple methods
- Order status updates with history tracking
- Shipping and tracking integration
- Real-time order update streaming

**Key features:**
- Complex order structure with line items
- Order totals calculation
- Multiple payment methods (cards, PayPal, Apple/Google Pay)
- Payment status tracking (authorized, captured, refunded)
- Fulfillment workflow (picking, packing, shipping, delivery)
- Shipping carrier integration
- Order history audit trail

### 5. Notification Service (`notifications/notifications.proto`)
Multi-channel notification system with 11 RPC methods:
- Single and batch notification sending
- User preference management
- Real-time notification streaming
- Client streaming bulk send
- Bidirectional notification channel

**Key features:**
- Multiple channels (in-app, email, push, SMS, webhook)
- Rich notifications with actions and images
- Priority levels (low, normal, high, urgent)
- Delivery tracking per channel
- User preferences and quiet hours
- Digest settings
- Notification grouping
- Retry logic for failed deliveries

## Import Graph

```
users.proto ─────┐
                 ├──→ common/types.proto
products.proto ──┤
                 │
orders.proto ────┼──→ users.proto
                 │    products.proto
                 │
notifications.proto ─┘
```

## Testing Scenarios

### Basic Functionality
1. Load all proto files from this directory
2. Verify all services are discovered
3. Verify all methods are indexed correctly
4. Verify all messages and enums are indexed
5. Verify imports are resolved correctly

### Comment Extraction
1. Service comments are extracted
2. Method comments are extracted
3. Message comments are extracted
4. Field comments are extracted
5. Enum comments are extracted
6. Enum value comments are extracted
7. Multi-line comments are preserved

### Type Resolution
1. Cross-package references resolve correctly
2. Nested message types resolve correctly
3. Map types are formatted correctly
4. Repeated fields are identified
5. Oneofs are identified
6. Well-known types are recognized

### Edge Cases
1. Deprecated fields are marked
2. Optional fields are identified
3. Empty messages are handled
4. Services with many methods render correctly
5. Deep nesting (messages within messages) works
6. Large enum lists render correctly

## Usage in Tests

```go
func TestLoadComprehensive(t *testing.T) {
    ctx := context.Background()
    root := "testdata/comprehensive"

    reg, err := descriptor.LoadDirectory(ctx, root, []string{root})
    if err != nil {
        t.Fatalf("Failed to load comprehensive testdata: %v", err)
    }

    // Verify services
    services := []string{
        "users.v1.UserService",
        "products.v1.ProductService",
        "orders.v1.OrderService",
        "notifications.v1.NotificationService",
    }

    for _, svc := range services {
        if _, exists := reg.FindService(svc); !exists {
            t.Errorf("Service %s not found", svc)
        }
    }

    // Verify cross-package imports
    if _, exists := reg.FindMessage("common.v1.Money"); !exists {
        t.Error("Common type Money not found")
    }

    // Verify streaming methods
    method, exists := reg.FindMethod("users.v1.UserService/StreamUsers")
    if !exists {
        t.Error("Streaming method not found")
    }
    if !method.IsStreamingServer() {
        t.Error("Expected server streaming method")
    }
}
```

## Statistics

- **Services**: 4
- **Total RPC Methods**: 40
- **Unary Methods**: 27
- **Server Streaming Methods**: 5
- **Client Streaming Methods**: 2
- **Bidirectional Streaming Methods**: 2
- **Messages**: ~100
- **Enums**: ~20
- **Lines of Code**: ~2,500
- **Comment Lines**: ~1,000
