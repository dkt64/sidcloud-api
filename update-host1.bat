ssh -i C:\Users\admin\.ssh\host1 root@185.157.81.123 /bin/bash -c 'rm -R /root/go/src/github.com/dkt64/sidcloud-api/dist'
c:\tools\pscp -P 22 -i %USERPROFILE%\.ssh\host1.ppk -r -v dist root@185.157.81.123:/root/go/src/github.com/dkt64/sidcloud-api
c:\tools\pscp -P 22 -i %USERPROFILE%\.ssh\host1.ppk sidcloud.go root@185.157.81.123:/root/go/src/github.com/dkt64/sidcloud-api
ssh -i C:\Users\admin\.ssh\host1 root@185.157.81.123 /bin/bash -c 'systemctl restart sidcloud'