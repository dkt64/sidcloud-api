ssh -i C:\Users\admin\.ssh\host1 root@185.157.81.123 /bin/bash -c 'rm -R /root/sidcloud-test/dist'
c:\tools\pscp -i %USERPROFILE%\.ssh\host1.ppk -r -v dist root@185.157.81.123:/root/sidcloud-test
c:\tools\pscp -i %USERPROFILE%\.ssh\host1.ppk sidcloud.go root@185.157.81.123:/root/sidcloud-test