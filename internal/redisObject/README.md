Package Object defines an object struct that represent all basic data types.

### RESP

To communicate with the Redis server, Redis client uses a protocol called **REdis Serialization Protocol**.
In RESP, the first byte of the data determines its type.
<br><br>
New RESP connections should begin the session by calling the HELLO command.
The HELLO command returns information about the server and the protocol that the client can use for different goals.
A successful reply to the HELLO command is a map reply.
The information in the reply is partly server-dependent, but certain fields are mandatory for all the RESP3 implementations:

- server: "redis" (or other software name).
- version: the server's version.
- proto: the highest supported version of the RESP protocol.

For example, if we were to send `PING` to the server from the client side, `PING` has to be encoded into

> `*1\r\n\$4\r\nPING\r\n`
> - 
> - '*' Means that we received an array.
> - 1 stands for the size of the array.
> - \r\n (CRLF) is the terminator of each part in RESP.
> - The backslash before $4 is the escape character for the $ sign.
> - $4 tells you that the following is a bulk string type of four characters long.
> - PING is the string itself.

> - `+PONG` should be the string the PING command returned. The plus sign tells you that it's a simple string type.
