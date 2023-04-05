# GoChat
GoChat 是一款使用 Golang 实现的简易 IM 服务器

### 压测

名词解释：
- PV（页面浏览量）：用户每打开一个网站页面，记录一个 PV，用户多次打开同一页面，PV 值累计多次
- UV（网站独立访客）：通过互联网访问、流量网站的自然人。1天内相同访客多次访问网站，只计算为1个独立访客

压测指标：
- 压测原则：每天 80% 的访问量集中在 20% 的时间内，这 20% 的时间就叫峰值
- 公式：（总 PV * 80%）/ (86400 * 20%) = 峰值期间每秒请求数（QPS）
- 峰值时间 QPS / 单台机器 QPS = 需要的机器数量
- 举例：网站用户数 100W，每天约 3000W PV，单机承载，这台机器的性能需要多少 QPS？
> (3000 0000 * 0.8) / (86400 * 0.2) ≈ 1398 QPS
- 假设单机 QPS 为 70，需要多少台机器来支撑？
> 1398 / 70 ≈ 20

建立 websocket 连接：
1. 客户端发送 HTTP 请求，但是携带升级协议的头部信息，路由：/ws
2. 服务端接收到升级协议的请求，创建 conn 对象，创建两个 goroutine，一个负责读，一个负责写，然后返回升级响应
3. 客户端收到响应信息，成功建立 websocket 连接
```text
A                      Server
   -    HTTP upgrade   ->
   <-    Response     - 
```

客户端登录：
1. 客户端携带 `pb.Input{type: CT_Login, data: proto.Marshal(pb.LoginMsg{token:""})}` 进行登录
2. 服务端进行处理后，回复 ACK
   1. 标记 userId 到 conn 的映射
   2. Redis 记录 userId 发送消息所在 rpc 地址，跨节点连接能通过该 rpc 地址发送数据到其他节点
   3. 将该 conn 加入到 connMap 进行心跳超时检测
   4. 回复 ACK `pb.Output{type: CT_ACK, data: proto.Marshal(pb.ACKMsg{type: AT_Login})`
3. 客户端收到 ACK，暂时不进行处理

```text
A                       Server
   -       Login      ->
   <-      ACK        - 
```

心跳和超时检测：
1. 客户端间隔时间携带 `pb.Input{type: CT_Heartbeat, data: proto.Marshal(pb.HeartbeatMsg{})}` 发送
2. 服务端收到心跳，啥也不干
3. 服务端维护的 conn 在每次读数据或写数据后会更新心跳时间，所以收到心跳消息，会更新 conn 活跃时间
4. 服务端定期进行超时检测，间隔时间获取全部连接信息，检测连接是否存活，及时清除超时连接

```text
A                      Server
   -       Heartbeat   ->
```

上行（客户端发送给服务端）消息投递：  
（上行消息依靠 clientId + ACK + 超时重传实现了消息可靠性和有序性，即：不丢、不重、有序）
1. 客户端发送消息，消息格式 `pb.Input{type: CT_Message, data: proto.Marshal(pb.UpMsg{ClientId: x, Msg:proto.Marshal(pb.Message{})}}` 
2. 客户端每次发送消息，clientId++，并启动消息计时器，超时时间内未收到 ACK，再次重发消息
3. 服务端收到消息，处理后回复 ACK
   1. 当且仅当 clientID = maxClientId+1，服务端接收此消息，并更新 maxClientId++
   2. 进行相应处理
   3. 回复客户端 ACK `pb.Output{type: CT_ACK, data: proto.Marshal(pb.ACKMsg{type: AT_Up, ClientId: x, Seq: y})`
4. 客户端收到 ACK 后，取消超时重传，更新 seq（离线消息同步用到）

```text
A                       Server
   -      Message       ->
   <-       ACK         -
```

单聊、群聊消息处理：
- 单聊消息处理：获取接收者id的 seq（单调递增），并将消息存入 Message 表，进行下行消息推送
- 群聊使用写扩散，即当一个群成员发送消息，需要向所有群成员的消息列表插入一条记录（同上单聊）
  - 优点是每个用户只需要维护一个序列号(Seq)和消息列表，拉取离线消息时只需要拉取自己的消息列表即可获取全部消息
  - 缺点是每次转发时，群组有多少人，就需要插入多少条数据


下行消息投递：  
(考虑到性能问题，下行消息投递暂未实现消息可靠性和有序性)   
下行消息涉及到一个问题：A 和 Server1 进行通信，投递消息给位于 Server2 的 B 该如何实现？
1. Server1 和 Server2 启动时，启动各自的 RPC 服务，当前 Server 通过调用其他 Server 的 RPC 方法，能将消息投递到其他 Server
2. Server1 处理完 A 发送的消息，组装出下行消息：`pb.Output{type: CT_Message, data: proto.Marshal(pb.PushMsg{Msg:proto.Marshal(pb.Message{})})`
3. 消息转发流程：
   1. 根据 Redis 查询 userId 是否在线，如果不在线，不进行推送
   2. 根据 connMap 查询是否在本地，如果在本地，进行本地推送
   3. 如果不在本地，调用 RPC 方法 DeliverMessage 进行推送
```text
A                   Server1               Server2               B  
    -      Message      ->
   <-       ACK         -
                        --  DeliverMessage  >
                                             --      Message   ->
```


离线消息同步：
1. 客户端请求离线消息同步，消息格式：`pb.Input{type: CT_Sync, data: proto.Marshal(pb.SyncInputMsg{seq: x})}}}`
2. 服务端收到客户端请求，拉取该 userId 大于 seq 的消息列表前 n 条，返回：`pb.Output{type:CT_Sync, data: proto.Marshal(pb.SyncOutputMsg{Messages: "", hasMore: bool})}`
3. 客户端根据返回值 hasMore 是否为 true，更新本地 seq 后决定是否再次拉取消息
```text
A                 Server
   -    Sync       ->
   <-  返回离线消息  -
```
