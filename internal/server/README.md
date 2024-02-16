Package server offers the entry point of the Redis server.

Sockets are endpoints for network communication.
Key sockets concept include:
- Socket address
- Server sockets
- Client sockets
- Socket communication: Enable bidirectional data communication.

Create a server:
1. Listen for incoming connections from clients.
2. Handle client connections.
3. Data I/O
4. Respond to client.
5. Connection session done, close connection.



### Client Timeouts

> By default, Redis don't close the connection with the client if the client is idle for many seconds: the connection will remain open forever.

We only close the connection when an error occurs.
=> `Sequential Network request pattern`.

With sequential data fetching, requests in a route are dependent on each other and therefore create waterfalls. 
There may be cases where you want this pattern because one fetch depends on the result of the other, or you want a condition to be satisfied before the next fetch to save resources. 
However, this behavior can also be unintentional and lead to longer loading times.