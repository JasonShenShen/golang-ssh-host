#golang ssh host 连接华为NE5k路由器并支持more分页符
* 利用golang/x/crypto/ssh实现并参考作者: CodyGuo代码
* 登录华为路由器执行命令display version和display arp statistics all
* 对终端屏宽产生的分页符做处理
* 可用作单台设备执行命令打印输出之用

#重点code说明
```
...
in <- "display version"  //执行输入goroutine
in <- "display arp statistics all"
...
go func() {
		for cmd := range in {
			wg.Add(1)
			w.Write([]byte(cmd + "\n"))
			wg.Wait() //控制每次goroutine执行一条命令
		}
	}()
go func() {
		var (
			buf [1024 * 1024]byte
			t   int
		)
		for {
			n, err := r.Read(buf[t:])
			if err != nil {
				fmt.Println(err.Error())
				close(in)
				close(out)
				return
			}
			t += n
			result := string(buf[:t])
			//循环处理设备分页符
			if strings.Contains(string(buf[t-n:t]), "More") {
				w.Write([]byte("\n"))
			}
			//匹配等待符完成一条操作清空goroutine
			if strings.Contains(result, "username:") ||
				strings.Contains(result, "password:") ||
				strings.Contains(result, ">") {
				out <- string(buf[:t])
				t = 0
				wg.Done()
			}
		}
	}()
...
```

#TestUnit
```
display version
Huawei Versatile Routing Platform Software
VRP (R) software, Version 5.90 (NE40E&80E V600R003C00SPCa00)
Copyright (C) 2000-2012 Huawei Technologies Co., Ltd.
HUAWEI NE80E uptime is 1695 days, 20 hours, 57 minutes
NE80E version information:

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

BKP 1 version information:
  PCB         Version : CR52BKPA REV B
  MPU  Slot  Quantity : 2
  SRU  Slot  Quantity : 0
  SFU  Slot  Quantity : 4
  LPU  Slot  Quantity : 16
...
...
...
<HK-HK-CW-F-1.CN2>
display arp statistics all
Dynamic: 345     Static: 0    

<HK-HK-CW-F-1.CN2>
```


