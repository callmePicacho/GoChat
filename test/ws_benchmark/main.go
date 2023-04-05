package main

import (
	"flag"
)

/*
------ 目标：单机群聊压测 ------
注：本机压本机
系统：Windows 10 19045.2604
CPU: 2.90 GHz AMD Ryzen 7 4800H with Radeon Graphics
内存: 16.0 GB
群成员总数: 500人
在线人数: 300人
每秒/次发送消息数量: 100条
发送消息次数: 30次
响应消息数量：90 0000条
Message 表数量总数：150 0000条
丢失消息数量: 0条
总耗时: 73 000ms 73s
平均每100条消息发送/转发在线人员/在线人员接收总耗时: 2 433ms
*/

var (
	// 起始phone_num
	pn = flag.Int64("pn", 100000, "First phone num")
	// 群成员总数
	gn = flag.Int64("gn", 500, "群成员总数")
	// 在线成员数量
	on = flag.Int64("on", 300, "在线成员数量")
	// 每次发送消息数量（不同在线成员发送一次，总共 500 个在线成员每人发送一次，总共发送 500 次）
	sn = flag.Int64("sn", 100, "每次发送消息数量")
	// 发送消息次数
	tn = flag.Int64("tn", 30, "发送消息次数")
)

func main() {
	flag.Parse()

	mgr := NewManager(*pn, *on, *sn, *gn, *tn)
	mgr.Run()

	select {}
}
