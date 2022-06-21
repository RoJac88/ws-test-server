const ws = new WebSocket("ws://localhost:8080/")

ws.addEventListener('open', () => console.log("connected to server"))
ws.addEventListener('message', (event) => console.log(`server says: ${event.data}`))
ws.addEventListener('close'), () => console.log("Connection lost")
