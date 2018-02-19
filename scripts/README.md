# Example startup scripts for pideo-server
* pideo-server - SysV style startup script (for Raspbian wheezy)
* pideo-server.service - systemd style startup service

```bash
chmod +x pideo-server
sudo cp pideo-server /etc/init.d
sudo update-rc.d pideo-server defaults
```

```bash
sudo cp pideo-server.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable pideo-server
sudo systemctl start pideo-server
```
