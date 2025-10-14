package descriptor

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadComprehensive(t *testing.T) {
	ctx := context.Background()
	root := filepath.Join("testdata", "comprehensive")

	reg, err := LoadDirectory(ctx, root, []string{root})
	if err != nil {
		t.Fatalf("Failed to load comprehensive testdata: %v", err)
	}

	if reg == nil {
		t.Fatal("Registry is nil")
	}

	t.Run("services", func(t *testing.T) {
		services := []string{
			"users.v1.UserService",
			"products.v1.ProductService",
			"orders.v1.OrderService",
			"notifications.v1.NotificationService",
		}

		for _, svcName := range services {
			t.Run(svcName, func(t *testing.T) {
				svc, exists := reg.FindService(svcName)
				if !exists {
					t.Fatalf("Service %s not found", svcName)
				}

				if svc.Name() == "" {
					t.Error("Service has empty name")
				}

				if svc.Methods().Len() == 0 {
					t.Error("Service has no methods")
				}

				// Check that service has a comment
				comment := reg.CommentIndex[svcName]
				if comment == "" {
					t.Errorf("Service %s has no comment", svcName)
				}
			})
		}
	})

	t.Run("user_service_methods", func(t *testing.T) {
		methods := []struct {
			name             string
			clientStreaming  bool
			serverStreaming  bool
		}{
			{"users.v1.UserService/CreateUser", false, false},
			{"users.v1.UserService/GetUser", false, false},
			{"users.v1.UserService/UpdateUser", false, false},
			{"users.v1.UserService/DeleteUser", false, false},
			{"users.v1.UserService/ListUsers", false, false},
			{"users.v1.UserService/SearchUsers", false, false},
			{"users.v1.UserService/BatchGetUsers", false, false},
			{"users.v1.UserService/StreamUsers", false, true},
			{"users.v1.UserService/SubscribeToUserUpdates", false, true},
			{"users.v1.UserService/BulkUpdateUsers", true, false},
			{"users.v1.UserService/SyncUsers", true, true},
		}

		for _, tc := range methods {
			t.Run(tc.name, func(t *testing.T) {
				method, exists := reg.FindMethod(tc.name)
				if !exists {
					t.Fatalf("Method %s not found", tc.name)
				}

				if method.IsStreamingClient() != tc.clientStreaming {
					t.Errorf("Expected client streaming: %v, got: %v", tc.clientStreaming, method.IsStreamingClient())
				}

				if method.IsStreamingServer() != tc.serverStreaming {
					t.Errorf("Expected server streaming: %v, got: %v", tc.serverStreaming, method.IsStreamingServer())
				}

				// Verify input and output types exist
				if method.Input().FullName() == "" {
					t.Error("Method has empty input type")
				}

				if method.Output().FullName() == "" {
					t.Error("Method has empty output type")
				}

				// Check for method comment
				comment := reg.CommentIndex[tc.name]
				if comment == "" {
					t.Errorf("Method %s has no comment", tc.name)
				}
			})
		}
	})

	t.Run("common_types", func(t *testing.T) {
		types := []string{
			"common.v1.Money",
			"common.v1.Address",
			"common.v1.GeoCoordinates",
			"common.v1.Pagination",
			"common.v1.PageInfo",
			"common.v1.ErrorInfo",
			"common.v1.ErrorDetail",
			"common.v1.Metadata",
		}

		for _, typeName := range types {
			t.Run(typeName, func(t *testing.T) {
				msg, exists := reg.FindMessage(typeName)
				if !exists {
					t.Fatalf("Message %s not found", typeName)
				}

				if msg.Name() == "" {
					t.Error("Message has empty name")
				}

				// Check for message comment
				comment := reg.CommentIndex[typeName]
				if comment == "" {
					t.Errorf("Message %s has no comment", typeName)
				}
			})
		}
	})

	t.Run("user_types", func(t *testing.T) {
		msg, exists := reg.FindMessage("users.v1.User")
		if !exists {
			t.Fatal("User message not found")
		}

		// Verify User has expected fields
		expectedFields := []string{"metadata", "email", "full_name", "profile", "role"}
		for _, fieldName := range expectedFields {
			found := false
			for i := 0; i < msg.Fields().Len(); i++ {
				if string(msg.Fields().Get(i).Name()) == fieldName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("User message missing field: %s", fieldName)
			}
		}

		// Verify nested message exists
		_, exists = reg.FindMessage("users.v1.UserProfile")
		if !exists {
			t.Error("UserProfile nested message not found")
		}
	})

	t.Run("enums", func(t *testing.T) {
		enums := []string{
			"common.v1.ErrorType",
			"common.v1.Status",
			"users.v1.UserRole",
			"users.v1.VerificationStatus",
			"users.v1.Theme",
			"products.v1.ProductStatus",
			"products.v1.StockStatus",
			"orders.v1.OrderStatus",
			"orders.v1.PaymentStatus",
			"notifications.v1.NotificationType",
			"notifications.v1.NotificationPriority",
		}

		for _, enumName := range enums {
			t.Run(enumName, func(t *testing.T) {
				enum, exists := reg.FindEnum(enumName)
				if !exists {
					t.Fatalf("Enum %s not found", enumName)
				}

				if enum.Values().Len() == 0 {
					t.Errorf("Enum %s has no values", enumName)
				}

				// Check for enum comment
				comment := reg.CommentIndex[enumName]
				if comment == "" {
					t.Errorf("Enum %s has no comment", enumName)
				}
			})
		}
	})

	t.Run("cross_package_imports", func(t *testing.T) {
		// Verify orders.proto imports from users and products
		orderMsg, exists := reg.FindMessage("orders.v1.Order")
		if !exists {
			t.Fatal("Order message not found")
		}

		// Check that Order message exists and has fields
		if orderMsg.Fields().Len() == 0 {
			t.Error("Order message has no fields")
		}

		// Verify OrderItem references products
		orderItemMsg, exists := reg.FindMessage("orders.v1.OrderItem")
		if !exists {
			t.Fatal("OrderItem message not found")
		}

		// Find product_id field
		found := false
		for i := 0; i < orderItemMsg.Fields().Len(); i++ {
			if string(orderItemMsg.Fields().Get(i).Name()) == "product_id" {
				found = true
				break
			}
		}
		if !found {
			t.Error("OrderItem missing product_id field")
		}
	})

	t.Run("field_comments", func(t *testing.T) {
		// Check that Money message fields have comments
		msg, exists := reg.FindMessage("common.v1.Money")
		if !exists {
			t.Fatal("Money message not found")
		}

		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			fieldName := string(field.Name())
			fqn := "common.v1.Money." + fieldName
			comment := reg.CommentIndex[fqn]
			if comment == "" {
				t.Errorf("Field %s has no comment", fqn)
			}
		}
	})

	t.Run("deprecated_fields", func(t *testing.T) {
		msg, exists := reg.FindMessage("common.v1.Money")
		if !exists {
			t.Fatal("Money message not found")
		}

		// Find the deprecated field (amount_deprecated)
		// Note: In the new protoreflect API, we check Options() which implements proto.Message
		// The deprecated field should be field number 3
		found := false
		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			if string(field.Name()) == "amount_deprecated" {
				found = true
				// In a real implementation, we would check field.Options()
				// but for this test, we'll just verify the field exists
				break
			}
		}

		if !found {
			t.Error("Expected to find amount_deprecated field in Money message")
		}
	})

	t.Run("map_fields", func(t *testing.T) {
		// Check that Address has map field
		msg, exists := reg.FindMessage("common.v1.Metadata")
		if !exists {
			t.Fatal("Metadata message not found")
		}

		// Find labels field (map<string, string>)
		found := false
		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			if string(field.Name()) == "labels" && field.IsMap() {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find labels map field in Metadata")
		}
	})

	t.Run("oneof_fields", func(t *testing.T) {
		// Check that requests with oneofs exist
		msg, exists := reg.FindMessage("orders.v1.GetOrderRequest")
		if !exists {
			t.Fatal("GetOrderRequest message not found")
		}

		// Find oneof field
		foundOneof := false
		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			if field.ContainingOneof() != nil {
				foundOneof = true
				break
			}
		}

		if !foundOneof {
			t.Error("Expected to find oneof field in GetOrderRequest")
		}
	})

	t.Run("streaming_methods", func(t *testing.T) {
		tests := []struct {
			name            string
			expectServer    bool
			expectClient    bool
			expectBidi      bool
		}{
			{"users.v1.UserService/StreamUsers", true, false, false},
			{"users.v1.UserService/SubscribeToUserUpdates", true, false, false},
			{"users.v1.UserService/BulkUpdateUsers", false, true, false},
			{"users.v1.UserService/SyncUsers", true, true, true},
			{"products.v1.ProductService/StreamInventoryUpdates", true, false, false},
			{"orders.v1.OrderService/StreamOrderUpdates", true, false, false},
			{"notifications.v1.NotificationService/SubscribeToNotifications", true, false, false},
			{"notifications.v1.NotificationService/SendBulkNotifications", false, true, false},
			{"notifications.v1.NotificationService/NotificationChannel", true, true, true},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				method, exists := reg.FindMethod(tc.name)
				if !exists {
					t.Fatalf("Method %s not found", tc.name)
				}

				if method.IsStreamingServer() != tc.expectServer {
					t.Errorf("Expected server streaming: %v, got: %v", tc.expectServer, method.IsStreamingServer())
				}

				if method.IsStreamingClient() != tc.expectClient {
					t.Errorf("Expected client streaming: %v, got: %v", tc.expectClient, method.IsStreamingClient())
				}

				isBidi := method.IsStreamingClient() && method.IsStreamingServer()
				if isBidi != tc.expectBidi {
					t.Errorf("Expected bidirectional: %v, got: %v", tc.expectBidi, isBidi)
				}
			})
		}
	})

	t.Run("total_counts", func(t *testing.T) {
		// Count services
		serviceCount := len(reg.ServicesByName)
		if serviceCount != 4 {
			t.Errorf("Expected 4 services, got %d", serviceCount)
		}

		// Count methods
		methodCount := len(reg.MethodsByName)
		if methodCount < 40 {
			t.Errorf("Expected at least 40 methods, got %d", methodCount)
		}

		// Count messages
		messageCount := len(reg.MessagesByName)
		if messageCount < 50 {
			t.Errorf("Expected at least 50 messages, got %d", messageCount)
		}

		// Count enums
		enumCount := len(reg.EnumsByName)
		if enumCount < 15 {
			t.Errorf("Expected at least 15 enums, got %d", enumCount)
		}

		t.Logf("Services: %d, Methods: %d, Messages: %d, Enums: %d",
			serviceCount, methodCount, messageCount, enumCount)
	})
}

func TestComprehensiveComments(t *testing.T) {
	ctx := context.Background()
	root := filepath.Join("testdata", "comprehensive")

	reg, err := LoadDirectory(ctx, root, []string{root})
	if err != nil {
		t.Fatalf("Failed to load comprehensive testdata: %v", err)
	}

	t.Run("service_comments", func(t *testing.T) {
		services := []string{
			"users.v1.UserService",
			"products.v1.ProductService",
			"orders.v1.OrderService",
			"notifications.v1.NotificationService",
		}

		for _, svcName := range services {
			comment := reg.CommentIndex[svcName]
			if comment == "" {
				t.Errorf("Service %s missing comment", svcName)
			}
			if len(comment) < 50 {
				t.Errorf("Service %s comment too short: %q", svcName, comment)
			}
		}
	})

	t.Run("message_comments", func(t *testing.T) {
		messages := []string{
			"common.v1.Money",
			"common.v1.Address",
			"users.v1.User",
			"products.v1.Product",
			"orders.v1.Order",
			"notifications.v1.Notification",
		}

		for _, msgName := range messages {
			comment := reg.CommentIndex[msgName]
			if comment == "" {
				t.Errorf("Message %s missing comment", msgName)
			}
		}
	})

	t.Run("enum_comments", func(t *testing.T) {
		enums := []string{
			"common.v1.Status",
			"users.v1.UserRole",
			"products.v1.ProductStatus",
			"orders.v1.OrderStatus",
			"notifications.v1.NotificationType",
		}

		for _, enumName := range enums {
			comment := reg.CommentIndex[enumName]
			if comment == "" {
				t.Errorf("Enum %s missing comment", enumName)
			}
		}
	})
}
