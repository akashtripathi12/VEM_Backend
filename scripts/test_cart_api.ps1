# Cart System Test Script
# Run this after the server is running and database is seeded

# Variables
$baseUrl = "http://localhost:8080/api/v1"
$eventId = "YOUR_EVENT_ID_HERE"  # Replace with actual event ID from your database
$token = "YOUR_JWT_TOKEN_HERE"   # Replace with actual JWT token from login

Write-Host "=== TBO Cart System API Tests ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Add Room to Cart
Write-Host "Test 1: Adding room to cart (wishlist)..." -ForegroundColor Yellow
$addRoomBody = @{
    type = "room"
    refId = "YOUR_ROOM_OFFER_ID"
    quantity = 2
    notes = "Ocean view preferred"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart" `
        -Method POST `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        } `
        -Body $addRoomBody
    
    Write-Host "✓ Room added successfully" -ForegroundColor Green
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ Failed to add room: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 2: Add Banquet to Cart
Write-Host "Test 2: Adding banquet to cart..." -ForegroundColor Yellow
$addBanquetBody = @{
    type = "banquet"
    refId = "YOUR_BANQUET_ID"
    quantity = 1
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart" `
        -Method POST `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        } `
        -Body $addBanquetBody
    
    Write-Host "✓ Banquet added successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to add banquet: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: Get Cart (Wishlist) - Hierarchical Response
Write-Host "Test 3: Getting cart with hierarchical grouping..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart?status=wishlist" `
        -Method GET `
        -Headers @{
            "Authorization" = "Bearer $token"
        }
    
    Write-Host "✓ Cart fetched successfully" -ForegroundColor Green
    Write-Host "Response structure (should be grouped by hotel):" -ForegroundColor Cyan
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ Failed to get cart: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 4: Update Cart Item
Write-Host "Test 4: Updating cart item quantity..." -ForegroundColor Yellow
$cartItemId = "YOUR_CART_ITEM_ID"  # Replace with actual cart item ID
$updateBody = @{
    quantity = 3
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart/$cartItemId" `
        -Method PATCH `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        } `
        -Body $updateBody
    
    Write-Host "✓ Cart item updated successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to update cart item: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 5: Approve Wishlist (Convert to Final Cart)
Write-Host "Test 5: Approving wishlist (convert to final cart)..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart/approve" `
        -Method POST `
        -Headers @{
            "Authorization" = "Bearer $token"
        }
    
    Write-Host "✓ Wishlist approved successfully" -ForegroundColor Green
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ Failed to approve wishlist: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 6: Get Approved Cart
Write-Host "Test 6: Getting approved cart..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart?status=approved" `
        -Method GET `
        -Headers @{
            "Authorization" = "Bearer $token"
        }
    
    Write-Host "✓ Approved cart fetched successfully" -ForegroundColor Green
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ Failed to get approved cart: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 7: Remove from Cart
Write-Host "Test 7: Removing item from cart..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest `
        -Uri "$baseUrl/events/$eventId/cart/$cartItemId" `
        -Method DELETE `
        -Headers @{
            "Authorization" = "Bearer $token"
        }
    
    Write-Host "✓ Cart item removed successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to remove cart item: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Tests Complete ===" -ForegroundColor Cyan
