# anatid

anatid is a [node-red](https://nodered.org) tribune proxy.

![Screenshot](screenshot.png)

## Endpoints

Anatid exposes two endpoint for polling and posting tribune messages.

### /anatid/poll

```javascript
var ws = new WebSocket("ws://localhost:8000/anatid/poll", "anatid");
ws.onmessage = function (event) {
  console.log(event.data);
};
```

### /anatid/post

```bash
curl -X POST -d "message=plop" -d "tribune=euromussels" 'http://localhost:8000/anatid/post'
```