#golang ssh host ���ӻ�ΪNE5k·������֧��more��ҳ��
* ����golang/x/crypto/sshʵ�ֲ��ο�����: CodyGuo����
* ��¼��Ϊ·����ִ������display version��display arp statistics all
* ���ն���������ķ�ҳ��������
* ��������̨�豸ִ�������ӡ���֮��

#�ص�code˵��
```
...
in <- "display version"  //ִ������goroutine
in <- "display arp statistics all"
...
go func() {
		for cmd := range in {
			wg.Add(1)
			w.Write([]byte(cmd + "\n"))
			wg.Wait() //����ÿ��goroutineִ��һ������
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
			//ѭ�������豸��ҳ��
			if strings.Contains(string(buf[t-n:t]), "More") {
				w.Write([]byte("\n"))
			}
			//ƥ��ȴ������һ���������goroutine
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


