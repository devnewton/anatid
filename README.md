# anatid

anatid is a [go](https://golang.org/) tribune proxy.

## Endpoints

Anatid exposes two endpoint for polling and posting tribune messages.

### /anatid/poll

```javascript
var ws = new WebSocket("ws://localhost:6666/anatid/poll", "anatid");
ws.onmessage = function (event) {
  console.log(event.data);
};
```

### /anatid/post

```bash
curl -X POST -d "message=plop" -d "tribune=euromussels" 'http://localhost:6666/anatid/post'
```
