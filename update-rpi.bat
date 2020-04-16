rem ssh -i C:\Users\admin\.ssh\pi pi@192.168.1.158 /bin/bash -c 'rm -R /home/pi/sidcloud-api/dist'
c:\tools\pscp -i %USERPROFILE%\.ssh\pi.ppk -r -v dist pi@192.168.1.158:/home/pi/sidcloud-api
c:\tools\pscp -i %USERPROFILE%\.ssh\pi.ppk sidcloud.go pi@192.168.1.158:/home/pi/sidcloud-api
rem ssh -i C:\Users\admin\.ssh\pi pi@192.168.1.158 /bin/bash -c 'systemctl restart sidcloud'