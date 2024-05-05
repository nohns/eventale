# Todo list

1. ~~Refactor decode encode into one single buffer~~
1. ~~Refactor conn.Listen() to be conn.Recv() and keep all handling of frame out
   of connection package~~
1. Implement AES encryption on the wire.
1. Implement authenication mechanism for clients.
1. Implement unary request-response
1. Refactor encryption so legacy (AES where key is sent to client using public
   key encryption) and TLS are valid options. TLS be the default form of
   encryption.
1. Event persistence using SQLite database.
1. Receiving events from clients.
1. Implement message queue, on which clients can listen to.
