# GoChat
GoChat 是一款使用 Golang 实现的简易 IM 服务器


上行消息一致性：
1. 客户端和服务端建立长连接，客户端初始化 clientID = 0
2. 客户端每次发送消息，clientID++，并启动消息计时器，等待 ACK 回复，或者超时后重传
3. 服务端收到消息，验证收到的 clientID = MaxClient + 1, 生成 SeqID（TODO）后，将消息丢入 MQ 后，回复客户端 ACK
4. 客户端收到 ACK 后，取消超时重传

目标：
实现离线消息同步

```text
A                Server                 B
   -     Login    ->
   <-    ACK   -    
   
   -     SYNC     ->
   <-  返回离线消息  -
   -      ACK      ->
   
   -    HeartBeat ->
   
   -   Message    ->
   <-    ACK   -    
                      -   Message   ->
                      <-   ACK      -
```

1. 客户端本地维护 seq，可能登录时 seq = 5
2. 客户端进行登录，暂且不讨论
3. 离线同步
   1. 客户端携带 seq 进行离线同步
   2. 服务端返回所有大于 seq 的消息列表
4. 客户端间隔时间发送心跳，暂且不讨论
5. 单聊
   1. 客户端 A 携带 clientId = x 发送消息给 B，clientId++，并启动消息计时器，等待 ACK 回复，或者超时重传
   2. 服务端收到消息，验证收到的 clientID = maxClientId+1，才继续进行，获取 B 的下一个 SeqId（Redis存seq：key:userId,value:seqId，lua），连同消息内容存入 Message，丢入 Mq 后，回复客户端 A ACK
   3. 在服务端 MQ 的消费者处理中，先对 Message 进行落库，然后查询 B 是否在线，如果在线，进行推送，并启动超时重传（TODO），如果不在线，啥也不干
   4. 客户端 A 收到上行消息的 ACK，取消超时重传
   5. 客户端 B 收到下行消息，从下行消息中获取 seq 更新本地 seq，并返回给服务端 ACK(TODO)
6. 群聊