# Example startup scripts for pideo-server
* pideo-server - SysV style startup script (for Raspbian wheezy)
```bash
chmod +x pideo-server
sudo cp pideo-server /etc/init.d
sudo update-rc.d pideo-server defaults
```
