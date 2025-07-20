# Webhook Integration Guide

This guide explains how to integrate with PromptPal's webhook system to receive real-time notifications when prompts are executed.

## Overview

PromptPal sends webhook notifications to your configured endpoints when specific events occur. Currently, the system supports the `onPromptFinished` event, which is triggered whenever a prompt execution completes (successfully or with errors).

## Setting Up Webhooks

Webhooks are configured per project and must be enabled to receive notifications. Each webhook has:
- A target URL where notifications will be sent
- An event type (`onPromptFinished`)
- An enabled/disabled status

## Webhook Request Details

### HTTP Method
```
POST
```

### Headers
PromptPal sends the following headers with each webhook request:

```http
Content-Type: application/json
User-Agent: PromptPal-Webhook@{version}
```

### Timeout
Webhook requests have a **10-second timeout**. Your endpoint must respond within this timeframe to avoid timeout errors.

## Payload Structure

The webhook payload is sent as JSON in the request body with the following structure:

```json
{
  "event": "onPromptFinished",
  "projectId": 123,
  "promptId": 456,
  "userId": "user-123",
  "result": 0,
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": 1250,
  "tokens": {
    "prompt": 150,
    "completion": 300,
    "total": 450
  },
  "cached": false,
  "ip": "192.168.1.1",
  "userAgent": "Mozilla/5.0...",
  "providerId": 789,
  "providerDefaultModel": "gpt-4"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `event` | string | Always `"onPromptFinished"` for prompt completion events |
| `projectId` | number | ID of the project containing the prompt |
| `promptId` | number | ID of the executed prompt |
| `userId` | string | ID of the user who executed the prompt |
| `result` | number | Execution result: `0` for success, `1` for failure |
| `timestamp` | string | ISO 8601 timestamp when the prompt finished executing |
| `duration` | number | Execution duration in milliseconds |
| `tokens.prompt` | number | Number of tokens used in the prompt |
| `tokens.completion` | number | Number of tokens in the AI response |
| `tokens.total` | number | Total tokens used (prompt + completion) |
| `cached` | boolean | Whether the response was served from cache |
| `ip` | string | IP address of the client that executed the prompt |
| `userAgent` | string | User agent string of the client |
| `providerId` | number | ID of the AI provider used (optional) |
| `providerDefaultModel` | string | Default model of the provider (optional) |

## Expected Response

Your webhook endpoint should respond with:

### Success Response
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "received"
}
```

### Acceptable Status Codes
Any HTTP status code in the **2xx range** (200-299) is considered successful.

### Error Handling
If your endpoint returns a non-2xx status code or times out, PromptPal will:
- Log the failure for monitoring purposes
- Record the webhook call details in the database
- **Not retry** the webhook call

## Implementation Examples

### Node.js (Express)
```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.post('/webhook/promptpal', (req, res) => {
  const payload = req.body;
  
  console.log('Received webhook:', {
    event: payload.event,
    projectId: payload.projectId,
    promptId: payload.promptId,
    result: payload.result === 0 ? 'success' : 'failure',
    duration: payload.duration,
    tokens: payload.tokens.total
  });
  
  // Process the webhook data
  // ... your custom logic here ...
  
  res.status(200).json({ status: 'received' });
});

app.listen(3000);
```

### Python (Flask)
```python
from flask import Flask, request, jsonify
import json

app = Flask(__name__)

@app.route('/webhook/promptpal', methods=['POST'])
def handle_webhook():
    payload = request.get_json()
    
    print(f"Received webhook: {payload['event']}")
    print(f"Project: {payload['projectId']}, Prompt: {payload['promptId']}")
    print(f"Result: {'success' if payload['result'] == 0 else 'failure'}")
    print(f"Duration: {payload['duration']}ms, Tokens: {payload['tokens']['total']}")
    
    # Process the webhook data
    # ... your custom logic here ...
    
    return jsonify({"status": "received"}), 200

if __name__ == '__main__':
    app.run(port=3000)
```

### Go
```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type WebhookPayload struct {
    Event     string `json:"event"`
    ProjectID int    `json:"projectId"`
    PromptID  int    `json:"promptId"`
    Result    int    `json:"result"`
    Duration  int64  `json:"duration"`
    Tokens    struct {
        Total int `json:"total"`
    } `json:"tokens"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    var payload WebhookPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    fmt.Printf("Received webhook: %s\n", payload.Event)
    fmt.Printf("Project: %d, Prompt: %d\n", payload.ProjectID, payload.PromptID)
    fmt.Printf("Result: %s\n", map[int]string{0: "success", 1: "failure"}[payload.Result])
    fmt.Printf("Duration: %dms, Tokens: %d\n", payload.Duration, payload.Tokens.Total)
    
    // Process the webhook data
    // ... your custom logic here ...
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func main() {
    http.HandleFunc("/webhook/promptpal", webhookHandler)
    log.Println("Webhook server listening on :3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}
```

## Security Considerations

### HTTPS
Always use HTTPS endpoints for production webhooks to ensure data transmission security.

### Validation
Consider implementing request validation:
- Verify the request contains expected fields
- Validate data types and ranges
- Check for required fields

### Rate Limiting
Implement rate limiting on your webhook endpoint to prevent abuse.

### Logging
Log webhook receipts for monitoring and debugging purposes.

## Testing Your Webhook

1. Set up a webhook endpoint that logs incoming requests
2. Configure the webhook URL in your PromptPal project
3. Execute a prompt in your project
4. Verify that your endpoint receives the webhook notification

## Troubleshooting

### Common Issues

**Webhook not received:**
- Verify the webhook is enabled in your project settings
- Check that your endpoint URL is accessible from the internet
- Ensure your endpoint responds within 10 seconds

**Timeout errors:**
- Optimize your webhook processing to respond quickly
- Consider processing webhook data asynchronously
- Return a 200 response immediately and process data in the background

**SSL/TLS errors:**
- Ensure your HTTPS certificate is valid
- Use a trusted certificate authority
- Test your endpoint with online SSL checkers

### Webhook Call Logs

PromptPal records detailed logs of all webhook calls, including:
- Request headers and body
- Response status code, headers, and body
- Execution timing
- Error messages (if any)

These logs can help diagnose webhook delivery issues.

## Rate Limits

Currently, there are no specific rate limits for webhooks, but consider that:
- Each prompt execution can trigger multiple webhook calls (if multiple webhooks are configured)
- High-frequency prompt executions will result in high-frequency webhook calls
- Design your webhook endpoint to handle the expected load

## Support

If you encounter issues with webhook integration, check:
1. Your endpoint logs for incoming requests
2. PromptPal webhook call logs for delivery status
3. Network connectivity between PromptPal and your endpoint
4. SSL certificate validity (for HTTPS endpoints)