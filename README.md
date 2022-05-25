# 这是一个即时聊天系统的小案例

## 记录下socket编程的学习过程

## 学习来源 https://www.bilibili.com/video/BV1gf4y1r79E?p=37

## 一些笔记

* 用户退出时,释放资源后

>> 关闭通道,造成broken pipe的错误,这是因为关闭通道后,通道读取将不再阻塞,导致user端的监听变成死循环

>> 同时造成了向已关闭连接写入的错误

## 实现的一些操作

![op1](https://github.com/duan1v/practice_go/blob/master/imgs/op1.png)

![op2](https://github.com/duan1v/practice_go/blob/master/imgs/op2.png)  