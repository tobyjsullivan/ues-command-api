# UES Command API

This service is the Command API for a Universal Event Store. It
handles all state-mutating (write) requests for the service. 
Read-only operations are performed via a separate Query API.

## Running with Docker

TK

## API

This service exposes a simple HTTP API.

### Request and Response Formats

The API tries to follow REST-inspired practices; however, is not itself a RESTful API. 

#### Authorization

Some commands require authorization (as noted in the 
documentation for the endpoint). This is presently facilitated 
via a Bearer token.

Include the following header in all requests requiring 
Authorization (replace the bearer token with your own).

`Authorization: Bearer yourAuthToken`

#### POST request format

All parameters are sent as URL Form Encoded key-value pairs. As 
such, a `Content-Type` header of 
`application/x-www-form-urlencoded` is expected.

#### Response format

All endpoints respond with JSON content. The following example 
includes all possible root elements that may be included in a 
response; however, any or all may be absent from any particular 
response.

```json
{
  "payload": ...,
  "pagination": {
    "nextOffset": ""
  },
  "error": {
    "message": "A human readable error message."
  }
}
```

#### Status Codes

Each endpoint will define its own set of possible status codes depending on the nature of the command. Additionally, the following may occur from any request.

- `401 UNAUTHORIZED` will result from any attempt to make a request against an authenticated endpoint without providing a valid authentication code.
- `404 NOT_FOUND` The command you have requested does not exist.
- `500 INTERNAL_SERVER_ERROR` The service experienced an unexpected error while processing the request. Perhaps retry in the future. Definitely file a bug report if you can reproduce.
- Other status codes may occur. Absense from the documentation doesn't neccessarily mean it can never happen - it will just likely result from external code we rely on rather than intentionally coded logic.

### GET /commands

Returns a list of available commands which the API supports.

Example response:

```json
{
  "payload": [
    {
      "label": "View all commands",
      "path": "/commands",
      "method": "GET",
      "params": []
    },
    {
      "label": "Create an account",
      "path": "/commands/create-account",
      "method": "POST",
      "params": [
        {
          "key": "email",
          "valueType": "string"
        },
        {
          "key": "password",
          "valueType": "string"
        }
      ]
    }
  ]
}
```

### POST /commands/create-account

Authentication: NONE

This endpoint creates a new account.

#### Request Params

| Param | Required | Description |
|---|---|---|
| Email | true | The email address to be associated with the account. |
| Password | true | The password to be associated with the account. |

#### Status Codes

- `202 ACCEPTED` The command passed validation and has been accepted. This most likely means the account will be created.
- `400 BAD_REQUEST` There is something wrong with your request such as a missing field or badly formatted value. Check `error` in the response for specific error information.
- `409 CONFLICT` This account would conflict with another. Most likely the email address is already in use by another account.

### POST /commands/create-store

Authentication: Account Token

Params:

- `name` The DNS-friendly name for the store

### POST /commands/commit-event

Authentication: Store Write Token

Params:

- `entityId` The entity which the event applies to. Must be a 128 bit value represented as a UUID (e.g., `cc0ed51b-2ae3-4608-a47d-fdb5d040b848`).
- `type` The event type identifier such as "AccountOpened".
- `data` The JSON event data encoded with Base64.
- `version` (optional) The version number of the event. The request will be rejected if this is not the next available version in the entity log.
