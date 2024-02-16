# Write Your Own Redis Server

This is the solution for [Redis Challenge](https://codingchallenges.fyi/challenges/challenge-redis)
implemented using Go.

## About

Redis stands for Remote Dictionary Server. Similar to byte arrays, Redis strings store sequences of bytes, including
text, serialized objects, counter values and binary arrays.

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


## Idea
1. Be a copy cat. Translate the first version of Redis ( written in TCL ) to Go. Understand how redis is designed.
2. Initiate a tcp server that receives data frames from any client.
3. Decode the client payload using RESP, return error message if payload is not a valid RESP.
4. Depending on the command, check which corresponding action to perform.
5. Send data back to the client.

## Steps

1. Build functionality to serialise and de-serialise messages. The Redis server should follow the **Redis Serialisation
   Protocol (RESP)**.<br><br>
2. Create a **Light Memory-Mapped Database(LMDB)** server that listens on port <ins>6379</ins>, which is usually
   implemented as an embedded transactional key-value database. The connection uses TCP.<br><br>
3. Implement the core functions of Redis.
   <br><br>
    - [ ] Concurrent CRUD.
    - [ ] Connector command line interface ( Interactive REPL )<br><br>
    - Data types<br><br>

      |          | First Byte | 
      |:---------|:----------:|
      | Strings  |     +      |
      | Errors   |     -      |
      | Integers |     :      |
      | Arrays   |     *      |
      | Nulls    |     _      |
      | Booleans |     #      |
      | Maps     |     %      |
      | Sets     |     ~      |
      <br><br>

    - Commands
    - [ ] Check server status ( **PING** and **PONG** )
    - [ ] Store and retrieve data ( **SET** and **GET** )
    - [ ] Altering and deleting data ( **SET** and **DEL** )
    - [ ] Incrementing and decrementing stored number ( **INCR** amd **DECR** )
    - [ ] Insert all the values and the head ( **LPUSH** ) or tail(**RPUSH**) of a list.
    - [ ] Check whether a data exists ( **EXISTS** )
    - [ ] Set key expiration ( **EXPIRE KEY**, **PX**, **EAXT** and **PXAT**)
    - [ ] Scan **keyspace** to get a list of keys ( **SCAN** )
    - [ ] Check data type and existence ( **TYPE** )
    - [ ] Show help about existing commands ( **HELP** )
    - [ ] Clearing the terminal screen ( **CLEAR** )
    - [ ] Save the database state to disk. ( **SAVE** )
      <br><br>
4. Test with the official Redis Client

## References

### Tools:

### Video:

### Reads:

- [LMDB -- First version of Redis written in Tcl](https://gist.github.com/antirez/6ca04dd191bdb82aad9fb241013e88a8)
- [Reference counting: Harder than it sound](https://www.playingwithpointers.com/blog/refcounting-harder-than-it-sounds.html)
- [Introduction to Redis](https://redis.io/docs/about/)
- [RESP protocol spec](https://redis.io/docs/reference/protocol-spec/)
- [理解 Redis 的 RESP 協議](https://moelove.info/2017/03/05/理解-Redis-的-RESP-协议/)
- [How to implement instant messaging with WebSockets in Go](https://yalantis.com/blog/how-to-build-websockets-in-go/)
- [Redis file persistence](https://redis.io/docs/management/persistence/)
- [How to test TCP/UDP connection in Go](https://dev.to/williamhgough/how-to-test-tcpudp-connections-in-go---part-1-3bga)
- [Understanding the Redis protocol](https://subscription.packtpub.com/book/data/9781783988167/1/ch01lvl1sec17/understanding-the-redis-protocol)
- [Introduction to RESP](https://medium.com/@dassomnath/introduction-to-resp-redis-serialization-protocol-f3d0b8bd9cdc)
- [Data Fetching Patterns](https://nextjs.org/docs/app/building-your-application/data-fetching/patterns)
- [The Complete Guide to TCP/IP Connections in Golang](https://okanexe.medium.com/the-complete-guide-to-tcp-ip-connections-in-golang-1216dae27b5a)
- [Redis Client Handling](https://redis.io/docs/reference/clients/)
- [How to test TCP/UDP connections in Go](https://dev.to/williamhgough/how-to-test-tcpudp-connections-in-go---part-1-3bga)