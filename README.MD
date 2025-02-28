# URL Shortner

This is a simple URL shortner service built with Go language. The service provides API endpoints for shortening URLs and retrieving the original URL from the shortened URL.

## API Endpoints

### Shorten a URL

* **POST /shorten**
    + Request Body: `{"longUrl": "https://example.com/long/url"}`
    + Response: `{"shortCode": "short-code", "longUrl": "https://example.com/long/url", "createdAt": "2023-02-20T14:30:00Z"}`
    + Status Codes:
        - 201 Created: URL shortened successfully
        - 400 Bad Request: Invalid request body
        - 500 Internal Server Error: Unable to shorten URL

### Retrieve original URL from shortened URL

* **GET /shorten/{shortCode}**
    + Path Parameters: `shortCode=short-code`
    + Response: `{"shortCode": "short-code", "longUrl": "https://example.com/long/url", "createdAt": "2023-02-20T14:30:00Z"}`
    + Status Codes:
        - 200 OK: Original URL retrieved successfully
        - 404 Not Found: Shortened URL not found
        - 500 Internal Server Error: Unable to retrieve original URL

### Create a task to process all URLs in the database

* **GET /shorten**
    + Response: `{"taskId": "task-id", "status": "pending", "result": null}`
    + Status Codes:
        - 201 Created: Task created successfully
        - 500 Internal Server Error: Unable to create task

### Get the result of a task

* **GET /task/{taskId}**
    + Path Parameters: `taskId=task-id`
    + Response: `{"taskId": "task-id", "status": "completed", "result": [{"shortCode": "short-code", "longUrl": "https://example.com/long/url"}]}`
    + Status Codes:
        - 200 OK: Task result retrieved successfully
        - 404 Not Found: Task not found
        - 500 Internal Server Error: Unable to retrieve task result

## Running the Service

To run the service, execute the following commands in the root directory of the project:

1. Run the migration:
    ```sh
    make migrate
    ```

2. Start the application:
    ```sh
    make run
    ```

## Environment Variables

Ensure to set the following environment variables before running the service:

- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `PORT`: Port on which the service will run

You can set these variables in a `.env` file in the root directory of the project.

## Secrets Management

For managing secrets, you can use a secrets management tool like AWS Secrets Manager, Azure Key Vault, or HashiCorp Vault. Ensure that your application is configured to retrieve secrets from the chosen tool.