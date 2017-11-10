# go-pcurl
cURL in parallel way, Written in golang

## Progressing bar

*Downloading resource from nginx runs locally*

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/bar1.gif)

*Downloading resource from website*

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/bar2.gif)

## Keep-Alive

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/img5.png)

## Testing Sample, compared w/ traditional wget way

### Common wget way took 6m21s+, used only 1 connection

```

(Laptop)simon@Simon-MBp:~/src$time wget https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm  -SO /tmp/docker.rpm
--2017-09-30 15:21:03--  https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm
Resolving download.docker.com... 54.239.132.250, 54.239.132.174, 54.239.132.84, ...
Connecting to download.docker.com|54.239.132.250|:443... connected.
HTTP request sent, awaiting response...
  HTTP/1.1 200 OK
  Content-Type: application/x-redhat-package-manager
  Content-Length: 22157896
  Connection: keep-alive
  Date: Wed, 27 Sep 2017 01:55:11 GMT
  Last-Modified: Wed, 27 Sep 2017 01:47:41 GMT
  x-amz-version-id: Y76xcrpq2VKOnT7JLWgG5L65_DXLCww4
  ETag: "5e7d5e5afcc6cda75771533dc58b2749-3"
  Server: AmazonS3
  X-Cache: RefreshHit from cloudfront
  Via: 1.1 ae162f6796e551002447afd7c07ec67a.cloudfront.net (CloudFront)
  X-Amz-Cf-Id: kQiBxdtLGBmSlogdglwGJHFYT_G9IjHjR_SMZQq64isGqGp4A3MhWg==
Length: 22157896 (21M) [application/x-redhat-package-manager]
Saving to: ‘/tmp/docker.rpm’

/tmp/docker.rpm                              100%[=============================================================================================>]  21.13M   100KB/s    in 6m 20s

2017-09-30 15:27:25 (56.9 KB/s) - ‘/tmp/docker.rpm’ saved [22157896/22157896]


real	6m21.471s
user	0m0.168s
sys	0m0.560s
(Laptop)simon@Simon-MBp:~/src$md5 /tmp/docker.rpm
MD5 (/tmp/docker.rpm) = 647b4bb14e61bec73ddd137f6a40edac

```

### pcurl took only 47s,  used 20 connections

```
(Laptop)simon@Simon-MBp:~/src$time ./pcurl https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm  /tmp/docker.rpm
2017/09/30 15:34:02 Created tmpdir: /tmp/gotemp430902692
2017/09/30 15:34:21 Created tmpfile: /tmp/gotemp430902692/13.223914333
2017/09/30 15:34:21 Created tmpfile: /tmp/gotemp430902692/15.058328607
2017/09/30 15:34:23 Created tmpfile: /tmp/gotemp430902692/0.228534679
2017/09/30 15:34:23 Created tmpfile: /tmp/gotemp430902692/12.122017740
2017/09/30 15:34:23 Created tmpfile: /tmp/gotemp430902692/7.808832630
2017/09/30 15:34:24 Created tmpfile: /tmp/gotemp430902692/6.203386216
2017/09/30 15:34:26 Created tmpfile: /tmp/gotemp430902692/18.738236867
2017/09/30 15:34:28 Created tmpfile: /tmp/gotemp430902692/1.387619003
2017/09/30 15:34:28 Created tmpfile: /tmp/gotemp430902692/5.490759238
2017/09/30 15:34:29 Created tmpfile: /tmp/gotemp430902692/8.978373601
2017/09/30 15:34:29 Created tmpfile: /tmp/gotemp430902692/17.806482164
2017/09/30 15:34:30 Created tmpfile: /tmp/gotemp430902692/2.045384128
2017/09/30 15:34:30 Created tmpfile: /tmp/gotemp430902692/16.242278130
2017/09/30 15:34:34 Created tmpfile: /tmp/gotemp430902692/4.063984553
2017/09/30 15:34:37 Created tmpfile: /tmp/gotemp430902692/9.871300318
2017/09/30 15:34:37 Created tmpfile: /tmp/gotemp430902692/11.116061194
2017/09/30 15:34:43 Created tmpfile: /tmp/gotemp430902692/3.103707373
2017/09/30 15:34:43 Created tmpfile: /tmp/gotemp430902692/19.037809331
2017/09/30 15:34:45 Created tmpfile: /tmp/gotemp430902692/10.886562072
2017/09/30 15:34:46 Created tmpfile: /tmp/gotemp430902692/14.389013669
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/0.228534679
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/1.387619003
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/2.045384128
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/3.103707373
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/4.063984553
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/5.490759238
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/6.203386216
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/7.808832630
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/8.978373601
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/9.871300318
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/10.886562072
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/11.116061194
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/12.122017740
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/13.223914333
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/14.389013669
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/15.058328607
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/16.242278130
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/17.806482164
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/18.738236867
2017/09/30 15:34:46 Cleaned tmpfile: /tmp/gotemp430902692/19.037809331
2017/09/30 15:34:46 Downloaded: from https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm to /tmp/docker.rpm
2017/09/30 15:34:46 Removed tmpdir: /tmp/gotemp430902692

real	0m47.058s
user	0m0.497s
sys	0m0.902s

(Laptop)simon@Simon-MBp:~/src$md5 /tmp/docker.rpm
MD5 (/tmp/docker.rpm) = 647b4bb14e61bec73ddd137f6a40edac
```

### iftop outputs when running wget (in 1 connection)

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/img2.png)

### iftop outputs when running pcurl (in 4 connections)

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/img1.png)

### iftop outputs when running pcurl (in more connections)

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/img3.png)

![Progress](https://github.com/thimoonxy/go-pcurl/blob/master/misc/img4.png)


### TODO

- [ ] Need more Parameter flags instead of $@ ;
- [ ] BW controling in each connection;
- [x] More human readable outputs instead of ugly log.print things;
- [x] Processing bar;
- [ ] Detect and select faster ip to use, if resource can be resolved to several IPs records;
- [x] Local optimization, decreasing the cost of native Mem/disk IO and so forth; 
- [x] Keep-Alive; 