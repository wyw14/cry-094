# Bug 是什么

后台投递器成功处理事件后没有结束该事件的队列生命周期，同一事件会被再次投递。

# 如何触发

入队一条 analysis.completed 事件，启动返回成功的投递处理器，并在首次成功后继续观察队列。

# 错误信息

`successful event was delivered 2 times`
